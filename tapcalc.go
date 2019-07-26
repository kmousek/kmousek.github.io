package service

import (
	"github.com/main/go/service"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
  "math"
  "strconv"
  "strings"
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
func CalcChargeAmount(stub shim.ChaincodeStubInterface, actAgt jsonStruct.AgreementForCal, tapRt *TapResult) error {
	Log_add("calcChargeAmount")
	Log_add(tapRt.CallType)
	
	subAgt := jsonStruct.Agreement{}  //계약 서브 구조체 (past와 current중 하나 Agreement매핑)
	nowDate := tapRt.LocalTimeStamp[:8]

	// 처리할 tap이 agreement의 past인지 current인지 확인, imsi cap/commitment적용대상인지 확인
	subAgt, bImsiCapFlag, bCommitmentFlag := searchAgtIdx(&actAgt, nowDate)
	
	// 정율 계산 
	for i:=0; i< len(subAgt.Basic); i++ {
		if tapRt.CallType == subAgt.Basic[i].CallType && (tapRt.CallType == gCT_MOC_LOCAL || tapRt.CallType == gCT_MOC_HOME || tapRt.CallType == gCT_MOC_INTL || tapRt.CallType == gCT_MTC) {
			tapRt.Charge = calcVoiceItem(subAgt.Basic[i].Unit, subAgt.Basic[i].Rate, subAgt.Basic[i].Volume, tapRt.TotalCallEventDuration)
			tapRt.SetCharge = tapRt.Charge
			break
		}else if tapRt.CallType == subAgt.Basic[i].CallType && (tapRt.CallType == gCT_SMS_MO || tapRt.CallType == gCT_SMS_MT ) {
			tapRt.Charge = calcSmsItem(subAgt.Basic[i].Unit, subAgt.Basic[i].Rate)
			tapRt.SetCharge = tapRt.Charge
			break
		}else if tapRt.CallType == subAgt.Basic[i].CallType && tapRt.CallType == gCT_DATA {
			tapRt.Charge = calcDataItem(subAgt.Basic[i].Unit, subAgt.Basic[i].Rate, subAgt.Basic[i].Volume, tapRt.TotalCallEventDuration)
			tapRt.SetCharge = tapRt.Charge
			break
		}
	}

    fmt.Println(bImsiCapFlag)
    fmt.Println(bCommitmentFlag)

    //Imsi Cap 계산
	//calcImsiCap(&subAgt, &tapRt)
	if bImsiCapFlag == true {
		f64ImsiCapCharge := calcImsiCap(stub, &subAgt, tapRt)
        tapRt.SetCharge = f64ImsiCapCharge
		return nil
	}

    fmt.Println(f64ImsiCapCharge)
/*	
	//calcCommitment(&tapRt)
	if bCommitmentFlag == true {

	}
*/

	return nil
 }


 func calcImsiCap(stub shim.ChaincodeStubInterface, subAgt *jsonStruct.Agreement, tapRt *TapResult) float64 {
	//per Imsi 사용량 누적 조회 
	
	//조건 체크가 금액인지 사용량인지 구분
	/*
		subAgt.ImsiCap.THRSMIN
		subAgt.ImsiCap.THRSMAX
		subAgt.ImsiCap.THRUNIT
		subAgt.ImsiCap.FIXAMT
		subAgt.ImsiCap.Basic
	*/

	//imsiCap누적 조회해서 없으면 charge로 imsiCap 조건 체크하고 있으면 settCharge + tap의 setCharge로 imsiCap 조건 체크
	queryKey := strings.Fields(gQuryType_ImsiUsage+tapRt.LocalTimeStamp[:8]+tapRt.Imsi)
	imsiCapBytes, err := Block_Query(stub, queryKey)
	var f64ImsiCapCharge float64
	if err != nil{
		//error처리
	}else if imsiCapBytes == nil{ //조회된 누적량이 없을 경우
		f64ImsiCapCharge = tapRt.Charge
	}else{
		stImsiUsage := new(jsonStruct.ImsiUsage)
		err = json.Unmarshal(imsiCapBytes[0], stImsiUsage)
		if err != nil{  
			return getErrorReturnValue(err, "json Unmarshal error") 
		}
		
		if tapRt.CallType == gCT_MOC_LOCAL{
			f64ImsiCapCharge=stImsiUsage.TapCal.MOCLocal.CalculDetail.Charge + tapRt.Charge
		}else if tapRt.CallType == gCT_MOC_HOME{
			f64ImsiCapCharge=stImsiUsage.TapCal.MOCHome.CalculDetail.Charge + tapRt.Charge
		}else if tapRt.CallType == gCT_MOC_INTL{
			f64ImsiCapCharge=stImsiUsage.TapCal.MOCInt.CalculDetail.Charge + tapRt.Charge
		}else if tapRt.CallType == gCT_MTC{
			f64ImsiCapCharge=stImsiUsage.TapCal.MOCInt.CalculDetail.Charge + tapRt.Charge
		}else if tapRt.CallType == gCT_SMS_MO{
			f64ImsiCapCharge=stImsiUsage.TapCal.SMSMO.CalculDetail.Charge + tapRt.Charge
		}else if tapRt.CallType == gCT_SMS_MT{
			f64ImsiCapCharge=stImsiUsage.TapCal.SMSMT.CalculDetail.Charge + tapRt.Charge
		}else if tapRt.CallType == gCT_DATA{
			f64ImsiCapCharge=stImsiUsage.TapCal.GPRS.CalculDetail.Charge + tapRt.Charge
		}
	}
		
	
	//change rate, fixed charge, special rule인지 구분,,,,
/*	
	if subAgt.ImsiCap.Basic[0].TypeCD != gNotExgtValue{  //Change Rate
		if 
	}else{ 

	}
*/

	return f64ImsiCapCharge
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
func calcVoiceItem (unit string, rate float64, volume float64, totCallDurat float64) float64 {
	Log_add("calcVoiceItem")

	if unit =="min"{
		return math.Ceil(totCallDurat/volume * gVoiceUnit) * rate
	}else if unit =="sec"{
		return math.Ceil(totCallDurat/ volume) * rate
	}else{
		return 0
	}
}

//SMS 계산 함수
func calcSmsItem (unit string, rate float64) float64 {
	Log_add("calcSmsItem")
	return rate
}

//DATA 계산 함수
func calcDataItem (unit string, rate float64, volume float64, totCallDurat float64) float64 {
	Log_add("calcDataItem")

	if unit =="mbytes"{
		return math.Ceil(totCallDurat/ (volume * gDataUnit)) * rate
	}else if unit =="kbytes"{
		return math.Ceil(totCallDurat/ volume) * rate
	}else if unit =="bytes"{
		return math.Ceil(totCallDurat/ (volume / gDataUnit)) * rate
	}else{
		return 0
	}
}