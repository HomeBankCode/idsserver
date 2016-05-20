package main

import (
	"encoding/json"
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

// Block represents a CLAN conversation block
type Block struct {
	ClanFile  string `json:"clan-file"`
	Index     int    `json:"block-index"`
	Clips     []Clip `json:"clips"`
	FanOrMan  bool   `json:"fan-or-man"`
	DontShare bool   `json:"dont-share"`
	ID        string `json:"id"`
	Coder     string `json:"coder"`
	LabKey    string `json:"lab-key"`
	LabName   string `json:"lab-name"`
}

func (block *Block) encode() ([]byte, error) {
	data, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return data, err
	}
	return data, nil
}

func (block *Block) appendClip(clip Clip) {
	block.Clips = append(block.Clips, clip)
}

// Clip represent a single tier from a conversation block
type Clip struct {
	Index           int    `json:"clip-index"`
	Tier            string `json:"clip-tier"`
	Multiline       bool   `json:"multiline"`
	MultiTierParent string `json:"multi-tier-parent"`
	StartTime       string `json:"start-time"`
	OffsetTime      string `json:"offset-time"`
	TimeStamp       string `json:"timestamp"`
	Classification  string `json:"classification"`
	LabelDate       string `json:"label-date"`
	Coder           string `json:"coder"`
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

// Close closes the database
func (db *LabelsDB) Close() {
	db.db.Close()
}

func (db *LabelsDB) addBlock(block Block) error {
	encodedBlock, err := block.encode()
	if err != nil {
		return err
	}

	updateErr := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labelsBucket))
		err := bucket.Put([]byte(block.ID), encodedBlock)
		return err
	})
	return updateErr
}
