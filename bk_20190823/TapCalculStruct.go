package jsonStruct

import (
	"encoding/json"
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"reflect"
)

const RecordUsage_Type = "recordUsage"
const ImsiUsage_Type = "imsiUsage"
const DateUsage_Type = "dateUsage"
const TotalUsage_Type = "totalUsage"
const TapfileUsage_Type = "tapfileUsage"


var RecordUsage_Key = []string{"Date","TAP_File_Name","RecordID"}
var ImsiUsage_Key = []string{"Date","IMSI"}
var DateUsage_Key = []string{"HPMN","VPMN","Date"}
var TotalUsage_Key = []string{"HPMN","VPMN","ContractD"}
var TapfileUsage_Key = []string{"Date","TAP_File_Name"}

type Usager interface{
	// 블록 구조에 맞는 Key string array 반환
	GetKeyArray() []string

	// 블록 구조에 맞는 Key 구조 반환
	GetBlockKey(stub shim.ChaincodeStubInterface) (string, error)

	// TapCal 값 변경
	SetTapCal(cal TapCal)

	//해당 struct 구조의 블럭을 Add
	AddBlock(shim.ChaincodeStubInterface) error

	//인터페이스 Type 반환
	Type() string
}


// *********************************************************
// *******************   RecordUsage  **********************
// *********************************************************
//Tap Record별 사용량 요약 정보
type RecordUsage struct {
	Date				string		`json:"Date""`
	FileCreationDate 	string 		`json:"FileCreationDate"`
	TapFileName		 	string 		`json:"TapFileName"`
	RecordID		 	string 		`json:"RecordID"`
	InvoiceID        	string 		`json:"InvoiceID"`
	CalMonth         	string 		`json:"CalMonth"`
	Currency		 	string		`json:"Currency"`
	TapCal			 	TapCal		`json:"TapCal"`
}

func (r RecordUsage) GetKeyArray() []string{
	return []string{RecordUsage_Type, r.Date, r.TapFileName, r.RecordID}
}

func (r *RecordUsage) SetTapCal(cal TapCal){
	r.TapCal = cal
}

func (r RecordUsage) GetBlockKey(stub shim.ChaincodeStubInterface) (string, error){
	return stub.CreateCompositeKey(RecordUsage_Type, []string{r.Date, r.TapFileName, r.RecordID})
}


func (r RecordUsage) AddBlock(stub shim.ChaincodeStubInterface) error{
	key, _ := stub.CreateCompositeKey(RecordUsage_Type, r.GetKeyArray()[1:])
	asBytes, _ := json.Marshal(r)
	err := Block_Insert(stub, key, asBytes)
	return err
}


func (r RecordUsage) Type() string{
	return RecordUsage_Type
}


// *********************************************************
// *********************   ImsiUsage  **********************
// *********************************************************
//임시 별 사용량 요약 정보
type ImsiUsage struct {
	Date             	string		`json:"Date"`
	IMSI             	string 		`json:"IMSI"`
	HPMN             	string 		`json:"HPMN"`
	VPMN             	string 		`json:"VPMN"`
	InvoiceID        	string 		`json:"InvoiceID"`
	CalMonth         	string 		`json:"CalMonth"`
	TapCal			 	TapCal		`json:"TapCal"`
}

func (i ImsiUsage) GetKeyArray() []string{
	return []string{ImsiUsage_Type, i.Date, i.IMSI}
}

func (i *ImsiUsage) SetTapCal(cal TapCal){
	i.TapCal = cal
}

func (i ImsiUsage) GetBlockKey(stub shim.ChaincodeStubInterface) (string, error){
	return stub.CreateCompositeKey(ImsiUsage_Type, []string{i.Date, i.IMSI})
}

func (i ImsiUsage) AddBlock(stub shim.ChaincodeStubInterface) error{
	key, _ := stub.CreateCompositeKey(ImsiUsage_Type, i.GetKeyArray()[1:])
	asBytes, _ := json.Marshal(i)
	err := Block_Insert(stub, key, asBytes)
	return err
}

func (r ImsiUsage) Type() string{
	return ImsiUsage_Type
}

