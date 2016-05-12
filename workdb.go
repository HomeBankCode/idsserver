package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const (
	// path to the WorkDB file
	workDBPath = "db/work.db"

	// name of the database's work bucket
	workBucket = "Work"
)

// WorkQueue is the list of WorkGroups
type WorkQueue struct {
	WorkGroups []WorkGroup `json:"work-groups"`
}

/*
WorkGroup is the unit of work that gets distributed.
It contains a list of WorkItems. Each WorkItem contains
1 block from a single CLAN file
*/
type WorkGroup struct {
	WorkItems []WorkItem `json:"work-items"`
}

/*
NewWorkGroup returns a new WorkGroup containing
(numBlocksToSend) WorkItems uniquely selected to
be from distinct files and not currently being
worked on (i.e. haven't been send to any coders yet)
*/
/*func NewWorkGroup(wgRequest WorkGroupRequest) WorkGroup {*/
//workItems, err := chooseUniqueWorkItems(wgRequest)
//if err != nil {
//log.Fatal(err)
//}
//return WorkGroup{WorkItems: workItems}
/*}*/

/*
WorkItem represents a work item at the granularity of
a single CLAN file. Each WorkItem signifies a single
block from that specific clan file
*/
type WorkItem struct {
	ID        string `json:"id"`
	FileName  string `json:"filename"`
	Block     int    `json:"block"`
	Active    bool   `json:"active"`
	BlockPath string `json:"block-path"`
}

// WorkDB is a wrapper around a boltDB
type WorkDB struct {
	db *bolt.DB
}

// LoadWorkDB loads the global workDB
func LoadWorkDB() *WorkDB {
	localWorkDB := &WorkDB{db: new(bolt.DB)}
	err := localWorkDB.Open()
	if err != nil {
		return nil
	}
	return localWorkDB
}

// Open opens the database and returns error on failure
func (db *WorkDB) Open() error {
	workDB, openErr := bolt.Open(workDBPath, 0600, nil)

	if openErr != nil {
		log.Fatal(openErr)
		return openErr
	}

	db.db = workDB

	err := db.db.Update(func(tx *bolt.Tx) error {
		_, updateErr := tx.CreateBucketIfNotExists([]byte(workBucket))
		if updateErr != nil {
			log.Fatal(updateErr)
			return updateErr
		}
		return updateErr
	})

	return err
}

// Close closes the database
func (db *WorkDB) Close() {
	db.db.Close()
}

/*
fillWithDataMap fills the global workDB with the active/inactive
map of all the work items. Keys are WorkItem ID's and the values
are the bool values from the map.

true = active
false = inactive

*/
func (db *WorkDB) fillWithItemMap(itemMap WorkItemMap) {

	for id, item := range itemMap {

		// turn WorkItem into []byte
		encodedItem, encodeErr := item.encode()
		if encodeErr != nil {
			log.Fatal(encodeErr)
		}

		updateErr := db.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(workBucket))
			err := bucket.Put([]byte(id), encodedItem)
			return err
		})

		if updateErr != nil {
			log.Fatal(updateErr)
		}
	}
}

/*
loadItemMap reads the WorkItemMap from the workDB.
*/
func (db *WorkDB) loadItemMap() WorkItemMap {
	var itemMap = make(WorkItemMap)

	err := db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(workBucket))

		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			fmt.Println("\nWorkItem from DB:")
			fmt.Printf("key=%s, value=%s\n", k, v)
			currItem, err := decodeWorkItemJSON(v)
			if err != nil {
				log.Fatal(err)
			}
			itemMap[currItem.ID] = *currItem
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return itemMap
}

func (wg *WorkGroup) encode() ([]byte, error) {
	enc, err := json.MarshalIndent(wg, "", " ")
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func (wi *WorkItem) encode() ([]byte, error) {
	enc, err := json.MarshalIndent(wi, "", " ")
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decodeWorkItemJSON(data []byte) (*WorkItem, error) {
	var workItem *WorkItem
	err := json.Unmarshal(data, &workItem)
	if err != nil {
		return nil, err
	}
	return workItem, nil

}

func decodeWorkGroupJSON(data []byte) (*WorkGroup, error) {
	var wg *WorkGroup
	err := json.Unmarshal(data, &wg)
	if err != nil {
		return nil, err
	}
	return wg, nil
}

/*
workItemIsActive checks to see if a WorkItem
is part of the global activeWorkItems map.
*/
func workItemIsActive(item WorkItem) bool {
	value := workItemMap[item.ID]
	if value.Active {
		return true
	}
	return false
}

/*
inactivateWorkItem sets the WorkItem to true
in the workItemMap
*/
func inactivateWorkItem(item WorkItem) {
	value := workItemMap[item.ID]
	value.Active = false
	workItemMap[item.ID] = value

}

/*
activateWorkItem sets the WorkItem active status to true
in the workItemMap.

Also adds the work item to the User's checked out WorkItem
list
*/
func activateWorkItem(item WorkItem, request IDSRequest) {
	value := workItemMap[item.ID]
	value.Active = false

	// update the workItemMap (in memory)
	workItemMap[item.ID] = value

	// update the WorkItem value on disk
	workDB.persistWorkItem(value)

	// update the User's WorkItem list on disk
	user := labsDB.getUser(request.LabKey, request.Username)
	user.addWorkItem(item)
	labsDB.setUser(user)
}

func fileExistsInWorkItemArray(file string, array []WorkItem) bool {
	for _, item := range array {
		if item.FileName == file {
			return true
		}
	}
	return false
}

func chooseUniqueWorkItem(request IDSRequest) (WorkItem, error) {
	var workItem WorkItem
	for _, item := range workItemMap {

		if !item.Active && blockAppropriateForUser(request) {
			activateWorkItem(item, request)

			workItem = item
		}
	}
	return workItem, nil
}

func blockAppropriateForUser(request IDSRequest) bool {
	return true
}

func persistActiveMapChange(item WorkItem) {

}

func (db *WorkDB) compareWithWorkItemMap(itemMap WorkItemMap) []WorkItem {
	// missmatched WorkItems
	var diffs []WorkItem

	for key, value := range itemMap {
		var itemBytes []byte
		db.db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(workBucket))
			itemBytes = bucket.Get([]byte(key))

			workItem, err := decodeWorkItemJSON(itemBytes)
			if err != nil {
				log.Fatal(err)
			}

			switch {
			case workItem.Block != value.Block:
				diffs = append(diffs, value)
				break
			case workItem.BlockPath != value.BlockPath:
				diffs = append(diffs, value)
				break
			case workItem.FileName != value.FileName:
				diffs = append(diffs, value)
				break
			}
			return nil
		})
	}
	return diffs
}

func (db *WorkDB) persistWorkItem(item WorkItem) {

	// turn WorkItem into []byte
	encodedItem, err := item.encode()
	if err != nil {
		log.Fatal(err)
	}

	updateErr := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(workBucket))
		err := bucket.Put([]byte(item.ID), encodedItem)
		return err
	})

	if updateErr != nil {
		log.Fatal(updateErr)
	}
}

func (db *WorkDB) persistWorkItemMap(itemMap WorkItemMap) {
	for _, item := range itemMap {
		db.persistWorkItem(item)
	}
}
