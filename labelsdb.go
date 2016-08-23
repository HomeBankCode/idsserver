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
)

var (
	// ErrCouldntFindLabeledBlock means a block ID wasn't
	// found in the database
	ErrCouldntFindLabeledBlock = errors.New("Couldn't Find Labeled Block")

	// ErrBlockGroupFull means that this block has already been coded
	// numBlockPasses times.
	ErrBlockGroupFull = errors.New("This block has already been coded through all passes")

	// ErrAddBlockFailed means that something prevented a Block from
	// being added to a BlockGroup
	ErrAddBlockFailed = errors.New("Adding block to BlockGroup failed")
)

const (
	// name of the database's labels bucket
	labelsBucket = "Labels"

	numRealBlockPasses = 2
)

// LabelsDB is a wrapper around a boltDB
type LabelsDB struct {
	db *bolt.DB
}

// // BlockGroup is a collection of blocks
// type BlockGroup struct {
// 	Blocks []BlockGroup `json:"blocks"`
// }
//
// func (bg *BlockGroup) append(block Block) {
// 	bg.Blocks = append(bg.Blocks, block)
// }

/*
BlockGroupArray is an array of BlockGroups
*/
type BlockGroupArray []BlockGroup

func (blockArray *BlockGroupArray) addBlockGroup(group BlockGroup) {
	*blockArray = append(*blockArray, group)
}

/*
BlockArray is an array of Blocks
*/
type BlockArray []Block

func (blockArray *BlockArray) addBlock(block Block) {
	*blockArray = append(*blockArray, block)
}

/*
BlockIDList is a list of Block ID's.
It's used for looking up the actual Block
data.
*/
type BlockIDList []string

func (idList *BlockIDList) addID(id string) {
	*idList = append(*idList, id)
}

/*
BlockGroup is a container struct for multiple instances
of the same block. So you can have a single block coded
multiple times by different (or identical) users.
*/
type BlockGroup struct {
	ID          string  `json:"block-id"`
	Blocks      []Block `json:"blocks"`
	Training    bool    `json:"training"`
	Reliability bool    `json:"reliability"`
}

func (group *BlockGroup) addBlock(block Block) error {
	if block.ID != group.ID {
		return ErrAddBlockFailed
	}
	if !block.Training && !block.Reliability {

		if len(group.Blocks) == numRealBlockPasses {
			return ErrBlockGroupFull
		}
		group.Blocks = append(group.Blocks, block)
		return nil
	} else if block.Reliability {
		if group.coderPresent(block.LabKey, block.Coder) {
			return ErrBlockGroupFull
		}
		group.Blocks = append(group.Blocks, block)
		return nil

	} else if block.Training {
		group.Blocks = append(group.Blocks, block)
		return nil
	}
	return ErrAddBlockFailed
}

func (group *BlockGroup) getUsersBlocks(labKey, username string) []Block {
	var blocks []Block
	for _, item := range group.Blocks {
		if item.LabKey == labKey && item.Coder == username {
			blocks = append(blocks, item)
		}
	}
	return blocks
}

func (group *BlockGroup) coderPresent(labKey, coder string) bool {
	for _, element := range group.Blocks {
		if element.LabKey == labKey && element.Coder == coder {
			return true
		}
	}
	return false
}

func (group *BlockGroup) encode() ([]byte, error) {
	data, err := json.MarshalIndent(group, "", "  ")
	if err != nil {
		return data, err
	}
	return data, nil
}

func decodeBlockGroupJSON(data []byte) (*BlockGroup, error) {
	var group *BlockGroup
	err := json.Unmarshal(data, &group)
	if err != nil {
		return nil, err
	}
	return group, nil
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

// func (db *LabelsDB) addBlock(block Block) error {
// 	encodedBlock, err := block.encode()
// 	if err != nil {
// 		return err
// 	}
//
// 	updateErr := db.db.Update(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(labelsBucket))
// 		err := bucket.Put([]byte(block.ID), encodedBlock)
// 		return err
// 	})
// 	return updateErr
// }

// func (db *LabelsDB) getBlock(blockID string) (*Block, error) {
//
// 	fmt.Println("Trying to retrieve block data: ")
// 	fmt.Println(blockID)
//
// 	var exists bool
// 	var BlockGroup []byte
//
// 	db.db.View(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(labelsBucket))
// 		BlockGroup = bucket.Get([]byte(blockID))
//
// 		// lab key doesn't exist
// 		if BlockGroup == nil {
// 			exists = false
// 		} else {
// 			exists = true
// 		}
// 		return nil
// 	})
//
// 	if !exists {
// 		return &Block{}, ErrWorkItemDoesntExist
// 	}
//
// 	block, err := decodeBlockJSON(BlockGroup)
// 	//fmt.Println(labData)
// 	if err != nil {
// 		return &Block{}, ErrWorkItemDoesntExist
// 	}
// 	return block, nil
// }

func (db *LabelsDB) getBlockGroup(blockIDs []string) ([]BlockGroup, error) {
	var blocks []BlockGroup

	for _, id := range blockIDs {
		block, err := db.getBlock(id)
		if err != nil {
			return blocks, ErrCouldntFindLabeledBlock
		}
		blocks = append(blocks, *block)
		//blocks.append(*block)
	}
	return blocks, nil
}

func (db *LabelsDB) addBlock(block Block) error {
	fmt.Println("Trying to retrieve block data: ")
	fmt.Println(block.ID)

	var exists bool
	var groupData []byte

	db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labelsBucket))
		groupData = bucket.Get([]byte(block.ID))

		// block group doesn't exist
		if groupData == nil {
			exists = false
		} else {
			exists = true
		}
		return nil
	})

	if !exists {
		newBlockGroup := BlockGroup{ID: block.ID}
		newBlockGroup.addBlock(block)
		if block.Training {
			newBlockGroup.Training = true
		}
		if block.Reliability {
			newBlockGroup.Reliability = true
		}

		newEncodedBlockGroup, newBlockEncodeErr := newBlockGroup.encode()
		if newBlockEncodeErr != nil {
			return newBlockEncodeErr
		}

		updateErr := db.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(labelsBucket))
			err := bucket.Put([]byte(newBlockGroup.ID), newEncodedBlockGroup)
			return err
		})
		return updateErr
	}

	blockGroup, blockDecodeErr := decodeBlockGroupJSON(groupData)
	if blockDecodeErr != nil {
		return blockDecodeErr
	}

	addBlockErr := blockGroup.addBlock(block)
	if addBlockErr != nil {
		return addBlockErr
	}

	encodedBlockGroup, blockEncodeErr := blockGroup.encode()
	if blockEncodeErr != nil {
		return blockEncodeErr
	}

	updateErr := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labelsBucket))
		err := bucket.Put([]byte(blockGroup.ID), encodedBlockGroup)
		return err
	})
	return updateErr
}

func (db *LabelsDB) getBlock(blockID string) (*BlockGroup, error) {
	fmt.Println("Trying to retrieve block data: ")
	fmt.Println(blockID)

	var exists bool
	var groupData []byte

	db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labelsBucket))
		groupData = bucket.Get([]byte(blockID))

		// block group doesn't exist
		if groupData == nil {
			exists = false
		} else {
			exists = true
		}
		return nil
	})

	if !exists {
		return &BlockGroup{}, ErrWorkItemDoesntExist
	}

	blockGroup, err := decodeBlockGroupJSON(groupData)
	//fmt.Println(labData)
	if err != nil {
		return &BlockGroup{}, ErrWorkItemDoesntExist
	}
	return blockGroup, nil
}