// *********************************************************
// *********************   DateUsage  **********************
// *********************************************************
//날짜 별 사용량 요약 정보
type DateUsage struct {
	VPMN             	string 		`json:"VPMN"`
	HPMN             	string 		`json:"HPMN"`
	Date             	string		`json:"Date"`
	InvoiceID        	string 		`json:"InvoiceID"`
	CalMonth         	string 		`json:"CalMonth"`
	Currency       	 	string 		`json:"Currency"`
	FileSequenceNumbers	[2]string	`json:"FileSequenceNumbers"`
	TapCal			 	TapCal		`json:"TapCal"`
}
func (d DateUsage) GetKeyArray() []string{
	return []string{DateUsage_Type, d.VPMN, d.HPMN, d.Date}
}

func (r *DateUsage) SetTapCal(cal TapCal){
	r.TapCal = cal
}

func (d DateUsage) GetBlockKey(stub shim.ChaincodeStubInterface) (string, error){
	return stub.CreateCompositeKey(DateUsage_Type, []string{d.VPMN, d.HPMN, d.Date})
}

func (d DateUsage) AddBlock(stub shim.ChaincodeStubInterface) error{
	key, _ := stub.CreateCompositeKey(DateUsage_Type, d.GetKeyArray()[1:])
	asBytes, _ := json.Marshal(d)
	err := Block_Insert(stub, key, asBytes)
	return err
}

func (d DateUsage) Type() string{
	return DateUsage_Type
}

// *********************************************************
// *********************   TotalUsage  *********************
// *********************************************************
//계약 기간 별 사용량 요약 정보
type TotalUsage struct {
	HPMN             	string 		`json:"HPMN"`
	VPMN             	string 		`json:"VPMN"`
	ContractID      	string 		`json:"ContractID"`
	Peoriod          	[2]string	`json:"Peoriod"`
	TapCal			 	TapCal		`json:"TapCal"`
}

func (t TotalUsage) GetKeyArray() []string{
	return []string{TotalUsage_Type, t.VPMN, t.HPMN, t.ContractID}
}

func (r *TotalUsage) SetTapCal(cal TapCal){
	r.TapCal = cal
}

func (t TotalUsage) GetBlockKey(stub shim.ChaincodeStubInterface) (string, error){
	return stub.CreateCompositeKey(TotalUsage_Type, []string{t.VPMN, t.HPMN, t.ContractID})
}

func (t TotalUsage) AddBlock(stub shim.ChaincodeStubInterface) error{
	key, _ := stub.CreateCompositeKey(TotalUsage_Type, t.GetKeyArray()[1:])
	asBytes, _ := json.Marshal(t)
	err := Block_Insert(stub, key, asBytes)
	return err
}

func (t TotalUsage) Type() string{
	return TotalUsage_Type
}

// *********************************************************
// ********************   TapfileUsage  ********************
// *********************************************************
//Tap file별 사용량 요약 정보
type TapfileUsage struct {
	Date             	string		`json:"Date"`
	FileCreationDate 	string 		`json:"FileCreationDate"`
	TapFileName		 	string 		`json:"TapFileName"`
	HPMN             	string 		`json:"HPMN"`
	VPMN             	string 		`json:"VPMN"`
	ContractID      	string 		`json:"ContractID"`
	Peoriod          	[2]string	`json:"Peoriod"`
	InvoiceID        	string 		`json:"InvoiceID"`
	CalMonth         	string 		`json:"CalMonth"`
	Currency       	 	string 		`json:"Currency"`
	TapCal			 	TapCal		`json:"TapCal"`
}

func (t TapfileUsage) GetKeyArray() []string{
	return []string{TapfileUsage_Type, t.Date, t.TapFileName}
}

func (r *TapfileUsage) SetTapCal(cal TapCal){
	r.TapCal = cal
}

func (t TapfileUsage) GetBlockKey(stub shim.ChaincodeStubInterface) (string, error){
	return stub.CreateCompositeKey(TapfileUsage_Type, []string{t.Date, t.TapFileName})
}

