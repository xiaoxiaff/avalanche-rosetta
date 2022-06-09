package p

import (
	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/service/chain"
	"github.com/ava-labs/avalanchego/utils/crypto"
)

type Backend struct {
	chain.ConstructionBackend
	chain.NetworkBackend

	fac     crypto.FactorySECP256K1R
	pClient client.PChainClient
}

func NewBackend(pClient client.PChainClient) *Backend {
	return &Backend{
		fac:     crypto.FactorySECP256K1R{},
		pClient: pClient,
	}
}
