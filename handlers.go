package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}

/*
WorkItemDataReq is a request for the label data
for a particular work item from the database.
*/
type WorkItemDataReq struct {
	ItemID string `json:"item-id"`
	LabKey string `json:"lab-key"`
}

/*
WorkItemReleaseReq is a request to inactivate a
group of blocks without coding them.
*/
type WorkItemReleaseReq struct {
	LabKey   string   `json:"lab-key"`
	LabName  string   `json:"lab-name"`
	Username string   `json:"username"`
	BlockIds []string `json:"blocks"`
}

func (br *IDSRequest) userID() string {
	return br.LabKey + ":::" + br.Username
}

func (br *IDSRequest) userFromDB() (User, error) {
	user, err := labsDB.getUser(br.LabKey, br.Username)
	return user, err
}

/*
ShutdownRequest is a JSON encoded request to
shutdown the server. This will tell the server
to persist the current state to disk and shut down
*/
type ShutdownRequest struct {
	AdminKey string `json:"admin-key"`
}

func getBlockHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
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
		http.Error(w, err.Error(), 404)
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
		http.Error(w, err.Error(), 400)
		return
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
		http.Error(w, err.Error(), 400)
		return
	}

	var labInfoReq IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &labInfoReq)

	lab, getLabErr := labsDB.getLab(labInfoReq.LabKey)
	if getLabErr != nil {
		http.Error(w, getLabErr.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(lab)

}

func allLabInfoHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
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
		http.Error(w, err.Error(), 400)
		return
	}

	var addUserReq IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &addUserReq)
	fmt.Println(addUserReq)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsRegistered(addUserReq.LabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}
	labsDB.addUser(addUserReq.LabKey, addUserReq.LabName, addUserReq.Username)

}

func submitLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a submission request")
	var block Block
	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	json.Unmarshal(jsonDataFromHTTP, &block)
	fmt.Printf("\n%#v", block)

	if !labsDB.userExists(block.LabKey, block.Coder) {
		http.Error(w, ErrUserDoesntExist.Error(), 500)
		return
	}

	addBlockErr := labelsDB.addBlock(block)
	if addBlockErr != nil {
		http.Error(w, addBlockErr.Error(), 400)
		return
	}

	workItem := workItemMap[block.ID]
	request := IDSRequest{
		LabKey:   block.LabKey,
		LabName:  block.LabName,
		Username: block.Coder,
	}

	inactivateWorkItem(workItem, request)

}

func getLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a request for work item data")
	var workItemReq WorkItemDataReq

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &workItemReq)
	fmt.Println(workItemReq)

	// requestedWI, exists := workItemMap[workItemReq.ItemID]
	// if !exists {
	// 	http.Error(w, ErrWorkItemDoesntExist.Error(), 400)
	// 	return
	// }

	block, getBlockErr := labelsDB.getBlock(workItemReq.ItemID)
	if getBlockErr != nil {
		http.Error(w, ErrWorkItemDoesntExist.Error(), 400)
		return
	}

	fmt.Println(block)
	json.NewEncoder(w).Encode(block)
}

func getLabLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a request for work item data")
	var workItemReq WorkItemDataReq

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &workItemReq)
	fmt.Println(workItemReq)

	// requestedWI, exists := workItemMap[workItemReq.ItemID]
	// if !exists {
	// 	http.Error(w, ErrWorkItemDoesntExist.Error(), 400)
	// 	return
	// }

	block, getBlockErr := labelsDB.getBlock(workItemReq.ItemID)
	if getBlockErr != nil {
		http.Error(w, ErrWorkItemDoesntExist.Error(), 400)
		return
	}

	fmt.Println(block)
	json.NewEncoder(w).Encode(block)
}

func getAllLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a request for work item data")
	var idsRequest IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &idsRequest)
	fmt.Println(idsRequest)

}

func submitWOLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a request for work item data")
	var workItemRelReq WorkItemReleaseReq

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &workItemRelReq)
	fmt.Println(workItemReq)

	// requestedWI, exists := workItemMap[workItemReq.ItemID]
	// if !exists {
	// 	http.Error(w, ErrWorkItemDoesntExist.Error(), 400)
	// 	return
	// }

	block, getBlockErr := labelsDB.getBlock(workItemReq.ItemID)
	if getBlockErr != nil {
		http.Error(w, ErrWorkItemDoesntExist.Error(), 400)
		return
	}

	fmt.Println(block)
	json.NewEncoder(w).Encode(block)
}

// func backupHandler(w http.ResponseWriter, r *http.Request) {
// 	err := db.View(func(tx *bolt.Tx) error {
// 		w.Header().Set("Content-Type", "application/octet-stream")
// 		w.Header().Set("Content-Disposition", `attachment; filename="my.db"`)
// 		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
// 		_, err := tx.WriteTo(w)
// 		return err
// 	})
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }
