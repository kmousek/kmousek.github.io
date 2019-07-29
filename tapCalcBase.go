package service

import (
	"github.com/main/go/jsonStruct"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"math"
	"strconv"
)



//전역변수
const gVoiceUnit float64 = 60
const gDataUnit float64 = 1024
const gPowTenOfTen float64 = 10000000000
const gNotExgtValue string = "null"
var gQuryType_ImsiUsage = "imsiUsage"
var gCT_MOC_LOCAL = "MOC-local"
var gCT_MOC_HOME = "MOC-home"
var gCT_MOC_INTL = "MOC-int"
var gCT_MTC = "MTC"
var gCT_SMS_MO = "SMS-mo"
var gCT_SMS_MT = "SMS-mt"
var gCT_DATA = "GRPC"

//tap 요율 계산 처리 main
func CalculChargeAmount(stub shim.ChaincodeStubInterface, tapRt *jsonStruct.TapResult) error {
	Log_add("calcChargeAmount")
	Log_add(tapRt.CallType)

	subAgt := jsonStruct.Agreement{}  //계약 서브 구조체 (past와 current중 하나 Agreement매핑)
	nowDate := tapRt.LocalTimeStamp[:8]
	Log_add(nowDate)

	//1. active인 요율 계산용 agreement 조회
	actAgt, err := Agreement_getActive(stub, nowDate, tapRt.VPMN, tapRt.HPMN)
	if err != nil{
		//return shim.Error("Agreement_getActive error")
	}


	// 처리할 tap이 agreement의 past인지 current인지 확인, imsi cap/commitment적용대상인지 확인
	subAgt, bImsiCapFlag, bCommitmentFlag := searchAgtIdx(&actAgt, nowDate)

	// 정율 계산
	for i:=0; i< len(subAgt.Basic); i++ {
		if tapRt.CallType == subAgt.Basic[i].TypeCD && (tapRt.CallType == gCT_MOC_LOCAL || tapRt.CallType == gCT_MOC_HOME || tapRt.CallType == gCT_MOC_INTL || tapRt.CallType == gCT_MTC) {
			tapRt.Charge = calcVoiceItem(subAgt.Basic[i].Unit, subAgt.Basic[i].Rate, subAgt.Basic[i].Volume, tapRt.TotalCallEventDuration)
			tapRt.SetCharge = tapRt.Charge
			break
		}else if tapRt.CallType == subAgt.Basic[i].TypeCD && (tapRt.CallType == gCT_SMS_MO || tapRt.CallType == gCT_SMS_MT ) {
			tapRt.Charge = calcSmsItem(subAgt.Basic[i].Unit, subAgt.Basic[i].Rate)
			tapRt.SetCharge = tapRt.Charge
			break
		}else if tapRt.CallType == subAgt.Basic[i].TypeCD && tapRt.CallType == gCT_DATA {
			tapRt.Charge = calcDataItem(subAgt.Basic[i].Unit, subAgt.Basic[i].Rate, subAgt.Basic[i].Volume, tapRt.TotalCallEventDuration)
			tapRt.SetCharge = tapRt.Charge
			break
		}
	}

	fmt.Println(bImsiCapFlag)
	fmt.Println(bCommitmentFlag)


	fmt.Println(f64ImsiCapCharge)


	return nil
}


//tap이 past인지 current인지 확인
func searchAgtIdx(actAgt *jsonStruct.AgreementForCal, nowDate string) (jsonStruct.Agreement, bool, bool) {
	returnAgt := jsonStruct.Agreement{}
	var bImsiCapFlag boolean
	var bCommitmentFlag boolean

	if actAgt.AgreementInfo.Past.Period[0] <= nowDate && actAgt.AgreementInfo.Past.Period[1] >= nowDate{
		returnAgt = agt.AgreementInfo.Past
	}else if actAgt.AgreementInfo.Current.Period[0] <= nowDate && actAgt.AgreementInfo.Current.Period[1] >= nowDate {
		returnAgt = agt.AgreementInfo.Current
	}

	if returnAgt.ImsiCap.THRSMIN == gNotExgtValue {
		bImsiCapFlag = false
	}else{
		bImsiCapFlag = true
	}

	if returnAgt.Commitment.THRSMIN == gNotExgtValue {
		bCommitmentFlag = false
	}else{
		bCommitmentFlag = true
	}

	return returnAgt, bImsiCapFlag, bCommitmentFlag
}

//음성 계산 함수
func calcVoiceItem (unit string, rate string, volume string, totCallDurat float64) float64 {
	Log_add("calcVoiceItem")

	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0 //에러처리
	}

	f64Volume, err := strconv.ParseFloat(volume, 64)
	if err != nil{
		return 0 //에러처리
	}

	if unit =="min"{
		return math.Ceil(totCallDurat/f64Volume * gVoiceUnit) * f64Rate
	}else if unit =="sec"{
		return math.Ceil(totCallDurat/ f64Volume) * f64Rate
	}else{
		return 0
	}
}

//SMS 계산 함수
func calcSmsItem (unit string, rate string) float64 {
	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0 //에러처리
	}

	Log_add("calcSmsItem")
	return f64Rate
}

//DATA 계산 함수
func calcDataItem (unit string, rate string, volume string, totCallDurat float64) float64 {
	Log_add("calcDataItem")

	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0 //에러처리
	}

	f64Volume, err := strconv.ParseFloat(volume, 64)
	if err != nil{
		return 0 //에러처리
	}

	if unit =="mbytes"{
		return math.Ceil(totCallDurat/ (f64Volume * gDataUnit)) * f64Rate
	}else if unit =="kbytes"{
		return math.Ceil(totCallDurat/ f64Volume) * f64Rate
	}else if unit =="bytes"{
		return math.Ceil(totCallDurat/ (f64Volume / gDataUnit)) * f64Rate
	}else{
		return 0
	}
}