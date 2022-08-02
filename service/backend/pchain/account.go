package pchain

import (
	"context"
	"errors"
	pmapper "github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	"github.com/ava-labs/avalanchego/vms/platformvm/stakeable"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"strconv"

	"github.com/ava-labs/avalanche-rosetta/service/backend/common"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
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

	// fetch height before the balance fetch
	preHeight, err := b.pClient.GetHeight(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get height")
	}

	balanceResponse, err := b.pClient.GetBalance(ctx, []ids.ShortID{addr})
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
		balance, err := b.fetchBalance(ctx, req.AccountIdentifier.Address)
		if err != nil {
			return nil, service.WrapError(service.ErrInvalidInput, "unable to get balance from input address")
		}
		if balance < 0 {
			return nil, service.WrapError(service.ErrInvalidInput, "negative balance")
		}
		balanceValue = uint64(balance)
	}

	blockIdentifier := &types.BlockIdentifier{}
	if req.BlockIdentifier != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to fetch historical balance")
	}

	height, err := b.pClient.GetHeight(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get height")
	}
	if height != preHeight {
		return nil, service.WrapError(service.ErrNotReady, "block number changed, pls retry")
	}
	_, hash, _, err := b.getBlock(ctx, int64(height), "")
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get height")
	}
	blockIdentifier.Hash = hash
	blockIdentifier.Index = int64(height)

	return &types.AccountBalanceResponse{
		BlockIdentifier: blockIdentifier,
		Balances: []*types.Amount{
			{
				Value:    strconv.FormatUint(balanceValue, 10),
				Currency: mapper.AvaxCurrency,
			},
		},
	}, nil
}

func (b *Backend) AccountCoins(ctx context.Context, req *types.AccountCoinsRequest) (*types.AccountCoinsResponse, *types.Error) {
	if req.AccountIdentifier == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "account identifier is not provided")
	}
	addr, err := address.ParseToID(req.AccountIdentifier.Address)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to convert address")
	}

	currencyAssetIDs, wrappedErr := b.buildCurrencyAssetIDs(ctx, req)
	if err != nil {
		return nil, wrappedErr
	}

	var coins []*types.Coin

	// Used for pagination
	var startAddr ids.ShortID
	var startUTXOID ids.ID

	// fetch height before the balance fetch
	preHeight, err := b.pClient.GetHeight(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get height")
	}

	for {
		var utxos [][]byte
		var utxoType string
		if req.AccountIdentifier.SubAccount != nil {
			utxoType = req.AccountIdentifier.SubAccount.Address
		}

		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}

		if utxoType == pmapper.UTXOTypeSharedMemory {
			// it is an atomic utxo, then
			// GetAtomicUTXOs for both C and X Chain

			// C Chain
			utxos, _, _, err = b.pClient.GetAtomicUTXOs(ctx, []ids.ShortID{addr}, "C", b.getUTXOsPageSize, startAddr, startUTXOID)
			// convert raw UTXO bytes to Rosetta Coins
			coinsPage, err := b.processUtxos(currencyAssetIDs, utxos, true)
			if err != nil {
				return nil, service.WrapError(service.ErrInternalError, err)
			}
			coins = append(coins, coinsPage...)

			// X Chain
			utxos, _, _, err = b.pClient.GetAtomicUTXOs(ctx, []ids.ShortID{addr}, "X", b.getUTXOsPageSize, startAddr, startUTXOID)
			// convert raw UTXO bytes to Rosetta Coins
			coinsPage, err = b.processUtxos(currencyAssetIDs, utxos, true)
			if err != nil {
				return nil, service.WrapError(service.ErrInternalError, err)
			}
			coins = append(coins, coinsPage...)
		} else {
			// P Chain
			// GetUTXOs controlled by addr
			utxos, startAddr, startUTXOID, err = b.pClient.GetAtomicUTXOs(ctx, []ids.ShortID{addr}, "", b.getUTXOsPageSize, startAddr, startUTXOID)
			if err != nil {
				return nil, service.WrapError(service.ErrInternalError, "unable to get UTXOs")
			}

			// convert raw UTXO bytes to Rosetta Coins
			coinsPage, err := b.processUtxos(currencyAssetIDs, utxos, false)
			if err != nil {
				return nil, service.WrapError(service.ErrInternalError, err)
			}

			coins = append(coins, coinsPage...)
		}

		// Fetch next page only if there may be more UTXOs
		if len(utxos) < int(b.getUTXOsPageSize) {
			break
		}
	}

	height, err := b.pClient.GetHeight(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get height")
	}
	if height != preHeight {
		return nil, service.WrapError(service.ErrNotReady, "block number changed, pls retry")
	}
	_, hash, _, err := b.getBlock(ctx, int64(height), "")
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to get height")
	}

	return &types.AccountCoinsResponse{
		//TODO: return block identifier once AvalancheGo exposes an API for it
		// BlockIdentifier: ...
		BlockIdentifier: &types.BlockIdentifier{
			Hash:  hash,
			Index: int64(height),
		},
		Coins: common.SortUnique(coins),
	}, nil
}

