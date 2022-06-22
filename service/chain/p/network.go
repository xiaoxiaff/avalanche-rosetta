package p

import (
	"context"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func (c *Backend) NetworkIdentifier() *types.NetworkIdentifier {
	return c.networkIdentifier
}

func (c *Backend) NetworkStatus(ctx context.Context, req *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}

func (c *Backend) NetworkOptions(ctx context.Context, request *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion:    types.RosettaAPIVersion,
			NodeVersion:       service.NodeVersion,
			MiddlewareVersion: types.String(service.MiddlewareVersion),
		},
		Allow: &types.Allow{
			OperationStatuses:       mapper.OperationStatuses,
			OperationTypes:          mapper.PChainOperationTypes,
			CallMethods:             mapper.PChainCallMethods,
			Errors:                  service.Errors,
			HistoricalBalanceLookup: false,
		},
	}, nil
}
