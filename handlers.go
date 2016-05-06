package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

/*
BlockRequest is a struct representing a request
sent to the server asking for a single block (i.e. WorkItem)
*/
type BlockRequest struct {
	LabKey   string `json:"lab-key"`
	Username string `json:"username"`
	//NumItems int    `json:"num-items"`
}

func (br *BlockRequest) userID() string {
	return br.LabKey + ":::" + br.Username
}

func (br *BlockRequest) userFromDB() User {
	user := labsDB.getUser(br.LabKey, br.Username)
	return user
}

/*
WorkGroupRequest is a struct representing a request
sent to the server asking for a new WorkGroup
*/
type WorkGroupRequest struct {
	LabKey   string `json:"lab-key"`
	Username string `json:"username"`
	NumItems int    `json:"num-items"`
}

func (wgr *WorkGroupRequest) toBlockRequest() BlockRequest {
	return BlockRequest{LabKey: wgr.LabKey, Username: wgr.Username}
}

func getBlockHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var blockRequest BlockRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &blockRequest)

	fmt.Println(blockRequest)

	fmt.Println(blockRequest.userID())

	workItem, err := chooseUniqueWorkItem(blockRequest)

	blockPath := workItem.BlockPath
	blockName := path.Base(blockPath)

	dispositionString := "attachment; filename=" + blockName

	fmt.Println(workItem)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", dispositionString)

	http.ServeFile(w, r, workItem.BlockPath)

}

func getWorkGroupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var workGroupRequest WorkGroupRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &workGroupRequest)

	fmt.Println(workGroupRequest)

	newWorkGroup := NewWorkGroup(workGroupRequest)

	fmt.Println(newWorkGroup)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename='file.zip'")

}
