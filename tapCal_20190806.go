package service

import (
	"../jsonStruct"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"math"
	"strconv"
	"strings"
)



//전역변수
const gVoiceUnit float64 = 60
const gDataUnit float64 = 1024
const gPowTenOfTen float64 = 10000000000
const gNotExgtValue string = "null"
const gQuryType_ImsiUsage = "imsiUsage"
const gQuryType_Commitment = "commitment"
const gCT_MOC_LOCAL = "MOC-local"
const gCT_MOC_HOME = "MOC-home"
const gCT_MOC_INTL = "MOC-int"
const gCT_MTC = "MTC"
const gCT_SMS_MO = "SMS-mo"
const gCT_SMS_MT = "SMS-mt"
const gCT_DATA = "GRPC"
const gUnitByte = "B"
const gUnitKbyte = "KB"
const gUnitMbyte = "MB"
const gUnitSec = "sec"
const gUnitMin = "min"
const gScTypeChRate = "ChangeRate"
const gScTypeFixedChrg = "FixedCharge"
const gScTypeSpRule = "SpecialRule"
const gAddFeeTypeCallSetFee = "CallSetypFee"
const gModelTypeImsiCap = "ImsiCap"
const gModelTypeCommit = "Commitment"

/* return 값
type TapCalculreturnValue struct{
	AgreementID	string
	Peoriod		[2]string
	Currency	string
}


// GW -> BC 데이터 전송
type TapRecord struct {
	Header struct {
		VPMN                    string `json:"VPMN"`
		HPMN                    string `json:"HPMN"`
		FILE_TYPE_CD                string `json:"FILE_TYPE_CD"`
		FILE_DIV_CD           string `json:"FILE_DIV_CD"`
		FILE_SEQ_NO      string `json:"FILE_SEQ_NO"`
		FILE_CRET_DT_VAL   string `json:"FILE_CRET_DT_VAL"`
		FILE_UTC_OFFSET string `json:"FILE_UTC_OFFSET"`
		RECD_CNT             string `json:"RECD_CNT"`
	} `json:"header"`
	CdrInfos CdrInfosGW `json:"cdrInfos"`
}

type CdrInfosGW struct {
	RECD_SEQ                     string `json:"RECD_SEQ"`
	CALL_TYPE_ID               string `json:"CALL_TYPE_ID"`
	IMSI_ID                   string `json:"IMSI_ID"`
	CHAGE_ID			   string `json:"CHAGE_ID"`
	CALL_NO           string `json:"CALL_NO"`
	LOCAL_TIME         string `json:"LOCAL_TIME"`
	RECD_UTC_OFFSET          string `json:"RECD_UTC_OFFSET"`
	TOT_CALL_DURAT string `json:"TOT_CALL_DURAT"`
	IMEI_ID                   string `json:"IMEI_ID"`
	CALLG_NO          string `json:"CALLG_NO"`
	DATA_VLM_INPT_VAL     string `json:"DATA_VLM_INPT_VAL"`
	DATA_VLM_OUTPUT_VAL     string `json:"DATA_VLM_OUTPUT_VAL"`
	Charge					float64
	SetCharge				float64
}
*/



/*
package service

import (
	"../jsonStruct"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"math"
	"strconv"
	"strings"
)

func CalculChargeAmount(stub shim.ChaincodeStubInterface, cal *jsonStruct.TapRecord, recordMemory jsonStruct.RecordMemory) (jsonStruct.TapCalculreturnValue, error){

	cal.CdrInfos.Charge = 100
	cal.CdrInfos.SetCharge = 100

	returnSample := jsonStruct.TapCalculreturnValue{}
	returnSample.Peoriod = [2]string{"20150715","20200715"}
	returnSample.Currency = "USD"
	//계약 시작일 확인해서 아이디값 만들어 줘야 합니다.
	returnSample.ContractID = "contract KORKF CHNCT 20150715"

	return returnSample, nil
}



*/


