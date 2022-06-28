package p

import (
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/service/chain"
)

type Backend struct {
	chain.ConstructionBackend
	chain.NetworkBackend

	networkIdentifier *types.NetworkIdentifier
	fac               *crypto.FactorySECP256K1R
	pClient           client.PChainClient
	getUTXOsPageSize  uint32
	codec             codec.Manager
	assetID           ids.ID
}

func NewBackend(
	pClient client.PChainClient,
	assetID ids.ID,
	networkIdentifier *types.NetworkIdentifier,
) *Backend {
	return &Backend{
		fac:               &crypto.FactorySECP256K1R{},
		pClient:           pClient,
		networkIdentifier: networkIdentifier,
		getUTXOsPageSize:  1024,
		codec:             platformvm.Codec,
		assetID:           assetID,
	}
}
