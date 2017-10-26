/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Baggage Chaincode implementation
type Baggage struct {
}

//ALL PARTY CONSTANTS
const AIRLINE = "AIRLINE"
const ORIGINAIRPORT = "ORIGINAIRPORT"
const TRANSITAIRPORT = "TRANSITAIRPORT"
const DESTINATIONAIRPORT = "DESTINATIONAIRPORT"

//ALL STATUS CONSTANTS
const CHECKEDIN = "CHECKEDIN"
const BAGTAGGING = "BAGTAGGING"
const ONBOARDINGORIGIN = "ONBOARDINGORIGIN"
const OFFBOARDINGTRANSIT = "OFFBOARDINGTRANSIT"
const ONBOARDINGTRANSIT = "ONBOARDINGTRANSIT"
const OFFBOARDINGDESTINATION = "OFFBOARDINGDESTINATION"
const CLAIM = "CLAIM"

func (t *Baggage) Init(stub shim.ChaincodeStubInterface) pb.Response {

	_, args := stub.GetFunctionAndParameters()

	if len(args) < 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	return shim.Success(nil)
}

//create a new check in request
func (t *Baggage) createCheckInRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	recBytes := args[0]
	pnrNumber := args[1]
	updatedAt := args[2]
	identity := args[3]

	//Checking whether the user has the authority to call the method
	if identity != AIRLINE {
		return shim.Error("You are not authorized to createCheckInRequest")
	}

	//==== Check if Pnr already exists ====
	fetchedPnrDetails, err := stub.GetState("PNR:" + pnrNumber)
	if err != nil {
		return shim.Error("Failed to get PNR details: " + err.Error())
	} else if fetchedPnrDetails != nil {
		fmt.Println("This PNR already exists: " + pnrNumber)
		return shim.Error("This PNR already exists: " + pnrNumber)
	}

	var lineItemRecordMapArray []map[string]interface{}
	lineItemRecordMapArray = make([]map[string]interface{}, 0)

	err = json.Unmarshal([]byte(recBytes), &lineItemRecordMapArray)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}

	for _, item := range lineItemRecordMapArray {

		item["status"] = CHECKEDIN
		item["pnr"] = pnrNumber
		item["updatedAt"] = updatedAt

		outputMapBytes, _ := json.Marshal(item)
		//Store the records
		stub.PutState("ITEM:"+getSafeString(item["itemId"]), outputMapBytes)
		if err != nil {
			return shim.Error(err.Error())
		}

	}

	return shim.Success([]byte("SUCCESS"))
}

//update the status of the item
func (t *Baggage) updateItemStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	recBytes := args[0]
	updatedAt := args[1]
	//identity := args[2]

	var lineItemRecordMap map[string]interface{}

	err := json.Unmarshal([]byte(recBytes), &lineItemRecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}

	//==== Check if Pnr already exists ====
	fetchedPnrDetails, err := stub.GetState("ITEM:" + getSafeString(lineItemRecordMap["itemId"]))
	if err != nil {
		return shim.Error("Failed to get ITEM details: " + err.Error())
	} else if fetchedPnrDetails == nil {
		fmt.Println("This ITEM does not exists:" + getSafeString(lineItemRecordMap["itemId"]))
		return shim.Error("This ITEM does not exists:" + getSafeString(lineItemRecordMap["itemId"]))
	}

	var itemMap map[string]interface{}
	err = json.Unmarshal(fetchedPnrDetails, &itemMap)
	if err != nil {
		return shim.Error("Failed to unmarshal item")
	}
	itemMap["status"] = getSafeString(lineItemRecordMap["status"])
	itemMap["updatedAt"] = updatedAt

	outputMapBytes, _ := json.Marshal(itemMap)
	//Store the records
	stub.PutState("ITEM:"+getSafeString(itemMap["itemId"]), outputMapBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("SUCCESS"))
}

//get the item
func (t *Baggage) getLineItem(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var err error
	var ItemRecordMap map[string]interface{}

	recBytes := args[0]
	//who := args[1]

	err = json.Unmarshal([]byte(recBytes), &ItemRecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal ItemRecordMap ")
	}

	itemId := getSafeString(ItemRecordMap["itemId"])

	itemDetails, err := stub.GetState("ITEM:" + itemId)
	if err != nil {
		return shim.Error("Failed to get ITEM: " + itemId)
	} else if itemDetails == nil {
		fmt.Println("This ITEM does not exist: " + itemId)
		return shim.Error("This ITEM does not exist: " + itemId)
	}

	fmt.Println("ITEM:", string(itemDetails))
	return shim.Success(itemDetails)
}

//get all the items by status
func (t *Baggage) getAllItemByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var err error
	var ItemRecordMap map[string]interface{}

	recBytes := args[0]
	//who := args[1]

	err = json.Unmarshal([]byte(recBytes), &ItemRecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal ItemRecordMap ")
	}

	status := getSafeString(ItemRecordMap["status"])

	var queryString string

	queryString = "{\"selector\":{\"status\":\"" + status + "\"}}"

	query := []string{queryString}

	return t.queryDetails(stub, query)

}