//tap 요율 계산 처리 main
func CalculChargeAmount(stub shim.ChaincodeStubInterface, tapRd *jsonStruct.TapRecord, recordMemory jsonStruct.RecordMemory ) (jsonStruct.TapCalculreturnValue, error) {
	Log_add("calcChargeAmount")
	Log_add(tapRd.CdrInfos.CALL_TYPE_ID)

	var actContract jsonStruct.ContractForCal
	var stTapCalcReturn jsonStruct.TapCalculreturnValue //return구조체,,,tap정보는 pointer로 넘어온 값을 치환하여 처리(call by ref,,,)
	var stSubContract jsonStruct.Contract   //계약 서브 구조체 (past와 current중 하나 Agreement매핑)
	var bIsMonetary bool  // true : 금액 base, false : 사용량 base
	var f64ImsiCapCharge float64
	var sScImsiType string

	var f64Charge float64
	var bImsiCapFlag bool

//	var sScCommitType string
	var bCommitmentFlag bool
//	var f64CommitCharge float64
	var err error

	var stSubCtrtCalcSpImsiCap jsonStruct.CalcSpcl
	var stSubCtrtCalcSpCommit jsonStruct.CalcSpcl

	nowDate := tapRd.CdrInfos.LOCAL_TIME[:8]
	Log_add(nowDate)

	//active인 요율 계산용 agreement 조회

	actContract, err = Contract_getActive(stub, nowDate, tapRd.Header.VPMN, tapRd.Header.HPMN)
	if err != nil{
		Log_add("Agreement_getActive 조회오류")
		return stTapCalcReturn, errors.New("Agreement_getActive 조회오류")
	}

	// 처리할 tap이 agreement의 past인지 current인지 확인, imsi cap/commitment적용대상인지 확인
	stSubContract = searchAgtIdx(actContract, nowDate)
	fmt.Println(bCommitmentFlag)
	Log_add("after searchAgtIdx")

	//return구조체 값 매핑
	stTapCalcReturn.ContractID = stSubContract.CONTID
	stTapCalcReturn.Peoriod[0] = stSubContract.ContDtl.CONTSTDATE
	stTapCalcReturn.Peoriod[1] = stSubContract.ContDtl.CONTEXPDATE
	stTapCalcReturn.Currency = stSubContract.ContDtl.CONTCURCD


	// 정율 계산, additional fee 처리
	//jsonStruct.Usage와 tap record struct 인자값
	var stCalcBas jsonStruct.CalcBas

	for i:=0;i<len(stSubContract.ContDtl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stSubContract.ContDtl.CalcBas[i].CALLTYPECD {
			stCalcBas = stSubContract.ContDtl.CalcBas[i]
			f64Charge, err = calculBaseRate(stCalcBas, tapRd, tapRd.CdrInfos.TOT_CALL_DURAT)
			if err != nil{
				return stTapCalcReturn, errors.New(err.Error())
			}

			tapRd.CdrInfos.Charge = f64Charge
			break
		}
	}

	//특수조건 타입 저장
	//sScImsiType = subContract.ContDataReq.ContDtl.CalcSpcl.MODELTYPECD
	//sScCommitType = subContract.Commitment.APPLY
	for i:=0;i<len(stSubContract.ContDtl.CalcSpcl);i++{
		if stSubContract.ContDtl.CalcSpcl[i].MODELTYPECD == gModelTypeImsiCap {
			stSubCtrtCalcSpImsiCap = stSubContract.ContDtl.CalcSpcl[i]
			bImsiCapFlag = true
		}else{
			bImsiCapFlag = false
		}

		if stSubContract.ContDtl.CalcSpcl[i].MODELTYPECD == gModelTypeCommit {
			stSubCtrtCalcSpCommit = stSubContract.ContDtl.CalcSpcl[i]
			bCommitmentFlag = true
		}else{
			bCommitmentFlag = false
		}
	}


	fmt.Println(stSubCtrtCalcSpCommit)
	//Imsi Cap 계산
	//calcImsiCap(&subContract, &tapRd)
	/*
		var gScTypeChRate = "ChangeRate"
		var gScTypeFixedChrg = "FixedCharge"
		var gScTypeSpRule = "SpecialRule"
	*/

	if bImsiCapFlag == true {
		// 사용량 base인지 금액 base인지 체크
		if stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitByte || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitKbyte || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitMbyte ||
			stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitSec  || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitMin {
			//사용량 base check,,,,
			bIsMonetary = false
		}else{
			//금액 base check,,,,
			bIsMonetary = true
		}

		if sScImsiType == gScTypeChRate && bIsMonetary == false{    //change Rate and 사용량 base
			f64ImsiCapCharge, err = calcImsiCapDuration(stub, recordMemory, stCalcBas, stSubCtrtCalcSpImsiCap, tapRd)
			if err != nil{
				//에러처리
				return stTapCalcReturn, errors.New( err.Error())
			}
			tapRd.CdrInfos.SetCharge = f64ImsiCapCharge
		}else if sScImsiType == gScTypeChRate && bIsMonetary == true{  //change Rate and 금액 base
			// perImsi 금액 기준으로 check
			f64ImsiCapCharge, err = calcImsiCapMonetary(stub, recordMemory, stCalcBas, stSubCtrtCalcSpImsiCap, tapRd)
			if err != nil{
				//에러처리
				return stTapCalcReturn, errors.New( err.Error())
			}
			tapRd.CdrInfos.SetCharge = f64ImsiCapCharge
		}
	}

/*
	if bCommitmentFlag == true {
		// 사용량 base인지 금액 base인지 체크
		if stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitByte || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitKbyte || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitMbyte ||
			stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitSec  || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitMin {
			//사용량 base check,,,,
			bIsMonetary = false
		}else{
			//금액 base check,,,,
			bIsMonetary = true
		}

		if sScCommitType == gScTypeChRate && bIsMonetary == false{    //change Rate and 사용량 base
			f64CommitCharge, err = calcCommitDuration(stub, stCalcBas, stSubCtrtCalcSpCommit, tapRd)
			if err != nil{
				//에러처리
				return stTapCalcReturn, errors.New( err.Error())
			}

		}else if sScCommitType == gScTypeChRate && bIsMonetary == true{  //change Rate and 금액 base
			// perImsi 금액 기준으로 check
			f64CommitCharge = calcCommitMonetary(stub, stCalcBas, stSubCtrtCalcSpCommit, tapRd)
		}
	}
	//	fmt.Println(f64ImsiCapCharge)
*/



	return stTapCalcReturn, nil
}





