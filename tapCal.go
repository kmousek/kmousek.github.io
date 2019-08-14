package service

import (
	"github.com/main/go/jsonStruct"
	"encoding/json"
	"errors"
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
const gQuryType_ImsiUsage = "imsiUsage"
const gQuryType_Commitment = "commitment"
const gCallTypeMocLocal = "MOC-local"
const gCallTypeMocHome = "MOC-home"
const gCallTypeMocInt = "MOC-int"
const gCallTypeMtc = "MTC"
const gCallTypeSmsMo = "SMS-mo"
const gCallTypeSmsMt = "SMS-mt"
const gCallTypeData = "GPRS"
const gCallTypeAll = "All-Services"
const gUnitByte = "B"
const gUnitKbyte = "KB"
const gUnitMbyte = "MB"
const gUnitSec = "sec"
const gUnitMin = "min"
const gUnitOcc = "occ"
const gApplyTypeChRate = "ChangeRate"
const gApplyTypeFixedChrg = "FixedCharge"
const gApplyTypeSpRule = "SpecialRule"
const gAddFeeTypeCallSetFee = "CallSetypFee"
const gModelTypeImsiCap = "Imsicap"
const gModelTypeCommit = "Commitment"
const gY = "Y"
const gBaseRate = "B"
const gSpecialRate = "S"


//tap 요율 계산 처리 main
func CalculChargeAmount(stub shim.ChaincodeStubInterface, tapRd *jsonStruct.TapRecord, recordMemory jsonStruct.RecordMemory ) (jsonStruct.TapCalculreturnValue, error) {
	Log_add("========================================================================================")
	Log_add("======================function : CalculChargeAmount")
	Log_add("========================================================================================")
	Log_add(tapRd.CdrInfos.CALL_TYPE_ID)

	var actContract jsonStruct.ContractForCal
	var stTapCalcReturn jsonStruct.TapCalculreturnValue //return구조체,,,tap정보는 pointer로 넘어온 값을 치환하여 처리(call by ref,,,)
	var stSubContract jsonStruct.Contract   //계약 서브 구조체 (past와 current중 하나 Agreement매핑)
	//var bIsMonetary bool  // true : 금액 base, false : 사용량 base
	//var f64ImsiCapCharge float64
	//	var sScImsiType string

	var f64Charge float64
	//var bImsiCapFlag bool

	//	var sScCommitType string
	var bCommitmentFlag bool
	//	var f64CommitCharge float64
	var err error

	var stSubCtrtCalcSpImsiCap jsonStruct.CalcSpcl
	var stSubCtrtCalcSpCommit jsonStruct.CalcSpcl
	var stCalcBas jsonStruct.CalcBas
	var f64TaxPercent float64
	var f64TaxCharge float64
	var nowDate string
	var f64TapActDurat float64

	nowDate = tapRd.CdrInfos.LOCAL_TIME[:8]
	Log_add(nowDate)


	//active인 요율 계산용 agreement 조회
	actContract, err = Contract_getActive(stub, nowDate, tapRd.Header.VPMN, tapRd.Header.HPMN)
	if err != nil{
		Log_add("Agreement_getActive 조회오류")
		return stTapCalcReturn, errors.New("Agreement_getActive 조회오류")
	}

	// 처리할 tap이 agreement의 past인지 current인지 확인
	stSubContract = searchAgtIdx(actContract, nowDate)
	fmt.Println(bCommitmentFlag)
	Log_add("after searchAgtIdx")

	//Tax 부가 여부 flag
	if stSubContract.ContDtl.TAXAPLYPECNT == gNotExgtValue { //값이 "null"이면 tax percent를 0으로 셋팅
		f64TaxPercent = 0
	}else{
		f64TaxPercent, err = strconv.ParseFloat(stSubContract.ContDtl.TAXAPLYPECNT, 64)
		if err != nil{
			return stTapCalcReturn, errors.New("f64TaxPercent : ParseFloat Error") //에러처리
		}
	}

	//return구조체 값 매핑
	stTapCalcReturn.ContractID = stSubContract.CONTID
	stTapCalcReturn.Peoriod[0] = stSubContract.ContDtl.CONTSTDATE
	stTapCalcReturn.Peoriod[1] = stSubContract.ContDtl.CONTEXPDATE
	stTapCalcReturn.Currency = stSubContract.ContDtl.CONTCURCD
	Log_add("after stTapCalcReturn mapping")

	// 정율 계산, additional fee 처리
	//jsonStruct.Usage와 tap record struct 인자값

	f64TapActDurat, err = strconv.ParseFloat(tapRd.CdrInfos.TOT_CALL_DURAT,64)
	if err != nil{
		return stTapCalcReturn, errors.New("totDurat : parseFloat Error")
	}

	Log_add("tapRd.CdrInfos.CALL_TYPE_ID : [" + tapRd.CdrInfos.CALL_TYPE_ID + "]")


	/******************************************************************************************************
	정율 계산 처리
	*****************************************************************************************************/
	for i:=0;i<len(stSubContract.ContDtl.CalcBas);i++{
		Log_add("stSubContract.ContDtl.CalcBas[i].CALLTYPECD : [" + stSubContract.ContDtl.CalcBas[i].CALLTYPECD + "]")

		if tapRd.CdrInfos.CALL_TYPE_ID == stSubContract.ContDtl.CalcBas[i].CALLTYPECD {
			Log_add("tapRd.CdrInfos.CALL_TYPE_ID == stSubContract.ContDtl.CalcBas[i].CALLTYPECD")
			stCalcBas = stSubContract.ContDtl.CalcBas[i]
			f64Charge, f64TaxCharge, err = calculBaseRate(gBaseRate, stCalcBas, tapRd, f64TapActDurat, f64TaxPercent)
			if err != nil{
				return stTapCalcReturn, errors.New(err.Error())
			}
			tapRd.CdrInfos.CalculDetail.Charge = f64Charge
			tapRd.CdrInfos.CalculDetail.TAXCharge = f64TaxCharge
			tapRd.CdrInfos.CalculDetail.SetCharge = f64Charge
			tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64TaxCharge

			break
		}
	}

	Log_add("***********************Tap 정율 과금 처리 결과 시작***********************")
	Log_add("HPMN             : ["+tapRd.Header.HPMN+"]")
	Log_add("VPMN             : ["+tapRd.Header.VPMN+"]")
	Log_add("CALL_TYPE_ID     : ["+tapRd.CdrInfos.CALL_TYPE_ID+"]")
	Log_add("TOT_CALL_DURAT   : ["+tapRd.CdrInfos.TOT_CALL_DURAT+"]")
	Log_add("LOCAL_TIME       : ["+tapRd.CdrInfos.LOCAL_TIME+"]")
	Log_add("IMSI_ID          : ["+tapRd.CdrInfos.IMSI_ID+"]")
	Log_add("Record           : ["+strconv.Itoa(tapRd.CdrInfos.CalculDetail.Record)+"]")
	Log_add("Unit             : ["+tapRd.CdrInfos.CalculDetail.Unit+"]")
	Log_add("Duration         : ["+strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.Duration,'g',-1,64)+"]")
	Log_add("RoundedDuration  : ["+strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.RoundedDuration,'g',-1,64)+"]")
	Log_add("TAXINCLYN        : ["+tapRd.CdrInfos.CalculDetail.TAXINCLYN+"]")
	Log_add("Charge           : ["+strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.Charge,'g',-1,64)+"]")
	Log_add("TAXCharge        : ["+strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.TAXCharge,'g',-1,64)+"]")
	Log_add("SetCharge        : ["+strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.SetCharge,'g',-1,64)+"]")
	Log_add("TAXSETCharge     : ["+strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.TAXSETCharge,'g',-1,64)+"]")
	Log_add("***********************Tap 정율 과금 처리 결과 끝*************************")

	/******************************************************************************************************
	특수조건 처리 (per IMSI, Commitment)
	*****************************************************************************************************/
	//특수조건이 배열로 들어가 있으므로 for문 돌면서 Model type 체크해서 특수조건 처리 함수 호출
	for i:=0;i<len(stSubContract.ContDtl.CalcSpcl);i++{
		Log_add("stSubContract.ContDtl.CalcSpcl[i].MODELTYPECD : ["+stSubContract.ContDtl.CalcSpcl[i].MODELTYPECD+"]")
		if stSubContract.ContDtl.CalcSpcl[i].MODELTYPECD == gModelTypeImsiCap {
			Log_add("in stSubContract.ContDtl.CalcSpcl[i].MODELTYPECD == gModelTypeImsiCap")
			stSubCtrtCalcSpImsiCap = stSubContract.ContDtl.CalcSpcl[i]
			err = calculImsiCap(stub, recordMemory, stCalcBas, stSubCtrtCalcSpImsiCap, tapRd, f64TaxPercent)
			if err != nil {
				return stTapCalcReturn, errors.New(err.Error())
			}
		}

		if stSubContract.ContDtl.CalcSpcl[i].MODELTYPECD == gModelTypeCommit {
			stSubCtrtCalcSpCommit = stSubContract.ContDtl.CalcSpcl[i]
			calculCommitment(stub, recordMemory, stCalcBas, stSubCtrtCalcSpCommit, tapRd, f64TaxPercent, stSubContract.CONTID)
			if err != nil {
				return stTapCalcReturn, errors.New(err.Error())
			}
		}
	}


	Log_add("***********************return 값***********************")
	Log_add("stTapCalcReturn.ContractID : ["+stTapCalcReturn.ContractID+"]")
	Log_add("stTapCalcReturn.Peoriod[0] : ["+stTapCalcReturn.Peoriod[0]+"]")
	Log_add("stTapCalcReturn.Peoriod[1] : ["+stTapCalcReturn.Peoriod[1]+"]")
	Log_add("stTapCalcReturn.Currency   : ["+stTapCalcReturn.Currency+"]")
	Log_add("***********************return 값 끝********************")

	return stTapCalcReturn, nil
}

