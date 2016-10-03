package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

var (
	// path to the LabsDB file
	labsDBPath = mainConfig.LabsDBPath
)

const (
	labsBucket = "Labs"
)

var (
	// ErrUserDoesntExist means User doesn't exist in the LabsDB
	ErrUserDoesntExist = errors.New("User doesn't exist")

	// ErrLabDoesntExist means Lab doesn't exist in the LabsDB
	ErrLabDoesntExist = errors.New("Lab doesn't exist")

	// ErrUserNotAssignedWorkItem means that the user doesn't have
	// a given WorkItem in their ActiveWorkItems list.
	ErrUserNotAssignedWorkItem = errors.New("User wasn't assigned this work item")

	// ErrLabNotRegistered means that the lab is not one of the labs
	// approved to access this server
	ErrLabNotRegistered = errors.New("Lab is not registered")

	// ErrTrainBlockNotFound means that this user doesn't have
	// this training block's labels stored in it's CompleteTrainBlocks
	// field
	ErrTrainBlockNotFound = errors.New("Training block not found for this user")

	// ErrReliaBlockNotFound means that this user doesn't have
	// this reliability block's labels stored in it's CompleteRelBlocks
	// field
	ErrReliaBlockNotFound = errors.New("Reliability block not found for this user")
)

// Lab is a JSON serialization
// struct representing lab metadata
type Lab struct {
	Key     string          `json:"key"`
	LabName string          `json:"lab_name"`
	Users   map[string]User `json:"users"`
}

// User is a lab user
type User struct {
	Name                string      `json:"name"`
	ParentLab           string      `json:"parent_lab"`
	ActiveWorkItems     BlockIDList `json:"active_work_items"`
	PastWorkItems       BlockIDList `json:"finished_work_items"`
	CompleteTrainBlocks BlockIDList `json:"complete_train_blocks"`
	CompleteRelBlocks   BlockIDList `json:"complete_reliability_blocks"`
}

func (user *User) addWorkItem(itemID string) {
	// make sure not to add an ID more than once
	for _, userActiveID := range user.ActiveWorkItems {
		if itemID == userActiveID {
			return
		}
	}
	user.ActiveWorkItems.addID(itemID)
}

func (user *User) addCompleteTrainBlock(block Block) {
	// make sure not to add an ID more than once
	for _, id := range user.CompleteTrainBlocks {
		if block.ID == id {
			return
		}
	}
	// add ID
	user.CompleteTrainBlocks.addID(block.ID)
}

func (user *User) addCompleteRelBlock(block Block) {
	for _, id := range user.CompleteRelBlocks {
		if block.ID == id {
			return
		}
	}
	// add ID
	user.CompleteRelBlocks.addID(block.ID)
}

func (user *User) inactivateWorkItem(item WorkItem) error {
	var newActiveItems BlockIDList
	foundItemInActive := false

	for _, activeID := range user.ActiveWorkItems {
		if item.ID == activeID {
			foundItemInActive = true
		} else {
			newActiveItems.addID(activeID)
		}
	}
	if foundItemInActive {
		var itemAlreadyInPastList = false
		for _, userFinishedID := range user.PastWorkItems {
			if item.ID == userFinishedID {
				itemAlreadyInPastList = true
			}
		}
		if !itemAlreadyInPastList {
			user.PastWorkItems.addID(item.ID)
		}
		user.ActiveWorkItems = newActiveItems
	} else {
		return ErrUserNotAssignedWorkItem
	}
	return nil
}

func (user *User) inactivateIncompleteWorkItem(item WorkItem) error {
	var newActiveItems BlockIDList
	foundItem := false

	for _, activeID := range user.ActiveWorkItems {
		if item.ID == activeID {
			foundItem = true
		} else {
			newActiveItems.addID(activeID)
		}
	}
	//user.PastWorkItems = append(user.PastWorkItems, item)
	user.ActiveWorkItems = newActiveItems

	if !foundItem {
		return ErrUserNotAssignedWorkItem
	}
	return nil
}

func (user *User) getPastBlockInstanceMap() (InstanceMap, error) {
	instanceMap := make(InstanceMap)
	fmt.Println("inside user.getPastBlockInstanceMap ----- made the map")
	for _, blockID := range user.PastWorkItems {
		blockGroup, blockGroupErr := labelsDB.getBlock(blockID)
		if blockGroupErr != nil {
			return instanceMap, blockGroupErr
		}
		fmt.Println("inside user.getPastBlockInstanceMap ----- got a blockGroup")
		userInstances := blockGroup.getUsersBlocks(user.ParentLab, user.Name)
		for _, block := range userInstances {

			if _, exists := instanceMap[block.ID]; exists {
				instanceMap[block.ID].addInstance(block.Instance)
			} else {
				instanceMap[block.ID] = &InstanceList{}
				instanceMap[block.ID].addInstance(block.Instance)
			}
			fmt.Println("inside user.getPastBlockInstanceMap ----- added instance to map")
		}
	}
	fmt.Println("The instanceMap being returned from user.getPastBlockInstanceMap(): ")
	for key, value := range instanceMap {
		for _, instance := range *value {
			fmt.Println("\nkey: ", key, "\t", "inst: ", instance)
		}
	}
	fmt.Println("\n\n", instanceMap)
	return instanceMap, nil
}

