package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//c "github.com/main/go/controller"
	c "../controller"
	//"github.com/main/go/jsonStruct"
	"../jsonStruct"
)

//commitment 모델 확인해서 추가 정산금액 확인 function
func InvoiceCalcCommitment(stub shim.ChaincodeStubInterface, invoice *jsonStruct.Invoice) (error){
	/*
			type InvoiceInfo struct {
				InvoicedAmount       InvoicedAmount `json:"InvoicedAmount"`
				TapCal             	 TapCal   		`json:"TapCal"`
				AgreeSubCharge       string         `json:"AgreeSubCharge"`
				CommitmentTotalRate  string         `json:"Commitment_totalRate"`
				CommitmentPeriodRate string         `json:"Commitment_periodRate"`
			}

		   1. acitive계약 조회하여 계약 기간의 마지막 달인지 확인
	       2. 확인해서 마지막 달이 아니면 AgreeSubCharge는 0
	       3. 마지막 달이면 AgreeSubCharge 계산
	           계산방식은 사업자별로 상이,,,,
	       4. CommitmentTotalRate 계산
	       5. CommitmentPeriodRate 계산
	 */

	//active 계약 조회
	var sActContract jsonStruct.ContractForCal
	var sNowDate string
	var err error
	var sCtrtStDate string
	var sCtrtEdDate string
	var stInvoice jsonStruct.Invoice
	var f64TotalPostTaxAmount float64
	var stCalcSpcl []jsonStruct.CalcSpcl
	var bCommitmentFlag bool
	var bAllServicesFlag bool
	var bMocLocalFlag bool
	var bMocHomeFlag bool
	var bMocIntFlag bool
	var bMTCFlag bool
	var bSmsMoFlag bool
	var bSmsMtFlag bool
	var bGPRSFlag bool
	var bFixedChargeFlag bool
	var bSpecialRuleFlag bool
	var bIsMonetary bool
	var f64CommitTHRMIN float64
	var f64CommitTHRMAX float64
	var f64TotContPeriodCnt float64
	var f64NowContPeriodCnt float64
	var f64VoiceTotalDuration float64
	var f64VoiceTotalRoundedDuration float64
	var f64FixAmt float64
//	var sFixAmtCurCd string
	var f64ModiFactor float64

	bCommitmentFlag = false
	bAllServicesFlag = false
	bMocLocalFlag = false
	bMocHomeFlag = false
	bMocIntFlag = false
	bMTCFlag = false
	bSmsMoFlag = false
	bSmsMtFlag = false
	bGPRSFlag = false
	bFixedChargeFlag = false
	bSpecialRuleFlag = false
	bIsMonetary = false


	Log_add("invoice.InvoiceDate : "+invoice.InvoiceDate)
	sNowDate = invoice.InvoiceDate+"01" //invoice사용월에 01을 붙임


	Log_add("sNowDate : "+ sNowDate)
	Log_add("invoice.VPMN : "+ invoice.VPMN)
	Log_add("invoice.HPMN : "+ invoice.HPMN)

	sActContract, err = Contract_getActive(stub, sNowDate, invoice.VPMN, invoice.HPMN)
	if err != nil{
		Log_add("Agreement_getActive 조회오류")
	}

	Log_add("sActContract.ContractInfo.Past.ContDtl.CONTSTDATE : "+sActContract.ContractInfo.Past.ContDtl.CONTSTDATE)
	Log_add("sActContract.ContractInfo.Past.ContDtl.CONTEXPDATE : "+sActContract.ContractInfo.Past.ContDtl.CONTEXPDATE)
	Log_add("sActContract.ContractInfo.Current.ContDtl.CONTSTDATE : "+sActContract.ContractInfo.Current.ContDtl.CONTSTDATE)
	Log_add("sActContract.ContractInfo.Current.ContDtl.CONTEXPDATE : "+sActContract.ContractInfo.Current.ContDtl.CONTEXPDATE)


	//계약 시작일자
	if sActContract.ContractInfo.Past.ContDtl.CONTSTDATE == gNotExgtValue || sActContract.ContractInfo.Past.ContDtl.CONTSTDATE =="" {  //"null"
		sCtrtStDate = sActContract.ContractInfo.Current.ContDtl.CONTSTDATE
	}else{
		if sActContract.ContractInfo.Past.ContDtl.CONTSTDATE < sActContract.ContractInfo.Current.ContDtl.CONTSTDATE {
			sCtrtStDate = sActContract.ContractInfo.Past.ContDtl.CONTSTDATE
		}else{
			sCtrtStDate = sActContract.ContractInfo.Current.ContDtl.CONTSTDATE
		}
	}

	//계약 종료일자
	if sActContract.ContractInfo.Past.ContDtl.CONTEXPDATE == gNotExgtValue || sActContract.ContractInfo.Past.ContDtl.CONTEXPDATE ==""{  //"null"
		sCtrtEdDate = sActContract.ContractInfo.Current.ContDtl.CONTEXPDATE
	}else{
		if sActContract.ContractInfo.Past.ContDtl.CONTEXPDATE < sActContract.ContractInfo.Current.ContDtl.CONTEXPDATE {
			sCtrtEdDate = sActContract.ContractInfo.Past.ContDtl.CONTEXPDATE
		}else{
			sCtrtEdDate = sActContract.ContractInfo.Current.ContDtl.CONTEXPDATE
		}
	}

	Log_add("sCtrtStDate : ["+sCtrtStDate+"], sCtrtEdDate : ["+sCtrtEdDate+"]")

	stCalcSpcl = sActContract.ContractInfo.Current.ContDtl.CalcSpcl

	//commitment check
	for i:=0;i<len(stCalcSpcl);i++{
		if stCalcSpcl[i].MODELTYPECD == gModelTypeCommit{ //commitment
			bCommitmentFlag = true

			f64CommitTHRMIN, err = strconv.ParseFloat(stCalcSpcl[i].THRSMIN, 64)
			if err != nil {
				return errors.New("f64ImsiCapTHRMIN : string to float64 conv error")
			}

			f64CommitTHRMAX, err = strconv.ParseFloat(stCalcSpcl[i].THRSMAX, 64)
			if err != nil {
				return errors.New("f64ImsiCapTHRMAX : string to float64 conv error")
			}

			if stCalcSpcl[i].APLYTYPE == gApplyTypeFixedChrg {
				bFixedChargeFlag = true
				f64FixAmt, err = strconv.ParseFloat(stCalcSpcl[i].FIXAMT,64)
				if err != nil{
					return errors.New("f64FixAmt : parseFloat Error")
				}
				//sFixAmtCurCd = stCalcSpcl[i].FIXAMTCURCD

			}else if stCalcSpcl[i].APLYTYPE == gApplyTypeSpRule{
				bSpecialRuleFlag = true
			}

			if stCalcSpcl[i].THRSUNIT == gUnitKbyte || stCalcSpcl[i].THRSUNIT == gUnitMbyte ||
				stCalcSpcl[i].THRSUNIT == gUnitSec  || stCalcSpcl[i].THRSUNIT == gUnitMin {
				//사용량 base check,,,,
				bIsMonetary = false
				Log_add("bIsMonetary = false")
				if stCalcSpcl[i].THRSUNIT == gUnitMin{
					f64CommitTHRMIN = f64CommitTHRMIN *gVoiceUnit //min을 sec로 변경
					f64CommitTHRMAX = f64CommitTHRMAX *gVoiceUnit //min을 sec로 변경
				}

				if stCalcSpcl[i].THRSUNIT == gUnitMbyte{
					f64CommitTHRMIN = f64CommitTHRMIN *gDataUnit //MB를 KB로 변경
					f64CommitTHRMAX = f64CommitTHRMAX *gVoiceUnit //MB를 KB로 변경
				}
			}else{
				//금액 base check,,,,
				bIsMonetary = true
				Log_add("bIsMonetary = true")
			}


			for j:=0;j<len(stCalcSpcl[i].CALLTYPECD);j++{
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeAll{
					bAllServicesFlag = true
					break
				}
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeMocLocal{
					bMocLocalFlag = true
				}
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeMocHome{
					bMocHomeFlag = true
				}
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeMocInt{
					bMocIntFlag = true
				}
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeMtc{
					bMTCFlag = true
				}
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeSmsMo{
					bSmsMoFlag = true
				}
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeSmsMt{
					bSmsMtFlag = true
				}
				if stCalcSpcl[i].CALLTYPECD[j] == gCallTypeData{
					bGPRSFlag = true
				}
			}
		}
	}


	// invoice발행 내역 조회
  	/*
 		발행 대상 월은 인자값(invoce)로 넘어옴
 		계약시작월~ 발행 대상 월 이전월 까지 summary + invoice 해서 금액을 만들고
 		누적 달성률, 기간 달성률을 계산
  	 */
	layout := "200601"
	tInvoiceUseDate, _ := time.Parse(layout, invoice.UsedDate)
	sInvoicePreDate := tInvoiceUseDate.AddDate(0,-1,0).Format(layout) //yyyymm

	Log_add("sInvoicePreDate : " + sInvoicePreDate)

	queryMonthList := c.GetBetweenMonth(sCtrtStDate[:6], sInvoicePreDate)

	f64TotalPostTaxAmount = 0

	for i:=0;i<len(queryMonthList);i++{
		queryKey := []string{jsonStruct.Invoice_TYPE, invoice.VPMN, queryMonthList[i]}
		invoiceBytes, err := Block_Query(stub, queryKey)

		if invoiceBytes != nil {
			err = json.Unmarshal(invoiceBytes[0], &stInvoice)
			if err != nil{
				Log_add("json Unmarshal error")
				return errors.New("json Unmarshal error")
			}

			if bAllServicesFlag == true {
				f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.InvoicedAmount.PostTaxAmount
				f64VoiceTotalDuration = 0 //unit이 다르기 때문에 해당 케이스는 존재 불가함...
				f64VoiceTotalRoundedDuration = 0 //unit이 다르기 때문에 해당 케이스는 존재 불가함...
			}else{
				if bMocLocalFlag == true{
					f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.TapCal.MOCLocal.SetCharge + stInvoice.InvoiceInfo.TapCal.MOCLocal.TAXSETCharge
					if stInvoice.InvoiceInfo.TapCal.MOCLocal.Unit == gUnitMin {
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MOCLocal.Duration*gVoiceUnit
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MOCLocal.RoundedDuration*gVoiceUnit
					} else{
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MOCLocal.Duration
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MOCLocal.RoundedDuration
					}
				}
				if bMocHomeFlag == true{
					f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.TapCal.MOCHome.SetCharge + stInvoice.InvoiceInfo.TapCal.MOCHome.TAXSETCharge
					if stInvoice.InvoiceInfo.TapCal.MOCHome.Unit == gUnitMin {
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MOCHome.Duration*gVoiceUnit
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MOCHome.RoundedDuration*gVoiceUnit
					} else{
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MOCHome.Duration
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MOCHome.RoundedDuration
					}
				}
				if bMocIntFlag == true{
					f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.TapCal.MOCInt.SetCharge + stInvoice.InvoiceInfo.TapCal.MOCInt.TAXSETCharge
					if stInvoice.InvoiceInfo.TapCal.MOCInt.Unit == gUnitMin {
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MOCInt.Duration*gVoiceUnit
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MOCInt.RoundedDuration*gVoiceUnit
					} else{
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MOCInt.Duration
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MOCInt.RoundedDuration
					}
				}
				if bMTCFlag == true{
					f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.TapCal.MTC.SetCharge + stInvoice.InvoiceInfo.TapCal.MTC.TAXSETCharge
					if stInvoice.InvoiceInfo.TapCal.MTC.Unit == gUnitMin {
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MTC.Duration*gVoiceUnit
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MTC.RoundedDuration*gVoiceUnit
					} else{
						f64VoiceTotalDuration = f64VoiceTotalDuration + stInvoice.InvoiceInfo.TapCal.MTC.Duration
						f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + stInvoice.InvoiceInfo.TapCal.MTC.RoundedDuration
					}
				}
				if bSmsMoFlag == true{
					f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.TapCal.SMSMO.SetCharge + stInvoice.InvoiceInfo.TapCal.SMSMO.TAXSETCharge
				}
				if bSmsMtFlag == true{
					f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.TapCal.SMSMT.SetCharge + stInvoice.InvoiceInfo.TapCal.SMSMT.TAXSETCharge
				}
				if bGPRSFlag == true{
					f64TotalPostTaxAmount = f64TotalPostTaxAmount + stInvoice.InvoiceInfo.TapCal.GPRS.SetCharge + stInvoice.InvoiceInfo.TapCal.GPRS.TAXSETCharge
				}
			}
		}else if err != nil && err.Error() == "wrong key data" {
			//no row selected
			Log_add("no row selected")
		}else if err != nil{
			Log_add("Invoice 조회 오류 : "+err.Error())
			return errors.New("Invoice 조회 오류")
		}
	}

	//현재 발행될 invoice 금액을 sum
	if bAllServicesFlag == true {
		f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.InvoicedAmount.PostTaxAmount
	}else{
		if bMocLocalFlag == true{
			f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.TapCal.MOCLocal.SetCharge + invoice.InvoiceInfo.TapCal.MOCLocal.TAXSETCharge
			if invoice.InvoiceInfo.TapCal.MOCLocal.Unit == gUnitMin {
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MOCLocal.Duration*gVoiceUnit
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MOCLocal.RoundedDuration*gVoiceUnit
			} else{
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MOCLocal.Duration
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MOCLocal.RoundedDuration
			}
		}
		if bMocHomeFlag == true{
			f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.TapCal.MOCHome.SetCharge + invoice.InvoiceInfo.TapCal.MOCHome.TAXSETCharge
			if invoice.InvoiceInfo.TapCal.MOCHome.Unit == gUnitMin {
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MOCHome.Duration*gVoiceUnit
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MOCHome.RoundedDuration*gVoiceUnit
			} else{
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MOCHome.Duration
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MOCHome.RoundedDuration
			}
		}
		if bMocIntFlag == true{
			f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.TapCal.MOCInt.SetCharge + invoice.InvoiceInfo.TapCal.MOCInt.TAXSETCharge
			if invoice.InvoiceInfo.TapCal.MOCInt.Unit == gUnitMin {
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MOCInt.Duration*gVoiceUnit
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MOCInt.RoundedDuration*gVoiceUnit
			} else{
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MOCInt.Duration
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MOCInt.RoundedDuration
			}
		}
		if bMTCFlag == true{
			f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.TapCal.MTC.SetCharge + invoice.InvoiceInfo.TapCal.MTC.TAXSETCharge
			if invoice.InvoiceInfo.TapCal.MTC.Unit == gUnitMin {
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MTC.Duration*gVoiceUnit
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MTC.RoundedDuration*gVoiceUnit
			} else{
				f64VoiceTotalDuration = f64VoiceTotalDuration + invoice.InvoiceInfo.TapCal.MTC.Duration
				f64VoiceTotalRoundedDuration = f64VoiceTotalRoundedDuration + invoice.InvoiceInfo.TapCal.MTC.RoundedDuration
			}
		}
		if bSmsMoFlag == true{
			f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.TapCal.SMSMO.SetCharge + invoice.InvoiceInfo.TapCal.SMSMO.TAXSETCharge
		}
		if bSmsMtFlag == true{
			f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.TapCal.SMSMT.SetCharge + invoice.InvoiceInfo.TapCal.SMSMT.TAXSETCharge
		}
		if bGPRSFlag == true{
			f64TotalPostTaxAmount = f64TotalPostTaxAmount + invoice.InvoiceInfo.TapCal.GPRS.SetCharge + invoice.InvoiceInfo.TapCal.GPRS.TAXSETCharge
		}
	}

    //계약에 commitment가 있으면 commitment달성률 계산 아니면 0
	if bCommitmentFlag == true {
		Log_add("IN bCommitmentFlag == true")
		//계약의 마지막 달인지 체크 ( invoice사용월 == 계약 current의 종료월 )
		if invoice.InvoiceDate == sCtrtEdDate[:6] {
			Log_add("IN invoice.InvoiceDate == sCtrtEdDate[:6]")
			//AgreeSubCharge 계산
			//fixed charge, special rule,,,,적용
			//hongkong 전체 traffic에 대한 commitment적용 ( 0~ 900,000USD이면 정산금액이 900,000USD)
			if bFixedChargeFlag == true && bIsMonetary == true{
				if f64CommitTHRMIN > f64TotalPostTaxAmount && f64TotalPostTaxAmount <= f64CommitTHRMAX{
					invoice.InvoiceInfo.AgreeSubCharge = strconv.FormatFloat(c.RoundOff(f64FixAmt - f64TotalPostTaxAmount,6),'g',-1,64)
				}else{
					invoice.InvoiceInfo.AgreeSubCharge = "0"
				}
			}

			//China  3710000분 미만시 사용금액에 보정계수 곱한 금액
			if bSpecialRuleFlag == true && bIsMonetary == false{
				if f64CommitTHRMIN < f64VoiceTotalDuration && f64VoiceTotalDuration <= f64CommitTHRMAX{
					f64ModiFactor = f64CommitTHRMAX/f64VoiceTotalDuration
					invoice.InvoiceInfo.AgreeSubCharge = strconv.FormatFloat(c.RoundOff(f64VoiceTotalDuration*f64ModiFactor-f64VoiceTotalDuration,6),'g',-1,64)
				}else{
					invoice.InvoiceInfo.AgreeSubCharge = "0"
				}
			}
		}else{
			invoice.InvoiceInfo.AgreeSubCharge = "0"
		}

		// CommitmentTotalRate 계산
		/*
		  누적 달성률(%) : 현재까지 누적/commitment량 * 100
		  ex) 1년 commitment가 12만원이고, 4월 traffic까지 금액 합계가 4만원 일 경우
		     누적 달성률 : 33.3% = 4만/12만*100(소수점 둘째자리에서 반올림하여 첫째자리까지 표현)

		     commitment 금액 확인   --> contract
		     현재까지 정산 금액 확인 --> 계약별 누적량 조회하여 CalMonth기준으로 summary?
		*/
		invoice.InvoiceInfo.CommitmentTotalRate = strconv.FormatFloat(c.RoundOff(f64TotalPostTaxAmount/f64CommitTHRMIN*100,1),'g',-1,64)

		// CommitmentPeriodRate 계산
		/*
		  기간 달성률(%) : 현재까지 누적/((commitment량/계약개월 수)*누적개월 수)) * 100
		  ex) 1년 commitment가 12만원이고, 4월 traffic까지 금액 합계가 4만원 일 경우
		     기간 달성률 : 100% = 4만/((12만/12개월)*4)*100(소수점 둘째자리에서 반올림하여 첫째자리까지 표현)
		*/

		f64TotContPeriodCnt = float64(len(c.GetBetweenMonth(sCtrtStDate, sCtrtEdDate)))
		f64NowContPeriodCnt = float64(len(c.GetBetweenMonth(sCtrtStDate, invoice.UsedDate[:6])))

		invoice.InvoiceInfo.CommitmentPeriodRate = strconv.FormatFloat(c.RoundOff(f64TotalPostTaxAmount/((f64CommitTHRMIN/f64TotContPeriodCnt)*f64NowContPeriodCnt)*100,1),'g',-1,64)
	}else{
		Log_add("ELSE bCommitmentFlag == true")
		invoice.InvoiceInfo.AgreeSubCharge = "0"
		invoice.InvoiceInfo.CommitmentTotalRate = "0"
		invoice.InvoiceInfo.CommitmentPeriodRate = "0"
	}

	return nil
}