/************************************************************************************************************/
//Imsi cap 처리
/************************************************************************************************************/
func calculImsiCap(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stSubCtrtCalcSpImsiCap jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord, f64TaxPercent float64) error {
	Log_add("======================function : calculImsiCap")

	var bIsMonetary bool
	var err error

	Log_add("stSubCtrtCalcSpImsiCap.THRSUNIT: " + stSubCtrtCalcSpImsiCap.THRSUNIT)
	// 사용량 base인지 금액 base인지 체크
	if stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitKbyte || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitMbyte ||
		stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitSec  || stSubCtrtCalcSpImsiCap.THRSUNIT == gUnitMin {
		//사용량 base check,,,,
		bIsMonetary = false
		Log_add("bIsMonetary = false")
	}else{
		//금액 base check,,,,
		bIsMonetary = true
		Log_add("bIsMonetary = true")
	}

	Log_add("stSubCtrtCalcSpImsiCap.APLYTYPE : ["+stSubCtrtCalcSpImsiCap.APLYTYPE+"]")
	if stSubCtrtCalcSpImsiCap.APLYTYPE == gApplyTypeChRate && bIsMonetary == false{    //change Rate and 사용량 base
		Log_add("imsicap change Rate and 사용량 base")
		err = calcImsiCapDuration(stub, recordMemory, stCalcBas, stSubCtrtCalcSpImsiCap, tapRd, f64TaxPercent)
		if err != nil{
			//에러처리
			Log_add(err.Error())
			return errors.New( err.Error())
		}

	}else if stSubCtrtCalcSpImsiCap.APLYTYPE == gApplyTypeChRate && bIsMonetary == true{  //change Rate and 금액 base
		Log_add("imsicap change Rate and 금액 base")
		// perImsi 금액 기준으로 check
		err = calcImsiCapMonetary(stub, recordMemory, stCalcBas, stSubCtrtCalcSpImsiCap, tapRd, f64TaxPercent)
		if err != nil{
			//에러처리
			Log_add(err.Error())
			return errors.New( err.Error())
		}
	}
	return nil
}


/************************************************************************************************************/
//Commitment 처리
/************************************************************************************************************/
func calculCommitment(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stSubCtrtCalcSpCommit jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord, f64TaxPercent float64, sContractId string) error {
	Log_add("======================function : calculCommitment")
	var bIsMonetary bool
	//var f64CommitCharge float64
	var err error

	// 사용량 base인지 금액 base인지 체크
	if stSubCtrtCalcSpCommit.THRSUNIT == gUnitKbyte || stSubCtrtCalcSpCommit.THRSUNIT == gUnitMbyte ||
		stSubCtrtCalcSpCommit.THRSUNIT == gUnitSec  || stSubCtrtCalcSpCommit.THRSUNIT == gUnitMin {
		//사용량 base check,,,,
		bIsMonetary = false
	}else{
		//금액 base check,,,,
		bIsMonetary = true
	}

	if stSubCtrtCalcSpCommit.APLYTYPE == gApplyTypeChRate && bIsMonetary == false{    //change Rate and 사용량 base
		err = calcCommitDuration(stub, recordMemory, stCalcBas, stSubCtrtCalcSpCommit, tapRd, f64TaxPercent, sContractId)
		if err != nil{
			//에러처리
			return errors.New( err.Error())
		}

	}else if stSubCtrtCalcSpCommit.APLYTYPE == gApplyTypeChRate && bIsMonetary == true{  //change Rate and 금액 base
		// perImsi 금액 기준으로 check
		err = calcCommitMonetary(stub, recordMemory, stCalcBas, stSubCtrtCalcSpCommit, tapRd, f64TaxPercent, sContractId)
		if err != nil{
			return errors.New( err.Error())
		}

	}

	//	fmt.Println(f64ImsiCapCharge)

	return nil
}


