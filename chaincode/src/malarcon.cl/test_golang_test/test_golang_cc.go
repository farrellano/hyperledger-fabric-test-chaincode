package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"go.uber.org/zap"
)

// TestChaincode structure for defining the shim
type TestChaincode struct {
	logger *zap.SugaredLogger
}

// BiometricStruct data store representing a Document
type BiometricStruct struct {
	TokenBiometric        	string `json:"Token"`
	TypeBiometric      		string `json:"TypeBiometric"` //huella,visual,facial,voz
	KeyBiometric 			string `json:"KeyBiometric"`
	ActivationDate       	string `json:"ActivationDate"`
	ExpiredDate			 	string `json:"ExpiredDate"`
	ProviderBiometric		string `json:"ProviderBiometric"` //Local,Veridas,Facephi
}

func main() {
	zl, _ := zap.NewProduction()
	logger := zl.With(zap.String("module", "test-chaincode-go")).Sugar()
	logger.Info("Starting Test chaincode GO")

	chaincode := &TestChaincode{
		logger: logger,
	}
	err := shim.Start(chaincode)
	if err != nil {
		logger.Errorf("Error starting Test chaincode: %s", err)
	}
}

// Init function used to initialize the Closest Snapshot in-memory Index
func (c *TestChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	c.logger.Info("Initializing Test Chaincode")

	return shim.Success(nil)
}

// Invoke accepts all invoke commands from the blockchain and decides which function to call based on the inputs
func (c *TestChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	c.logger.Debug("V1.0")

	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "init":
		return c.Init(stub)
	case "search":
		return c.searchBiometricRecord(stub, args[0])
	case "insertBiometricRecord":
		return c.insertBiometricRecord(stub, args[0])
	case "deleteBiometricRecord":
		return c.insertBiometricRecord(stub, args[0])
	case "queryAll":
		return c.searchAllBiometricRecords(stub)
	}

	c.logger.Errorf("Invalid Test Invoke Function: %s", function)
	return shim.Error(fmt.Sprint("Invalid Invoke Function: ", function))
}

// ---------------------------- //
// Common Shim usage functions  //
// ---------------------------- //

// query executes a query on the Worldstate DB
func query(stub shim.ChaincodeStubInterface, jsonSnip string) ([]string, error) {
	//create iterator from selector key
	iter, err := stub.GetQueryResult(jsonSnip)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var outArray []string
	for iter.HasNext() {
		data, err := iter.Next()
		if err != nil {
			return nil, err
		}
		outArray = append(outArray, string(data.Value))
	}

	return outArray, nil
}

// convertQueryResultsToJSONByteArray converts an array of JSON Strings to a
// byte array that is also a JSON Array
func convertQueryResultsToJSONByteArray(rows []string) []byte {
	var buffer bytes.Buffer

	buffer.WriteString("[")
	if len(rows) > 0 {
		buffer.WriteString(strings.Join(rows, ","))
	}
	buffer.WriteString("]")

	return buffer.Bytes()
}

// ----------------------------- //
// Chaincode Business functions  //
// ----------------------------- //

// find executes a Selector query against the WorldState and returns the results in JSON format.
func (c *TestChaincode) searchBiometricRecord(stub shim.ChaincodeStubInterface, jsonSnip string) pb.Response {
	c.logger.Info("Beginning JSON query")

	results, err := query(stub, jsonSnip)
	if err != nil {
		c.logger.Error(err)
		return shim.Error("Error executing query")
	}

	outBytes := convertQueryResultsToJSONByteArray(results)
	return shim.Success(outBytes)
}

// insert record the BiometricStruct as either an insert or an update transaction
func (c *TestChaincode) insertBiometricRecord(stub shim.ChaincodeStubInterface, jsonSnip string) pb.Response {
	incoming := BiometricStruct{}
	err := json.Unmarshal([]byte(jsonSnip), &incoming)
	if err != nil {
		c.logger.Error("Error in store(): ", err)
		return shim.Error("Error parsing input")
	}

	if len(incoming.TokenBiometric) == 0 || len(incoming.KeyBiometric) == 0 {
		c.logger.Error("Invalid Token or Key ", err)
		return shim.Error("Invalid Token or Key")
	}

	key := incoming.KeyBiometric + "_" + incoming.TypeBiometric + "_" + incoming.ProviderBiometric
	incoming.KeyBiometric = key

	bytes, err := json.Marshal(incoming)
	if err != nil {
		c.logger.Info("Error marshalling BiometricStruct: ", incoming.KeyBiometric)
		return shim.Error("Error marshalling BiometricStruct" + incoming.KeyBiometric)
	}
	
	err = stub.PutState(key , bytes)
	if err != nil {
		c.logger.Info("Error writing BiometricStruct: ", incoming.KeyBiometric)
		return shim.Error("Error writing BiometricStruct: " + incoming.KeyBiometric)
	}
	c.logger.Info("Loaded StockSymbol: ", incoming.KeyBiometric)

	return shim.Success(nil)
}

//delete record the BiometricStruct from store
func (c *TestChaincode) deleteBiometricRecord(stub shim.ChaincodeStubInterface, ID string) pb.Response {
	assetBytes, err := stub.GetState(ID)
	if err != nil {
		c.logger.Error("failed to get asset %s: %v", ID,err)
		return shim.Error("failed to get asset")
	}
	if assetBytes == nil {
		c.logger.Error("asset %s does not exist", ID)
		return shim.Error("asset does not exist")
	}

	err = stub.DelState(ID)
	if err != nil {
		return shim.Error("Ocurred a problem asset not deleted")
	}
	return shim.Success(nil)
}

//search all BiometricStruct 
func (c *TestChaincode) searchAllBiometricRecords(stub shim.ChaincodeStubInterface) pb.Response {
	startKey := ""
	endKey := ""

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)

	if err != nil {
		c.logger.Error(err)
		return shim.Error("Error executing query")
	}
	defer resultsIterator.Close()

	var outArray []string
	for resultsIterator.HasNext() {
		data, err := resultsIterator.Next()
		if err != nil {
			c.logger.Error(err)
			return shim.Error("Error executing hash iterator")
		}
		outArray = append(outArray, string(data.Value))
	}

	outBytes := convertQueryResultsToJSONByteArray(outArray)
	return shim.Success(outBytes)

}


