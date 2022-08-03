package pchain

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"
	"golang.org/x/sync/errgroup"

	"github.com/ava-labs/avalanche-rosetta/mapper"
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

	block, err := b.getBlockDetails(ctx, blockIndex, hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	transactions, err := b.parseTransactions(ctx, request.NetworkIdentifier, block.Txs)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	resp := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: blockIndex,
				Hash:  block.BlockID.String(),
			},
			ParentBlockIdentifier: &types.BlockIdentifier{
				Index: blockIndex - 1,
				Hash:  block.ParentID.String(),
			},
			Timestamp:    block.Timestamp,
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

	transactions, err := b.parseTransactions(ctx, networkIdentifier, genesisBlock.Txs)
	if err != nil {
		return nil, err
	}

	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       b.genesisBlockIdentifier,
			ParentBlockIdentifier: b.genesisBlockIdentifier,
			Transactions:          transactions,
			Timestamp:             genesisBlock.Timestamp,
			Metadata: map[string]interface{}{
				pmapper.MetadataMessage: genesisBlock.Message,
			},
		},
	}, err
}

// BlockTransaction implements the /block/transaction endpoint.
func (b *Backend) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	block, err := b.getBlockDetails(ctx, request.BlockIdentifier.Index, request.BlockIdentifier.Hash)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	transactions, err := b.parseTransactions(ctx, request.NetworkIdentifier, block.Txs)
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
	dependencyTxs, err := b.fetchDependencyTxs(ctx, txs)
	if err != nil {
		return nil, err
	}

	parser, err := b.newTxParser(ctx, networkIdentifier, dependencyTxs)
	if err != nil {
		return nil, err
	}

	var transactions []*types.Transaction
	for _, tx := range txs {
		err = common.InitializeTx(b.codecVersion, b.codec, *tx)
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

func (b *Backend) fetchDependencyTxs(ctx context.Context, txs []*platformvm.Tx) (map[string]*pmapper.DependencyTx, error) {
	dependencyTxIDs := []ids.ID{}

	for _, tx := range txs {
		inputTxsIds, err := pmapper.GetDependencyTxIDs(tx.UnsignedTx)
		if err != nil {
			return nil, err
		}
		dependencyTxIDs = append(dependencyTxIDs, inputTxsIds...)
	}

	dependencyTxChan := make(chan *pmapper.DependencyTx, len(dependencyTxIDs))
	eg, ctx := errgroup.WithContext(ctx)

	dependencyTxs := make(map[string]*pmapper.DependencyTx)
	for i := range dependencyTxIDs {
		txID := dependencyTxIDs[i]
		eg.Go(func() error {
			return b.fetchDependencyTx(ctx, txID, dependencyTxChan)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	close(dependencyTxChan)

	for dTx := range dependencyTxChan {
		dependencyTxs[dTx.ID.String()] = dTx
	}

	return dependencyTxs, nil
}

func (b *Backend) fetchDependencyTx(ctx context.Context, txID ids.ID, out chan *pmapper.DependencyTx) error {
	txBytes, err := b.pClient.GetTx(ctx, txID)
	if err != nil {
		return err
	}

	var tx platformvm.Tx
	_, err = b.codec.Unmarshal(txBytes, &tx)
	if err != nil {
		return err
	}

	err = common.InitializeTx(0, platformvm.Codec, tx)
	if err != nil {
		return err
	}

	utxoBytes, err := b.pClient.GetRewardUTXOs(ctx, &api.GetTxArgs{
		TxID:     txID,
		Encoding: formatting.Hex,
	})

	utxos := []*avax.UTXO{}
	for _, bytes := range utxoBytes {
		utxo := avax.UTXO{}
		_, err = b.codec.Unmarshal(bytes, &utxo)
		if err != nil {
			return err
		}
		utxos = append(utxos, &utxo)
	}
	out <- &pmapper.DependencyTx{
		ID:          txID,
		Tx:          &tx,
		RewardUTXOs: utxos,
	}

	return nil
}

func (b *Backend) newTxParser(
	ctx context.Context,
	networkIdentifier *types.NetworkIdentifier,
	dependencyTxs map[string]*pmapper.DependencyTx,
) (*pmapper.TxParser, error) {
	hrp, err := mapper.GetHRP(networkIdentifier)
	if err != nil {
		return nil, err
	}

	chainIDs, err := b.getChainIDs(ctx)
	if err != nil {
		return nil, err
	}

	inputAddresses, err := pmapper.GetAccountsFromUTXOs(hrp, dependencyTxs)
	if err != nil {
		return nil, err
	}

	return pmapper.NewTxParser(false, hrp, chainIDs, inputAddresses, dependencyTxs), nil
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

func (b *Backend) getBlockDetails(ctx context.Context, index int64, hash string) (*indexer.ParsedBlock, error) {
	if index <= 0 && hash == "" {
		return nil, errMissingBlockIndexHash
	}

	var parsedBlock *indexer.ParsedBlock
	var err error
	// Extract block id from hash parameter if it is non-empty, or from index if stated
	if hash != "" {
		parsedBlock, err = b.indexerParser.ParseBlockWithHash(ctx, hash)
	} else if index > 0 {
		parsedBlock, err = b.indexerParser.ParseBlockAtIndex(ctx, uint64(index))
	}
	if err != nil {
		return nil, err
	}

	return parsedBlock, nil
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