func (user *User) deletePastItem(blockID string) {
	var newBlockIDList BlockIDList
	var newTrainIDList BlockIDList
	var newReliaIDList BlockIDList

	// clear PastWorkItems
	for _, pastBlockID := range user.PastWorkItems {
		if !(pastBlockID == blockID) {
			newBlockIDList.addID(pastBlockID)
		}
	}
	user.PastWorkItems = newBlockIDList

	// clear CompleteTrainBlocks
	for _, pastTrainID := range user.CompleteTrainBlocks {
		if !(pastTrainID == blockID) {
			newTrainIDList.addID(pastTrainID)
		}
	}
	user.CompleteTrainBlocks = newTrainIDList

	// clear CompleteRelBlocks
	for _, pastReliaID := range user.CompleteRelBlocks {
		if !(pastReliaID == blockID) {
			newReliaIDList.addID(pastReliaID)
		}
	}
	user.CompleteRelBlocks = newReliaIDList
}

func (user *User) hasThisBlock(blockID string) bool {
	for _, userItemID := range user.ActiveWorkItems {
		if userItemID == blockID {
			return true
		}
	}
	return false

}

func (user *User) prevCodedRelia(blockID string) bool {
	for _, userItem := range user.CompleteRelBlocks {
		if userItem == blockID {
			return true
		}
	}
	return false

}

func (user *User) prevCoded(blockID string) bool {
	for _, userItem := range user.PastWorkItems {
		if userItem == blockID {
			return true
		}
	}
	return false

}

func (lab *Lab) encode() ([]byte, error) {
	enc, err := json.MarshalIndent(lab, "", " ")
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decodeLabJSON(data []byte) (*Lab, error) {
	var lab *Lab
	err := json.Unmarshal(data, &lab)
	if err != nil {
		return nil, err
	}
	return lab, nil
}

func (lab *Lab) addUser(user User) {
	user.ParentLab = lab.Key
	if _, exists := lab.Users[user.Name]; exists {
		return
	}
	lab.Users[user.Name] = user
}

func (lab *Lab) deleteUser(user string) {
	delete(lab.Users, user)
}

func (lab *Lab) getPastBlockInstanceMap() (InstanceMap, error) {
	instanceMap := make(InstanceMap)

	for _, user := range lab.Users {
		for _, blockID := range user.PastWorkItems {
			blockGroup, blockGroupErr := labelsDB.getBlock(blockID)
			if blockGroupErr != nil {
				fmt.Println("\nlabelsDB.getBlock() failed in lab.getPastBlockInstanceMap()")
				return instanceMap, blockGroupErr
			}
			userInstances := blockGroup.getUsersBlocks(user.ParentLab, user.Name)
			for _, block := range userInstances {
				if _, exists := instanceMap[block.ID]; exists {
					instanceMap[block.ID].addInstance(block.Instance)
				} else {
					instanceMap[block.ID] = &InstanceList{}
					instanceMap[block.ID].addInstance(block.Instance)
				}

				// fmt.Println("\n\n\nbefore map access in getPastBlockInstanceMap")
				// instanceMap[block.ID].addInstance(block.Instance)
				// fmt.Println("\n\n\nafter map access in getPastBlockInstanceMap")
			}
			fmt.Printf("\nthe instanceMap from lab.getPastBlockInstanceMap():  \n\n")
			fmt.Println(instanceMap)
		}
	}
	return instanceMap, nil
}

// LabsDB is a wrapper around a boltdb
type LabsDB struct {
	db *bolt.DB
}

// LoadLabsDB loads the global LabsDB
func LoadLabsDB() *LabsDB {
	localLabsDB := &LabsDB{db: new(bolt.DB)}
	err := localLabsDB.Open()
	if err != nil {
		return nil
	}
	return localLabsDB
}

// Open opens the database and returns error on failure
func (db *LabsDB) Open() error {
	labsDB, openErr := bolt.Open(labsDBPath, 0600, nil)

	if openErr != nil {
		log.Fatal(openErr)
		return openErr
	}

	db.db = labsDB

	err := db.db.Update(func(tx *bolt.Tx) error {
		_, updateErr := tx.CreateBucketIfNotExists([]byte(labsBucket))
		if updateErr != nil {
			log.Fatal(updateErr)
			return updateErr
		}
		return updateErr
	})

	return err
}

// Close closes the database
func (db *LabsDB) Close() {
	db.db.Close()
}

func (db *LabsDB) addUser(labKey, labName, username string) {
	newUser := User{Name: username,
		ParentLab:       labKey,
		ActiveWorkItems: make(BlockIDList, 0)}

	if db.labExists(labKey) {
		if db.userExists(labKey, username) {
			fmt.Println("This user already exists")
			return
		}
		lab, _ := db.getLab(labKey)
		lab.addUser(newUser)
		db.setLab(labKey, lab)
	} else {
		db.createLabAddUser(labKey, labName, username)
	}
}

func (db *LabsDB) labExists(labKey string) bool {
	var exists bool

	fmt.Println("Check for lab key existence: ")
	fmt.Println(labKey)

	db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))
		lab := bucket.Get([]byte(labKey))

		// lab key doesn't exist
		if lab == nil {
			exists = false
		} else {
			exists = true
		}
		return nil
	})
	return exists
}

