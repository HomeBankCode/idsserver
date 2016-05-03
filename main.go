package main

import (
	"fmt"
	"net/http"
	"os"
)

var labsDB *LabsDB

/*
dataPath is the path to where all the
CLAN files and audio blocks are going
to be stored
*/
const dataPath = "data"

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "this is the mainHandler")
}

func main() {

	manifestFile = os.Args[1]

	fmt.Println(manifestFile)
	dataMap := fillDataMap()
	fmt.Printf("\n\n")
	fmt.Println(dataMap["30_13_coderJS_final-8"])

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
