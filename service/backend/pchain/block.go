package pchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	pmapper "github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
)

var (
	errMissingBlockIndexHash = errors.New("a positive block index, a block hash or both must be specified")
	errTxInitialize          = errors.New("tx initalize error")
)

// Block implements the /block endpoint
func (b *Backend) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	isGenesisBlockRequest, err := b.isGenesisBlockRequest(ctx, request.BlockIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	if isGenesisBlockRequest {
		block, err := b.buildGenesisBlockResponse(ctx, request.NetworkIdentifier)
		if err != nil {
			return nil, service.WrapError(service.ErrClientError, err)
		}
		return block, nil
	}

	var blockIndex int64
	var hash string

	if request.BlockIdentifier.Index != nil {
		blockIndex = *request.BlockIdentifier.Index
	}

	if request.BlockIdentifier.Hash != nil {
		hash = *request.BlockIdentifier.Hash
	}

	block, blockHash, err := b.getBlock(ctx, blockIndex, hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var txs []*platformvm.Tx
	switch v := block.(type) {
	// No transactions in the following
	case *platformvm.AbortBlock, *platformvm.CommitBlock, *platformvm.AtomicBlock:
	// Single transaction in the proposal blocks
	case *platformvm.ProposalBlock:
		txs = append(txs, &v.Tx)
	// 0..n transactions in standard block
	case *platformvm.StandardBlock:
		txs = append(txs, v.Txs...)
	default:
		log.Printf("unknown %s", reflect.TypeOf(v))
	}

	transactions, err := b.parseTransactions(ctx, request.NetworkIdentifier, txs)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	resp := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Height()),
				Hash:  blockHash,
			},
			ParentBlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Height()) - 1,
				Hash:  block.Parent().String(),
			},
			//TODO: Find a way to get block timestamp. The following causes panic as there is no vm defined for the block
			//Timestamp:    block.Timestamp().UnixMilli(),
			Transactions: transactions,
		},
	}

	return resp, nil
}

func (b *Backend) buildGenesisBlockResponse(ctx context.Context, networkIdentifier *types.NetworkIdentifier) (*types.BlockResponse, error) {
	genesisBlock, err := b.getGenesisBlock(ctx)
	if err != nil {
		return nil, err
	}

	txs, err := b.indexerParser.GenesisToTxs(genesisBlock)
	if err != nil {
		return nil, err
	}

	transactions, err := b.parseTransactions(ctx, networkIdentifier, txs)
	if err != nil {
		return nil, err
	}

	genesisBlockIdentifier := b.buildGenesisBlockIdentifier(genesisBlock)
	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       genesisBlockIdentifier,
			ParentBlockIdentifier: genesisBlockIdentifier,
			Transactions:          transactions,
			Timestamp:             mapper.UnixToUnixMilli(genesisBlock.Timestamp),
			Metadata: map[string]interface{}{
				pmapper.MetadataMessage: genesisBlock.Message,
			},
		},
	}, err
}

// BlockTransaction implements the /block/transaction endpoint.
func (b *Backend) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	block, _, err := b.getBlock(ctx, request.BlockIdentifier.Index, request.BlockIdentifier.Hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var txs []*platformvm.Tx
	switch typedBlock := block.(type) {
	case *platformvm.StandardBlock:
		txs = append(txs, typedBlock.Txs...)
	case *platformvm.ProposalBlock:
		txs = append(txs, &typedBlock.Tx)
	}

	transactions, err := b.parseTransactions(ctx, request.NetworkIdentifier, txs)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	for i := range transactions {
		transaction := transactions[i]
		if transaction.TransactionIdentifier.Hash == request.TransactionIdentifier.Hash {
			return &types.BlockTransactionResponse{
				Transaction: transaction,
			}, nil
		}
	}

	return nil, service.ErrTransactionNotFound
}

func (b *Backend) parseTransactions(
	ctx context.Context,
	networkIdentifier *types.NetworkIdentifier,
	txs []*platformvm.Tx,
) ([]*types.Transaction, error) {
	parser, err := b.newTxParser(ctx, networkIdentifier)
	if err != nil {
		return nil, err
	}

	var transactions []*types.Transaction
	for _, tx := range txs {
		err := common.InitializeTx(b.codecVersion, b.codec, *tx)
		if err != nil {
			return nil, errTxInitialize
		}

		t, err := parser.Parse(tx.UnsignedTx)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, t)
	}
	return transactions, nil
}

func (b *Backend) newTxParser(ctx context.Context, networkIdentifier *types.NetworkIdentifier) (*pmapper.TxParser, error) {
	hrp, err := mapper.GetHRP(networkIdentifier)
	if err != nil {
		return nil, err
	}

	chainIDs, err := b.getChainIDs(ctx)
	if err != nil {
		return nil, err
	}

	return pmapper.NewTxParser(false, hrp, chainIDs), nil
}

func (b *Backend) getChainIDs(ctx context.Context) (map[string]string, error) {
	if b.chainIDs == nil {
		b.chainIDs = map[string]string{
			ids.Empty.String(): mapper.PChainNetworkIdentifier,
		}

		cChainID, err := b.pClient.GetBlockchainID(ctx, mapper.CChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}
		b.chainIDs[cChainID.String()] = mapper.CChainNetworkIdentifier

		xChainID, err := b.pClient.GetBlockchainID(ctx, mapper.XChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}
		b.chainIDs[xChainID.String()] = mapper.XChainNetworkIdentifier
	}

	return b.chainIDs, nil
}

func (b *Backend) getBlock(ctx context.Context, index int64, hash string) (platformvm.Block, string, error) {
	var blockId ids.ID
	var err error

	if index <= 0 && hash == "" {
		return nil, "", errMissingBlockIndexHash
	}

	// Extract block id from hash parameter if it is non-empty, or from index if stated
	if hash != "" {
		blockId, err = ids.FromString(hash)
		if err != nil {
			return nil, "", err
		}
	} else if index > 0 {
		blockIndex := uint64(index)
		parsedBlock, err := b.indexerParser.ParseBlockAtIndex(ctx, blockIndex)
		if err != nil {
			return nil, "", err
		}

		blockId = parsedBlock.BlockID
	}

	// Get the block bytes
	var blockBytes []byte
	blockBytes, err = b.pClient.GetBlock(ctx, blockId)
	if err != nil {
		return nil, "", err
	}

	// Unmarshal the block
	var block platformvm.Block
	if _, err := b.codec.Unmarshal(blockBytes, &block); err != nil {
		return nil, "", err
	}

	// Verify block height matches specified height - if there is one
	if index > 0 && block.Height() != uint64(index) {
		return nil, "", fmt.Errorf("requested block index: %d, found: %d for block %s", index, block.Height(), blockId.String())
	}

	return block, blockId.String(), nil
}

func (b *Backend) isGenesisBlockRequest(ctx context.Context, id *types.PartialBlockIdentifier) (bool, error) {
	genesisBlock, err := b.getGenesisBlock(ctx)
	if err != nil {
		return false, err
	}
	genesisBlockIdentifier := b.buildGenesisBlockIdentifier(genesisBlock)

	if number := id.Index; number != nil {
		return *number == genesisBlockIdentifier.Index, nil
	}
	if hash := id.Hash; hash != nil {
		return *hash == genesisBlockIdentifier.Hash, nil
	}
	return false, nil
}
