package chain

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// Chain specific /construction/* implementations
type ConstructionBackend interface {
	ConstructionDerive(ctx context.Context, req *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error)
	ConstructionPreprocess(ctx context.Context, req *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error)
	ConstructionMetadata(ctx context.Context, req *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error)
	ConstructionPayloads(ctx context.Context, req *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error)
	ConstructionParse(ctx context.Context, req *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error)
	ConstructionCombine(ctx context.Context, req *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error)
	ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error)
	ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error)
}

// Chain specific /network/ implementations
type NetworkBackend interface {
	NetworkIdentifier() *types.NetworkIdentifier
	NetworkStatus(ctx context.Context, request *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error)
	NetworkOptions(ctx context.Context, request *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error)
}

// Chain specific /account/ implementations
type AccountBackend interface {
	AccountBalance(ctx context.Context, req *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error)
	AccountCoins(ctx context.Context, req *types.AccountCoinsRequest) (*types.AccountCoinsResponse, *types.Error)
}
