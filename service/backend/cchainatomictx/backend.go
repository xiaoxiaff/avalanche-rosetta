package cchainatomictx

import (
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/coreth/plugin/evm"

	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/service"
)

type Backend struct {
	service.AccountBackend
	service.ConstructionBackend

	fac              *crypto.FactorySECP256K1R
	cClient          client.Client
	getUTXOsPageSize uint32
	codec            codec.Manager
	codecVersion     uint16
	avaxAssetId      ids.ID
}

func NewBackend(cClient client.Client, avaxAssetId ids.ID) *Backend {
	return &Backend{
		fac:              &crypto.FactorySECP256K1R{},
		cClient:          cClient,
		avaxAssetId:      avaxAssetId,
		getUTXOsPageSize: 1024,
		codec:            evm.Codec,
		codecVersion:     0,
	}
}
