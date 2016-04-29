package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const workDBPath = "db/work.db"
const workBucket = "Work"

// WorkQueue is the list of WorkGroups
type WorkQueue struct {
	WorkGroups []WorkGroup `json:"work-groups"`
}

// WorkGroup is the unit of work that gets distributed
type WorkGroup struct {
	ID       int    `json:"id"`
	ClanFile string `json:"clanfile"`
	Blocks   int    `json:"blocks"`
	Active   bool   `json:"active"`
}

// WorkDB is a wrapper around a boltdb
type WorkDB struct {
	db *bolt.DB
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
