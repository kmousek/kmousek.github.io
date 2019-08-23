package main

import (
	"github.com/main/go/jsonStruct"
	"github.com/main/go/service"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)


type SmartContract struct {
}

func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// GW쪽에서 호출 하는 function mapping
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "blockInsert" {
		return s.blockInsert(stub, args)
	}else if function == "blockQuery" {
		return s.blockQueryForPrint(stub, args)
	}else if function == "agreementAgree"{
		return s.agreementAgree(stub, args)
	}else if function == "block_getActiveAgreement"{
		//Test용입니다. 실제 GW에서 호출 안함
		return s.block_getActiveAgreement(stub, args)
	}else if function == "agreementExpiration"{
		return s.agreementExpiration(stub, args)
	}else if function == "tapInsert"{
		return s.tapInsert(stub, args)
	}else if function == "tapProcess"{
		return s.tapProcess(stub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}





//tap calc test
func (s *SmartContract) tapProcess(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var tapRt jsonStruct.TapRecord

	tapRt.Header.VPMN                        = "KORKF"
	tapRt.Header.HPMN                        = "CHNCT"
	tapRt.Header.FileType                    = "TransferBatch"
	tapRt.Header.CdTdlndicator               = "TD"
	tapRt.Header.FileSequenceNumber          = "0001"
	tapRt.Header.FileCreationTimeStamp       = "201907221018"
	tapRt.Header.FileCreateUtcTimeOffset     = ""
	tapRt.Header.Recordcount                 = "1"
	tapRt.CdrInfos.ID	                     = "0412"
	tapRt.CdrInfos.CallType	                 = "MOC"
	tapRt.CdrInfos.Imsi	                     = "0821088106346"
	tapRt.CdrInfos.CalledNumber	             = "062222222"
	tapRt.CdrInfos.LocalTimeStamp	         = "20190725135959"
	tapRt.CdrInfos.UtcTimeOffset	         = ""
	tapRt.CdrInfos.TotalCallEventDuration	 = "485"
	tapRt.CdrInfos.Imei	                     = ""
	tapRt.CdrInfos.CallingNumber	         = "01011111111"
	tapRt.CdrInfos.DataBolumeIncoming	     = ""
	tapRt.CdrInfos.DataVolumeOutgoing	     = ""
	tapRt.CdrInfos.Charge                    = 0
	tapRt.CdrInfos.SetCharge                 = 0



	err := service.CalculChargeAmount(stub, &tapRt)
	if err != nil{
		return shim.Error("Calc Err")
	}

	service.Log_add(tapRt.Header.VPMN)
	return shim.Success(getSuccessReturnValue("success"))
}

//Tap insert
func (s *SmartContract) tapInsert(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	returnValue, err := service.Tap_Insert(stub, args)
	if err != nil{
		return shim.Error(err.Error())
	}
	return shim.Success(returnValue)
}

//Block data 저장
func (s *SmartContract) blockInsert(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	//args[0] objectType check function 추가 예정
	objectType := args[0]
	var returnValue []byte
	var err error
	switch objectType{
	case "agreement":
		returnValue, err = service.Agreement_Insert(stub,args)
	}
	if err != nil{
		return shim.Error(err.Error())
	}
	return shim.Success(returnValue)
}



// for Test
func (s *SmartContract) block_getActiveAgreement(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	temp, err := service.Agreement_getActive(stub, args[0], args[1], args[2])
	if err != nil{
		return shim.Error(err.Error())
	}
	tempBytes, _ := json.Marshal(temp)

	return shim.Success(tempBytes)
}

//Block data 조회
func (s *SmartContract) blockQueryForPrint(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	responseValue, err := service.Block_Query_Print(stub,args)
	if err != nil{
		return shim.Error(err.Error())
	}
	return shim.Success(responseValue)
}

//계약 승인
func (s *SmartContract) agreementAgree(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	returnValue, err := service.Agreement_Agree(stub, args)
	if err != nil{
		return shim.Error(string(returnValue))
	}
	return shim.Success(returnValue)
}

//계약 만료
func (s *SmartContract) agreementExpiration(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	returnValue, err := service.Agreement_Expiration(stub, args)
	if err != nil{
		return shim.Error(string(returnValue))
	}
	return shim.Success(returnValue)
}

// main함수는 테스트에서만 사용이 됩니다.
func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}