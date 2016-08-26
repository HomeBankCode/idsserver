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
	LabKey          string `json:"lab-key"`
	LabName         string `json:"lab-name"`
	Username        string `json:"username"`
	Training        bool   `json:"training"`
	Reliability     bool   `json:"reliability"`
	TrainingPackNum int    `json:"train-pack-num"`
}

/*
WorkItemDataReq is a request for the label data
for a particular work item from the database.
*/
type WorkItemDataReq struct {
	ItemID      string `json:"item-id"`
	LabKey      string `json:"lab-key"`
	Username    string `json:"username"`
	Training    bool   `json:"training"`
	Reliability bool   `json:"reliability"`
	Instance    int    `json:"instance"`
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

/*
DeleteBlockRequest is a request to delete the submitted
labels for a collection of block/instances. The BlockMap
if a map of Block ID's to an array of instance numbers
*/
type DeleteBlockRequest struct {
	LabKey   string           `json:"lab-key"`
	Coder    string           `json:"coder"`
	Type     string           `json:"delete-type"`
	BlockMap map[string][]int `json:"block-map"`
	BlockID  string           `json:"block-id"`
	Instance int              `json:"instance"`
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

	var workItem WorkItem
	var chooseWIErr error

	if blockRequest.Training {
		workItem, chooseWIErr = chooseTrainingWorkItem(blockRequest)
		fmt.Println("got past chooseTrainingWorkItem()")
		if chooseWIErr != nil {
			fmt.Println("returning error http code from chooseTrainingWorkItem()")
			http.Error(w, chooseWIErr.Error(), 404)
			return
		}
	} else if blockRequest.Reliability {
		workItem, chooseWIErr = chooseReliabilityWorkItem(blockRequest)
		if chooseWIErr != nil {
			http.Error(w, chooseWIErr.Error(), 404)
			return
		}
	} else {
		workItem, chooseWIErr = chooseRegularWorkItem(blockRequest)
		if chooseWIErr != nil {
			http.Error(w, chooseWIErr.Error(), 404)
			return
		}
	}

	fmt.Println(workItem)
	blockPath := workItem.BlockPath
	blockName := path.Base(blockPath)
	filename := path.Join(workItem.FileName, blockName)

	dispositionString := "attachment; filename=" + filename

	fmt.Println(workItem)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", dispositionString)

	http.ServeFile(w, r, workItem.BlockPath)

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
		fmt.Println("userExists failed")
		http.Error(w, ErrUserDoesntExist.Error(), 500)
		return
	}

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

	blockGroup, getBlockErr := labelsDB.getBlock(workItemReq.ItemID)
	if getBlockErr != nil {
		http.Error(w, ErrWorkItemDoesntExist.Error(), 400)
		return
	}
	blocks := blockGroup.getUsersBlocks(workItemReq.LabKey, workItemReq.Username)

	fmt.Println(blocks)
	json.NewEncoder(w).Encode(blocks)
}

