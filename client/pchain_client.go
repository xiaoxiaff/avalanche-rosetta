package client

import (
	"context"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

// Interface compliance
var _ PChainClient = &pchainClient{}

type PChainClient interface {
	// indexer.Client methods

	GetContainerByIndex(ctx context.Context, index uint64, options ...rpc.Option) (indexer.Container, error)
	GetLastAccepted(context.Context, ...rpc.Option) (indexer.Container, error)

	// platformvm.Client methods

	GetUTXOs(
		ctx context.Context,
		addrs []ids.ShortID,
		limit uint32,
		startAddress ids.ShortID,
		startUTXOID ids.ID,
		options ...rpc.Option,
	) ([][]byte, ids.ShortID, ids.ID, error)
	GetBalance(ctx context.Context, addrs []ids.ShortID, options ...rpc.Option) (*platformvm.GetBalanceResponse, error)
	GetTx(ctx context.Context, txID ids.ID, options ...rpc.Option) ([]byte, error)
	GetBlock(ctx context.Context, blockID ids.ID, options ...rpc.Option) ([]byte, error)
	IssueTx(ctx context.Context, tx []byte, options ...rpc.Option) (ids.ID, error)

	// avm.Client methods

	GetAssetDescription(ctx context.Context, assetID string, options ...rpc.Option) (*avm.GetAssetDescriptionReply, error)
}

type indexerClient = indexer.Client
type platformvmClient = platformvm.Client

type pchainClient struct {
	platformvmClient
	indexerClient
	xChainClient avm.Client
}

// NewPChainClient returns a new client for Avalanche APIs related to P-chain
func NewPChainClient(ctx context.Context, endpoint string) PChainClient {
	endpoint = strings.TrimSuffix(endpoint, "/")

	return pchainClient{
		platformvmClient: platformvm.NewClient(endpoint),
		indexerClient:    indexer.NewClient(endpoint),
		xChainClient:     avm.NewClient(endpoint, "X"),
	}
}

func (p pchainClient) GetAssetDescription(ctx context.Context, assetID string, options ...rpc.Option) (*avm.GetAssetDescriptionReply, error) {
	return p.xChainClient.GetAssetDescription(ctx, assetID, options...)
}
