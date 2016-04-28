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
	key   string
	users []string
}

// LabsDB is a struct carrying the bolt database
// for lab metadata
type LabsDB struct {
	db *bolt.DB
}

func (l *Lab) encode() ([]byte, error) {
	enc, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return enc, nil
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

func (db *LabsDB) addUser(labKey, user string) {
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(labsBucket))
		v := b.Get([]byte(labKey))
		fmt.Printf("The answer is: %s\n", v)
		return nil
	})

	db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(labsBucket))
		err := b.Put([]byte(labKey), []byte("42"))
		return err
	})
}

func (db *LabsDB) lookupUser(labKey string, user string) {
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(labsBucket))
		v := b.Get([]byte(labKey))
		fmt.Printf("The answer is: %s\n", v)
		return nil
	})
}

func (db *LabsDB) close() {
	db.db.Close()
}
