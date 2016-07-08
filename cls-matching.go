package main

import (
	"errors"
	"fmt"
	"strconv"
        "encoding/base64"
	"encoding/json"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"bytes"
	"crypto/sha512"
	"io"
	
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
	PId         string
	CPId        string
	Bic         string
	Cbic        string
	RoleType    string
	CurrDetails Currency
}

type Currency struct {
	ISOBVol  float64
	ISOSVol  float64
	ISOBCurr string
	ISOSCurr string
}

type matchingEngine interface {
	match(instr1 *SimpleChaincode, instr2 *SimpleChaincode) bool
}

func match(instr1 *SimpleChaincode, instr2 *SimpleChaincode) bool {

	fmt.Println(" Matching Started ... ")

	isMatched := strings.EqualFold(instr1.PId, instr2.CPId) && strings.EqualFold(instr1.CPId, instr2.PId) && strings.EqualFold(instr1.Bic, instr2.Cbic) && strings.EqualFold(instr1.CurrDetails.ISOBCurr, instr2.CurrDetails.ISOSCurr) && instr1.CurrDetails.ISOBVol == instr2.CurrDetails.ISOSVol && instr1.CurrDetails.ISOSVol == instr2.CurrDetails.ISOBVol

	fmt.Println("Match Result %s", isMatched)
	return isMatched

}
func getHashFromInstr(instrStr string) string {
	h512 := sha512.New()
	io.WriteString(h512, instrStr)

	//	fmt.Printf("SHA512 checksum : %s\n", base64.URLEncoding.EncodeToString(h512.Sum(nil)))

	return base64.URLEncoding.EncodeToString(h512.Sum(nil))
}

func convertInstrToHashEligString(instr SimpleChaincode) string {
	var buffer bytes.Buffer

	RoleType := instr.RoleType
	if strings.Compare(RoleType, "I") == 0 {
		fmt.Println("--------------In Initiator ------")
		PId := instr.PId
		buffer.WriteString(PId)

		CPId := instr.CPId
		buffer.WriteString(CPId)

		Bic := instr.Bic
		buffer.WriteString(Bic)

		Cbic := instr.Cbic
		buffer.WriteString(Cbic)

		ISOBVol := instr.CurrDetails.ISOBVol
		buffer.WriteString(strconv.FormatFloat(ISOBVol, 'E', -1, 64))

		ISOSVol := instr.CurrDetails.ISOSVol
		buffer.WriteString(strconv.FormatFloat(ISOSVol, 'E', -1, 64))

		ISOBCurr := instr.CurrDetails.ISOBCurr
		buffer.WriteString(ISOBCurr)

		ISOSCurr := instr.CurrDetails.ISOSCurr
		buffer.WriteString(ISOSCurr)

	} else if strings.Compare(RoleType, "V") == 0 {
		fmt.Println("--------------In Validator ------")
		CPId := instr.CPId
		buffer.WriteString(CPId)

		PId := instr.PId
		buffer.WriteString(PId)

		Cbic := instr.Cbic
		buffer.WriteString(Cbic)

		Bic := instr.Bic
		buffer.WriteString(Bic)

		ISOSVol := instr.CurrDetails.ISOSVol
		buffer.WriteString(strconv.FormatFloat(ISOSVol, 'E', -1, 64))

		ISOBVol := instr.CurrDetails.ISOBVol
		buffer.WriteString(strconv.FormatFloat(ISOBVol, 'E', -1, 64))

		ISOSCurr := instr.CurrDetails.ISOSCurr
		buffer.WriteString(ISOSCurr)

		ISOBCurr := instr.CurrDetails.ISOBCurr
		buffer.WriteString(ISOBCurr)

	}
	str := buffer.String()
	fmt.Println("Hash Eligible Str " + str)
	return str

}




func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	var instStr string
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	instStr = args[0]
	encoded := base64.StdEncoding.EncodeToString([]byte(instStr))
	fmt.Println(encoded)
         
	//instrBytes, _ := json.Marshal(encoded)

	//err = stub.PutState("instr", instrBytes)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function == "match" {
		// Deletes an entity from its state
		//		return t.delete(stub, args)
	}

	var stateInstr,chainCodeInstr SimpleChaincode
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}


        fmt.Println("Invoke json string argument  :: %s  ",args[0])
        chainCodeJsonStr := []byte(args[0])
       

	err = json.Unmarshal(chainCodeJsonStr, &chainCodeInstr)

	instrStr:=convertInstrToHashEligString(chainCodeInstr)
	
	hashInstr:=getHashFromInstr(instrStr)
	fmt.Println("HashString...%s",hashInstr)


	
	// Get the state from the ledger
	//  will be nice to have a GetAllState call to ledger
	InstrBytes, err := stub.GetState(hashInstr)
	
	//fmt.Println("Input State from Hyperledger ", InstrBytes)


	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	if err != nil {
		return nil, errors.New("Failed to get state")
	}
//	if InstrBytes == nil {
//		return nil, errors.New("Entity not found")
//	}


	if InstrBytes==nil {

	        fmt.Println("There is no instruction with hash %s",hashInstr)
		encoded := base64.StdEncoding.EncodeToString([]byte(chainCodeJsonStr))
		instrBytes, _ := json.Marshal(encoded)
		err = stub.PutState(hashInstr, instrBytes)
		if err != nil {
			return nil, err
		}
	}else{

		fmt.Println("Matching Instruction found for hash %s ",hashInstr)
        	stateStr:= string(InstrBytes)
        	fmt.Println("State Str %s  ",stateStr)
		decoded, err := base64.StdEncoding.DecodeString(stateStr[1 : len(stateStr)-1])
		if err != nil {
			fmt.Println("decode error:", err)
			//	return
		}	

		fmt.Println("Decoded String ====  ",string(decoded))
		bs := []byte(string(decoded))	
        	err=json.Unmarshal(bs, &stateInstr)

		fmt.Println("--------------")
		fmt.Println(stateInstr.PId)
		fmt.Println(stateInstr.CPId)
		fmt.Println(stateInstr.Bic)
		fmt.Println(stateInstr.Cbic)
		fmt.Println(stateInstr.RoleType)
		fmt.Println(stateInstr.CurrDetails.ISOBCurr)
		fmt.Println(stateInstr.CurrDetails.ISOBVol)
		fmt.Println(stateInstr.CurrDetails.ISOSCurr)
		fmt.Println(stateInstr.CurrDetails.ISOSVol)
		fmt.Println(err)	
		//err = json.Unmarshal(InstrBytes, &instr1)
       
	     }		
	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var instrJson string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	instrJson = args[0]
	fmt.Printf("Query Input :%s\n", instrJson)

	// Get the state from the ledger
	instrBytes, err := stub.GetState("instr")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + instrJson + "\"}"
		return nil, errors.New(jsonResp)
	}

	if instrBytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + instrJson + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + instrJson + "\",\"Amount\":\"" + string(instrBytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	fmt.Printf("Finally getting Results")
	return instrBytes, nil
}


func main() {
        err := shim.Start(new(SimpleChaincode))
        if err != nil {
                fmt.Printf("Error starting Simple chaincode: %s", err)
        }



}

