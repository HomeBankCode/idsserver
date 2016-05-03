package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

// path to the path_manifests.csv file
var manifestFile string

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
DataActive is a map of ID's to booleans.
The bool values represent whether a
CLAN file is actively being worked on
or not.
*/
type DataActive map[uint]bool

/*
fillDataMap reads a path_manifest.csv file and
fills a DataMap with all the paths to the CLAN
files and blocks
*/
func fillDataMap() DataMap {
	file, _ := os.Open(manifestFile)
	fmt.Println(manifestFile)
	reader := csv.NewReader(bufio.NewReader(file))

	dataMap := make(DataMap)

	var currDataGroup = &DataGroup{}

	var currFile = ""

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		fmt.Printf("\ncurrFile: %s", currFile)
		fmt.Printf("\ncurrGroup: %+v", currDataGroup)

		if record[0] != currFile {
			currFile = record[0]
			currDataGroup = &DataGroup{ClanFile: currFile, BlockPaths: make(map[int]string)}
			dataMap[currFile] = currDataGroup
		} else {

			index, err := strconv.Atoi(record[1])
			if err != nil {
				log.Fatal(err)
			}
			currDataGroup.BlockPaths[index] = record[2]
		}
	}

	return dataMap
}
