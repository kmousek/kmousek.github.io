/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Chaincode is the definition of the chaincode structure.
type Chaincode struct {
}

type Tap struct {
	RecdSeq       string
	TrmPlmnId     string
	RcvPlmnId     string
	FileDivCd     string
	FileSeqNo     string
	FileCretDtVal string
	CallTypeId    string
	RecNo         string
	ImsiId        string
	MsisdnNo      string
	CalldNo       string
	DialNo        string
	LocalTime     string
	TotCallDurat  int
	CallgNo       string
}

type TapResult struct {
	RecdSeq       string
	TrmPlmnId     string
	RcvPlmnId     string
	FileDivCd     string
	FileSeqNo     string
	FileCretDtVal string
	CallTypeId    string
	RecNo         string
	ImsiId        string
	MsisdnNo      string
	CalldNo       string
	DialNo        string
	LocalTime     string
	TotCallDurat  int
	CallgNo       string
	Charge        int
	CapCharge     int
	CommitCharge  int
}

type AgreementRate struct {
	Agreement   string
	TrmPlmnId   string
	RcvPlmnId   string
	ContStDate  string
	ContexpDate string
	CallTypeCd  string
	StelTarif   int
	StelVlm     int
	StelUnit    string
	IsPerImsi   string
	IsCommit    string
}

// Init is called when the chaincode is instantiated by the blockchain network.
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Init()", fcn, params)

	return shim.Success(nil)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()

	fmt.Println("Invoke()", fcn, params)

	if fcn == "tapProcessing" {
		return cc.tapProcessing(stub, params)
	} else if fcn == "getData" {
		return cc.getData(stub, params)
	}

	return shim.Error("Invalid Chain code function name.")
}

func (cc *Chaincode) tapProcessing(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	agtRt := AgreementRate{"11111111111", "KT", "CHNCM", "20190101000000", "20191231235959", "VOICE-LOCAL", 20, 1, "sec", "Y", "N"}

	totCallDurat, err := strconv.Atoi(params[13])
	if err != nil {
		return shim.Error(err.Error())
	}

	tap := Tap{RecdSeq: params[0],
		TrmPlmnId:     params[1],
		RcvPlmnId:     params[2],
		FileDivCd:     params[3],
		FileSeqNo:     params[4],
		FileCretDtVal: params[5],
		CallTypeId:    params[6],
		RecNo:         params[7],
		ImsiId:        params[8],
		MsisdnNo:      params[9],
		CalldNo:       params[10],
		DialNo:        params[11],
		LocalTime:     params[12],
		TotCallDurat:  totCallDurat,
		CallgNo:       params[14],
	}

	fmt.Println("[tap] :", tap)
	fmt.Println("[totCallDurat] :", totCallDurat)

	charge := calcChargeAmount(agtRt, params[6], totCallDurat)

	capCharge := calcCapAmount(agtRt, params[6], totCallDurat)

	commitCharge := calcCommitAmount(agtRt, params[6], totCallDurat)

	tapResult := TapResult{RecdSeq: params[0],
		TrmPlmnId:     params[1],
		RcvPlmnId:     params[2],
		FileDivCd:     params[3],
		FileSeqNo:     params[4],
		FileCretDtVal: params[5],
		CallTypeId:    params[6],
		RecNo:         params[7],
		ImsiId:        params[8],
		MsisdnNo:      params[9],
		CalldNo:       params[10],
		DialNo:        params[11],
		LocalTime:     params[12],
		TotCallDurat:  totCallDurat,
		CallgNo:       params[14],
		Charge:        charge,
		CapCharge:     capCharge,
		CommitCharge:  commitCharge,
	}

	tapResultAsBytes, err := json.Marshal(tapResult)

	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(params[0], tapResultAsBytes)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (cc *Chaincode) getData(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	valueJSONAsBytes, err := stub.GetState(params[0])

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(valueJSONAsBytes)
}

func calcChargeAmount(agt AgreementRate, callTypeCd string, totCallDurat int) int {

	if callTypeCd == "VOICE-LOCAL" {
		return 10
	} else if callTypeCd == "VOICE-HOME" {
		return 20
	} else if callTypeCd == "VOICE-INTL" {
		return 30
	} else if callTypeCd == "VOICE-MTC" {
		return 40
	} else if callTypeCd == "SMS-MO" {
		return 50
	} else if callTypeCd == "SMS-MT" {
		return 60
	} else if callTypeCd == "DATA" {
		return 70
	}
	return 0
}

func calcCapAmount(agt AgreementRate, callTypeCd string, totCallDurat int) int {
	return 90
}

func calcCommitAmount(agt AgreementRate, callTypeCd string, totCallDurat int) int {
	return 100
}
