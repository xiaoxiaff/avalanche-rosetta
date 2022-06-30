package p

import (
	"context"
	"log"
	"reflect"

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

	blockBytes, err := b.pClient.GetBlock(ctx, id)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var block platformvm.Block
	if _, err := platformvm.Codec.Unmarshal(blockBytes, &block); err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var txs []*platformvm.Tx
	switch v := block.(type) {
	case *platformvm.AbortBlock, *platformvm.CommitBlock:
	case *platformvm.ProposalBlock:
		txs = append(txs, &v.Tx)
	case *platformvm.StandardBlock:
		txs = append(txs, v.Txs...)
	default:
		log.Printf("unknown %s", reflect.TypeOf(v))
	}

	var transactions []*types.Transaction
	for _, tx := range txs {
		t, err := mapper.Transaction(tx.UnsignedTx)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
		transactions = append(transactions, t)
	}

	resp := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: 0,
				Hash:  block.ID().String(),
			},
			Transactions: transactions,
		},
	}

	return resp, nil
}

// BlockTransaction implements the /block/transaction endpoint.
func (b *Backend) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	id, err := ids.FromString(request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	txBytes, err := b.pClient.GetTx(ctx, id)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var tx platformvm.Tx
	if _, err := platformvm.Codec.Unmarshal(txBytes, &tx); err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	t, err := mapper.Transaction(tx.UnsignedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	resp := &types.BlockTransactionResponse{
		Transaction: t,
	}
	return resp, nil
}