func (b *Backend) buildCurrencyAssetIDs(ctx context.Context, req *types.AccountCoinsRequest) (map[ids.ID]struct{}, *types.Error) {
	currencyAssetIDs := make(map[ids.ID]struct{})
	for _, reqCurrency := range req.Currencies {
		description, err := b.pClient.GetAssetDescription(ctx, reqCurrency.Symbol)
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

func (b *Backend) processUtxos(currencyAssetIDs map[ids.ID]struct{}, utxos [][]byte, isAtomic bool) ([]*types.Coin, error) {
	var coins []*types.Coin
	for _, utxoBytes := range utxos {
		utxo := avax.UTXO{}
		_, err := b.codec.Unmarshal(utxoBytes, &utxo)
		if err != nil {
			return nil, errUnableToParseUTXO
		}

		// Skip UTXO if req.Currencies is specified but it doesn't contain the UTXOs asset
		if _, ok := currencyAssetIDs[utxo.AssetID()]; len(currencyAssetIDs) > 0 && !ok {
			continue
		}

		transferableOut, ok := utxo.Out.(avax.TransferableOut)
		if !ok {
			return nil, errUnableToGetUTXOOut
		}

		coin := &types.Coin{
			CoinIdentifier: &types.CoinIdentifier{Identifier: utxo.UTXOID.String()},
			Amount: &types.Amount{
				Value:    strconv.FormatUint(transferableOut.Amount(), 10),
				Currency: mapper.AvaxCurrency,
				Metadata: map[string]interface{}{
					pmapper.IsAtomicUTXO: isAtomic,
				},
			},
		}
		coins = append(coins, coin)
	}
	return coins, nil
}

func (b *Backend) fetchBalance(ctx context.Context, addr string) (int64, *types.Error) {
	balance := int64(0)

	parsedAddr, err := address.ParseToID(addr)
	if err != nil {
		return 0, service.WrapError(service.ErrInvalidInput, "unable to convert address")
	}

	// Used for pagination
	var startAddr ids.ShortID
	var startUTXOID ids.ID

	for {
		var utxos [][]byte

		// GetUTXOs controlled by addr
		utxos, startAddr, startUTXOID, err = b.pClient.GetUTXOs(ctx, []ids.ShortID{parsedAddr}, b.getUTXOsPageSize, startAddr, startUTXOID)
		if err != nil {
			return 0, service.WrapError(service.ErrInternalError, "unable to get UTXOs")
		}

		for _, utxoBytes := range utxos {
			utxo := avax.UTXO{}
			_, err := b.codec.Unmarshal(utxoBytes, &utxo)
			if err != nil {
				return 0, service.WrapError(service.ErrInvalidInput, "unable to parse UTXO")
			}

			outIntf := utxo.Out
			if lockedOut, ok := outIntf.(*stakeable.LockOut); ok {
				outIntf = lockedOut.TransferableOut
			}

			out, ok := outIntf.(*secp256k1fx.TransferOutput)
			if !ok {
				return 0, service.WrapError(service.ErrBlockInvalidInput, "output type assertion failed")
			}

			// ignore multisig
			if len(out.OutputOwners.Addrs) > 1 {
				continue
			}

			balance += int64(out.Amt)
		}

		// Fetch next page only if there may be more UTXOs
		if len(utxos) < int(b.getUTXOsPageSize) {
			break
		}
	}

	_, stakeUTXOs, err := b.pClient.GetStake(ctx, []ids.ShortID{parsedAddr})
	if err != nil {
		return 0, service.WrapError(service.ErrInvalidInput, "unable to get stake")
	}

	staked := int64(0)

	for _, utxoBytes := range stakeUTXOs {
		utxo := avax.TransferableOutput{}
		_, err := b.codec.Unmarshal(utxoBytes, &utxo)
		if err != nil {
			return 0, service.WrapError(service.ErrInvalidInput, "unable to parse UTXO")
		}

		outIntf := utxo.Out
		if lockedOut, ok := outIntf.(*stakeable.LockOut); ok {
			outIntf = lockedOut.TransferableOut
		}

		out, ok := outIntf.(*secp256k1fx.TransferOutput)
		if !ok {
			return 0, service.WrapError(service.ErrBlockInvalidInput, "output type assertion failed")
		}

		// ignore multisig
		if len(out.OutputOwners.Addrs) > 1 {
			continue
		}

		staked += int64(out.Amt)
	}

	// TODO (stake should be in the balance or not)
	balance += staked

	return balance, nil
}
