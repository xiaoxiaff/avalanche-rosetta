package p

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	mapper "github.com/ava-labs/avalanche-rosetta/mapper/p"
	"github.com/ava-labs/avalanche-rosetta/service"
)

// Block implements the /block endpoint
func (b *Backend) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {

	id, err := ids.FromString(*request.BlockIdentifier.Hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	txBytes, err := b.pClient.GetBlock(ctx, id)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var block platformvm.Block
	if _, err := platformvm.Codec.Unmarshal(txBytes, &block); err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	t, err := mapper.Transaction(block)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}
	resp := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: 0,
				Hash:  block.ID().String(),
			},
			Timestamp:    0,
			Transactions: []*types.Transaction{t},
		},
		OtherTransactions: []*types.TransactionIdentifier{},
	}

	return resp, nil
}

// BlockTransaction implements the /block/transaction endpoint.
func (b *Backend) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	return nil, nil
}
