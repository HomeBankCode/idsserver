package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
	AddBlockReq is a Block along with some metadata about its
	submission.
*/
type AddBlockReq struct {
	ItemID      string `json:"block_id"`
	AdminLabKey string `json:"admin_lab_key"`
	LabKey      string `json:"lab_key"`
	Username    string `json:"username"`
	Training    bool   `json:"training"`
	Reliability bool   `json:"reliability"`
	Instance    int    `json:"instance"`
	Block       Block  `json:"block"`
}

type AddUserReq struct {
	AdminLabKey string `json:"admin_lab_key"`
	LabKey      string `json:"lab_key"`
	LabName     string `json:"lab_name"`
	User        string `json:"user"`
}

/*
	All of these functions are for database migration purposes.

	If there are updates in the format of the database which
	require an old database to be resubmitted in a new form,
	this will allow you to manually submit the data from the
	old database along with metadata in order to insert it into
	the new database.
*/

func migrateAddLabeledBlockHandler(w http.ResponseWriter, r *http.Request) {
	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		http.Error(w, parseFormErr.Error(), 400)
		return
	}

	fmt.Println("got a request to upload a block")
	var addBlockReq AddBlockReq

	jsonDataFromHTTP, readBodyErr := ioutil.ReadAll(r.Body)
	if readBodyErr != nil {
		http.Error(w, readBodyErr.Error(), 400)
		return
	}

	fmt.Println()
	unmarshalErr := json.Unmarshal(jsonDataFromHTTP, &addBlockReq)
	if unmarshalErr != nil {
		http.Error(w, unmarshalErr.Error(), 400)
		return
	}

	fmt.Println(addBlockReq)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsAdmin(addBlockReq.AdminLabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	var block = addBlockReq.Block

	workItem := workItemMap[block.ID]
	request := IDSRequest{
		LabKey:   block.LabKey,
		LabName:  block.LabName,
		Username: block.Coder,
	}

	fmt.Printf("\n\n\n\nThe block:")
	fmt.Println(block)
	if block.Training {

		user, getUserErr := labsDB.getUser(block.LabKey, block.Coder)
		if getUserErr != nil {
			fmt.Println("getUser failed")
			http.Error(w, getUserErr.Error(), 400)
			return
		}
		user.addCompleteTrainBlock(block)
		fmt.Println("got past:   user.addCompleteTrainBlock(block)")

		setUserErr := labsDB.setUser(user)
		if setUserErr != nil {
			fmt.Println("setUser failed")
			http.Error(w, getUserErr.Error(), 400)
			return
		}

	} else if block.Reliability {

		user, getUserErr := labsDB.getUser(block.LabKey, block.Coder)
		if getUserErr != nil {
			fmt.Println("getUser failed")
			http.Error(w, getUserErr.Error(), 400)
			return
		}

		user.addCompleteRelBlock(block)
		fmt.Println("got past:    user.addCompleteRelBlock(block)")

		setUserErr := labsDB.setUser(user)
		if setUserErr != nil {
			fmt.Println("setUser failed")
			http.Error(w, getUserErr.Error(), 400)
			return
		}
	}

	addBlockErr := labelsDB.addBlock(block)
	if addBlockErr != nil {
		http.Error(w, addBlockErr.Error(), 400)
		return
	}
	inactivateWorkItem(workItem, request)

}

func migrateAddUserHandler(w http.ResponseWriter, r *http.Request) {
	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		http.Error(w, parseFormErr.Error(), 400)
		return
	}

	fmt.Println("got a request to add a user to the database")
	var addUserReq AddUserReq

	jsonDataFromHTTP, readBodyErr := ioutil.ReadAll(r.Body)
	if readBodyErr != nil {
		http.Error(w, readBodyErr.Error(), 400)
		return
	}

	fmt.Println()
	unmarshalErr := json.Unmarshal(jsonDataFromHTTP, &addUserReq)
	if unmarshalErr != nil {
		http.Error(w, unmarshalErr.Error(), 400)
		return
	}

	fmt.Println(addUserReq)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsAdmin(addUserReq.AdminLabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	labsDB.addUser(addUserReq.LabKey, addUserReq.LabName, addUserReq.User)
}

func migrateSetActiveWorkItemHandler(w http.ResponseWriter, r *http.Request) {
	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		http.Error(w, parseFormErr.Error(), 400)
		return
	}

	fmt.Println("got a request to add active work item to user")
	var addItemReq AddBlockReq

	jsonDataFromHTTP, readBodyErr := ioutil.ReadAll(r.Body)
	if readBodyErr != nil {
		http.Error(w, readBodyErr.Error(), 400)
		return
	}

	fmt.Println()
	unmarshalErr := json.Unmarshal(jsonDataFromHTTP, &addItemReq)
	if unmarshalErr != nil {
		http.Error(w, unmarshalErr.Error(), 400)
		return
	}

	fmt.Println(addItemReq)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsAdmin(addItemReq.AdminLabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	user, getUserErr := labsDB.getUser(addItemReq.LabKey, addItemReq.Username)
	if getUserErr != nil {
		http.Error(w, getUserErr.Error(), 400)
		return
	}

	user.addWorkItem(addItemReq.ItemID)

	setUserErr := labsDB.setUser(user)
	if setUserErr != nil {
		http.Error(w, setUserErr.Error(), 400)
		return
	}

}
