package pchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/stakeable"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	pmapper "github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
	"github.com/ava-labs/avalanche-rosetta/service/backend/pchain/indexer"
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

	block, blockHash, blockTimestamp, err := b.getBlock(ctx, blockIndex, hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	var txs []*platformvm.Tx
	switch v := block.(type) {
	// No transactions in the following
	case *platformvm.AbortBlock, *platformvm.CommitBlock:
	case *platformvm.AtomicBlock:
		// Single transaction in the atomic blocks
		txs = append(txs, &v.Tx)
	case *platformvm.ProposalBlock:
		// Single transaction in the proposal blocks
		txs = append(txs, &v.Tx)
	// 0..n transactions in standard block
	case *platformvm.StandardBlock:
		txs = append(txs, v.Txs...)
	default:
		log.Printf("unknown %s", reflect.TypeOf(v))
	}

	transactions, typeErr := b.buildBlockTxs(ctx, request.NetworkIdentifier, txs, blockHash)
	if typeErr != nil {
		return nil, typeErr
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
			Timestamp:    time.Unix(blockTimestamp, 0).UnixMilli(),
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

	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       b.genesisBlockIdentifier,
			ParentBlockIdentifier: b.genesisBlockIdentifier,
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
	block, _, _, err := b.getBlock(ctx, request.BlockIdentifier.Index, request.BlockIdentifier.Hash)
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

func (b *Backend) getBlock(ctx context.Context, index int64, hash string) (platformvm.Block, string, int64, error) {
	var blockId ids.ID
	var err error

	if index <= 0 && hash == "" {
		return nil, "", 0, errMissingBlockIndexHash
	}

	var parsedBlock *indexer.ParsedBlock
	// Extract block id from hash parameter if it is non-empty, or from index if stated
	if hash != "" {
		parsedBlock, err = b.indexerParser.ParseBlockWithHash(ctx, hash)
		if err != nil {
			return nil, "", 0, err
		}

	} else if index > 0 {
		blockIndex := uint64(index)
		parsedBlock, err = b.indexerParser.ParseBlockAtIndex(ctx, blockIndex)
		if err != nil {
			return nil, "", 0, err
		}
	}

	blockId = parsedBlock.BlockID
	blockTimestamp := parsedBlock.Proposer.Timestamp

	// Unsigned blocks have empty proposer, in that case we return the block timestamp
	// which is currently set to the genesis timestamp
	if blockTimestamp == 0 {
		blockTimestamp = parsedBlock.Timestamp
	}

	// Get the block bytes
	var blockBytes []byte
	blockBytes, err = b.pClient.GetBlock(ctx, blockId)
	if err != nil {
		return nil, "", 0, err
	}

	// Unmarshal the block
	var block platformvm.Block
	if _, err := b.codec.Unmarshal(blockBytes, &block); err != nil {
		return nil, "", 0, err
	}

	// Verify block height matches specified height - if there is one
	if index > 0 && block.Height() != uint64(index) {
		return nil, "", 0, fmt.Errorf("requested block index: %d, found: %d for block %s", index, block.Height(), blockId.String())
	}

	return block, blockId.String(), blockTimestamp, nil
}

func (b *Backend) isGenesisBlockRequest(ctx context.Context, id *types.PartialBlockIdentifier) (bool, error) {
	_, err := b.getGenesisBlock(ctx)
	if err != nil {
		return false, err
	}

	if number := id.Index; number != nil {
		return *number == b.genesisBlockIdentifier.Index, nil
	}
	if hash := id.Hash; hash != nil {
		return *hash == b.genesisBlockIdentifier.Hash, nil
	}
	return false, nil
}

// buildBlockTxs parses transactions and generates operations.
// If it is reward operations, it will fetch the original
// tx and parse the rewards outputs. If it is input, it will
// fetch the previous UTXO and parse the output address.
// It will ignore import/export and multi-sig UTXO.
func (b *Backend) buildBlockTxs(ctx context.Context, networkIdentifier *types.NetworkIdentifier, txs []*platformvm.Tx, blockHash string) ([]*types.Transaction, *types.Error) {
	parser, err := b.newTxParser(ctx, networkIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, "tx parser error")
	}
	transactions := []*types.Transaction{}
	for _, tx := range txs {
		err := common.InitializeTx(b.codecVersion, b.codec, *tx)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "tx initalize error")
		}

		t, err := parser.Parse(tx.UnsignedTx)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
		if len(t.Operations) > 0 {
			newOperations := make([]*types.Operation, 0)
			for _, op := range t.Operations {
				opMetadata, err := pmapper.ParseOpMetadata(op.Metadata)
				if err != nil {
					return nil, service.WrapError(service.ErrBlockInvalidInput, err)
				}

				// skip multi-sig operations.
				if opMetadata.MultiSig {
					continue
				}
				switch opMetadata.Type {
				case pmapper.OpTypeImport, pmapper.OpTypeExport, pmapper.OpTypeCreateChain:
					// ignore import, export and create-chain operations
					continue
				case pmapper.OpTypeOutput, pmapper.OpTypeStakeOutput:
					// do nothing and append to the result directly
				case pmapper.OpTypeReward:
					// we fetch the original tx's reward UTXOs and parse into the operations
					txID, err := ids.FromString(opMetadata.StakingTxID)
					if err != nil {
						return nil, service.WrapError(service.ErrBlockInvalidInput, err)
					}
					rewardOps, typeErr := b.getRewardOps(ctx, txID, 0, pmapper.OpRewardValidator, parser)
					if typeErr != nil {
						return nil, typeErr
					}
					// ignore multi-sig rewards as well
					for _, rewardOp := range rewardOps {
						rewardMetadata, err := pmapper.ParseOpMetadata(rewardOp.Metadata)
						if err != nil {
							return nil, service.WrapError(service.ErrBlockInvalidInput, err)
						}
						if rewardMetadata.MultiSig {
							continue
						}
						newOperations = append(newOperations, rewardOp)
					}
					continue
				case pmapper.OpTypeInput:
					isMultiSigInput := false
					// get the previous UTXO
					utxoID, err := mapper.DecodeUTXOID(op.CoinChange.CoinIdentifier.Identifier)
					if err != nil {
						return nil, service.WrapError(service.ErrBlockInvalidInput, err)
					}

					txBytes, err := b.pClient.GetTx(ctx, utxoID.TxID)
					if err != nil {
						return nil, service.WrapError(service.ErrBlockInvalidInput, err)
					}

					var previousTx platformvm.Tx
					_, err = b.codec.Unmarshal(txBytes, &previousTx)
					if err != nil {
						return nil, service.WrapError(service.ErrBlockInvalidInput, err)
					}
					rosettaTx, err := parser.Parse(previousTx.UnsignedTx)
					if err != nil {
						return nil, service.WrapError(service.ErrBlockInvalidInput, err)
					}

					if len(rosettaTx.Operations) > 0 {
						rewardOps, typeErr := b.getRewardOps(ctx, utxoID.TxID, len(rosettaTx.Operations), rosettaTx.Operations[0].Type, parser)
						if typeErr != nil {
							return nil, typeErr
						}
						rosettaTx.Operations = append(rosettaTx.Operations, rewardOps...)
					}

					outputIndex := uint32(0)
					for _, previousOp := range rosettaTx.Operations {
						opMetadata, err := pmapper.ParseOpMetadata(previousOp.Metadata)
						if err != nil {
							return nil, service.WrapError(service.ErrBlockInvalidInput, err)
						}

						if opMetadata.Type == pmapper.OpTypeOutput || opMetadata.Type == pchain.OpTypeStakeOutput {
							if outputIndex == utxoID.OutputIndex {
								if opMetadata.MultiSig {
									isMultiSigInput = true
									break
								}
								// TODO figure the account for multisig input
								op.Account = previousOp.Account
								break
							}
							outputIndex++
						}
					}
					if isMultiSigInput {
						continue
					}
				}
				newOperations = append(newOperations, op)
			}
			for i, op := range newOperations {
				op.OperationIdentifier.Index = int64(i)
			}
			t.Operations = newOperations
		}

		transactions = append(transactions, t)
	}

	return transactions, nil
}

