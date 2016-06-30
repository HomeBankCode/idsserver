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
	labelsDBPath = mainConfig.LabelsDBPath

	// ErrCouldntFindLabeledBlock means a block ID wasn't
	// found in the database
	ErrCouldntFindLabeledBlock = errors.New("Couldn't Find Labeled Block")
)

const (
	// name of the database's labels bucket
	labelsBucket = "Labels"
)

// LabelsDB is a wrapper around a boltDB
type LabelsDB struct {
	db *bolt.DB
}

// BlockGroup is a collection of blocks
type BlockGroup struct {
	Blocks []Block `json:"blocks"`
}

func (bg *BlockGroup) append(block Block) {
	bg.Blocks = append(bg.Blocks, block)
}

// Block represents a CLAN conversation block
type Block struct {
	ClanFile    string `json:"clan-file"`
	Index       int    `json:"block-index"`
	Clips       []Clip `json:"clips"`
	FanOrMan    bool   `json:"fan-or-man"`
	DontShare   bool   `json:"dont-share"`
	ID          string `json:"id"`
	Coder       string `json:"coder"`
	LabKey      string `json:"lab-key"`
	LabName     string `json:"lab-name"`
	Username    string `json:"username"`
	Training    bool   `json:"training"`
	Reliability bool   `json:"reliability"`
}

func (block *Block) encode() ([]byte, error) {
	data, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return data, err
	}
	return data, nil
}

func decodeBlockJSON(data []byte) (*Block, error) {
	var block *Block
	err := json.Unmarshal(data, &block)
	if err != nil {
		return nil, err
	}
	return block, nil
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
	GenderLabel     string `json:"gender-label"`
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

func (db *LabelsDB) getBlock(blockID string) (*Block, error) {

	fmt.Println("Trying to retrieve block data: ")
	fmt.Println(blockID)

	var exists bool
	var blockData []byte

	db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labelsBucket))
		blockData = bucket.Get([]byte(blockID))

		// lab key doesn't exist
		if blockData == nil {
			exists = false
		} else {
			exists = true
		}
		return nil
	})

	if !exists {
		return &Block{}, ErrWorkItemDoesntExist
	}

	block, err := decodeBlockJSON(blockData)
	//fmt.Println(labData)
	if err != nil {
		return &Block{}, ErrWorkItemDoesntExist
	}
	return block, nil

}

func (db *LabelsDB) getBlockGroup(blockIDs []string) (BlockGroup, error) {
	var blocks BlockGroup

	for _, id := range blockIDs {
		block, err := db.getBlock(id)
		if err != nil {
			return BlockGroup{}, ErrCouldntFindLabeledBlock
		}
		blocks.append(*block)
	}
	return blocks, nil
}
