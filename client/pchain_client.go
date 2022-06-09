package client

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"strings"
)

// Interface compliance
var _ PChainClient = &pchainClient{}

type PChainClient interface {
	// indexer.Client methods

	GetContainerByIndex(ctx context.Context, index uint64, options ...rpc.Option) (indexer.Container, error)
	GetLastAccepted(context.Context, ...rpc.Option) (indexer.Container, error)

	// platformvm.Client methods

	GetBalance(ctx context.Context, addrs []ids.ShortID, options ...rpc.Option) (*platformvm.GetBalanceResponse, error)
	GetTx(ctx context.Context, txID ids.ID, options ...rpc.Option) ([]byte, error)
	GetBlock(ctx context.Context, blockID ids.ID, options ...rpc.Option) ([]byte, error)
	IssueTx(ctx context.Context, tx []byte, options ...rpc.Option) (ids.ID, error)
}

type indexerClient = indexer.Client
type platformvmClient = platformvm.Client

type pchainClient struct {
	platformvmClient
	indexerClient
}

// NewPChainClient returns a new client for Avalanche APIs related to P-chain
func NewPChainClient(ctx context.Context, endpoint string) PChainClient {
	endpoint = strings.TrimSuffix(endpoint, "/")

	return pchainClient{
		platformvmClient: platformvm.NewClient(endpoint),
		indexerClient:    indexer.NewClient(endpoint),
	}
}
