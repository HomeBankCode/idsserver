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
	WorkItems []WorkItem
}

/*
WorkItem represents a work item at the granularity of
a single CLAN file. Each object signifies a group of
{numBlocksSent} blocks from that specific clan file
*/
type WorkItem struct {
	ID       int    `json:"id"`
	FileName string `json:"filename"`
	Blocks   []int  `json:"blocks"`
	Active   bool   `json:"active"`
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
