package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

var (
	// path to the WorkDB file
	workDBPath = mainConfig.WorkDBPath
)

const (

	// name of the database's work bucket
	workBucket = "Work"
)

var (

	// ErrRanOutOfItems means chooseUniqueWorkItem can't
	// find a suitable WorkItem for this particular user
	ErrRanOutOfItems = errors.New("Block checkout reached capacity for this user")

	// ErrWorkItemDoesntExist is thrown when a workItemMap access is
	// made with a key that doesn't exist
	ErrWorkItemDoesntExist = errors.New("This work Item doesn't exist")
)

/*
WorkItem represents a work item at the granularity of
a single CLAN file. Each WorkItem signifies a single
block from that specific clan file
*/
type WorkItem struct {
	ID              string `json:"id"`
	FileName        string `json:"filename"`
	Block           int    `json:"block"`
	Active          bool   `json:"active"`
	BlockPath       string `json:"block_path"`
	TimesCoded      int    `json:"times_coded"`
	Training        bool   `json:"training"`
	Reliability     bool   `json:"reliability"`
	TrainingPackNum int    `json:"train_pack_num"`
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
			//fmt.Println("\nWorkItem from DB:")
			//fmt.Printf("key=%s, value=%s\n", k, v)
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
inactivateWorkItem sets the WorkItem to false
in the workItemMap
*/
func inactivateWorkItem(item WorkItem, request IDSRequest) {
	value := workItemMap[item.ID]
	value.Active = false
	//value.TimesCoded++
	workItemMap[item.ID] = value
	workDB.persistWorkItem(value)

	// update the User's WorkItem list on disk
	user, getUsrError := labsDB.getUser(request.LabKey, request.Username)
	if getUsrError != nil {
		return
	}
	user.inactivateWorkItem(value)
	labsDB.setUser(user)
}

func inactivateIncompleteWorkItem(item WorkItem, request IDSRequest) {
	value := workItemMap[item.ID]
	value.Active = false
	workItemMap[item.ID] = value
	workDB.persistWorkItem(value)

	// update the User's WorkItem list on disk
	user, getUsrError := labsDB.getUser(request.LabKey, request.Username)
	if getUsrError != nil {
		return
	}

	user.inactivateIncompleteWorkItem(value)
	labsDB.setUser(user)
}

/*
activateWorkItem sets the WorkItem active status to true
in the workItemMap.

Also adds the work item to the User's checked out WorkItem
list
*/
func activateWorkItem(item WorkItem, request BlockReq) {
	value := workItemMap[item.ID]
	value.Active = true

	// update the workItemMap (in memory)
	workItemMap[item.ID] = value

	// update the WorkItem value on disk
	workDB.persistWorkItem(value)

	// update the User's WorkItem list on disk
	user, getUsrError := labsDB.getUser(request.LabKey, request.Username)
	if getUsrError != nil {
		return
	}
	user.addWorkItem(item.ID)
	labsDB.setUser(user)
}

func chooseRegularWorkItem(request BlockReq) (WorkItem, error) {
	var workItem WorkItem
	user, getUsrErr := labsDB.getUser(request.LabKey, request.Username)
	if getUsrErr != nil {
		return WorkItem{}, ErrUserDoesntExist
	}
	for _, item := range workItemMap {
		if blockAppropriateForUser(item, request, user) {
			activateWorkItem(item, request)
			fmt.Println("Selected Item: ")
			fmt.Println(item)
			return item, nil
		}
	}

	fmt.Println("\nRan out of unique items for this user")
	return workItem, ErrRanOutOfItems
}

func blockAppropriateForUser(item WorkItem, request BlockReq, user User) bool {
	if item.Active {
		return false
	} else if item.TimesCoded >= numRealBlockPasses {
		fmt.Println("item has been coded already")
		return false
	} else if item.Training {
		return false
	} else if item.Reliability {
		return false
	} else if user.prevCoded(item.ID) {
		return false
	}
	return true
}

func userHasBlockFromFile(item WorkItem, request BlockReq, user User) bool {
	/*
		Check if user already has a block
		from the same file
	*/
	for _, userItem := range user.ActiveWorkItems {
		userWorkItem := workItemMap[userItem]
		if userWorkItem.FileName == item.FileName {
			return true
		}
	}
	return false
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

func chooseSpecificBlock(req BlockReq) (WorkItem, error) {
	var workItem WorkItem

	workItem, exists := workItemMap[req.ItemID]
	if !exists {
		return workItem, ErrWorkItemDoesntExist
	}
	if !workItem.Training && !workItem.Reliability && workItem.TimesCoded >= numRealBlockPasses {
		return workItem, ErrBlockGroupFull
	}
	activateWorkItem(workItem, req)

	return workItem, nil
}

func chooseTrainingWorkItem(request BlockReq) (WorkItem, error) {
	var workItem WorkItem
	user, getUsrErr := labsDB.getUser(request.LabKey, request.Username)
	if getUsrErr != nil {
		return WorkItem{}, ErrUserDoesntExist
	}

	for _, item := range workItemMap {
		if blockAppropriateForUserTraining(item, request, user) {
			activateWorkItem(item, request)
			fmt.Println("Selected Item: ")
			fmt.Println(item)
			return item, nil
		}
	}

	fmt.Println("\nRan out of unique items for this user")
	return workItem, ErrRanOutOfItems
}

func chooseSpecificTrainingBlock(req BlockReq) (WorkItem, error) {
	var workItem WorkItem

	workItem, exists := workItemMap[req.ItemID]
	if !exists {
		return workItem, ErrWorkItemDoesntExist
	}
	activateWorkItem(workItem, req)

	return workItem, nil
}

func chooseReliabilityWorkItem(request BlockReq) (WorkItem, error) {
	var workItem WorkItem
	user, getUsrErr := labsDB.getUser(request.LabKey, request.Username)
	if getUsrErr != nil {
		return WorkItem{}, ErrUserDoesntExist
	}
	for _, item := range workItemMap {

		if blockAppropriateForUserReliability(item, request, user) {
			activateWorkItem(item, request)
			fmt.Println("Selected Item: ")
			fmt.Println(item)
			return item, nil
		}
	}

	fmt.Println("\nRan out of unique items for this user")
	return workItem, ErrRanOutOfItems
}

func blockAppropriateForUserTraining(item WorkItem, request BlockReq, user User) bool {
	if !item.Training {
		return false
	} else if user.hasThisBlock(item.ID) {
		fmt.Println("user has this block already")
		return false
	}
	return true
}

func blockAppropriateForUserReliability(item WorkItem, request BlockReq, user User) bool {
	if !item.Reliability {
		return false
	} else if user.prevCodedRelia(item.ID) {
		fmt.Println("item has been coded already by this user")
		return false
	} else if user.hasThisBlock(item.ID) {
		fmt.Println("user has block from this file")
		return false
	}
	return true
}
