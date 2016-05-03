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
	ClanFile   string
	BlockPaths map[int]string
}

/*
DataMap is a map of integer ID's to
DataGroups. DataGroups give info for path
lookups to the relevant files
*/
type DataMap map[string]*DataGroup

/*
ActiveDataQueue is a map of ID's to booleans.
The bool values represent whether a
CLAN file is actively being worked on
or not.
*/
type ActiveDataQueue map[uint]bool

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
		} else {
			index, err := strconv.Atoi(line[1])
			if err != nil {
				log.Fatal(err)
			}
			currDataGroup.BlockPaths[index] = line[2]
		}
	}
	return dataMap
}

func (dataMap DataMap) partitionIntoWorkItems() []WorkItem {
	var (
		workItems    []WorkItem
		currWorkItem WorkItem
	)

	for key, value := range dataMap {
		fmt.Printf("key: \n%s\n\nvalue: \n%v\n\n", key, value)

		currWorkItem.FileName = value.ClanFile

		count := 0

		var (
			groupLength    = len(value.BlockPaths)
			numFullItems   = groupLength / numBlocksSent
			itemsRemainder = groupLength % numBlocksSent
		)

		fmt.Printf("groupLength: %d\n", groupLength)
		fmt.Printf("numFullItems: %d\n", numFullItems)
		fmt.Printf("itemsRemainder: %d\n", itemsRemainder)

		for blockKey, blockValue := range value.BlockPaths {
			if count == numBlocksSent-1 {
				workItems = append(workItems, currWorkItem)
				currWorkItem = WorkItem{FileName: value.ClanFile}
				count = 0
			}
			fmt.Printf("\nblockKey: %d\nblockValue: %v\n", blockKey, blockValue)
			count++
		}
	}
	return workItems
}