/************************************************************************************************************/
//계약 조회 (tap이 past인지 current인지 확인)
/************************************************************************************************************/
func searchAgtIdx(actContract jsonStruct.ContractForCal, nowDate string) (jsonStruct.Contract) {
	Log_add("======================function : searchAgtIdx")
	var returnAgt jsonStruct.Contract

	Log_add("nowDate : " + nowDate)
	Log_add("past stdt : [" + actContract.ContractInfo.Past.ContDtl.CONTSTDATE + "]")
	Log_add("past eddt : [" + actContract.ContractInfo.Past.ContDtl.CONTEXPDATE + "]")
	Log_add("curr stdt : [" + actContract.ContractInfo.Current.ContDtl.CONTSTDATE + "]")
	Log_add("curr eddt : [" + actContract.ContractInfo.Current.ContDtl.CONTEXPDATE + "]")

	if actContract.ContractInfo.Past.ContDtl.CONTSTDATE <= nowDate && actContract.ContractInfo.Past.ContDtl.CONTEXPDATE >= nowDate{
		Log_add("PAST")
		returnAgt = actContract.ContractInfo.Past
	}else if actContract.ContractInfo.Current.ContDtl.CONTSTDATE <= nowDate && actContract.ContractInfo.Current.ContDtl.CONTEXPDATE >= nowDate {
		Log_add("Current")
		returnAgt = actContract.ContractInfo.Current
	}

	return returnAgt
}


/************************************************************************************************************/
//정율 과금 처리
/************************************************************************************************************/
func calculBaseRate(sRateType string, stCalcBas jsonStruct.CalcBas, tapRd *jsonStruct.TapRecord, f64TapActDurat float64, f64TaxPercent float64) (float64, float64, error) {
	Log_add("======================function : calculBaseRate")
	var f64CallSetFee float64
	var err error
	var f64Charge float64
	var f64TaxCharge float64
	var f64RoundedDuration float64

	//tap record에 tax incl yn, 정율 unit 매핑
	//tapRd.CdrInfos.CalculDetail.TAXINCLYN = stCalcBas.TAXINCLYN
	//tapRd.CdrInfos.CalculDetail.Unit = stCalcBas.STELUNIT

	Log_add("check ADTNFEETYPECD st")
	if stCalcBas.ADTNFEETYPECD == gAddFeeTypeCallSetFee { //추가 요금 적용이 있으면
		f64CallSetFee, err = strconv.ParseFloat(stCalcBas.ADTNFEEAMT, 64)
		if err != nil {
			return 0, 0, errors.New("string to float64 conv error")
		}
	}else if stCalcBas.ADTNFEETYPECD == gNotExgtValue{ // "null"
		f64CallSetFee = 0
	}
	Log_add("check ADTNFEETYPECD ed")
	Log_add("tapRd.CdrInfos.CALL_TYPE_ID : [" + tapRd.CdrInfos.CALL_TYPE_ID + "]")

	if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocLocal || tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocHome ||
		tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocInt || tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMtc {
		Log_add("if VOICE")
		f64Charge, f64RoundedDuration, err = calcVoiceItem(stCalcBas.STELUNIT, stCalcBas.STELTARIF, stCalcBas.STELVLM, f64TapActDurat)
		if err !=nil{
			return 0, 0, errors.New( err.Error())
		}

		if stCalcBas.TAXINCLYN == gY && f64TaxPercent > 0 {
			f64TaxCharge = f64Charge/f64TaxPercent
		}else{
			f64TaxCharge = 0
		}

		if sRateType == gBaseRate {
			tapRd.CdrInfos.CalculDetail.Unit = gUnitSec
			tapRd.CdrInfos.CalculDetail.Duration = f64TapActDurat
			tapRd.CdrInfos.CalculDetail.RoundedDuration = f64RoundedDuration
			tapRd.CdrInfos.CalculDetail.TAXINCLYN = stCalcBas.TAXINCLYN
		}

		f64Charge = f64Charge + f64CallSetFee
		return f64Charge, f64TaxCharge, nil
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMo || tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMt  {
		Log_add("if SMS")

		f64Charge, err = calcSmsItem(stCalcBas.STELUNIT, stCalcBas.STELTARIF)
		if err !=nil{
			return 0, 0, errors.New( err.Error())
		}

		if stCalcBas.TAXINCLYN == gY && f64TaxPercent > 0 {
			f64TaxCharge = f64Charge/f64TaxPercent
		}else{
			f64TaxCharge = 0
		}

		if sRateType == gBaseRate {
			tapRd.CdrInfos.CalculDetail.Unit = gUnitOcc
			tapRd.CdrInfos.CalculDetail.Duration = f64TapActDurat
			tapRd.CdrInfos.CalculDetail.RoundedDuration = f64TapActDurat
			tapRd.CdrInfos.CalculDetail.TAXINCLYN = stCalcBas.TAXINCLYN
		}

		f64Charge = f64Charge + f64CallSetFee
		return f64Charge, f64TaxCharge, nil

	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeData {
		Log_add("if gCallTypeData")

		f64Charge, f64RoundedDuration, err= calcDataItem(stCalcBas.STELUNIT, stCalcBas.STELTARIF, stCalcBas.STELVLM, f64TapActDurat)
		if err != nil{
			return 0, 0, errors.New( err.Error())
		}
		if stCalcBas.TAXINCLYN == gY && f64TaxPercent > 0 {
			f64TaxCharge = f64Charge/f64TaxPercent
		}else{
			f64TaxCharge = 0
		}

		if sRateType == gBaseRate {
			tapRd.CdrInfos.CalculDetail.Unit = gUnitKbyte
			tapRd.CdrInfos.CalculDetail.Duration = f64TapActDurat
			tapRd.CdrInfos.CalculDetail.RoundedDuration = f64RoundedDuration
			tapRd.CdrInfos.CalculDetail.TAXINCLYN = stCalcBas.TAXINCLYN
		}

		f64Charge = f64Charge + f64CallSetFee
		return f64Charge, f64TaxCharge, nil
	}

	Log_add("Not matched Call type")
	return 0, 0, errors.New("Not matched Call type")
}


/************************************************************************************************************/
//음성 계산 함수
/************************************************************************************************************/
func calcVoiceItem (unit string, rate string, volume string, f64TapActDurat float64) (float64, float64, error) {
	Log_add("======================function : calcVoiceItem")

	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0,0, errors.New("f64Rate : ParseFloat error") //에러처리
	}

	f64Volume, err := strconv.ParseFloat(volume, 64)
	if err != nil{
		return 0,0, errors.New("f64Volume : ParseFloat error") //에러처리
	}

	Log_add("rate : ["+rate+"]")
	Log_add("volume : ["+volume+"]")
	Log_add("unit : ["+unit+"]")

	if unit ==gUnitMin{
		Log_add("if gUnitMin")
		Log_add("charge        : ["+strconv.FormatFloat(math.Ceil(f64TapActDurat/(f64Volume * gVoiceUnit)) * f64Rate,'g',-1,64)+"]")
		Log_add("roundingDurat : ["+strconv.FormatFloat(math.Ceil(f64TapActDurat/(f64Volume * gVoiceUnit)) * f64Volume,'g',-1,64)+"]")
		return math.Ceil(f64TapActDurat/(f64Volume * gVoiceUnit)) * f64Rate, math.Ceil(f64TapActDurat/(f64Volume * gVoiceUnit)) * f64Volume, nil
	}else if unit ==gUnitSec{
		Log_add("if gUnitSec")
		return math.Ceil(f64TapActDurat/ f64Volume) * f64Rate, math.Ceil(f64TapActDurat/ f64Volume) * f64Volume, nil
	}else{
		Log_add("else")
		return 0,0,nil
	}
}


/************************************************************************************************************/
//SMS 계산 함수
/************************************************************************************************************/
func calcSmsItem (unit string, rate string) (float64,error) {
	Log_add("======================function : calcSmsItem")

	Log_add("rate : ["+rate+"]")
	Log_add("unit : ["+unit+"]")

	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0, errors.New("f64Rate : ParseFloat error") //에러처리
	}


	return f64Rate, nil
}


