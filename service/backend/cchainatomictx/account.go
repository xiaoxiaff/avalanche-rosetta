package cchainatomictx

import (
	"context"
	"errors"
	"strconv"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/utils/math"

	"github.com/ava-labs/avalanche-rosetta/service/backend/common"

	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
)

var (
	errUnableToParseUTXO  = errors.New("unable to parse UTXO")
	errUnableToGetUTXOOut = errors.New("unable to get UTXO output")
)

func (b *Backend) AccountBalance(ctx context.Context, req *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	if req.AccountIdentifier == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "account identifier is not provided")
	}
	preHeader, err := service.BlockHeaderFromInput(ctx, b.cClient, req.BlockIdentifier)
	if err != nil {
		return nil, err
	}

	coins, wrappedErr := b.getAccountCoins(ctx, req.AccountIdentifier.Address)
	if wrappedErr != nil {
		return nil, wrappedErr
	}

	postHeader, terr := service.BlockHeaderFromInput(ctx, b.cClient, req.BlockIdentifier)
	if err != nil {
		return nil, err
	}

	if postHeader.Number.Int64() != preHeader.Number.Int64() {
		return nil, service.WrapError(service.ErrInternalError, "new block added while fetching utxos")
	}

	var balanceValue uint64

	for _, coin := range coins {
		amountValue, err := types.AmountValue(coin.Amount)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "unable to extract amount from UTXO")
		}

		balanceValue, err = math.Add64(balanceValue, amountValue.Uint64())
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "overflow while calculating balance")
		}
	}

	if terr != nil {
		return nil, terr
	}
	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: postHeader.Number.Int64(),
			Hash:  postHeader.Hash().String(),
		},
		Balances: []*types.Amount{
			{
				Value:    strconv.FormatUint(balanceValue, 10),
				Currency: mapper.AtomicAvaxCurrency,
			},
		},
	}, nil
}

func (b *Backend) AccountCoins(ctx context.Context, req *types.AccountCoinsRequest) (*types.AccountCoinsResponse, *types.Error) {
	if req.AccountIdentifier == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "account identifier is not provided")
	}
	coins, wrappedErr := b.getAccountCoins(ctx, req.AccountIdentifier.Address)
	if wrappedErr != nil {
		return nil, wrappedErr
	}

	// get the tip
	blockHeader, err := b.cClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get tip")
	}

	if blockHeader == nil {
		return nil, service.WrapError(service.ErrClientError, "latest block not found")
	}

	return &types.AccountCoinsResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: blockHeader.Number.Int64(),
			Hash:  blockHeader.Hash().String(),
		},
		Coins: common.SortUnique(coins),
	}, nil
}

func (b *Backend) getAccountCoins(ctx context.Context, address string) ([]*types.Coin, *types.Error) {
	var coins []*types.Coin
	sourceChains := []string{
		mapper.PChainNetworkIdentifier,
		mapper.XChainNetworkIdentifier,
	}

	for _, chain := range sourceChains {
		chainCoins, wrappedErr := b.fetchCoinsFromChain(ctx, address, chain)
		if wrappedErr != nil {
			return nil, wrappedErr
		}
		coins = append(coins, chainCoins...)
	}

	return coins, nil
}

func (b *Backend) fetchCoinsFromChain(ctx context.Context, address string, sourceChain string) ([]*types.Coin, *types.Error) {
	var coins []*types.Coin

	// Used for pagination
	var lastUtxoIndex api.Index

	for {

		// GetUTXOs controlled by addr
		utxos, newUtxoIndex, err := b.cClient.GetAtomicUTXOs(ctx, []string{address}, sourceChain, b.getUTXOsPageSize, lastUtxoIndex.Address, lastUtxoIndex.UTXO)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "unable to get UTXOs")
		}

		// convert raw UTXO bytes to Rosetta Coins
		coinsPage, err := b.processUtxos(sourceChain, utxos)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}

		coins = append(coins, coinsPage...)

		// Fetch next page only if there may be more UTXOs
		if len(utxos) < int(b.getUTXOsPageSize) {
			break
		}

		lastUtxoIndex = newUtxoIndex
	}

	return coins, nil
}

func (b *Backend) processUtxos(sourceChain string, utxos [][]byte) ([]*types.Coin, error) {
	var coins []*types.Coin
	for _, utxoBytes := range utxos {
		utxo := avax.UTXO{}
		_, err := platformvm.Codec.Unmarshal(utxoBytes, &utxo)
		if err != nil {
			return nil, errUnableToParseUTXO
		}

		transferableOut, ok := utxo.Out.(avax.TransferableOut)
		if !ok {
			return nil, errUnableToGetUTXOOut
		}

		coin := &types.Coin{
			CoinIdentifier: &types.CoinIdentifier{Identifier: utxo.UTXOID.String()},
			Amount: &types.Amount{
				Value:    strconv.FormatUint(transferableOut.Amount(), 10),
				Currency: mapper.AtomicAvaxCurrency,
				Metadata: map[string]interface{}{
					"source_chain": sourceChain,
				},
			},
		}
		coins = append(coins, coin)
	}
	return coins, nil
}