func (t TapfileUsage) AddBlock(stub shim.ChaincodeStubInterface) error{
	key, _ := stub.CreateCompositeKey(TapfileUsage_Type, t.GetKeyArray()[1:])
	asBytes, _ := json.Marshal(t)
	err := Block_Insert(stub, key, asBytes)
	return err
}

func (t TapfileUsage) Type() string{
	return TapfileUsage_Type
}

// *********************************************************
// *******************   RecordMemory  *********************
// *********************************************************
type RecordMemory struct{
	RecordUsageList 	[]RecordUsage 	`json:"RecordUsageList"`
	ImsiUsageList 		[]ImsiUsage 	`json:"ImsiUsageList"`
	DateUsage 			DateUsage 		`json:"DateUsageList"`
	TotalUsage 			TotalUsage 		`json:"TotalUsageList"`
	TapfileUsage 		TapfileUsage 	`json:"TapfileUsageList"`
}

// Imsi값이 있는지 확인
// -1일시 찾은 값 없음.
// 그 외 array index 번호 반환
func (r RecordMemory) IsInImsi(imsi string) int{
	for i:=0; i< len(r.ImsiUsageList); i++{
		if r.ImsiUsageList[i].IMSI == imsi{
			return i
		}
	}
	return -1
}

// 메모리에 key값을 이용해 데이터 조회
func (r RecordMemory) Query(query []string) ([]byte, error){
	//return []string{TotalUsage_Type, t.VPMN, t.HPMN, t.ContractID}
	switch query[0]{
	case RecordUsage_Type:
		usageList :=  r.RecordUsageList
		for i:=0; i< len(usageList); i++{
			if reflect.DeepEqual(usageList[i].GetKeyArray() , query){
				returnValue, _:=json.Marshal(usageList[i])
				return returnValue, nil
			}
		}
	case ImsiUsage_Type:
		usageList :=  r.ImsiUsageList
		for i:=0; i< len(usageList); i++{
			if reflect.DeepEqual(usageList[i].GetKeyArray() , query){
				returnValue, _:=json.Marshal(usageList[i])
				return returnValue, nil
			}
		}
	case DateUsage_Type:
		usageList :=  r.DateUsage
		if reflect.DeepEqual(usageList.GetKeyArray() , query){
			returnValue, _:=json.Marshal(usageList)
			return returnValue, nil
		}
	case TotalUsage_Type:
		usageList :=  r.TotalUsage
		if reflect.DeepEqual(usageList.GetKeyArray() , query){
			returnValue, _:=json.Marshal(usageList)
			return returnValue, nil
		}
	case TapfileUsage_Type:
		usageList :=  r.TapfileUsage
		if reflect.DeepEqual(usageList.GetKeyArray() , query){
			returnValue, _:=json.Marshal(usageList)
			return returnValue, nil
		}
	default:
		return nil, errors.New("query type error")
	}
	return nil, errors.New("RecordMemory Query Error")
}

// 메모리에 있는 모든 데이터 블록 생성
func (r RecordMemory) AddBlockAll(stub shim.ChaincodeStubInterface) error{
	RecordUsageList := r.RecordUsageList
	ImsiUsageList := r.ImsiUsageList
	dateUsage := r.DateUsage
	totalUsage := r.TotalUsage
	tapfileUsage := r.TapfileUsage

	for i:=0; i< len(RecordUsageList); i++{
		RecordUsageList[i].AddBlock(stub)
	}
	for i:=0; i< len(ImsiUsageList); i++{
		ImsiUsageList[i].AddBlock(stub)
	}
	if dateUsage.Date != ""{
		//Tap Record 번호 업데이트
		returnValueByte, err := Block_Query(stub, dateUsage.GetKeyArray())
		if err == nil{
			dateUsageValue := DateUsage{}
			json.Unmarshal(returnValueByte[0],&dateUsageValue)

			if dateUsageValue.FileSequenceNumbers[0]==""{
				dateUsage.FileSequenceNumbers[0] = dateUsage.FileSequenceNumbers[1]
			}else{
				dateUsage.FileSequenceNumbers[0] = dateUsageValue.FileSequenceNumbers[0]
			}

		}else{
			dateUsage.FileSequenceNumbers[0] = dateUsage.FileSequenceNumbers[1]
		}
		dateUsage.AddBlock(stub)
	}
	if totalUsage.VPMN != ""{
		totalUsage.AddBlock(stub)
	}
	if tapfileUsage.TapFileName != ""{
		tapfileUsage.AddBlock(stub)
	}
	return nil
}

