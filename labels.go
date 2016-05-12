package main

import (
	"log"

	"github.com/boltdb/bolt"
)

const (
	// path to the LabelsDB file
	labelsDBPath = "db/labels.db"

	// name of the database's labels bucket
	labelsBucket = "Labels"
)

// LabelsDB is a wrapper around a boltDB
type LabelsDB struct {
	db *bolt.DB
}

/*
WorkItemLabels represents the classification results
that are submitted by the users of IDSLabel
*/
type WorkItemLabels struct {
	ItemID     string     `json:"item-id"`
	BlockClips [][]string ``
}

// LoadLabelsDB loads the global workDB
func LoadLabelsDB() *LabelsDB {
	localLabelsDB := &LabelsDB{db: new(bolt.DB)}
	err := localLabelsDB.Open()
	if err != nil {
		return nil
	}
	return localLabelsDB
}

// Open opens the database and returns error on failure
func (db *LabelsDB) Open() error {
	labelsDB, openErr := bolt.Open(labelsDBPath, 0600, nil)

	if openErr != nil {
		log.Fatal(openErr)
		return openErr
	}

	db.db = labelsDB

	err := db.db.Update(func(tx *bolt.Tx) error {
		_, updateErr := tx.CreateBucketIfNotExists([]byte(labelsBucket))
		if updateErr != nil {
			log.Fatal(updateErr)
			return updateErr
		}
		return updateErr
	})

	return err
}
