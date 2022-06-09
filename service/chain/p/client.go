package p

import (
	"github.com/ava-labs/avalanche-rosetta/service/chain"
	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

type client struct {
	fac           crypto.FactorySECP256K1R
	pClient       platformvm.Client
	indexerClient indexer.Client
}

type Client interface {
	chain.ConstructionBackend
	chain.NetworkBackend
}

func NewClient(rpcEndpoint string) Client {
	return &client{
		fac:     crypto.FactorySECP256K1R{},
		pClient: platformvm.NewClient(rpcEndpoint),
	}
}
