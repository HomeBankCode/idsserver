package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "this is the mainHandler")
}

/*
IDSRequest is a struct representing a request
sent to the server asking for a single block (i.e. WorkItem)
*/
type IDSRequest struct {
	LabKey   string `json:"lab-key"`
	LabName  string `json:"lab-name"`
	Username string `json:"username"`
	//NumItems int    `json:"num-items"`
}

/*
SubmissionRequest is a struct representing the
classifications of a block being sent back to the
server
*/
type SubmissionRequest struct {
	Block Block `json:"block"`
}

func (br *IDSRequest) userID() string {
	return br.LabKey + ":::" + br.Username
}

func (br *IDSRequest) userFromDB() User {
	user := labsDB.getUser(br.LabKey, br.Username)
	return user
}

/*
ShutdownRequest is a JSON encoded request to
shutdown the server. This will tell the server
to persist the current state to disk and shut down
*/
type ShutdownRequest struct {
	AdminKey string `json:"admin-key"`
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

func (wgr *WorkGroupRequest) toIDSRequest() IDSRequest {
	return IDSRequest{LabKey: wgr.LabKey, Username: wgr.Username}
}

func getBlockHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var blockRequest IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &blockRequest)

	fmt.Println(blockRequest)

	fmt.Println(blockRequest.userID())

	workItem, err := chooseUniqueWorkItem(blockRequest)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("	HTTP status code returned"))
		return
	}

	blockPath := workItem.BlockPath
	blockName := path.Base(blockPath)
	filename := path.Join(workItem.FileName, blockName)

	dispositionString := "attachment; filename=" + filename

	fmt.Println(workItem)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", dispositionString)

	http.ServeFile(w, r, workItem.BlockPath)

}

func shutDownHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var shutdownReq ShutdownRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &shutdownReq)

	fmt.Println(shutdownReq)

	fmt.Println(shutdownReq.AdminKey)

	if shutdownReq.AdminKey == mainConfig.AdminKey {
		shutDown()
	}

}

func labInfoHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var labInfoReq IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &labInfoReq)

	lab := labsDB.getLab(labInfoReq.LabKey)

	fmt.Println(labsDB)

	json.NewEncoder(w).Encode(lab)

}

func allLabInfoHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var labInfoReq IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &labInfoReq)

	labs := labsDB.getAllLabs()

	fmt.Println(labsDB)

	json.NewEncoder(w).Encode(labs)

}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var addUserReq IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &addUserReq)
	fmt.Println(addUserReq)

	labsDB.addUser(addUserReq.LabKey, addUserReq.LabName, addUserReq.Username)

}

func submitLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("got a submission request")
	//submitReq := SubmissionRequest{Blocks: make(map[string][]map[string]string)}
	var block Block
	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &block)
	//blocks := submissionParse(submitReq)

	//fmt.Println(blocks)
	fmt.Println(block)

}
