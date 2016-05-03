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

	/*
	   workItemMap is a map of WorkItem to bool.
	   The boolean value represents whether or not
	   the particular work item is active or not.
	   Active means it's been sent out for coding
	   and has not been submitted back yet.
	*/
	workItemMap WorkItemMap

	/*
		activeWorkItems is a map of WorkItem ID's. All the ID's
		represent blocks which have been sent out to be worked on.
		(i.e. active blocks)

		format:

		map["30_13_coderJS_final-2:::6" : true , "32_13_coderJS_final-5:::16" : true, ......]

		The ID's are a concatenation of the name of the CLAN file of origin
		and the block index, separated by ":::".
	*/
	activeWorkItems ActiveDataQueue
)

const (
	/*
		dataPath is the path to where all the
		CLAN files and audio blocks are going
		to be stored
	*/
	dataPath = "data"

	/*
		numBlocksToSend is the number of blocks that will be sent
		from any given CLAN file to the end user upon request
	*/
	numBlocksToSend = 5
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

	workItemMap := dataMap.partitionIntoWorkItemsMap()

	fmt.Println("# of work items: ", len(workItems))
	fmt.Println("# of work items map: ", len(workItemMap))

	for key, value := range workItemMap {
		// data, err := json.MarshalIndent(item, "", "  ")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		fmt.Println(key)
		fmt.Println(value)
	}

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
	http.HandleFunc("/getblocks/", getWorkGroupHandler)
	http.ListenAndServe(":8080", nil)

}