// userExists assumes that the lab exists
func (db *LabsDB) userExists(labKey, username string) bool {
	var userExists = false
	err := db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))
		lab := bucket.Get([]byte(labKey))

		labData, err := decodeLabJSON(lab)
		if err != nil {
			log.Fatal(err)
		}

		_, exists := labData.Users[username]
		if exists {
			userExists = true
			fmt.Println("inside labsDB.userExists(): userExists = true")
		} else {
			userExists = false
		}
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return userExists
}

func (db *LabsDB) getLab(labKey string) (*Lab, error) {
	if !db.labExists(labKey) {
		return nil, ErrLabDoesntExist
	}
	var lab []byte
	db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))
		lab = bucket.Get([]byte(labKey))
		return nil
	})

	labData, err := decodeLabJSON(lab)
	//fmt.Println(labData)
	if err != nil {
		log.Fatal(err)
	}
	return labData, nil
}

func (db *LabsDB) setLab(labKey string, data *Lab) {
	encodedLab, err := data.encode()
	if err != nil {
		log.Fatal(err)
		return
	}

	db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))
		err := bucket.Put([]byte(labKey), encodedLab)
		return err
	})
}

func (db *LabsDB) createLab(labKey, labName string) {
	newLab := Lab{Key: labKey,
		LabName: labName,
		Users:   make(map[string]User)}
	encodedLab, err := newLab.encode()
	if err != nil {
		log.Fatal(err)
	}

	db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))
		err := bucket.Put([]byte(labKey), encodedLab)
		return err
	})
}

func (db *LabsDB) createLabAddUser(labKey, labName, username string) {
	newLab := Lab{Key: labKey,
		LabName: labName,
		Users:   make(map[string]User)}

	newUser := User{Name: username}
	newLab.addUser(newUser)
	encodedLab, err := newLab.encode()
	if err != nil {
		log.Fatal(err)
	}

	db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))
		err := bucket.Put([]byte(labKey), encodedLab)
		return err
	})
}

func (db *LabsDB) getAllLabs() []*Lab {
	var labs []*Lab
	err := db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))

		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			currLab, err := decodeLabJSON(v)
			if err != nil {
				log.Fatal(err)
			}
			labs = append(labs, currLab)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return labs
}

func (db *LabsDB) getUser(labKey, username string) (User, error) {
	fmt.Println("\n\n\ntrying to get lab:  " + labKey)
	lab, err := db.getLab(labKey)
	if err != nil {
		fmt.Println("\ngetLab in getUser failed")
		return User{}, err
	}

	fmt.Println("\n\nthe lab from db.getLab():")
	fmt.Println(lab)

	user, exists := lab.Users[username]
	if !exists {
		fmt.Println("lab[username] map access failed")
		fmt.Printf("tried to find username: ")
		fmt.Println(username)
		return user, ErrUserDoesntExist
	}
	return user, nil
}

func (db *LabsDB) setUser(user User) error {
	lab, getLabErr := db.getLab(user.ParentLab)
	if getLabErr != nil {
		return ErrLabDoesntExist
	}
	lab.Users[user.Name] = user
	db.setLab(lab.Key, lab)
	return nil
}

func (db *LabsDB) getCompletedBlocks(labKey string) (BlockIDList, error) {
	var blocks BlockIDList

	lab, getLabErr := db.getLab(labKey)
	if getLabErr != nil {
		return blocks, getLabErr
	}

	for _, user := range lab.Users {
		for _, blockID := range user.PastWorkItems {
			blocks.addID(blockID)
		}
	}
	return blocks, nil
}

func (db *LabsDB) getCompleteTrainBlocks(labKey string) (BlockIDList, error) {
	var blocks BlockIDList
	lab, err := db.getLab(labKey)
	if err != nil {
		return blocks, err
	}
	for _, user := range lab.Users {
		for _, block := range user.CompleteTrainBlocks {
			blocks.addID(block)
		}
	}
	return blocks, nil
}

func (db *LabsDB) getCompleteReliaBlocks(labKey string) (BlockIDList, error) {
	var blocks BlockIDList
	lab, err := db.getLab(labKey)
	if err != nil {
		return blocks, err
	}
	for _, user := range lab.Users {
		for _, block := range user.CompleteRelBlocks {
			blocks.addID(block)
		}
	}
	return blocks, nil
}

func (db *LabsDB) deleteUser(labKey, username string) error {
	deleteBlocksErr := labelsDB.deleteUserBlocks(labKey, username)
	if deleteBlocksErr != nil {
		return deleteBlocksErr
	}
	lab, getLabErr := db.getLab(labKey)
	if getLabErr != nil {
		return getLabErr
	}
	lab.deleteUser(username)
	labsDB.setLab(labKey, lab)
	return nil
}
