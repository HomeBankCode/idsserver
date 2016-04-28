package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
)

var labsDB *LabsDB

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "this is the mainHandler")
}

func main() {

	labsDB := LabsDB{db: new(bolt.DB)}

	labsDBOpenErr := labsDB.openDB()

	if labsDBOpenErr != nil {
		log.Fatal(labsDBOpenErr)
	}
	defer labsDB.close()

	http.HandleFunc("/", mainHandler)
	http.ListenAndServe(":8080", nil)

}