func getLabLabelsHandler(w http.ResponseWriter, r *http.Request) {
	parseErr := r.ParseForm()
	if parseErr != nil {
		http.Error(w, parseErr.Error(), 400)
		return
	}

	fmt.Println("got a request for all labeled blocks")
	var idsRequest IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &idsRequest)
	fmt.Println(idsRequest)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsRegistered(idsRequest.LabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	blockIDs, getIdsErr := labsDB.getCompletedBlocks(idsRequest.LabKey)
	if getIdsErr != nil {
		http.Error(w, getIdsErr.Error(), 400)
		fmt.Println("getCompletedBlocks failed")
		return
	}

	blocks, getBlocksErr := labelsDB.getBlockGroup(blockIDs)
	if getBlocksErr != nil {
		http.Error(w, getBlocksErr.Error(), 400)
		return
	}

	labBlocks, labBlocksErr := blocks.filterLab(idsRequest.LabKey)
	if labBlocksErr != nil {
		http.Error(w, labBlocksErr.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(labBlocks)
}

func getAllLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a request for all labeled blocks")
	var idsRequest IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &idsRequest)
	fmt.Println(idsRequest)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsRegistered(idsRequest.LabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	//blockIDs := getAllCompleteBlockIDs()
	blocks, getBlocksErr := labelsDB.getAllBlockGroups()

	if getBlocksErr != nil {
		http.Error(w, getBlocksErr.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(blocks)
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
	fmt.Println(workItemRelReq)

	if !labsDB.userExists(workItemRelReq.LabKey, workItemRelReq.Username) {
		http.Error(w, ErrUserDoesntExist.Error(), 500)
		return
	}

	for _, block := range workItemRelReq.BlockIds {

		workItem := workItemMap[block]
		request := IDSRequest{
			LabKey:   workItemRelReq.LabKey,
			LabName:  workItemRelReq.LabName,
			Username: workItemRelReq.Username,
		}

		inactivateIncompleteWorkItem(workItem, request)
	}
}

func getTrainingLabelsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a request for all training blocks")
	var idsRequest IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &idsRequest)
	fmt.Println(idsRequest)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsRegistered(idsRequest.LabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	// Get Block ID's for all training blocks completed by lab users
	blockIDs, getBlockIDsErr := labsDB.getCompleteTrainBlocks(idsRequest.LabKey)
	if getBlockIDsErr != nil {
		http.Error(w, getBlockIDsErr.Error(), 400)
		return
	}

	// Get all the BlockGroups with those ID's
	blockGroups, getGroupsErr := labelsDB.getBlockGroup(blockIDs)
	if getGroupsErr != nil {
		http.Error(w, getGroupsErr.Error(), 400)
		return
	}

	// Filter those BlockGroups just for this lab's entries.
	// labBlocks is a BlockArray (array of Block)
	labBlocks, labBlocksErr := blockGroups.filterLab(idsRequest.LabKey)
	if labBlocksErr != nil {
		http.Error(w, labBlocksErr.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(labBlocks)
}

func getReliabilityHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("got a request for all training blocks")
	var idsRequest IDSRequest

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)

	fmt.Println()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	json.Unmarshal(jsonDataFromHTTP, &idsRequest)
	fmt.Println(idsRequest)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsRegistered(idsRequest.LabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	// Get Block ID's for all reliability blocks completed by lab users
	blockIDs, getBlockIDsErr := labsDB.getCompleteReliaBlocks(idsRequest.LabKey)
	if getBlockIDsErr != nil {
		http.Error(w, getBlockIDsErr.Error(), 400)
		return
	}

	// Get all the BlockGroups with those ID's
	blockGroups, getGroupsErr := labelsDB.getBlockGroup(blockIDs)
	if getGroupsErr != nil {
		http.Error(w, getGroupsErr.Error(), 400)
		return
	}

	// Filter those BlockGroups just for this lab's entries.
	// labBlocks is a BlockArray (array of Block)
	labBlocks, labBlocksErr := blockGroups.filterLab(idsRequest.LabKey)
	if labBlocksErr != nil {
		http.Error(w, labBlocksErr.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(labBlocks)
}

func deleteBlockHandler(w http.ResponseWriter, r *http.Request) {
	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		http.Error(w, parseFormErr.Error(), 400)
		return
	}

	fmt.Println("got a delete blocks request")
	var deleteBlockReq DeleteBlockRequest

	jsonDataFromHTTP, readBodyErr := ioutil.ReadAll(r.Body)
	if readBodyErr != nil {
		http.Error(w, readBodyErr.Error(), 400)
		return
	}

	fmt.Println()
	unmarshalErr := json.Unmarshal(jsonDataFromHTTP, &deleteBlockReq)
	if unmarshalErr != nil {
		http.Error(w, unmarshalErr.Error(), 400)
		return
	}
	fmt.Println(deleteBlockReq)

	// make sure the lab is one of the approved labs
	if !mainConfig.labIsRegistered(deleteBlockReq.LabKey) {
		http.Error(w, ErrLabNotRegistered.Error(), 400)
		fmt.Println("Unauthorized Lab Key")
		return
	}

	if deleteBlockReq.Type == "single" {
		/*
			We need to delete a single instance of a block from the LabelsDB
			and delete the entry from the coder's PastWorkItems list. If the
			user submitted more than one instance of that particular block,
			then we leave the ID in the PastWorkItems list (only deleted one
			instance of it).
		*/

		// make map
		singleInstanceMap := make(InstanceMap)
		singleInstanceMap[deleteBlockReq.BlockID] = []int{deleteBlockReq.Instance}

		// Delete from labelsDB. Function might also delete the BlockGroup entirely,
		// so it returns a flag
		groupWasDeleted, _ := labelsDB.deleteBlocks(singleInstanceMap)

		if !groupWasDeleted {

			blockGroup, getGroupErr := labelsDB.getBlock(deleteBlockReq.BlockID)
			if getGroupErr != nil {
				http.Error(w, getGroupErr.Error(), 400)
				return
			}

			// if there's a block instance with this coder still, it means they submitted
			// more than one instance of the same block, so we keep the PastWorkItem entry,
			// otherwise we delete it
			if !blockGroup.coderPresent(deleteBlockReq.LabKey, deleteBlockReq.Coder) {
				user, getUserErr := labsDB.getUser(deleteBlockReq.LabKey, deleteBlockReq.Coder)
				if getUserErr != nil {
					http.Error(w, getUserErr.Error(), 400)
					return
				}
				fmt.Println("\n\nBefore deleting PastItem")
				fmt.Println(user.PastWorkItems)
				user.deletePastItem(deleteBlockReq.BlockID)
				fmt.Println("\nAfter deleting PastItem")
				fmt.Println(user.PastWorkItems)
				fmt.Printf("\n\n")
				labsDB.setUser(user)
			}

		} else {
			user, getUserErr := labsDB.getUser(deleteBlockReq.LabKey, deleteBlockReq.Coder)
			if getUserErr != nil {
				http.Error(w, getUserErr.Error(), 400)
				return
			}
			fmt.Println("\n\nBefore deleting PastItem")
			fmt.Println(user.PastWorkItems)
			user.deletePastItem(deleteBlockReq.BlockID)
			fmt.Println("\nAfter deleting PastItem")
			fmt.Println(user.PastWorkItems)
			fmt.Printf("\n\n")
			labsDB.setUser(user)
		}

	} else if deleteBlockReq.Type == "user" {
		/*
			We need to build an InstanceMap of all the user's completed
			instances of blocks, and then pass that map to labelsDB.deleteBlocks()
			function. Then we need to delete all of those block entries from the
			user's PastWorkItems list.
		*/

		// get the user
		user, getUserErr := labsDB.getUser(deleteBlockReq.LabKey, deleteBlockReq.Coder)
		if getUserErr != nil {
			http.Error(w, getUserErr.Error(), 400)
			return
		}

		// build user's block instance map
		userInstances, userInstanceErr := user.getPastBlockInstanceMap()
		if userInstanceErr != nil {
			http.Error(w, userInstanceErr.Error(), 400)
			return
		}

		// delete those instances
		_, deleteUserInstErr := labelsDB.deleteBlocks(userInstances)
		if deleteUserInstErr != nil {
			http.Error(w, deleteUserInstErr.Error(), 400)
		}

		// clear out user's PastWorkItems
		user.PastWorkItems = nil
		user.CompleteTrainBlocks = nil
		user.CompleteRelBlocks = nil
		labsDB.setUser(user)

	} else if deleteBlockReq.Type == "lab" {
		/*
			We need to build an InstanceMap of all the lab's completed
			instances of blocks, and then pass that map to labelsDB.deleteBlocks()
			function. Then we need to delete all block entries from all of the lab's
			user's PastWorkItems lists.
		*/

		// get the lab
		lab, getLabErr := labsDB.getLab(deleteBlockReq.LabKey)
		if getLabErr != nil {
			http.Error(w, getLabErr.Error(), 400)
			return
		}

		// get all block instances submitted by the lab
		labInstanceMap, labInstanceErr := lab.getPastBlockInstanceMap()
		if labInstanceErr != nil {
			http.Error(w, labInstanceErr.Error(), 400)
			return
		}

		// delete all those instances from LabelsDB
		_, deleteLabInstErr := labelsDB.deleteBlocks(labInstanceMap)
		if deleteLabInstErr != nil {
			http.Error(w, deleteLabInstErr.Error(), 400)
			return
		}

		// clear out all users' PastWorkItems
		for _, user := range lab.Users {
			user.PastWorkItems = nil
			user.CompleteTrainBlocks = nil
			user.CompleteRelBlocks = nil
		}

		labsDB.setLab(lab.Key, lab)
	}
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {

}

func deleteLabHandler(w http.ResponseWriter, r *http.Request) {

}