// TODO: we need to know if get reward sequence is deterministic
func (b *Backend) getRewardOps(
	ctx context.Context,
	txID ids.ID,
	startIndex int,
	opType string,
	parser *pmapper.TxParser,
) ([]*types.Operation, *types.Error) {
	rewardUTXOs, err := b.pClient.GetRewardUTXOs(ctx, &api.GetTxArgs{
		TxID:     txID,
		Encoding: formatting.Hex,
	})
	if err != nil {
		return nil, service.WrapError(service.ErrTransactionNotFound, err)
	}
	rewardOps := make([]*types.Operation, len(rewardUTXOs))
	for i, bytes := range rewardUTXOs {
		utxo := avax.UTXO{}
		_, err := b.codec.Unmarshal(bytes, &utxo)
		if err != nil {
			return nil, service.WrapError(service.ErrBlockInvalidInput, err)
		}

		outIntf := utxo.Out
		if lockedOut, ok := outIntf.(*stakeable.LockOut); ok {
			outIntf = lockedOut.TransferableOut
		}

		out, ok := outIntf.(*secp256k1fx.TransferOutput)
		if !ok {
			return nil, service.WrapError(service.ErrBlockInvalidInput, err)
		}

		op, err := parser.BuildOutputOperation(
			out,
			types.String(mapper.StatusSuccess),
			startIndex+len(rewardUTXOs)-i-1,
			opType,
			pchain.OpTypeOutput,
			mapper.PChainNetworkIdentifier,
		)
		if err != nil {
			return nil, service.WrapError(service.ErrBlockInvalidInput, err)
		}
		rewardOps[len(rewardUTXOs)-i-1] = op
	}

	return rewardOps, nil
}