// *********************************************************

type CalculDetail struct {
	Unit			string		`json:"Unit"`
	Record    		int 		`json:"Record"`
	Duration  		float64 	`json:"Duration"`
	RoundedDuration	float64		`json:"RoundedDuration"`
	Charge    		float64 	`json:"Charge"`
	SetCharge 		float64 	`json:"SetCharge"`
	TAXINCLYN    	string	 	`json:"TAX_INCL_YN"`
	TAXCharge    	float64 	`json:"TAX_Charge"`
	TAXSETCharge 	float64 	`json:"TAX_SET_Charge"`
}

func(self CalculDetail) New() CalculDetail{
	self.Record = 1
	self.RoundedDuration = 0
	self.Charge = 0
	self.SetCharge = 0
	self.TAXCharge = 0
	self.TAXSETCharge = 0
	return self
}

func(self CalculDetail) Add(cal CalculDetail) CalculDetail{
	self.Record = self.Record + cal.Record
	self.Duration = self.Duration + cal.Duration
	self.RoundedDuration = self.RoundedDuration + cal.RoundedDuration
	self.Charge = self.Charge + cal.Charge
	self.SetCharge = self.SetCharge + cal.SetCharge
	self.TAXCharge = self.TAXCharge + cal.TAXCharge
	self.TAXSETCharge = self.TAXSETCharge + cal.TAXSETCharge

	if self.Unit == ""{
		self.Unit = cal.Unit
	}
	if self.TAXINCLYN == ""{
		self.TAXINCLYN = cal.TAXINCLYN
	}

	return self
}


type TapCal struct{
	MOCLocal		CalculDetail
	MOCHome			CalculDetail
	MOCInt			CalculDetail
	MTC				CalculDetail
	SMSMO			CalculDetail
	SMSMT			CalculDetail
	GPRS			CalculDetail
}

func (t TapCal)Add(cal TapCal) TapCal{
	t.MOCLocal = t.MOCLocal.Add(cal.MOCLocal)
	t.MOCHome = t.MOCHome.Add(cal.MOCHome)
	t.MOCInt = t.MOCInt.Add(cal.MOCInt)
	t.MTC = t.MTC.Add(cal.MTC)
	t.SMSMO = t.SMSMO.Add(cal.SMSMO)
	t.SMSMT = t.SMSMT.Add(cal.SMSMT)
	t.GPRS = t.GPRS.Add(cal.GPRS)
	return t
}

func(self TapCal) GetTotalSum() (PreTaxAmount float64, PostTaxAmount float64){
	PreTaxAmount = self.MOCLocal.SetCharge + self.MOCHome.SetCharge + self.MOCInt.SetCharge + self.MTC.SetCharge + self.SMSMO.SetCharge + self.SMSMT.SetCharge + self.GPRS.SetCharge
	PostTaxAmount = self.MOCLocal.TAXSETCharge + self.MOCHome.TAXSETCharge + self.MOCInt.TAXSETCharge + self.MTC.TAXSETCharge + self.SMSMO.TAXSETCharge + self.SMSMT.TAXSETCharge + self.GPRS.TAXSETCharge
	return
}

type TapCalculreturnValue struct{
	ContractID		string
	Peoriod			[2]string
	Currency		string
}

// Block_Query(stub shim.ChaincodeStubInterface, args []string) ([][]byte, error)
// usage 조회
func TapRecordUsageQuery(stub shim.ChaincodeStubInterface, key []string, recordMemory RecordMemory)([]byte, error){
	returnValue, err := recordMemory.Query(key)
	if err == nil{
		return returnValue, nil
	}
	returnValueByte, err := Block_Query(stub, key)
	if len(returnValueByte) >1{
		return nil, errors.New("value is more than two.")
	}
	if err != nil{
		return nil, err
	}
	return returnValueByte[0], nil
}