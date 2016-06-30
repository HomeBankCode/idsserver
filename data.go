package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
)

/*
DataGroup is a collection of paths to
CLAN files and blocks (audio clips) which
are going to be served to users.
*/
type DataGroup struct {
	ClanFile    string
	Training    bool
	Reliability bool
	BlockPaths  map[int]string
}

/*
DataMap is a map of integer ID's to
DataGroups. DataGroups give info for path
lookups to the relevant files
*/
type DataMap map[string]*DataGroup

/*
ActiveDataQueue is a map of WorkItem
ID's. All these ID's represent the blocks
which have been sent out to be worked on.
(i.e. active blocks)
*/
type ActiveDataQueue map[string]WorkItem

/*
WorkItemMap is a map of WorkItem ID's to WorkItems.
*/
type WorkItemMap map[string]WorkItem

/*
fillDataMap reads a path_manifest.csv file and
fills a DataMap with all the paths to the CLAN
files and blocks
*/
func fillDataMap() DataMap {
	file, _ := os.Open(manifestFile)
	reader := csv.NewReader(bufio.NewReader(file))

	lines, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	dataMap := make(DataMap)
	var currDataGroup = &DataGroup{}
	var currFile = ""

	for i, line := range lines {
		// skip the header
		if i == 0 {
			continue
		}

		// we're on a new CLAN file
		if line[0] != currFile {
			// reset currFile
			currFile = line[0]
			// construct new DataGroup for the new file
			currDataGroup = &DataGroup{ClanFile: currFile, BlockPaths: make(map[int]string)}
			// assign a key/value to the dataMap for this new group
			dataMap[currFile] = currDataGroup
			// set the value of the first block of this new file
			index, err := strconv.Atoi(line[1])
			if err != nil {
				log.Fatal(err)
			}
			// set new BlockPaths index/path
			currDataGroup.BlockPaths[index] = line[2]
			// set Training and Reliability variables
			training, trainErr := strconv.ParseBool(line[3])

			currDataGroup.Training = training
			if trainErr != nil {
				log.Fatal("training parsing error")
			}
			reliability, reliaErr := strconv.ParseBool(line[4])
			if reliaErr != nil {
				log.Fatal("reliability parsing error")
			}
			currDataGroup.Reliability = reliability

		} else {
			index, err := strconv.Atoi(line[1])
			if err != nil {
				log.Fatal(err)
			}
			currDataGroup.BlockPaths[index] = line[2]

		}
		if currDataGroup.Training {
			fmt.Println(currDataGroup)
		}
	}
	return dataMap
}

/*
partitionIntoWorkItemsMap breaks up the DataMap into
an array of WorkItem's and returns it
*/
func (dataMap DataMap) partitionIntoWorkItemsMap() WorkItemMap {
	var (
		workItems    = make(WorkItemMap)
		currWorkItem = WorkItem{}
	)

	for key, value := range dataMap {
		currWorkItem = WorkItem{}
		currWorkItem.FileName = key
		currWorkItem.Training = value.Training
		currWorkItem.Reliability = value.Reliability

		for blockKey, blockValue := range value.BlockPaths {
			currWorkItem = WorkItem{}

			currWorkItem.ID = key + ":::" + strconv.Itoa(blockKey)
			currWorkItem.Block = blockKey
			currWorkItem.Active = false
			currWorkItem.FileName = value.ClanFile
			currWorkItem.BlockPath = blockValue
			currWorkItem.Training = value.Training
			currWorkItem.Reliability = value.Reliability

			workItems[currWorkItem.ID] = currWorkItem
			if currWorkItem.Training {
				fmt.Println(currWorkItem)
			}
		}
	}
	return workItems
}