/************************************************************************************************************/
//DATA 계산 함수
/************************************************************************************************************/
func calcDataItem (unit string, rate string, volume string, f64TapActDurat float64) (float64, float64, error) {
	Log_add("======================function : calcDataItem")
	var f64TapActDuratToKB float64

	f64Rate, err := strconv.ParseFloat(rate, 64)
	if err != nil{
		return 0,0, errors.New("f64Rate : ParseFloat error") //에러처리
	}

	f64Volume, err := strconv.ParseFloat(volume, 64)
	if err != nil{
		return 0,0, errors.New("f64Volume : ParseFloat error") //에러처리
	}

	f64TapActDuratToKB = f64TapActDurat/gDataUnit  //Byte를 Kbyte로 변환

	Log_add("rate : ["+rate+"]")
	Log_add("volume : ["+volume+"]")
	Log_add("unit : ["+unit+"]")

	if unit ==gUnitMbyte{
		Log_add("if gUnitMbyte")
		return math.Ceil(f64TapActDuratToKB/ (f64Volume * gDataUnit)) * f64Rate, math.Ceil(f64TapActDuratToKB/ (f64Volume * gDataUnit)) * f64Volume, nil
	}else if unit ==gUnitKbyte{
		Log_add("if gUnitKbyte")
		return math.Ceil(f64TapActDuratToKB/ f64Volume) * f64Rate, math.Ceil(f64TapActDuratToKB/ f64Volume) * f64Volume, nil
	}else{
		Log_add("else")
		return 0, 0, nil
	}
}


