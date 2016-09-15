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

	// ErrBlockAlreadyCodedByUser means that this user has already submitted
	// a response to this block
	ErrBlockAlreadyCodedByUser = errors.New("This block has already been coded by this user")

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

	// number of times a (real) block should be coded.
	// training and reliability blocks can be coded arbitrarily
	// many times
	numRealBlockPasses = 3
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
	ID          string     `json:"block_id"`
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
			return ErrBlockAlreadyCodedByUser
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
	ClanFile    string `json:"clan_file"`
	Index       int    `json:"block_index"`
	Instance    int    `json:"block_instance"`
	Clips       []Clip `json:"clips"`
	FanOrMan    bool   `json:"fan_or_man"`
	DontShare   bool   `json:"dont_share"`
	ID          string `json:"id"`
	Coder       string `json:"coder"`
	LabKey      string `json:"lab_key"`
	LabName     string `json:"lab_name"`
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
	Index           int    `json:"clip_index"`
	Tier            string `json:"clip_tier"`
	Multiline       bool   `json:"multiline"`
	MultiTierParent string `json:"multi_tier_parent"`
	StartTime       string `json:"start_time"`
	OffsetTime      string `json:"offset_time"`
	TimeStamp       string `json:"timestamp"`
	Classification  string `json:"classification"`
	LabelDate       string `json:"label_date"`
	Coder           string `json:"coder"`
	GenderLabel     string `json:"gender_label"`
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
		blockWorkItem := workItemMap[block.ID]
		blockWorkItem.TimesCoded = len(newBlockGroup.Blocks)
		workItemMap[block.ID] = blockWorkItem
		workDB.persistWorkItem(blockWorkItem)
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

	blockWorkItem := workItemMap[block.ID]
	blockWorkItem.TimesCoded = len(blockGroup.Blocks)
	workItemMap[block.ID] = blockWorkItem
	workDB.persistWorkItem(blockWorkItem)

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
	keyWasDeleted := false

	for blockID, instanceList := range instanceMap {
		fmt.Println("\n\n\ndealing with block: ", blockID)
		fmt.Printf("\n\n")
		// Get the requested BlockGroup
		blockGroup, getGroupErr := labelsDB.getBlock(blockID)
		if getGroupErr != nil {
			return false, getGroupErr
		}

		fmt.Println("before blockGroup.deleteInstances()")
		fmt.Printf("\n\n")
		fmt.Println(blockGroup)
		blockGroup.deleteInstances(instanceList)
		fmt.Println("\n\nafter blockGroup.deleteInstances()")
		fmt.Printf("\n\n")
		fmt.Println(blockGroup)

		/*
			If there are no more instances of the block left, then we
			need to delete the entire BlockGroup from the LabelsDB.
			We delete the Block ID from the keys of the LabelsDB
		*/
		if len(blockGroup.Blocks) == 0 {
			deleteKeyErr := db.db.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(labelsBucket))
				delKeyErr := bucket.Delete([]byte(blockGroup.ID))
				return delKeyErr
			})

			if deleteKeyErr != nil {
				return false, deleteKeyErr
			}
			fmt.Println("\n\nblockGroup key delete worked")
			keyWasDeleted = true

			blockWorkItem := workItemMap[blockGroup.ID]
			blockWorkItem.TimesCoded = len(blockGroup.Blocks)
			workItemMap[blockGroup.ID] = blockWorkItem
			workDB.persistWorkItem(blockWorkItem)

			continue
		}

		// Set the updated version of the group, with instance deleted
		setNewGroupErr := labelsDB.setBlockGroup(*blockGroup)
		if setNewGroupErr != nil {
			return keyWasDeleted, setNewGroupErr
		}

		blockWorkItem := workItemMap[blockGroup.ID]
		blockWorkItem.TimesCoded = len(blockGroup.Blocks)
		workItemMap[blockGroup.ID] = blockWorkItem
		workDB.persistWorkItem(blockWorkItem)

		fmt.Println("\n\ngot past labelsDB.setBlockGroup()")
	}
	fmt.Println("outside of labelsDB.deleteBlocks() for loop, about to return")
	return keyWasDeleted, nil
}

