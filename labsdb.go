package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const (
	labsDBPath = "db/labs.db"
	labsBucket = "Labs"
)

// Lab is a JSON serialization
// struct representing lab metadata
type Lab struct {
	Key     string          `json:"key"`
	LabName string          `json:"lab-name"`
	Users   map[string]User `json:"users"`
}

// User is a lab user
type User struct {
	Name          string     `json:"name"`
	ParentLab     string     `json:"parent-lab"`
	WorkItems     []WorkItem `json:"active-work-items"`
	PastWorkItems []WorkItem `json:"finished-work-items"`
}

func (user *User) addWorkItem(item WorkItem) {
	user.WorkItems = append(user.WorkItems, item)
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
	lab.Users[user.Name] = user
}

// LabsDB is wrapper around a boltdb
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
		ParentLab: labKey,
		WorkItems: make([]WorkItem, 0)}

	if db.labExists(labKey) {
		if db.userExists(labKey, username) {
			return
		}
		lab := db.getLab(labKey)
		lab.addUser(newUser)
		db.setLab(labKey, lab)
	} else {
		db.createLabAddUser(labKey, labName, username)
	}
}

func (db *LabsDB) labExists(labKey string) bool {
	var exists bool
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
		}
		userExists = false

		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return userExists
}

func (db *LabsDB) getLab(labKey string) *Lab {
	var lab []byte
	db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labsBucket))
		lab = bucket.Get([]byte(labKey))
		return nil
	})

	labData, err := decodeLabJSON(lab)
	if err != nil {
		log.Fatal(err)
	}
	return labData
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

func (db *LabsDB) getUser(labKey, username string) User {

	var lab = db.getLab(labKey)
	user := lab.Users[username]
	return user

}

func (db *LabsDB) setUser(user User) {
	var lab = db.getLab(user.ParentLab)
	lab.Users[user.Name] = user
	db.setLab(lab.Key, lab)

}
