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
func (s *Backend) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {

	id, err := ids.FromString(*request.BlockIdentifier.Hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	txBytes, err := s.pClient.GetBlock(ctx, id)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	pTx := platformvm.Tx{}
	if _, err := platformvm.Codec.Unmarshal(txBytes, &pTx); err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	// pTx.UnsignedTx.InitCtx(rc.SnowContext)
	tx := pTx.UnsignedTx

	t, err := mapper.Transaction(tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}
	resp := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: 0,
				Hash:  pTx.ID().String(),
			},
			Timestamp:    0,
			Transactions: []*types.Transaction{t},
		},
		OtherTransactions: []*types.TransactionIdentifier{},
	}

	return resp, nil
}

// BlockTransaction implements the /block/transaction endpoint.
func (s *Backend) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {

	return nil, nil
}
