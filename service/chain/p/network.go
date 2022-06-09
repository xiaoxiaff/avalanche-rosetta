package p

import (
	"context"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func (c *client) NetworkStatus(ctx context.Context, req *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}

func (c *client) NetworkOptions(ctx context.Context, request *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}
