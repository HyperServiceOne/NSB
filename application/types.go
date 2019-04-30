package nsb

import (
	"github.com/Myriad-Dreamin/NSB/merkmap"
	"github.com/Myriad-Dreamin/NSB/localstorage"
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/syndtr/goleveldb/leveldb"
)


type NSBApplication struct {
	types.BaseApplication
	state       *NSBState
	stateMap    *merkmap.MerkMap
	accMap      *merkmap.MerkMap
	txMap       *merkmap.MerkMap
	statedb     *leveldb.DB
	ValUpdates  []types.ValidatorUpdate
	logger      log.Logger
}


type NSBState struct {
	db dbm.DB
	StateRoot []byte `json:"action_root"`
	Height  int64  `json:"height"`
}
