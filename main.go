package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		configFile is the path to the main config file,
		which has the server admin keys and other metadata
	*/
	configFile string

	/*
		mainConfig is the Config struct produced from reading
		the configFile
	*/
	mainConfig Config

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

type Config struct {
	AdminKey string `json:"admin-key"`
}

func readConfigFile(path string) Config {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	json.Unmarshal(file, &config)

	return config
}

func shutDown() {

}

func main() {

	configFile = os.Args[1]
	manifestFile = os.Args[2]

	// Open the LabsDB
	labsDB = LoadLabsDB()
	defer labsDB.Close()

	workDB = LoadWorkDB()
	defer workDB.Close()

	mainConfig = readConfigFile(configFile)

	fmt.Println("mainConfig: ")
	fmt.Println(mainConfig)

	//	return

	dataMap := fillDataMap()

	workItemMap = dataMap.partitionIntoWorkItemsMap()

	//fmt.Println("# of work items: ", len(workItems))
	fmt.Println("# of work items map: ", len(workItemMap))

	// for key, value := range workItemMap {
	//
	// 	fmt.Println(key)
	// 	fmt.Println(value)
	// }

	labsDB.addUser("123456", "andrei")
	labsDB.addUser("123457", "alice")
	labsDB.addUser("123458", "bob")
	labsDB.addUser("123459", "sally")
	labsDB.addUser("123450", "joe")

	//	labs := labsDB.getAllLabs()

	/* for _, lab := range labs {*/
	//fmt.Println(*lab)
	/*}*/

	labsDB.addUser("1234567654321", "billybob")

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/getblock/", getBlockHandler)
	http.HandleFunc("/labinfo/", labInfoHandler)
	http.HandleFunc("/shutdown/", shutDownHandler)

	http.ListenAndServe(":8080", nil)

}
