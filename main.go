package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	/*
		labsDB is the wrapper around the lab user info database.
		It keeps track of all the users and the current jobs that
		have been handed out to them.
	*/
	labsDB *LabsDB

	/*
		workDB is the wrapper around the database that keeps track of
		all the jobs that have been assigned and yet to be assigned.
	*/
	workDB *WorkDB

	/*
	 manifestFile is the path to the path_manifests.csv file.
	 This contains the name of the clan files and the paths to
	 all the blocks that are a part of them.

	 format:

	 [clan_file, block_index, path_to_block]

	*/
	manifestFile string

	/*
		dataMap is the global map of CLAN files to block paths
	*/
	dataMap DataMap
)

const (
	/*
		dataPath is the path to where all the
		CLAN files and audio blocks are going
		to be stored
	*/
	dataPath = "data"

	/*
		numBlocksSent is the number of blocks that will be sent
		from any given CLAN file to the end user upon request
	*/
	numBlocksSent = 5
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "this is the mainHandler")
}

func main() {

	manifestFile = os.Args[1]

	dataMap := fillDataMap()

	//fmt.Printf("\n\n")
	//fmt.Println(dataMap["30_13_coderJS_final-8"])

	workItems := dataMap.partitionIntoWorkItems()

	fmt.Println("# of work items: ", len(workItems))

	// Open the LabsDB
	labsDB, err := LoadLabsDB()
	if err != nil {
		log.Fatal(err)
	}
	defer labsDB.Close()

	workDB, err := LoadWorkDB()
	if err != nil {
		log.Fatal(err)
	}
	defer workDB.Close()

	// labsDB := LabsDB{db: new(bolt.DB)}
	//
	// labsDBOpenErr := labsDB.openDB()
	//
	// if labsDBOpenErr != nil {
	// 	log.Fatal(labsDBOpenErr)
	// }
	// defer labsDB.Close()
	//
	// labsDB.addUser("123456", "andrei")
	// labsDB.addUser("123457", "alice")
	// labsDB.addUser("123458", "bob")
	// labsDB.addUser("123459", "sally")
	// labsDB.addUser("123450", "joe")
	//
	// labs := labsDB.getAllLabs()
	//
	// for _, lab := range labs {
	// 	fmt.Println(*lab)
	// }

	http.HandleFunc("/", mainHandler)
	http.ListenAndServe(":8080", nil)

}
