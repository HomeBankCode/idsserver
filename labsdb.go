package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const labsDBPath = "db/labs.db"
const labsBucket = "Labs"

// Lab is a JSON serialization
// struct representing lab metadata
type Lab struct {
	Key   string `json:"key"`
	Users []User `json:"users"`
}

// User is a lab user
type User struct {
	Name string `json:"name"`
}

func (lab *Lab) encode() ([]byte, error) {
	enc, err := json.Marshal(lab)
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
	lab.Users = append(lab.Users, user)
}

// LabsDB is a struct carrying the bolt database
// for lab metadata
type LabsDB struct {
	db *bolt.DB
}

func (db *LabsDB) openDB() error {
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

func (db *LabsDB) addUser(labKey, username string) {
	newUser := User{Name: username}

	if db.labExists(labKey) {
		lab := db.getLab(labKey)
		lab.addUser(newUser)
		db.setLab(labKey, lab)
	} else {
		db.createLabAddUser(labKey, username)
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
	fmt.Println(encodedLab)
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

func (db *LabsDB) createLab(labKey string) {
	newLab := Lab{Key: labKey}
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

func (db *LabsDB) createLabAddUser(labKey, username string) {
	newLab := Lab{Key: labKey}
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

// Close closes the boltdb
func (db *LabsDB) Close() {
	db.db.Close()
}
