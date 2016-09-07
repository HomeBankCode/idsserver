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
	LabKey      string `json:"lab_key"`
	Username    string `json:"username"`
	Training    bool   `json:"training"`
	Reliability bool   `json:"reliability"`
	Instance    int    `json:"instance"`
	Block       Block  `json:"block"`
}

/*
	addLabeledBlock is for database migration purposes. If there are
	updates in the format of the database which require an old database
	to be resubmitted in a new form, this will allow you to manually
	submit a Block response along with metadata of that submission
*/
func addLabeledBlock(w http.ResponseWriter, r *http.Request) {
	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		http.Error(w, parseFormErr.Error(), 400)
		return
	}

	fmt.Println("got a request to download a specific block")
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
	if !mainConfig.labIsAdmin(addBlockReq.LabKey) {
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

	if block.Training {

		user, getUserErr := labsDB.getUser(block.LabKey, block.Username)
		if getUserErr != nil {
			fmt.Println("getUser failed")
			http.Error(w, getUserErr.Error(), 400)
			return
		}
		user.addCompleteTrainBlock(block)

		setUserErr := labsDB.setUser(user)
		if setUserErr != nil {
			fmt.Println("setUser failed")
			http.Error(w, getUserErr.Error(), 400)
			return
		}

	} else if block.Reliability {

		user, getUserErr := labsDB.getUser(block.LabKey, block.Username)
		if getUserErr != nil {
			fmt.Println("getUser failed")
			http.Error(w, getUserErr.Error(), 400)
			return
		}

		user.addCompleteRelBlock(block)

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

	labelsDB.addBlock(block)

}
