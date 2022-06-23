package p

import (
	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/service/chain"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/coinbase/rosetta-sdk-go/types"
)

type Backend struct {
	chain.ConstructionBackend
	chain.NetworkBackend

	networkIdentifier *types.NetworkIdentifier
	fac               crypto.FactorySECP256K1R
	pClient           client.PChainClient
	getUTXOsPageSize  uint32
}

func NewBackend(pClient client.PChainClient, networkIdentifier *types.NetworkIdentifier) *Backend {
	return &Backend{
		fac:               crypto.FactorySECP256K1R{},
		pClient:           pClient,
		networkIdentifier: networkIdentifier,
		getUTXOsPageSize:  1024,
	}
}