/*
	deleteSingleBlock deletes a single instance of a block from the LabelsDB
	and deletes the block's entry from the coder's PastWorkItems list. If the
	user submitted more than one instance of that particular block,
	then we leave the ID in the PastWorkItems list (only deleted one
	instance of it).
*/
func (db *LabelsDB) deleteSingleBlock(labKey, coder, blockID string, instance int) error {
	// make map
	singleInstanceMap := make(InstanceMap)
	singleInstanceMap[blockID] = NewInstanceList(instance)

	// Delete from labelsDB. Function might also delete the BlockGroup entirely,
	// so it returns a flag
	groupWasDeleted, _ := db.deleteBlocks(singleInstanceMap)

	if !groupWasDeleted {
		fmt.Println("\n\ninside !groupWasDeleted")
		blockGroup, getGroupErr := db.getBlock(blockID)
		if getGroupErr != nil {
			return getGroupErr
		}
		// if there's a block instance with this coder still, it means they submitted
		// more than one instance of the same block, so we keep the PastWorkItem entry,
		// otherwise we delete it
		if !blockGroup.coderPresent(labKey, coder) {
			user, getUserErr := labsDB.getUser(labKey, coder)
			if getUserErr != nil {
				return getUserErr
			}
			user.deletePastItem(blockID)
			labsDB.setUser(user)
		}
	} else {
		fmt.Println("inside groupWasDeleted")
		user, getUserErr := labsDB.getUser(labKey, coder)
		fmt.Println("\n\nuser from labsDB.getUser():")
		fmt.Println("\n\n", user)

		if getUserErr != nil {
			return getUserErr
		}
		user.deletePastItem(blockID)
		labsDB.setUser(user)
	}
	return nil
}

/*
	deleteUserBlocks builds an InstanceMap of all the user's completed
	instances of blocks, and then pass that map to labelsDB.deleteBlocks()
	function. Then we need to delete all of those block entries from the
	user's PastWorkItems list.
*/
func (db *LabelsDB) deleteUserBlocks(labKey, username string) error {
	// get the user
	user, getUserErr := labsDB.getUser(labKey, username)
	if getUserErr != nil {
		return getUserErr
	}

	// build user's block instance map
	userInstances, userInstanceErr := user.getPastBlockInstanceMap()
	if userInstanceErr != nil {
		return userInstanceErr
	}

	// delete those instances
	_, deleteUserInstErr := labelsDB.deleteBlocks(userInstances)
	if deleteUserInstErr != nil {
		return deleteUserInstErr
	}

	// clear out user's PastWorkItems
	user.PastWorkItems = nil
	user.CompleteTrainBlocks = nil
	user.CompleteRelBlocks = nil
	labsDB.setUser(user)
	return nil
}

/*
	deleteLabBlocks builds an InstanceMap of all the lab's completed
	instances of blocks, and then pass that map to labelsDB.deleteBlocks()
	function. Then we need to delete all block entries from all of the lab's
	user's PastWorkItems lists.
*/
func (db *LabelsDB) deleteLabBlocks(labKey string) error {

	// get the lab
	lab, getLabErr := labsDB.getLab(labKey)
	if getLabErr != nil {
		return getLabErr
	}

	fmt.Println("inside labelsDB.deleteLabBlocks() ----- got the lab")
	// get all block instances submitted by the lab
	labInstanceMap, labInstanceErr := lab.getPastBlockInstanceMap()
	if labInstanceErr != nil {
		return labInstanceErr
	}

	// delete all those instances from LabelsDB
	_, deleteLabInstErr := db.deleteBlocks(labInstanceMap)
	if deleteLabInstErr != nil {
		return deleteLabInstErr
	}

	// clear out all users' PastWorkItems
	for index, user := range lab.Users {
		user.PastWorkItems = nil
		user.CompleteTrainBlocks = nil
		user.CompleteRelBlocks = nil

		lab.Users[index] = user
	}

	labsDB.setLab(lab.Key, lab)
	return nil
}
