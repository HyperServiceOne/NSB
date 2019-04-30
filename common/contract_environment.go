package common

import (
	"github.com/Myriad-Dreamin/NSB/localstorage"
	"github.com/Myriad-Dreamin/NSB/math"
)


type ContractEnvironment struct {
	Storage *localstorage.LocalStorage
	From []byte
	ContractAddress []byte
	Data []byte
	Value *math.Uint256
}