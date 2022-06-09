package p

import (
	"context"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func (c *client) ConstructionDerive(
	ctx context.Context,
	req *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	pub, err := c.fac.ToPublicKey(req.PublicKey.Bytes)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	chainIDAlias, hrp, getErr := mapper.GetAliasAndHRP(req.NetworkIdentifier)
	if getErr != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	addr, err := address.Format(chainIDAlias, hrp, pub.Address().Bytes())
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: addr,
		},
	}, nil
}

func (c *client) ConstructionPreprocess(ctx context.Context, req *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	//TODO implement me
	panic("implement me")
}

func (c *client) ConstructionMetadata(ctx context.Context, req *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	//TODO implement me
	panic("implement me")
}

func (c *client) ConstructionPayloads(ctx context.Context, req *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	//TODO implement me
	panic("implement me")
}

func (c *client) ConstructionParse(ctx context.Context, req *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	//TODO implement me
	panic("implement me")
}

func (c *client) ConstructionCombine(ctx context.Context, req *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	//TODO implement me
	panic("implement me")
}

func (c *client) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	//TODO implement me
	panic("implement me")
}

func (c *client) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	//TODO implement me
	panic("implement me")
}
