package p

import (
	"context"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func (c *Backend) ConstructionDerive(
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

func (c *Backend) ConstructionPreprocess(ctx context.Context, req *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}

func (c *Backend) ConstructionMetadata(ctx context.Context, req *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}

func (c *Backend) ConstructionPayloads(ctx context.Context, req *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}

func (c *Backend) ConstructionParse(ctx context.Context, req *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}

func (c *Backend) ConstructionCombine(ctx context.Context, req *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	return nil, service.ErrNotImplemented
}

func (c *Backend) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txHex := req.SignedTransaction
	txByte, err := formatting.Decode(formatting.Hex, txHex)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}
	txHash256 := hashing.ComputeHash256(txByte)
	pHash, err := formatting.EncodeWithChecksum(formatting.CB58, txHash256)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}
	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: pHash,
		},
	}, nil
}

func (c *Backend) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txHex := req.SignedTransaction
	txByte, err := formatting.Decode(formatting.Hex, txHex)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	txID, err := c.pClient.IssueTx(ctx, txByte)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txID.String(),
		},
	}, nil
}
