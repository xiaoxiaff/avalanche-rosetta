package c

import (
	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/service/chain"
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/coreth/plugin/evm"
)

type CChainAtomicTxBackend struct {
	chain.AccountBackend
	chain.ConstructionBackend

	fac              *crypto.FactorySECP256K1R
	cClient          client.Client
	getUTXOsPageSize uint32
	codec            codec.Manager
	codecVersion     uint16
	avaxAssetId      ids.ID
}

func NewAtomicTxBackend(cClient client.Client, avaxAssetId ids.ID) *CChainAtomicTxBackend {
	return &CChainAtomicTxBackend{
		fac:              &crypto.FactorySECP256K1R{},
		cClient:          cClient,
		avaxAssetId:      avaxAssetId,
		getUTXOsPageSize: 1024,
		codec:            evm.Codec,
		codecVersion:     0,
	}
}