//getAllItemByTemperatureStatus
func (t *Baggage) getAllItemByTemperatureStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var err error
	var ItemRecordMap map[string]interface{}

	recBytes := args[0]
	//who := args[1]

	err = json.Unmarshal([]byte(recBytes), &ItemRecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal ItemRecordMap ")
	}

	temperatureStatus := getSafeString(ItemRecordMap["temperatureStatus"])

	var queryString string

	queryString = "{\"selector\":{\"temperatureStatus\":\"" + temperatureStatus + "\"}}"

	query := []string{queryString}

	return t.queryDetails(stub, query)

}

//getAllItemByTemperature
func (t *Baggage) getAllItemByTemperature(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var err error
	var ItemRecordMap map[string]interface{}

	recBytes := args[0]
	//who := args[1]

	err = json.Unmarshal([]byte(recBytes), &ItemRecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal ItemRecordMap ")
	}

	temperature := getSafeString(ItemRecordMap["temperature"])

	var queryString string

	queryString = "{\"selector\":{\"temperature\":\"" + temperature + "\"}}"

	query := []string{queryString}

	return t.queryDetails(stub, query)

}

//get all the items by status
func (t *Baggage) getAllItemByPnr(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var err error
	var ItemRecordMap map[string]interface{}

	recBytes := args[0]
	//who := args[1]

	err = json.Unmarshal([]byte(recBytes), &ItemRecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal ItemRecordMap ")
	}

	pnr := getSafeString(ItemRecordMap["pnr"])

	var queryString string

	queryString = "{\"selector\":{\"pnr\":\"" + pnr + "\"}}"

	query := []string{queryString}

	return t.queryDetails(stub, query)

}

//updateItemTemperatureStatus
func (t *Baggage) updateItemTemperatureStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	recBytes := args[0]
	updatedAt := args[1]
	//identity := args[2]

	var lineItemRecordMap map[string]interface{}

	err := json.Unmarshal([]byte(recBytes), &lineItemRecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}

	//==== Check if Pnr already exists ====
	fetchedPnrDetails, err := stub.GetState("ITEM:" + getSafeString(lineItemRecordMap["itemId"]))
	if err != nil {
		return shim.Error("Failed to get ITEM details: " + err.Error())
	} else if fetchedPnrDetails == nil {
		fmt.Println("This ITEM does not exists:" + getSafeString(lineItemRecordMap["itemId"]))
		return shim.Error("This ITEM does not exists:" + getSafeString(lineItemRecordMap["itemId"]))
	}

	var itemMap map[string]interface{}
	err = json.Unmarshal(fetchedPnrDetails, &itemMap)
	if err != nil {
		return shim.Error("Failed to unmarshal item")
	}
	itemMap["temperatureStatus"] = getSafeString(lineItemRecordMap["temperatureStatus"])
	itemMap["updatedAt"] = updatedAt
	itemMap["temp"] = getSafeString(lineItemRecordMap["temperature"])
	itemMap["TempItemStatus"] = itemMap["status"] 

	outputMapBytes, _ := json.Marshal(itemMap)
	//Store the records
	stub.PutState("ITEM:"+getSafeString(itemMap["itemId"]), outputMapBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("SUCCESS"))
}

// ===== Example: Ad hoc rich query ========================================================
// queryMarbles uses a query string to perform a query for marbles.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *Baggage) queryDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

//check whether string has value or not
func getSafeString(input interface{}) string {
	var safeValue string
	var isOk bool

	if input == nil {
		safeValue = ""
	} else {
		safeValue, isOk = input.(string)
		if isOk == false {
			safeValue = ""
		}
	}
	return safeValue
}

func (t *Baggage) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//myLogger.Debug("Invoke Chaincode...")
	function, args := stub.GetFunctionAndParameters()
	if function == "createCheckInRequest" {
		//request a new check in
		return t.createCheckInRequest(stub, args)
	} else if function == "updateItemStatus" {
		//update the status
		return t.updateItemStatus(stub, args)
	} else if function == "getLineItem" {
		//get the item
		return t.getLineItem(stub, args)
	} else if function == "getAllItemByStatus" {
		//get all Item By Status
		return t.getAllItemByStatus(stub, args)
	} else if function == "getAllItemByPnr" {
		//get all Item By pnr
		return t.getAllItemByPnr(stub, args)
	} else if function == "updateItemTemperatureStatus" {
		//updateItemTemperatureStatus
		return t.updateItemTemperatureStatus(stub, args)
	} else if function == "getAllItemByTemperatureStatus" {
		//getAllItemByTemperatureStatus
		return t.getAllItemByTemperatureStatus(stub, args)
	} else if function == "getAllItemByTemperature" {
		//getAllItemByTemperature
		return t.getAllItemByTemperature(stub, args)
	}
	return shim.Error("Invalid invoke function name.")
}

func main() {
	err := shim.Start(new(Baggage))
	if err != nil {
		fmt.Printf("Error starting Baggage: %s", err)
	}
}