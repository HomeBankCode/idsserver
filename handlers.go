package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

/*
GetBlocksRequest is a struct representing a request
sent to the server asking for a new WorkGroup
*/
type GetBlocksRequest struct {
	LabKey   string `json:"lab-key"`
	Username string `json:"username"`
}

func getWorkGroupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var blocksRequest GetBlocksRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &blocksRequest)

	fmt.Println(blocksRequest)

	newWorkGroup := NewWorkGroup()

	fmt.Println(newWorkGroup)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename='file.zip'")

}