/************************************************************************************************************/
// perImsi 모델, 금액 base, change rate 계산
/************************************************************************************************************/
func calcImsiCapMonetary(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stCalcSpcl jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord, f64TaxPercent float64) (error) {
	Log_add("======================function : calcImsiCapMonetary")
	var stImsiUsage jsonStruct.ImsiUsage  //imsi별 누적량 구조체
	var stCalcSpBas jsonStruct.CalcBas  //Agreement의 usage 요율 구조체

	var f64ImsiCapUseAmount float64
	var f64NowCharge float64

	var f64ImsiCapTHRMIN float64
	var f64ImsiCapTHRMAX float64
	var f64NowTaxCharge float64

	//sc의 base 구조체 조회
	for i:=0;i<len(stCalcSpcl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stCalcSpcl.CalcBas[i].CALLTYPECD {
			stCalcSpBas = stCalcSpcl.CalcBas[i]
			break
		}
	}


	queryKey := []string{gQuryType_ImsiUsage, tapRd.CdrInfos.LOCAL_TIME[:8], tapRd.CdrInfos.IMSI_ID}
	//queryKey := strings.Fields(gQuryType_ImsiUsage+tapRd.CdrInfos.LOCAL_TIME[:8]+tapRd.CdrInfos.IMSI_ID)
	imsiCapBytes, err := jsonStruct.TapRecordUsageQuery(stub, queryKey, recordMemory)

	if err.Error() == "wrong key data" {
		//no row selected
	}else if err != nil{
		Log_add("ImsiCap누적 조회 오류")
		return errors.New("ImsiCap누적 조회 오류")
	}else if imsiCapBytes != nil {
		err = json.Unmarshal(imsiCapBytes, stImsiUsage)
		if err != nil{
			Log_add("json Unmarshal error")
			return errors.New("json Unmarshal error")
		}
	}

	//비교를 위헤 string을 float64로 변환
	f64ImsiCapTHRMIN, err = strconv.ParseFloat(stCalcSpcl.THRSMIN, 64)
	if err != nil {
		return errors.New("f64ImsiCapTHRMIN : string to float64 conv error")
	}

	f64ImsiCapTHRMAX, err = strconv.ParseFloat(stCalcSpcl.THRSMAX, 64)
	if err != nil {
		return errors.New("f64ImsiCapTHRMAX : string to float64 conv error")
	}

	// 비교할 사용량 조회
	if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocLocal{
		f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCLocal.Charge
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocHome{
		f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCHome.Charge
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocInt{
		f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCInt.Charge
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMtc{
		f64ImsiCapUseAmount=stImsiUsage.TapCal.MOCInt.Charge
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMo{
		f64ImsiCapUseAmount=stImsiUsage.TapCal.SMSMO.Charge
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMt{
		f64ImsiCapUseAmount=stImsiUsage.TapCal.SMSMT.Charge
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeData{
		f64ImsiCapUseAmount=stImsiUsage.TapCal.GPRS.Charge
	}

	//calculBaseRate(stCalcBas jsonStruct.Usage, tapRd jsonStruct.TapRecord) (flaot64, error)
	if f64ImsiCapUseAmount > f64ImsiCapTHRMIN && f64ImsiCapUseAmount <= f64ImsiCapTHRMAX{
		//특수 과금
		f64NowCharge, f64NowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, tapRd.CdrInfos.CalculDetail.Duration, f64TaxPercent)
		if err != nil{
			//에러처리
			return errors.New( err.Error())
		}

		tapRd.CdrInfos.CalculDetail.SetCharge = f64NowCharge
		tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64NowTaxCharge

	}else{ // 정율 과금,,,
		tapRd.CdrInfos.CalculDetail.SetCharge = tapRd.CdrInfos.CalculDetail.Charge
		tapRd.CdrInfos.CalculDetail.TAXSETCharge = tapRd.CdrInfos.CalculDetail.TAXCharge
		Log_add("IMSI CAP 정율과금")
	}

	return errors.New( "Not matched anything")
}


/************************************************************************************************************/
// perImsi 모델, 사용량 base, change rate 계산
/************************************************************************************************************/
func calcImsiCapDuration(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stCalcSpcl jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord, f64TaxPercent float64) (error) {
	Log_add("======================function : calcImsiCapDuration")

	var stImsiUsage jsonStruct.ImsiUsage  //imsi별 누적량 구조체
	//var stBaseUsage jsonStruct.Usage  //Agreement의 usage 요율 구조체
	var stCalcSpBas jsonStruct.CalcBas  //Agreement의 usage 요율 구조체

	var f64ImsiCapUseRoundedDurat float64

	var f64MinBeforeDurat float64
	var f64MinNowDurat float64
	var f64MaxNowDurat float64
	var f64MaxAfterDurat float64

	var f64MinBeforeCharge float64
	var f64MinNowCharge float64
	var f64MaxNowCharge float64
	var f64MaxAfterCharge float64
	var f64NowCharge float64


	var f64MinBeforeTaxCharge float64
	var f64MinNowTaxCharge float64
	var f64MaxNowTaxCharge float64
	var f64MaxAfterTaxCharge float64
	var f64NowTaxCharge float64

	var f64ImsiCapTHRMIN float64
	var f64ImsiCapTHRMAX float64
	//var f64TotalCallDurat float64

	//var sMinBeforeDurat string
	//var sMinNowDurat string
	//var sMaxNowDurat string
	//var sMaxAfterDurat string


	//sc의 base 구조체 조회
	for i:=0;i<len(stCalcSpcl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stCalcSpcl.CalcBas[i].CALLTYPECD {
			stCalcSpBas = stCalcSpcl.CalcBas[i]
			break
		}
	}
	//recordMemory jsonStruct.RecordMemory
	queryKey := []string{gQuryType_ImsiUsage, tapRd.CdrInfos.LOCAL_TIME[:8], tapRd.CdrInfos.IMSI_ID}
	//queryKey := strings.Fields(gQuryType_ImsiUsage+tapRd.CdrInfos.LOCAL_TIME[:8]+tapRd.CdrInfos.IMSI_ID)
	imsiCapBytes, err := jsonStruct.TapRecordUsageQuery(stub, queryKey, recordMemory)

	Log_add("TapRecordUsageQuery 조회")
	Log_add("err.Error() : ["+err.Error()+"]")

	if err.Error() == "wrong key data" {
		//no row selected
	}else if err != nil{
		Log_add("ImsiCap누적 조회 오류")
		return errors.New("ImsiCap누적 조회 오류")
	}else if imsiCapBytes != nil {
		err = json.Unmarshal(imsiCapBytes, stImsiUsage)
		if err != nil{
			Log_add("json Unmarshal error")
			return errors.New("json Unmarshal error")
		}
	}

	//비교를 위헤 string을 float64로 변환
	f64ImsiCapTHRMIN, err = strconv.ParseFloat(stCalcSpcl.THRSMIN, 64)
	if err != nil {
		return errors.New("f64ImsiCapTHRMIN : string to float64 conv error")
	}

	f64ImsiCapTHRMAX, err = strconv.ParseFloat(stCalcSpcl.THRSMAX, 64)
	if err != nil {
		return errors.New("f64ImsiCapTHRMAX : string to float64 conv error")
	}

	// 비교할 사용량 조회
	Log_add("tapRd.CdrInfos.CALL_TYPE_ID : ["+tapRd.CdrInfos.CALL_TYPE_ID+"]")
	if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocLocal{
		f64ImsiCapUseRoundedDurat=stImsiUsage.TapCal.MOCLocal.RoundedDuration
		if stCalcSpcl.THRSUNIT == gUnitMin {   //min이면 sec로 변환
			f64ImsiCapTHRMIN = f64ImsiCapTHRMIN * gVoiceUnit
			f64ImsiCapTHRMAX = f64ImsiCapTHRMAX * gVoiceUnit
		}
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocHome{
		f64ImsiCapUseRoundedDurat=stImsiUsage.TapCal.MOCHome.RoundedDuration
		if stCalcSpcl.THRSUNIT == gUnitMin {   //min이면 sec로 변환
			f64ImsiCapTHRMIN = f64ImsiCapTHRMIN * gVoiceUnit
			f64ImsiCapTHRMAX = f64ImsiCapTHRMAX * gVoiceUnit
		}
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocInt{
		f64ImsiCapUseRoundedDurat=stImsiUsage.TapCal.MOCInt.RoundedDuration
		if stCalcSpcl.THRSUNIT == gUnitMin {  //min이면 sec로 변환
			f64ImsiCapTHRMIN = f64ImsiCapTHRMIN * gVoiceUnit
			f64ImsiCapTHRMAX = f64ImsiCapTHRMAX * gVoiceUnit
		}
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMtc{
		f64ImsiCapUseRoundedDurat=stImsiUsage.TapCal.MOCInt.RoundedDuration
		if stCalcSpcl.THRSUNIT == gUnitMin {
			f64ImsiCapTHRMIN = f64ImsiCapTHRMIN * gVoiceUnit
			f64ImsiCapTHRMAX = f64ImsiCapTHRMAX * gVoiceUnit
		}
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMo{
		f64ImsiCapUseRoundedDurat=stImsiUsage.TapCal.SMSMO.RoundedDuration
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMt{
		f64ImsiCapUseRoundedDurat=stImsiUsage.TapCal.SMSMT.RoundedDuration
	}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeData{
		f64ImsiCapUseRoundedDurat=stImsiUsage.TapCal.GPRS.RoundedDuration
		if stCalcSpcl.THRSUNIT == gUnitMbyte {  //Mbyte이면 Kbyte로 변환
			f64ImsiCapTHRMIN = f64ImsiCapTHRMIN * gDataUnit
			f64ImsiCapTHRMAX = f64ImsiCapTHRMAX * gDataUnit
		}
	}

	//calculBaseRate(stCalcBas jsonStruct.Usage, tapRd jsonStruct.TapRecord) (flaot64, error)
	if f64ImsiCapUseRoundedDurat + tapRd.CdrInfos.CalculDetail.RoundedDuration > f64ImsiCapTHRMIN && f64ImsiCapUseRoundedDurat <= f64ImsiCapTHRMAX{
		Log_add("in if f64ImsiCapUseDuration + f64TotalCallDurat > f64ImsiCapTHRMIN && f64ImsiCapUseDuration <= f64ImsiCapTHRMAX")
		if f64ImsiCapUseRoundedDurat > f64ImsiCapTHRMIN {  //min에 걸치 호 처리
			Log_add("in if f64ImsiCapUseDuration > f64ImsiCapTHRMIN")
			f64MinBeforeDurat = tapRd.CdrInfos.CalculDetail.Duration-(tapRd.CdrInfos.CalculDetail.Duration+f64ImsiCapUseRoundedDurat-f64ImsiCapTHRMIN)
			f64MinNowDurat = tapRd.CdrInfos.CalculDetail.Duration-f64MinBeforeDurat

			//sMinBeforeDurat = strconv.FormatFloat(f64MinBeforeDurat, 'G', -1, 64)
			//sMinNowDurat = strconv.FormatFloat(f64MinNowDurat, 'G', -1, 64)

			f64MinBeforeCharge, f64MinBeforeTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, f64MinBeforeDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}
			f64MinNowCharge, f64MinNowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcBas, tapRd, f64MinNowDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}
			tapRd.CdrInfos.CalculDetail.SetCharge = f64MinBeforeCharge + f64MinNowCharge
			tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64MinBeforeTaxCharge + f64MinNowTaxCharge

		}else if f64ImsiCapUseRoundedDurat+tapRd.CdrInfos.CalculDetail.RoundedDuration > f64ImsiCapTHRMAX{  // max에 걸친 호 처리
			Log_add("in else if f64ImsiCapUseDuration+f64TotalCallDurat > f64ImsiCapTHRMAX")
			f64MaxNowDurat = tapRd.CdrInfos.CalculDetail.Duration-(tapRd.CdrInfos.CalculDetail.Duration+f64ImsiCapUseRoundedDurat-f64ImsiCapTHRMAX)
			f64MaxAfterDurat = tapRd.CdrInfos.CalculDetail.Duration-f64MaxNowDurat

			//sMaxNowDurat = strconv.FormatFloat(f64MaxNowDurat, 'G', -1, 64)
			//sMaxAfterDurat = strconv.FormatFloat(f64MaxAfterDurat, 'G', -1, 64)


			f64MaxNowCharge, f64MaxNowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, f64MaxNowDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}
			f64MaxAfterCharge, f64MaxAfterTaxCharge, err = calculBaseRate(gSpecialRate, stCalcBas, tapRd, f64MaxAfterDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}

			tapRd.CdrInfos.CalculDetail.SetCharge = f64MaxNowCharge + f64MaxAfterCharge
			tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64MaxNowTaxCharge + f64MaxAfterTaxCharge

		}else{  //특수 과금
			f64NowCharge, f64NowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, tapRd.CdrInfos.CalculDetail.Duration, f64TaxPercent)
			if err != nil{
				//에러처리
				Log_add(err.Error())
				return errors.New( err.Error())
			}

			tapRd.CdrInfos.CalculDetail.SetCharge = f64NowCharge
			tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64NowTaxCharge
			//return f64NowCharge, nil
		}
	}else{ // 정율 과금,,,
		Log_add("정율 과금")
		tapRd.CdrInfos.CalculDetail.SetCharge = tapRd.CdrInfos.CalculDetail.Charge
		tapRd.CdrInfos.CalculDetail.TAXSETCharge = tapRd.CdrInfos.CalculDetail.TAXCharge
		//return tapRd.CdrInfos.CalculDetail.Charge, nil
	}

	Log_add("Not matched anything")
	return errors.New( "Not matched anything")

}


/************************************************************************************************************/
// Commitment 모델, 금액 base, change rate 계산
/************************************************************************************************************/
func calcCommitMonetary(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stCalcSpcl jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord, f64TaxPercent float64, sContractId string) (error) {
	Log_add("======================function : calcCommitMonetary")
	var stTotalUsage jsonStruct.TotalUsage  //계약별 누적량 구조체
	var stCalcSpBas jsonStruct.CalcBas  //Agreement의 usage 요율 구조체

	var f64CommitUseAmount float64
	var f64NowCharge float64

	var f64CommitTHRMIN float64
	var f64CommitTHRMAX float64
	var f64NowTaxCharge float64

	//sc의 base 구조체 조회
	for i:=0;i<len(stCalcSpcl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stCalcSpcl.CalcBas[i].CALLTYPECD {
			stCalcSpBas = stCalcSpcl.CalcBas[i]
			break
		}
	}
	//계약별 누적데이터 조회 totalUsage+HPMN+VPMN+contract20190701 range
	queryKey := []string{gQuryType_Commitment, tapRd.Header.HPMN, tapRd.Header.VPMN, sContractId, tapRd.CdrInfos.LOCAL_TIME[:8]}
	commitmentBytes, err := jsonStruct.TapRecordUsageQuery(stub, queryKey, recordMemory)

	if err.Error() == "wrong key data" {
		//no row selected
	}else if err != nil{
		Log_add("계약 누적 조회 오류")
		return errors.New("계약 누적 조회 오류")
	}else if commitmentBytes != nil {
		err = json.Unmarshal(commitmentBytes, stTotalUsage)
		if err != nil{
			Log_add("json Unmarshal error")
			return errors.New("json Unmarshal error")
		}
	}

	Log_add(strconv.FormatFloat(f64CommitUseAmount,'g',-1,64))

	//commitment걸린 call type만큼 for문 돌아서 service type별로 계약누적에서 summary
	for i:=0;i<len(stCalcSpcl.CALLTYPECD);i++{
		if  stCalcSpcl.CALLTYPECD[i] == gCallTypeAll{
			f64CommitUseAmount = stTotalUsage.TapCal.MOCLocal.Charge+stTotalUsage.TapCal.MOCHome.Charge+stTotalUsage.TapCal.MOCInt.Charge+stTotalUsage.TapCal.MTC.Charge+stTotalUsage.TapCal.SMSMO.Charge+stTotalUsage.TapCal.SMSMT.Charge+stTotalUsage.TapCal.GPRS.Charge
		}else if stCalcSpcl.CALLTYPECD[i] == gCallTypeMocLocal{
			f64CommitUseAmount=f64CommitUseAmount+stTotalUsage.TapCal.MOCLocal.Charge
		}else if stCalcSpcl.CALLTYPECD[i] == gCallTypeMocHome{
			f64CommitUseAmount=f64CommitUseAmount+stTotalUsage.TapCal.MOCHome.Charge
		}else if stCalcSpcl.CALLTYPECD[i] == gCallTypeMocInt{
			f64CommitUseAmount=f64CommitUseAmount+stTotalUsage.TapCal.MOCInt.Charge
		}else if stCalcSpcl.CALLTYPECD[i] == gCallTypeMtc{
			f64CommitUseAmount=f64CommitUseAmount+stTotalUsage.TapCal.MTC.Charge
		}else if stCalcSpcl.CALLTYPECD[i] == gCallTypeSmsMo{
			f64CommitUseAmount=f64CommitUseAmount+stTotalUsage.TapCal.SMSMO.Charge
		}else if stCalcSpcl.CALLTYPECD[i] == gCallTypeSmsMt{
			f64CommitUseAmount=f64CommitUseAmount+stTotalUsage.TapCal.SMSMT.Charge
		}else if stCalcSpcl.CALLTYPECD[i] == gCallTypeData{
			f64CommitUseAmount=f64CommitUseAmount+stTotalUsage.TapCal.GPRS.Charge
		}
	}

	//비교를 위헤 string을 float64로 변환
	f64CommitTHRMIN, err = strconv.ParseFloat(stCalcSpcl.THRSMIN, 64)
	if err != nil {
		return errors.New("f64ImsiCapTHRMIN : string to float64 conv error")
	}

	f64CommitTHRMAX, err = strconv.ParseFloat(stCalcSpcl.THRSMAX, 64)
	if err != nil {
		return errors.New("f64ImsiCapTHRMAX : string to float64 conv error")
	}

	//calculBaseRate(stCalcBas jsonStruct.Usage, tapRd jsonStruct.TapRecord) (flaot64, error)
	if f64CommitUseAmount > f64CommitTHRMIN && f64CommitUseAmount <= f64CommitTHRMAX{
		//특수 과금
		f64NowCharge, f64NowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, tapRd.CdrInfos.CalculDetail.Duration, f64TaxPercent)
		if err != nil{
			//에러처리
			return errors.New( err.Error())
		}

		tapRd.CdrInfos.CalculDetail.SetCharge = f64NowCharge
		tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64NowTaxCharge
		//return f64NowCharge, nil
	}else{ // 정율 과금,,,
		tapRd.CdrInfos.CalculDetail.SetCharge = tapRd.CdrInfos.CalculDetail.Charge
		tapRd.CdrInfos.CalculDetail.TAXSETCharge = tapRd.CdrInfos.CalculDetail.TAXCharge
		Log_add("Commitment 정율과금")
	}

	return errors.New( "Not matched anything")
}


/************************************************************************************************************/
// Commitment 모델, 사용량 base, change rate 계산
/************************************************************************************************************/
func calcCommitDuration(stub shim.ChaincodeStubInterface, recordMemory jsonStruct.RecordMemory, stCalcBas jsonStruct.CalcBas, stCalcSpcl jsonStruct.CalcSpcl, tapRd *jsonStruct.TapRecord, f64TaxPercent float64, sContractId string) (error) {
	Log_add("======================function : calcCommitDuration")

	var stTotalUsage jsonStruct.TotalUsage  //계약별 누적량 구조체
	//var stBaseUsage jsonStruct.Usage  //Agreement의 usage 요율 구조체
	var stCalcSpBas jsonStruct.CalcBas  //Agreement의 usage 요율 구조체

	var f64CommitUseRoundedDurat float64

	var f64MinBeforeDurat float64
	var f64MinNowDurat float64
	var f64MaxNowDurat float64
	var f64MaxAfterDurat float64

	var f64MinBeforeCharge float64
	var f64MinNowCharge float64
	var f64MaxNowCharge float64
	var f64MaxAfterCharge float64
	var f64NowCharge float64


	var f64MinBeforeTaxCharge float64
	var f64MinNowTaxCharge float64
	var f64MaxNowTaxCharge float64
	var f64MaxAfterTaxCharge float64
	var f64NowTaxCharge float64

	var f64CommitTHRMIN float64
	var f64CommitTHRMAX float64
	//var f64TotalCallDurat float64

	//var sMinBeforeDurat string
	//var sMinNowDurat string
	//var sMaxNowDurat string
	//var sMaxAfterDurat string

	//sc의 base 구조체 조회
	for i:=0;i<len(stCalcSpcl.CalcBas);i++{
		if tapRd.CdrInfos.CALL_TYPE_ID == stCalcSpcl.CalcBas[i].CALLTYPECD {
			stCalcSpBas = stCalcSpcl.CalcBas[i]
			break
		}
	}

	//계약별 누적데이터 조회 totalUsage+HPMN+VPMN+contract20190701 range
	queryKey := []string{gQuryType_Commitment, tapRd.Header.HPMN, tapRd.Header.VPMN, sContractId, tapRd.CdrInfos.LOCAL_TIME[:8]}
	commitmentBytes, err := jsonStruct.TapRecordUsageQuery(stub, queryKey, recordMemory)


	Log_add("TapRecordUsageQuery 조회")
	Log_add("err.Error() : ["+err.Error()+"]")

	if err.Error() == "wrong key data" {
		//no row selected
	}else if err != nil{
		Log_add("계약별 누적 조회 오류")
		return errors.New("계약 누적 조회 오류")
	}else if commitmentBytes != nil {
		err = json.Unmarshal(commitmentBytes, stTotalUsage)
		if err != nil{
			Log_add("json Unmarshal error")
			return errors.New("json Unmarshal error")
		}
	}

	//비교를 위헤 string을 float64로 변환
	f64CommitTHRMIN, err = strconv.ParseFloat(stCalcSpcl.THRSMIN, 64)
	if err != nil {
		return errors.New("f64CommitTHRMIN : string to float64 conv error")
	}

	f64CommitTHRMAX, err = strconv.ParseFloat(stCalcSpcl.THRSMAX, 64)
	if err != nil {
		return errors.New("f64CommitTHRMAX : string to float64 conv error")
	}

	// Threshold를 기준 단위로 변환
	Log_add("tapRd.CdrInfos.CALL_TYPE_ID : ["+tapRd.CdrInfos.CALL_TYPE_ID+"]")

	for i:=0;i<len(stCalcSpcl.CALLTYPECD);i++{
		if  stCalcSpcl.CALLTYPECD[i] == gCallTypeAll {
			f64CommitUseRoundedDurat = stTotalUsage.TapCal.MOCLocal.RoundedDuration + stTotalUsage.TapCal.MOCHome.RoundedDuration + stTotalUsage.TapCal.MOCInt.RoundedDuration + stTotalUsage.TapCal.MTC.RoundedDuration + stTotalUsage.TapCal.SMSMO.RoundedDuration + stTotalUsage.TapCal.SMSMT.RoundedDuration + stTotalUsage.TapCal.GPRS.RoundedDuration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocLocal{
			f64CommitUseRoundedDurat=f64CommitUseRoundedDurat+stTotalUsage.TapCal.MOCLocal.RoundedDuration
			if stCalcSpcl.THRSUNIT == gUnitMin {   //min이면 sec로 변환
				f64CommitTHRMIN = f64CommitTHRMIN * gVoiceUnit
				f64CommitTHRMAX = f64CommitTHRMAX * gVoiceUnit
			}
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocHome{
			f64CommitUseRoundedDurat=stTotalUsage.TapCal.MOCHome.RoundedDuration
			if stCalcSpcl.THRSUNIT == gUnitMin {   //min이면 sec로 변환
				f64CommitTHRMIN = f64CommitTHRMIN * gVoiceUnit
				f64CommitTHRMAX = f64CommitTHRMAX * gVoiceUnit
			}
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMocInt{
			f64CommitUseRoundedDurat=stTotalUsage.TapCal.MOCInt.RoundedDuration
			if stCalcSpcl.THRSUNIT == gUnitMin {  //min이면 sec로 변환
				f64CommitTHRMIN = f64CommitTHRMIN * gVoiceUnit
				f64CommitTHRMAX = f64CommitTHRMAX * gVoiceUnit
			}
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeMtc{
			f64CommitUseRoundedDurat=stTotalUsage.TapCal.MOCInt.RoundedDuration
			if stCalcSpcl.THRSUNIT == gUnitMin {
				f64CommitTHRMIN = f64CommitTHRMIN * gVoiceUnit
				f64CommitTHRMAX = f64CommitTHRMAX * gVoiceUnit
			}
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMo{
			f64CommitUseRoundedDurat=stTotalUsage.TapCal.SMSMO.RoundedDuration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeSmsMt{
			f64CommitUseRoundedDurat=stTotalUsage.TapCal.SMSMT.RoundedDuration
		}else if tapRd.CdrInfos.CALL_TYPE_ID == gCallTypeData{
			f64CommitUseRoundedDurat=stTotalUsage.TapCal.GPRS.RoundedDuration
			if stCalcSpcl.THRSUNIT == gUnitMbyte {  //Mbyte이면 Kbyte로 변환
				f64CommitTHRMIN = f64CommitTHRMIN * gDataUnit
				f64CommitTHRMAX = f64CommitTHRMAX * gDataUnit
			}
		}
	}

	Log_add("f64CommitTHRMIN : " + strconv.FormatFloat(f64CommitTHRMIN, 'g',-1,64))
	Log_add("f64CommitTHRMAM : " + strconv.FormatFloat(f64CommitTHRMAX, 'g',-1,64))
	Log_add("f64CommitUseRoundedDurat : " + strconv.FormatFloat(f64CommitUseRoundedDurat, 'g',-1,64))
	Log_add("tapRd.CdrInfos.CalculDetail.RoundedDuration : " + strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.RoundedDuration, 'g',-1,64))

	//calculBaseRate(stCalcBas jsonStruct.Usage, tapRd jsonStruct.TapRecord) (flaot64, error)
	if f64CommitUseRoundedDurat + tapRd.CdrInfos.CalculDetail.RoundedDuration > f64CommitTHRMIN && f64CommitUseRoundedDurat <= f64CommitTHRMAX{
		Log_add("in if f64ImsiCapUseDuration + f64TotalCallDurat > f64ImsiCapTHRMIN && f64ImsiCapUseDuration <= f64ImsiCapTHRMAX")
		if f64CommitUseRoundedDurat > f64CommitTHRMIN {  //min에 걸치 호 처리
			Log_add("in if f64CommitUseRoundedDurat > f64CommitTHRMIN")
			f64MinBeforeDurat = tapRd.CdrInfos.CalculDetail.Duration-(tapRd.CdrInfos.CalculDetail.Duration+f64CommitUseRoundedDurat-f64CommitTHRMIN)
			f64MinNowDurat = tapRd.CdrInfos.CalculDetail.Duration-f64MinBeforeDurat

			Log_add("f64MinBeforeDurat : " + strconv.FormatFloat(f64MinBeforeDurat, 'g',-1,64))
			Log_add("f64MinNowDurat : " + strconv.FormatFloat(f64MinNowDurat, 'g',-1,64))

			f64MinBeforeCharge, f64MinBeforeTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, f64MinBeforeDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}
			f64MinNowCharge, f64MinNowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcBas, tapRd, f64MinNowDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}

			Log_add("f64MinBeforeCharge : " + strconv.FormatFloat(f64MinBeforeCharge, 'g',-1,64))
			Log_add("f64MinBeforeTaxCharge : " + strconv.FormatFloat(f64MinBeforeTaxCharge, 'g',-1,64))
			Log_add("f64MinNowCharge : " + strconv.FormatFloat(f64MinNowCharge, 'g',-1,64))
			Log_add("f64MinNowTaxCharge : " + strconv.FormatFloat(f64MinNowTaxCharge, 'g',-1,64))


			tapRd.CdrInfos.CalculDetail.SetCharge = f64MinBeforeCharge + f64MinNowCharge
			tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64MinBeforeTaxCharge + f64MinNowTaxCharge

		}else if f64CommitUseRoundedDurat+tapRd.CdrInfos.CalculDetail.RoundedDuration > f64CommitTHRMAX{  // max에 걸친 호 처리
			Log_add("in else if f64ImsiCapUseDuration+f64TotalCallDurat > f64ImsiCapTHRMAX")
			f64MaxNowDurat = tapRd.CdrInfos.CalculDetail.Duration-(tapRd.CdrInfos.CalculDetail.Duration+f64CommitUseRoundedDurat-f64CommitTHRMAX)
			f64MaxAfterDurat = tapRd.CdrInfos.CalculDetail.Duration-f64MaxNowDurat

			Log_add("f64MaxNowDurat : " + strconv.FormatFloat(f64MaxNowDurat, 'g',-1,64))
			Log_add("f64MaxAfterDurat : " + strconv.FormatFloat(f64MaxAfterDurat, 'g',-1,64))

			f64MaxNowCharge, f64MaxNowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, f64MaxNowDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}
			f64MaxAfterCharge, f64MaxAfterTaxCharge, err = calculBaseRate(gSpecialRate, stCalcBas, tapRd, f64MaxAfterDurat, f64TaxPercent)
			if err != nil{
				//에러처리
				return errors.New( err.Error())
			}

			Log_add("f64MaxNowCharge : " + strconv.FormatFloat(f64MaxNowCharge, 'g',-1,64))
			Log_add("f64MaxNowTaxCharge : " + strconv.FormatFloat(f64MaxNowTaxCharge, 'g',-1,64))
			Log_add("f64MaxAfterCharge : " + strconv.FormatFloat(f64MaxAfterCharge, 'g',-1,64))
			Log_add("f64MaxAfterTaxCharge : " + strconv.FormatFloat(f64MaxAfterTaxCharge, 'g',-1,64))

			tapRd.CdrInfos.CalculDetail.SetCharge = f64MaxNowCharge + f64MaxAfterCharge
			tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64MaxNowTaxCharge + f64MaxAfterTaxCharge

		}else{  //특수 과금
		    Log_add("특수과금")
			f64NowCharge, f64NowTaxCharge, err = calculBaseRate(gSpecialRate, stCalcSpBas, tapRd, tapRd.CdrInfos.CalculDetail.Duration, f64TaxPercent)
			if err != nil{
				//에러처리
				Log_add(err.Error())
				return errors.New( err.Error())
			}

			Log_add("f64NowCharge : " + strconv.FormatFloat(f64NowCharge, 'g',-1,64))
			Log_add("f64NowTaxCharge : " + strconv.FormatFloat(f64NowTaxCharge, 'g',-1,64))

			tapRd.CdrInfos.CalculDetail.SetCharge = f64NowCharge
			tapRd.CdrInfos.CalculDetail.TAXSETCharge = f64NowTaxCharge
		}
	}else{ // 정율 과금,,,
		Log_add("정율 과금")
		Log_add("tapRd.CdrInfos.CalculDetail.Charge : " + strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.Charge, 'g',-1,64))
		Log_add("tapRd.CdrInfos.CalculDetail.TAXCharge : " + strconv.FormatFloat(tapRd.CdrInfos.CalculDetail.TAXCharge, 'g',-1,64))

		tapRd.CdrInfos.CalculDetail.SetCharge = tapRd.CdrInfos.CalculDetail.Charge
		tapRd.CdrInfos.CalculDetail.TAXSETCharge = tapRd.CdrInfos.CalculDetail.TAXCharge
		//return tapRd.CdrInfos.CalculDetail.Charge, nil
	}

	Log_add("Not matched anything")
	return errors.New( "Not matched anything")

}

