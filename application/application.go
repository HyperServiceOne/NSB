package nsb

import (
	"fmt"
	"encoding/hex"
	"encoding/binary"
	"bytes"
	"errors"
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/version"
	"github.com/Myriad-Dreamin/NSB/merkmap"
	"github.com/Myriad-Dreamin/NSB/application/response"
)


func NewNSBApplication(dbDir string) (*NSBApplication, error) {
	name := "nsbstate"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		return nil, err
	}
	fmt.Println("loading state...")
	state := loadState(db)
	fmt.Println(state.String())

	var stmp *merkmap.MerkMap
	var statedb *leveldb.DB
	statedb, err = leveldb.OpenFile("./data/trienode.db", nil)
	if err != nil {
		return nil, err
	}
	stmp, err = merkmap.NewMerkMapFromDB(statedb, state.StateRoot, "00")
	if err != nil {
		return nil, err
	}

	return &NSBApplication{
		state:    state,
		logger:   log.NewNopLogger(),
		stateMap: stmp,
		accMap:   stmp.ArrangeSlot([]byte("acc:")),
		txMap:    stmp.ArrangeSlot([]byte("tx:")),
		statedb:  statedb,
	}, nil
}


func (nsb *NSBApplication) SetLogger(l log.Logger) {
	nsb.logger = l
}


func (nsb *NSBApplication) Revert() error {
	err := nsb.stateMap.Revert()
	if err != nil {
		return err
	}
	nsb.accMap = nsb.stateMap.ArrangeSlot([]byte("acc:"))
	nsb.txMap = nsb.stateMap.ArrangeSlot([]byte("tx:"))
	return nil
}


func (nsb *NSBApplication) Info(req types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{
		Data:       fmt.Sprintf(
			"{\"state_root\":%v, \"height\":%v, }",
			hex.EncodeToString(nsb.state.StateRoot),
			nsb.state.Height),
		Version:    version.ABCIVersion,
		AppVersion: NSBVersion.Uint64(),
	}
}


// Save the validators in the merkle tree
func (nsb *NSBApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := nsb.updateValidator(v)
		if r.IsErr() {
			nsb.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}


// Track the block hash and header information
func (nsb *NSBApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	// reset valset changes
	fmt.Println("BeginBlock")
	nsb.ValUpdates = make([]types.ValidatorUpdate, 0)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (nsb *NSBApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	fmt.Println("EndBlock")
	return types.ResponseEndBlock{ValidatorUpdates: nsb.ValUpdates}
}

func (nsb *NSBApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	fmt.Println("CheckTx")
	return types.ResponseCheckTx{Code: 0}
}

func (nsb *NSBApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	fmt.Println("DeliverTx")
	bytesTx := bytes.Split(tx, []byte("\x19"))
	var ret types.ResponseDeliverTx
	if len(bytesTx) != 2 {
		return *response.InvalidTxInputFormatWrongx19
	}
	switch string(bytesTx[0]) {

	case "validators": // nsb validators
		ret = nsb.execValidatorTx(bytesTx[1])

	case "sendTransaction": // transact contract methods
		ret = *nsb.parseFuncTransaction(bytesTx[1])

	case "transact": // send token
		ret = *nsb.transact(bytesTx[1])

	case "createContract": // create on-chain contracts
		ret = *nsb.parseCreateTransaction(bytesTx[1])

	default:
		return types.ResponseDeliverTx{Code: uint32(response.CodeInvalidTxType)}
	}
	if ret.Code != uint32(response.CodeOK) {
		err := nsb.Revert()
		if err != nil {
			fmt.Println(err)
		}
	}
	return ret
}

func (nsb *NSBApplication) Commit() types.ResponseCommit {
	fmt.Println("Commit")
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 32)
	binary.PutVarint(appHash, nsb.state.Height)
	var err error

	// accMap, txMap is the sub-Map of stateMap
	nsb.state.StateRoot, err = nsb.stateMap.Commit(nil)
	if err != nil {
		panic(err)
	}
	nsb.state.Height += 1
	saveState(nsb.state)
	return types.ResponseCommit{Data: nsb.state.StateRoot}
}

/*
type RequestQuery struct {
    Data                 []byte   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
    Path                 string   `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
    Height               int64    `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
    Prove                bool     `protobuf:"varint,4,opt,name=prove,proto3" json:"prove,omitempty"`
}
type ResponseQuery struct {
    Code uint32 `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
    // bytes data = 2; // use "value" instead.
    Log                  string        `protobuf:"bytes,3,opt,name=log,proto3" json:"log,omitempty"`
    Info                 string        `protobuf:"bytes,4,opt,name=info,proto3" json:"info,omitempty"`
    Index                int64         `protobuf:"varint,5,opt,name=index,proto3" json:"index,omitempty"`
    Key                  []byte        `protobuf:"bytes,6,opt,name=key,proto3" json:"key,omitempty"`
    Value                []byte        `protobuf:"bytes,7,opt,name=value,proto3" json:"value,omitempty"`
    Proof                *merkle.Proof `protobuf:"bytes,8,opt,name=proof" json:"proof,omitempty"`
    Height               int64         `protobuf:"varint,9,opt,name=height,proto3" json:"height,omitempty"`
    Codespace            string        `protobuf:"bytes,10,opt,name=codespace,proto3" json:"codespace,omitempty"`
}
type Proof struct {
    Ops                  []ProofOp `protobuf:"bytes,1,rep,name=ops" json:"ops"`
}
*/
func (nsb *NSBApplication) Query(req types.RequestQuery) (ret types.ResponseQuery) {
	if req.Prove {
		ret.Code = uint32(response.CodeOK)
		ret.Key = req.Data
		ret.Value = []byte(req.Path)
		ret.Log = fmt.Sprintf("asking Prove key: %v, value %v", req.Data, req.Path)
	} else {
		// start new ISC
		// add MerkleProof
		// add Action
		// insurance claim
		// settle contract
		// return/stake funds
		
		// 2ecddf60bb43e12eb402949337a4a0795480f1409e76b7f9cf52ef783532da0a

		ret.Code = uint32(response.CodeOK)
		ret.Key = req.Data
		ret.Value = []byte(req.Path)
		ret.Log = fmt.Sprintf("asking not Prove key: %v, value %v", req.Data, req.Path)
		ret.Info = nsb.QueryIndex(&req)
	}
	return 
}

func (nsb *NSBApplication) Stop() (err1 error, err2 error) {
	if nsb == nil {
		return errors.New("the NSB application is not started"), nil
	}
	err1 = nsb.state.Close()
	err2 = nsb.stateMap.Close()
	nsb.state = nil
	nsb.stateMap = nil
	return
}