//tap이 past인지 current인지 확인
func searchAgtIdx(actContract jsonStruct.ContractForCal, nowDate string) (jsonStruct.Contract) {
	Log_add("searchAgtIdx")
	var returnAgt jsonStruct.Contract

	Log_add("nowDate : " + nowDate)
	Log_add("past stdt :" + actContract.ContractInfo.Past.ContDtl.CONTSTDATE)
	Log_add("past eddt :" + actContract.ContractInfo.Past.ContDtl.CONTEXPDATE)
	Log_add("curr stdt :" + actContract.ContractInfo.Current.ContDtl.CONTSTDATE)
	Log_add("curr eddt :" + actContract.ContractInfo.Current.ContDtl.CONTEXPDATE)

	if actContract.ContractInfo.Past.ContDtl.CONTSTDATE <= nowDate && actContract.ContractInfo.Past.ContDtl.CONTEXPDATE >= nowDate{
		Log_add("PAST")
		returnAgt = actContract.ContractInfo.Past
	}else if actContract.ContractInfo.Current.ContDtl.CONTSTDATE <= nowDate && actContract.ContractInfo.Current.ContDtl.CONTEXPDATE >= nowDate {
		Log_add("Current")
		returnAgt = actContract.ContractInfo.Current
	}

	return returnAgt
}



