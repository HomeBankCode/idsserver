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

	var labInfoReq BlockRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &labInfoReq)

	//fmt.Printf("%+v", labInfoReq)

	lab := labsDB.getLab(labInfoReq.LabKey)

	fmt.Println(labsDB)

	//	fmt.Println(lab)
	json.NewEncoder(w).Encode(lab)

	//fmt.Println(labInfoReq.AdminKey)

}

/*func getWorkGroupHandler(w http.ResponseWriter, r *http.Request) {*/
//err := r.ParseForm()
//if err != nil {
//log.Fatal(err)
//}

//var workGroupRequest WorkGroupRequest

//jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

//if err != nil {
//panic(err)
//}

//fmt.Println()
//json.Unmarshal(jsonDataFromHTTP, &workGroupRequest)

//fmt.Println(workGroupRequest)

//newWorkGroup := NewWorkGroup(workGroupRequest)

//fmt.Println(newWorkGroup)

//w.Header().Set("Content-Type", "application/zip")
//w.Header().Set("Content-Disposition", "attachment; filename='file.zip'")

/*}*/
