package isc

import (
	"fmt"
	"encoding/hex"
	"encoding/json"
	cmn "github.com/Myriad-Dreamin/NSB/common"
	"github.com/Myriad-Dreamin/NSB/contract/isc/transaction"
	. "github.com/Myriad-Dreamin/NSB/common/contract_response"
)


func RigisteredMethod(env *cmn.ContractEnvironment) *cmn.ContractCallBackInfo {
	switch env.FuncName {
	case "a+b":
		return SafeAdd(env.Args)
	default:
		return InvalidFunctionType(env.FuncName)
	}
}

// func (nsb *NSBApplication) activeISC(byteJson []byte) (types.ResponseDeliverTx) {
// 	return types.ResponseDeliverTx{
// 		Code: uint32(CodeOK),
// 	}
// }

type ArgsCreateNewContract struct {
	IscOwners          [][]byte                        `json:"isc_owners"`
	Funds              []uint32                        `json:"required_funds"`
	VesSig             []byte                          `json:"ves_signature"`
	TransactionIntents []transaction.TransactionIntent `json:"transactionIntents"`
}
// // 0x637265617465495343197b226973635f6f776e657273223a5b22456a525765413d3d222c22456a5257654a6f3d225d2c2272657175697265645f66756e6473223a5b302c305d2c22566573536967223a22497a4d3d222c225472616e73616374696f6e496e74656e7473223a5b7b2266726f6d223a22456a525765413d3d222c22746f223a22456a5257654a6f3d222c22736571223a302c22616d74223a302c226d657461223a2249673d3d227d5d7d
// 2ecddf60bb43e12eb402949337a4a0795480f1409e76b7f9cf52ef783532da0a

func CreateNewContract(env *cmn.ContractEnvironment) (*cmn.ContractCallBackInfo) {
	var args ArgsCreateNewContract
	err := json.Unmarshal(env.Args, &args)
	if err != nil {
		return DecodeJsonError(err)
	}

	fmt.Print(string(env.Data))
	return &cmn.ContractCallBackInfo{
		CodeResponse: uint32(CodeOK),
		Info: fmt.Sprintf("create success , this contract is deploy at %v", hex.EncodeToString(env.ContractAddress)),
	}
}