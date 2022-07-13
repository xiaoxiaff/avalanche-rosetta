package p

import (
	"context"
	"errors"
	"strconv"

	"github.com/ava-labs/avalanche-rosetta/service/chain/common"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
)

func (c *Backend) AccountBalance(ctx context.Context, req *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	if req.AccountIdentifier == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "account indentifier is not provided")
	}
	addr, err := address.ParseToID(req.AccountIdentifier.Address)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to convert address")
	}

	var balanceType string
	if req.AccountIdentifier.SubAccount != nil {
		balanceType = req.AccountIdentifier.SubAccount.Address
	}

	balanceResponse, err := c.pClient.GetBalance(ctx, []ids.ShortID{addr})
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get balance from input address")
	}

	var balanceValue uint64

	switch balanceType {
	case "unlocked":
		balanceValue = uint64(balanceResponse.Unlocked)
	case "locked_stakeable":
		balanceValue = uint64(balanceResponse.LockedStakeable)
	case "locked_not_stakeable":
		balanceValue = uint64(balanceResponse.LockedNotStakeable)
	default:
		balanceValue = uint64(balanceResponse.Balance)
	}

	return &types.AccountBalanceResponse{
		//TODO: return block identifier once AvalancheGo exposes an API for it
		//BlockIdentifier: ...
		Balances: []*types.Amount{
			{
				Value:    strconv.FormatUint(balanceValue, 10),
				Currency: mapper.AvaxCurrency,
			},
		},
	}, nil
}

func (c *Backend) AccountCoins(ctx context.Context, req *types.AccountCoinsRequest) (*types.AccountCoinsResponse, *types.Error) {
	if req.AccountIdentifier == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "account identifier is not provided")
	}

	addr, err := address.ParseToID(req.AccountIdentifier.Address)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to convert address")
	}

	currencyAssetIDs, wrappedErr := c.buildCurrencyAssetIDs(ctx, req)
	if err != nil {
		return nil, wrappedErr
	}

	var coins []*types.Coin

	// Used for pagination
	var startAddr ids.ShortID
	var startUTXOID ids.ID

	for {
		var utxos [][]byte

		// GetUTXOs controlled by addr
		utxos, startAddr, startUTXOID, err = c.pClient.GetUTXOs(ctx, []ids.ShortID{addr}, c.getUTXOsPageSize, startAddr, startUTXOID)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "unable to get UTXOs")
		}

		// convert raw UTXO bytes to Rosetta Coins
		coinsPage, err := c.processUtxos(currencyAssetIDs, utxos)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}

		coins = append(coins, coinsPage...)

		// Fetch next page only if there may be more UTXOs
		if len(utxos) < int(c.getUTXOsPageSize) {
			break
		}
	}

	return &types.AccountCoinsResponse{
		//TODO: return block identifier once AvalancheGo exposes an API for it
		// BlockIdentifier: ...
		Coins: common.SortUnique(coins),
	}, nil
}

func (c *Backend) buildCurrencyAssetIDs(ctx context.Context, req *types.AccountCoinsRequest) (map[ids.ID]struct{}, *types.Error) {
	currencyAssetIDs := make(map[ids.ID]struct{})
	for _, reqCurrency := range req.Currencies {
		description, err := c.pClient.GetAssetDescription(ctx, reqCurrency.Symbol)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, "unable to get asset description")
		}
		if int32(description.Denomination) != reqCurrency.Decimals {
			return nil, service.WrapError(service.ErrInvalidInput, "incorrect currency decimals")
		}
		currencyAssetIDs[description.AssetID] = struct{}{}
	}

	return currencyAssetIDs, nil
}

func (c *Backend) processUtxos(currencyAssetIDs map[ids.ID]struct{}, utxos [][]byte) ([]*types.Coin, error) {
	var coins []*types.Coin
	for _, utxoBytes := range utxos {
		utxo := avax.UTXO{}
		_, err := platformvm.Codec.Unmarshal(utxoBytes, &utxo)
		if err != nil {
			return nil, errors.New("unable to parse UTXO")
		}

		// Skip UTXO if req.Currencies is specified but it doesn't contain the UTXOs asset
		if _, ok := currencyAssetIDs[utxo.AssetID()]; len(currencyAssetIDs) > 0 && !ok {
			continue
		}

		transferableOut, ok := utxo.Out.(avax.TransferableOut)
		if !ok {
			return nil, errors.New("unable to get UTXO output")
		}

		coin := &types.Coin{
			CoinIdentifier: &types.CoinIdentifier{Identifier: utxo.UTXOID.String()},
			Amount: &types.Amount{
				Value:    strconv.FormatUint(transferableOut.Amount(), 10),
				Currency: mapper.AvaxCurrency,
			},
		}
		coins = append(coins, coin)
	}
	return coins, nil
}