func calculBaseRate(stCalcBas jsonStruct.CalcBas, tapRd *jsonStruct.TapRecord, sTotalCallDurat string) (float64, error) {
	Log_add("calculBaseRate")
	var f64CallSetFee float64
	var err error


	if stCalcBas.ADTNFEETYPECD == gAddFeeTypeCallSetFee { //추가 요금 적용이 있으면
		f64CallSetFee, err = strconv.ParseFloat(stCalcBas.ADTNFEEAMT, 64)
		if err != nil {
			return 0, errors.New("string to float64 conv error")
		}
	}else if stCalcBas.ADTNFEETYPECD == gNotExgtValue{ // "null"
		f64CallSetFee = 0
	}

	if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_LOCAL || tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_HOME ||
		tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_INTL || tapRd.CdrInfos.CALL_TYPE_ID == gCT_MTC {
		return f64CallSetFee + calcVoiceItem(stCalcBas.STELUNIT, stCalcBas.STELTARIF, stCalcBas.STELVLM, sTotalCallDurat), nil
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MO || tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MT  {
		return f64CallSetFee + calcSmsItem(stCalcBas.STELUNIT, stCalcBas.STELTARIF), nil
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_DATA {
		return f64CallSetFee + calcDataItem(stCalcBas.STELUNIT, stCalcBas.STELTARIF, stCalcBas.STELVLM, sTotalCallDurat), nil
	}

	return 0, errors.New("Not matched Call type")
}


//음성 계산 함수
func calcVoiceItem (unit string, rate string, volume string, totCallDurat string) float64 {
	Log_add("calcVoiceItem")

	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0 //에러처리
	}

	f64Volume, err := strconv.ParseFloat(volume, 64)
	if err != nil{
		return 0 //에러처리
	}

	f64TotCallDurat, err := strconv.ParseFloat(totCallDurat, 64)
	if err != nil{
		return 0 //에러처리
	}

	if unit ==gUnitMin{
		return math.Ceil(f64TotCallDurat/f64Volume * gVoiceUnit) * f64Rate
	}else if unit ==gUnitSec{
		return math.Ceil(f64TotCallDurat/ f64Volume) * f64Rate
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
func calcDataItem (unit string, rate string, volume string, totCallDurat string) float64 {
	Log_add("calcDataItem")

	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0 //에러처리
	}

	f64Volume, err := strconv.ParseFloat(volume, 64)
	if err != nil{
		return 0 //에러처리
	}

	f64TotCallDurat, err := strconv.ParseFloat(totCallDurat, 64)
	if err != nil{
		return 0 //에러처리
	}

	if unit ==gUnitMbyte{
		return math.Ceil(f64TotCallDurat/ (f64Volume * gDataUnit)) * f64Rate
	}else if unit ==gUnitKbyte{
		return math.Ceil(f64TotCallDurat/ f64Volume) * f64Rate
	}else if unit ==gUnitByte{
		return math.Ceil(f64TotCallDurat/ (f64Volume / gDataUnit)) * f64Rate
	}else{
		return 0
	}
}




/*

//임시 별 사용량 요약 정보
type ImsiUsage struct {
	Date             string			`json:"Date"`
	IMSI             string 		`json:"IMSI"`
	HPMN             string 		`json:"HPMN"`
	VPMN             string 		`json:"VPMN"`
	InvoiceID        string 		`json:"InvoiceID"`
	CalMonth         string 		`json:"CalMonth"`
	TapCal			 TapCal			`json:"TapCal"`
}


type TapCal struct{
	MOCLocal	CalculDetail
	MOCHome		CalculDetail
	MOCInt		CalculDetail
	MTC			CalculDetail
	SMSMO		CalculDetail
	SMSMT		CalculDetail
	GPRS		CalculDetail
}

*/


/************************************************************************************************************/
// perImsi 모델, 금액 base, change rate 계산
/************************************************************************************************************/
func calcImsiCapMonetary(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stCalcSpcl jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord) (float64, error) {

	var stImsiUsage jsonStruct.ImsiUsage  //imsi별 누적량 구조체
	var stCalcSpBas jsonStruct.CalcBas  //Agreement의 usage 요율 구조체

	var f64ImsiCapUseAmount float64
	var f64NowCharge float64

	var f64ImsiCapTHRMIN float64
	var f64ImsiCapTHRMAX float64


	//sc의 base 구조체 조회
	for i:=0;i<len(stCalcSpcl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stCalcSpcl.CalcBas[i].CALLTYPECD {
			stCalcSpBas = stCalcSpcl.CalcBas[i]
			break
		}
	}

	queryKey := strings.Fields(gQuryType_ImsiUsage+tapRd.CdrInfos.LOCAL_TIME[:8]+tapRd.CdrInfos.IMSI_ID)
	imsiCapBytes, err := jsonStruct.TapRecordUsageQuery(stub, queryKey, recordMemory)

	if err != nil{
		return 0, errors.New("ImsiCap누적 조회 오류")
	}else if imsiCapBytes != nil {

		err = json.Unmarshal(imsiCapBytes, stImsiUsage)
		if err != nil{
			return 0, errors.New("json Unmarshal error")
		}

		// 비교할 사용량 조회
		if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_LOCAL{
			f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCLocal.Charge
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_HOME{
			f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCHome.Charge
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_INTL{
			f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCInt.Charge
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MTC{
			f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCInt.Charge
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MO{
			f64ImsiCapUseAmount=stImsiUsage.TapCal.SMSMO.Charge
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MT{
			f64ImsiCapUseAmount=stImsiUsage.TapCal.SMSMT.Charge
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_DATA{
			f64ImsiCapUseAmount=stImsiUsage.TapCal.GPRS.Charge
		}
	}

	//비교를 위헤 string을 float64로 변환
	f64ImsiCapTHRMIN, err = strconv.ParseFloat(stCalcSpcl.THRSMIN, 64)
	if err != nil {
		return 0, errors.New("f64ImsiCapTHRMIN : string to float64 conv error")
	}

	f64ImsiCapTHRMAX, err = strconv.ParseFloat(stCalcSpcl.THRSMAX, 64)
	if err != nil {
		return 0, errors.New("f64ImsiCapTHRMAX : string to float64 conv error")
	}

	//calculBaseRate(stCalcBas jsonStruct.Usage, tapRd jsonStruct.TapRecord) (flaot64, error)
	if f64ImsiCapUseAmount > f64ImsiCapTHRMIN && f64ImsiCapUseAmount <= f64ImsiCapTHRMAX{
		//특수 과금
		f64NowCharge, err = calculBaseRate(stCalcSpBas, tapRd, tapRd.CdrInfos.TOT_CALL_DURAT)
		if err != nil{
			//에러처리
			return 0, errors.New( err.Error())
		}
		return f64NowCharge, nil
	}else{ // 정율 과금,,,
		return tapRd.CdrInfos.Charge, nil
	}

	return 0, errors.New( "Not matched anything")
}


/************************************************************************************************************/
// perImsi 모델, 사용량 base, change rate 계산
/************************************************************************************************************/
func calcImsiCapDuration(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stCalcSpcl jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord) (float64, error) {

	var stImsiUsage jsonStruct.ImsiUsage  //imsi별 누적량 구조체
	//var stBaseUsage jsonStruct.Usage  //Agreement의 usage 요율 구조체
	var stCalcSpBas jsonStruct.CalcBas  //Agreement의 usage 요율 구조체

	var f64ImsiCapUseDuration float64

	var f64MinBeforeDurat float64
	var f64MinNowDurat float64
	var f64MaxNowDurat float64
	var f64MaxAfterDurat float64

	var f64MinBeforeCharge float64
	var f64MinNowCharge float64
	var f64MaxNowCharge float64
	var f64MaxAfterCharge float64
	var f64NowCharge float64


	var f64ImsiCapTHRMIN float64
	var f64ImsiCapTHRMAX float64
	var f64TotalCallDurat float64

	var sMinBeforeDurat string
	var sMinNowDurat string
	var sMaxNowDurat string
	var sMaxAfterDurat string


/*
	// base요율 구조체 조회
	for i:=0;i<len(subContract.Basic);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == subContract.Basic[i].TypeCD {
			stBaseUsage = subContract.Basic[i]
			break
		}
	}
*/
	//sc의 base 구조체 조회
	for i:=0;i<len(stCalcSpcl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stCalcSpcl.CalcBas[i].CALLTYPECD {
			stCalcSpBas = stCalcSpcl.CalcBas[i]
			break
		}
	}
	//recordMemory jsonStruct.RecordMemory
	queryKey := strings.Fields(gQuryType_ImsiUsage+tapRd.CdrInfos.LOCAL_TIME[:8]+tapRd.CdrInfos.IMSI_ID)
	imsiCapBytes, err := jsonStruct.TapRecordUsageQuery(stub, queryKey, recordMemory)

	if err != nil{
		return 0, errors.New("ImsiCap누적 조회 오류")
	}else if imsiCapBytes != nil {

		err = json.Unmarshal(imsiCapBytes, stImsiUsage)
		if err != nil{
			return 0, errors.New("json Unmarshal error")
		}

		// 비교할 사용량 조회
		if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_LOCAL{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCLocal.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_HOME{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCHome.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_INTL{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCInt.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MTC{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCInt.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MO{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.SMSMO.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MT{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.SMSMT.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_DATA{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.GPRS.Duration
		}
	}

	//비교를 위헤 string을 float64로 변환
	f64ImsiCapTHRMIN, err = strconv.ParseFloat(stCalcSpcl.THRSMIN, 64)
	if err != nil {
		return 0, errors.New("f64ImsiCapTHRMIN : string to float64 conv error")
	}

	f64ImsiCapTHRMAX, err = strconv.ParseFloat(stCalcSpcl.THRSMAX, 64)
	if err != nil {
		return 0, errors.New("f64ImsiCapTHRMAX : string to float64 conv error")
	}

	f64TotalCallDurat, err = strconv.ParseFloat(tapRd.CdrInfos.TOT_CALL_DURAT, 64)
	if err != nil {
		return 0, errors.New("f64TotalCallDurat : string to float64 conv error")
	}

	//calculBaseRate(stCalcBas jsonStruct.Usage, tapRd jsonStruct.TapRecord) (flaot64, error)
	if f64ImsiCapUseDuration + f64TotalCallDurat > f64ImsiCapTHRMIN && f64ImsiCapUseDuration <= f64ImsiCapTHRMAX{
		if f64ImsiCapUseDuration > f64ImsiCapTHRMIN {  //min에 걸치 호 처리
			f64MinBeforeDurat = f64TotalCallDurat-(f64TotalCallDurat+f64ImsiCapUseDuration-f64ImsiCapTHRMIN)
			f64MinNowDurat = f64TotalCallDurat-f64MinBeforeDurat

			sMinBeforeDurat = strconv.FormatFloat(f64MinBeforeDurat, 'G', -1, 64)
			sMinNowDurat = strconv.FormatFloat(f64MinNowDurat, 'G', -1, 64)


			f64MinBeforeCharge, err = calculBaseRate(stCalcSpBas, tapRd, sMinBeforeDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}
			f64MinNowCharge, err = calculBaseRate(stCalcBas, tapRd, sMinNowDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}

			return f64MinBeforeCharge + f64MinNowCharge, nil

		}else if f64ImsiCapUseDuration+f64TotalCallDurat > f64ImsiCapTHRMAX{  // max에 걸친 호 처리
			f64MaxNowDurat = f64TotalCallDurat-(f64TotalCallDurat+f64ImsiCapUseDuration-f64ImsiCapTHRMAX)
			f64MaxAfterDurat = f64TotalCallDurat-f64MaxNowDurat
			sMaxNowDurat = strconv.FormatFloat(f64MaxNowDurat, 'G', -1, 64)
			sMaxAfterDurat = strconv.FormatFloat(f64MaxAfterDurat, 'G', -1, 64)


			f64MaxNowCharge, err = calculBaseRate(stCalcSpBas, tapRd, sMaxNowDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}
			f64MaxAfterCharge, err = calculBaseRate(stCalcBas, tapRd, sMaxAfterDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}

			return f64MaxNowCharge + f64MaxAfterCharge, nil

		}else{  //특수 과금
			f64NowCharge, err = calculBaseRate(stCalcSpBas, tapRd, tapRd.CdrInfos.TOT_CALL_DURAT)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}
			return f64NowCharge, nil
		}
	}else{ // 정율 과금,,,
		return tapRd.CdrInfos.Charge, nil
	}

	return 0, errors.New( "Not matched anything")

}


/************************************************************************************************************/
// perImsi 모델, 사용량 base, change rate 계산
/************************************************************************************************************/
/*
func calcCommitDuration(stub shim.ChaincodeStubInterface, stCalcBas jsonStruct.CalcBas, stCalcSpcl jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord) (float64, error) {

	var stTotalUsage jsonStruct.TotalUsage   //계약 기간 별 사용량 요약 정보
	//var stBaseUsage jsonStruct.Usage  //Agreement의 usage 요율 구조체
	var stCalcSpBas jsonStruct.CalcBas  //Agreement의 usage 요율 구조체

	var f64CommitUseDuration float64

	var f64MinBeforeDurat float64
	var f64MinNowDurat float64
	var f64MaxNowDurat float64
	var f64MaxAfterDurat float64

	var f64MinBeforeCharge float64
	var f64MinNowCharge float64
	var f64MaxNowCharge float64
	var f64MaxAfterCharge float64
	var f64NowCharge float64


	var f64CommitTHRMIN float64
	var f64CommitTHRMAX float64
	var f64TotalCallDurat float64

	var sMinBeforeDurat string
	var sMinNowDurat string
	var sMaxNowDurat string
	var sMaxAfterDurat string


	//sc의 base 구조체 조회
	for i:=0;i<len(stCalcSpcl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stCalcSpcl.CalcBas[i].CALLTYPECD {
			stCalcSpBas = stCalcSpcl.CalcBas[i]
			break
		}
	}

	queryKey := strings.Fields(gQuryType_Commitment+tapRd.Header.VPMN+tapRd.Header.HPMN+)
	imsiCapBytes, err := Block_Query(stub, queryKey)

	if err != nil{
		return 0, errors.New("ImsiCap누적 조회 오류")
	}else if imsiCapBytes != nil {

		err = json.Unmarshal(imsiCapBytes[0], stImsiUsage)
		if err != nil{
			return 0, errors.New("json Unmarshal error")
		}

		// 비교할 사용량 조회
		if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_LOCAL{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCLocal.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_HOME{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCHome.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MOC_INTL{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCInt.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_MTC{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.MOCInt.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MO{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.SMSMO.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_SMS_MT{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.SMSMT.Duration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCT_DATA{
			f64ImsiCapUseDuration=stImsiUsage.TapCal.GPRS.Duration
		}
	}

	//비교를 위헤 string을 float64로 변환
	f64ImsiCapTHRMIN, err = strconv.ParseFloat(stCalcSpcl.THRSMIN, 64)
	if err != nil {
		return 0, errors.New("f64ImsiCapTHRMIN : string to float64 conv error")
	}

	f64ImsiCapTHRMAX, err = strconv.ParseFloat(stCalcSpcl.THRSMAX, 64)
	if err != nil {
		return 0, errors.New("f64ImsiCapTHRMAX : string to float64 conv error")
	}

	f64TotalCallDurat, err = strconv.ParseFloat(tapRd.CdrInfos.TOT_CALL_DURAT, 64)
	if err != nil {
		return 0, errors.New("f64TotalCallDurat : string to float64 conv error")
	}

	//calculBaseRate(stCalcBas jsonStruct.Usage, tapRd jsonStruct.TapRecord) (flaot64, error)
	if f64ImsiCapUseDuration + f64TotalCallDurat > f64ImsiCapTHRMIN && f64ImsiCapUseDuration <= f64ImsiCapTHRMAX{
		if f64ImsiCapUseDuration > f64ImsiCapTHRMIN {  //min에 걸치 호 처리
			f64MinBeforeDurat = f64TotalCallDurat-(f64TotalCallDurat+f64ImsiCapUseDuration-f64ImsiCapTHRMIN)
			f64MinNowDurat = f64TotalCallDurat-f64MinBeforeDurat

			sMinBeforeDurat = strconv.FormatFloat(f64MinBeforeDurat, 'G', -1, 64)
			sMinNowDurat = strconv.FormatFloat(f64MinNowDurat, 'G', -1, 64)


			f64MinBeforeCharge, err = calculBaseRate(stCalcSpBas, tapRd, sMinBeforeDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}
			f64MinNowCharge, err = calculBaseRate(stCalcBas, tapRd, sMinNowDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}

			return f64MinBeforeCharge + f64MinNowCharge, nil

		}else if f64ImsiCapUseDuration+f64TotalCallDurat > f64ImsiCapTHRMAX{  // max에 걸친 호 처리
			f64MaxNowDurat = f64TotalCallDurat-(f64TotalCallDurat+f64ImsiCapUseDuration-f64ImsiCapTHRMAX)
			f64MaxAfterDurat = f64TotalCallDurat-f64MaxNowDurat
			sMaxNowDurat = strconv.FormatFloat(f64MaxNowDurat, 'G', -1, 64)
			sMaxAfterDurat = strconv.FormatFloat(f64MaxAfterDurat, 'G', -1, 64)


			f64MaxNowCharge, err = calculBaseRate(stCalcSpBas, tapRd, sMaxNowDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}
			f64MaxAfterCharge, err = calculBaseRate(stCalcBas, tapRd, sMaxAfterDurat)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}

			return f64MaxNowCharge + f64MaxAfterCharge, nil

		}else{  //특수 과금
			f64NowCharge, err = calculBaseRate(stCalcSpBas, tapRd, tapRd.CdrInfos.TOT_CALL_DURAT)
			if err != nil{
				//에러처리
				return 0, errors.New( err.Error())
			}
			return f64NowCharge, nil
		}
	}else{ // 정율 과금,,,
		return tapRd.CdrInfos.Charge, nil
	}

	return 0, errors.New( "Not matched anything")

}
*/