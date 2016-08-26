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

	// ErrLabNotInBlockGroup means that this lab isn't in the BlockGroup
	ErrLabNotInBlockGroup = errors.New("Lab not in BlockGroup")

	// ErrInstanceNotInGroup means that the instance number was not found when
	// scanning the group's blocks
	ErrInstanceNotInGroup = errors.New("Instance not in group")
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

/*
BlockGroupArray is an array of BlockGroups
*/
type BlockGroupArray []BlockGroup

func (blockGroupArray *BlockGroupArray) addBlockGroup(group BlockGroup) {
	*blockGroupArray = append(*blockGroupArray, group)
}

func (blockGroupArray *BlockGroupArray) filterLab(labKey string) (BlockArray, error) {
	var blockArray BlockArray

	for _, group := range *blockGroupArray {
		for _, block := range group.Blocks {
			if block.LabKey == labKey {
				blockArray.addBlock(block)
			}
		}
	}
	if len(blockArray) == 0 {
		return blockArray, ErrLabNotInBlockGroup
	}
	return blockArray, nil
}

func (blockGroupArray *BlockGroupArray) filterUser(username string) (BlockArray, error) {
	var blockArray BlockArray

	for _, group := range *blockGroupArray {
		for _, block := range group.Blocks {
			if block.Coder == username {
				blockArray.addBlock(block)
			}
		}
	}
	if len(blockArray) == 0 {
		return blockArray, ErrLabNotInBlockGroup
	}
	return blockArray, nil
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
InstanceMap is a map of Block ID's to
instance numbers.
*/
type InstanceMap map[string]*InstanceList

/*
InstanceList is a list of Block instance numbers
*/
type InstanceList []int

// NewInstanceList return pointer to InstanceList with single entry
func NewInstanceList(firstInst int) *InstanceList {
	return &InstanceList{firstInst}
}

func (instList *InstanceList) addInstance(inst int) {
	*instList = append(*instList, inst)
}

func (instList *InstanceList) contains(inst int) bool {
	for _, instance := range *instList {
		if inst == instance {
			return true
		}
	}
	return false
}

/*
BlockGroup is a container struct for multiple instances
of the same block. So you can have a single block coded
multiple times by different (or identical) users.
*/
type BlockGroup struct {
	ID          string     `json:"block-id"`
	Blocks      BlockArray `json:"blocks"`
	Training    bool       `json:"training"`
	Reliability bool       `json:"reliability"`
}

func (group *BlockGroup) addBlock(block Block) error {
	if block.ID != group.ID {
		return ErrAddBlockFailed
	}
	if !block.Training && !block.Reliability {

		if len(group.Blocks) == numRealBlockPasses {
			return ErrBlockGroupFull
		}
		block.Instance = len(group.Blocks)
		group.Blocks.addBlock(block)
		return nil
	} else if block.Reliability {
		if group.coderPresent(block.LabKey, block.Coder) {
			return ErrBlockGroupFull
		}
		block.Instance = len(group.Blocks)
		group.Blocks.addBlock(block)
		return nil

	} else if block.Training {
		block.Instance = len(group.Blocks)
		group.Blocks.addBlock(block)
		return nil
	}
	return ErrAddBlockFailed
}

func (group *BlockGroup) deleteInstance(instance int) error {
	var newBlocks BlockArray
	for _, block := range group.Blocks {
		if block.Instance != instance {
			block.Instance = len(newBlocks)
			newBlocks.addBlock(block)
		}
	}
	if len(newBlocks) == len(group.Blocks) {
		return ErrInstanceNotInGroup
	}
	group.Blocks = newBlocks
	return nil
}

func (group *BlockGroup) deleteInstances(instances *InstanceList) error {
	var newBlocks BlockArray

	for _, block := range group.Blocks {
		if !instances.contains(block.Instance) {
			block.Instance = len(newBlocks)
			newBlocks.addBlock(block)
		}
	}

	if len(newBlocks) == len(group.Blocks) {
		return ErrInstanceNotInGroup
	}
	group.Blocks = newBlocks
	return nil
}

func (group *BlockGroup) getUsersBlocks(labKey, username string) BlockArray {
	var blocks BlockArray
	for _, block := range group.Blocks {
		if block.LabKey == labKey && block.Coder == username {
			blocks.addBlock(block)
		}
	}
	return blocks
}

func (group *BlockGroup) coderPresent(labKey, coder string) bool {
	for _, block := range group.Blocks {
		if block.LabKey == labKey && block.Coder == coder {
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
	Instance    int    `json:"block-instance"`
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

func (db *LabelsDB) getBlockGroup(blockIDs []string) (BlockGroupArray, error) {
	var blocks BlockGroupArray

	for _, id := range blockIDs {
		block, err := db.getBlock(id)
		if err != nil {
			return blocks, ErrCouldntFindLabeledBlock
		}
		blocks.addBlockGroup(*block)
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
	if err != nil {
		return &BlockGroup{}, ErrWorkItemDoesntExist
	}
	return blockGroup, nil
}

func (db *LabelsDB) setBlockGroup(group BlockGroup) error {
	encodedBlockGroup, encodeErr := group.encode()
	if encodeErr != nil {
		return encodeErr
	}

	updateErr := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(labelsBucket))
		err := bucket.Put([]byte(group.ID), encodedBlockGroup)
		return err
	})
	return updateErr
}

func (db *LabelsDB) getAllBlockGroups() (BlockGroupArray, error) {
	var blockGroupArray BlockGroupArray

	scanErr := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(labelsBucket))
		c := b.Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			blockGroup, groupDecodeErr := decodeBlockGroupJSON(value)
			if groupDecodeErr != nil {
				return groupDecodeErr
			}
			blockGroupArray.addBlockGroup(*blockGroup)
			//fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	if scanErr != nil {
		return blockGroupArray, scanErr
	}
	return blockGroupArray, nil
}

func (db *LabelsDB) deleteBlocks(instanceMap InstanceMap) (bool, error) {
	fmt.Println("\n\n\ninside of labelsDB.deleteBlocks()")

	for blockID, instanceList := range instanceMap {

		// Get the requested BlockGroup
		blockGroup, getGroupErr := labelsDB.getBlock(blockID)
		if getGroupErr != nil {
			return false, getGroupErr
		}
		fmt.Println("\n\ngot the block: ")
		fmt.Println(blockGroup)
		fmt.Printf("\n\n")

		blockGroup.deleteInstances(instanceList)

		// for _, instance := range instanceList {
		// 	fmt.Println("\nbefore blockGroup.deleteInstance()")
		// 	fmt.Println(blockGroup)
		// 	// Delete the requested instances of the block
		// 	blockGroup.deleteInstance(instance)
		// 	fmt.Println("\nafter blockGroup.deleteInstance()")
		// 	fmt.Println(blockGroup)
		// }

		/*
			If there are no more instances of the block left, then we
			need to delete the entire BlockGroup from the LabelsDB.
			We delete the Block ID from the keys of the LabelsDB
		*/
		if len(blockGroup.Blocks) == 0 {
			fmt.Println("\n\ninside of len(blockGroup.Blocks) == 0")
			deleteKeyErr := db.db.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(labelsBucket))
				delKeyErr := bucket.Delete([]byte(blockGroup.ID))
				return delKeyErr
			})

			if deleteKeyErr != nil {
				return false, deleteKeyErr
			}
			return true, nil
		}

		// Set the updated version of the group, with instance deleted
		setNewGroupErr := labelsDB.setBlockGroup(*blockGroup)
		if setNewGroupErr != nil {
			return false, setNewGroupErr
		}
	}
	return false, nil
}
