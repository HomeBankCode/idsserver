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
5 blocks from a single CLAN file
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
func NewWorkGroup(numItems int) WorkGroup {
	workItems, err := chooseUniqueWorkItems(numItems)
	if err != nil {
		log.Fatal(err)
	}
	return WorkGroup{WorkItems: workItems}
}

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
func LoadWorkDB() (*WorkDB, error) {
	workDB := &WorkDB{db: new(bolt.DB)}
	err := workDB.Open()
	if err != nil {
		return nil, err
	}
	return workDB, nil
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

func (wg *WorkGroup) encode() ([]byte, error) {
	enc, err := json.MarshalIndent(wg, "", " ")
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decodeWorkJSON(data []byte) (*WorkGroup, error) {
	var wg *WorkGroup
	err := json.Unmarshal(data, &wg)
	if err != nil {
		return nil, err
	}
	return wg, nil
}

func (db *WorkDB) getAllWorkGroups() []*WorkGroup {
	var workGroups []*WorkGroup
	err := db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(workBucket))

		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			currGroup, err := decodeWorkJSON(v)
			if err != nil {
				log.Fatal(err)
			}
			workGroups = append(workGroups, currGroup)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return workGroups
}

/*
workItemIsActive checks to see if a WorkItem
is part of the global activeWorkItems map.
*/
func workItemIsActive(item WorkItem) bool {
	_, exists := activeWorkItems[item.ID]
	if exists {
		return true
	}
	return false
}

/*
inactivateWorkItem sets the WorkItem to true
in the workItemMap
*/
func inactivateWorkItem(item WorkItem) {
	workItemMap[item] = false
}

/*
activateWorkItem sets the WorkItem to false
in the workItemMap
*/
func activateWorkItem(item WorkItem) {
	workItemMap[item] = true
}

func chooseUniqueWorkItems(numItems int) ([]WorkItem, error) {
	var workItems []WorkItem
	for item, active := range workItemMap {
		if len(workItems) == numItems {
			break
		}
		if !active && !fileExistsInWorkItemArray(item.FileName, workItems) {
			activateWorkItem(item)
			workItems = append(workItems, item)
		}
	}
	return workItems, nil
}

func fileExistsInWorkItemArray(file string, array []WorkItem) bool {
	for _, item := range array {
		if item.FileName == file {
			return true
		}
	}
	return false
}
