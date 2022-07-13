package p

import (
	"context"
	"strings"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	mocks "github.com/ava-labs/avalanche-rosetta/mocks/client"
	"github.com/ava-labs/avalanche-rosetta/service/chain/common"
)

type utxo struct {
	id     string
	amount uint64
}

var utxos = []utxo{
	{"NGcWaGCzBUtUsD85wDuX1DwbHFkvMHwJ9tDFiN7HCCnVcB9B8:0", 1000000000},
	{"pyQfA1Aq9vLaDETjeQe5DAwVxr2KAYdHg4CHzawmaj9oA6ppn:0", 2000000000},
}

func TestAccountBalance(t *testing.T) {
	pChainMock := &mocks.PChainClient{}
	ctx := context.Background()
	pChainMock.Mock.On("GetNetworkID", ctx).Return(uint32(5), nil)

	service, _ := NewBackend(ctx, pChainMock, ids.Empty, nil)

	t.Run("Account Balance Test", func(t *testing.T) {
		pChainAddr := "P-fuji1wmd9dfrqpud6daq0cde47u0r7pkrr46ep60399"
		addr, _ := address.ParseToID(pChainAddr)
		mockGetBalanceResponse := &platformvm.GetBalanceResponse{
			Balance:            1000000000,
			Unlocked:           0,
			LockedStakeable:    0,
			LockedNotStakeable: 0,
			UTXOIDs:            []*avax.UTXOID{},
		}
		pChainMock.Mock.On("GetBalance", ctx, []ids.ShortID{addr}).Return(mockGetBalanceResponse, nil)

		resp, err := service.AccountBalance(
			ctx,
			&types.AccountBalanceRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Network: mapper.FujiNetwork,
					SubNetworkIdentifier: &types.SubNetworkIdentifier{
						Network: mapper.PChainNetworkIdentifier,
					},
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: pChainAddr,
				},
				Currencies: []*types.Currency{
					mapper.AvaxCurrency,
				},
			},
		)

		expected := &types.AccountBalanceResponse{
			Balances: []*types.Amount{
				{
					Value:    "1000000000",
					Currency: mapper.AvaxCurrency,
				},
			},
		}

		assert.Nil(t, err)
		assert.Equal(
			t,
			expected.Balances,
			resp.Balances,
		)
	})
}

func TestAccountCoins(t *testing.T) {
	pChainMock := &mocks.PChainClient{}
	ctx := context.Background()
	pChainMock.Mock.On("GetNetworkID", ctx).Return(uint32(5), nil)

	service, _ := NewBackend(ctx, pChainMock, ids.Empty, nil)

	t.Run("Account Coins Test", func(t *testing.T) {
		pChainAddr := "P-fuji1wmd9dfrqpud6daq0cde47u0r7pkrr46ep60399"

		// Mock on GetAssetDescription
		mockAssetDescription := &avm.GetAssetDescriptionReply{
			Name:         "Avalanche",
			Symbol:       mapper.AvaxCurrency.Symbol,
			Denomination: 9,
		}
		pChainMock.Mock.On("GetAssetDescription", ctx, mapper.AvaxCurrency.Symbol).Return(mockAssetDescription, nil)

		// Mock on GetUTXOs
		utxo0Bytes := makeUtxoBytes(t, service, utxos[0].id, utxos[0].amount)
		utxo1Bytes := makeUtxoBytes(t, service, utxos[1].id, utxos[1].amount)

		addr, errp := address.ParseToID(pChainAddr)
		assert.Nil(t, errp)

		var startAddr ids.ShortID
		var startUTXOID ids.ID
		utxo1idShortID, _ := ids.FromString(strings.Split(utxos[1].id, ":")[0])
		pChainMock.Mock.On("GetUTXOs", ctx, []ids.ShortID{addr}, uint32(1024), startAddr, startUTXOID).
			Return([][]byte{utxo0Bytes, utxo1Bytes}, addr, utxo1idShortID, nil)

		resp, err := service.AccountCoins(
			ctx,
			&types.AccountCoinsRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Network: mapper.FujiNetwork,
					SubNetworkIdentifier: &types.SubNetworkIdentifier{
						Network: mapper.PChainNetworkIdentifier,
					},
				},
				AccountIdentifier: &types.AccountIdentifier{
					Address: pChainAddr,
				},
				Currencies: []*types.Currency{
					mapper.AvaxCurrency,
				},
			})

		expected := &types.AccountCoinsResponse{
			Coins: []*types.Coin{
				{
					CoinIdentifier: &types.CoinIdentifier{
						Identifier: "NGcWaGCzBUtUsD85wDuX1DwbHFkvMHwJ9tDFiN7HCCnVcB9B8:0",
					},
					Amount: &types.Amount{
						//Value:    "9000000",
						Value:    "1000000000",
						Currency: mapper.AvaxCurrency,
					},
				},
				{
					CoinIdentifier: &types.CoinIdentifier{
						Identifier: "pyQfA1Aq9vLaDETjeQe5DAwVxr2KAYdHg4CHzawmaj9oA6ppn:0",
					},
					Amount: &types.Amount{
						//Value:    "2877137500",
						Value:    "2000000000",
						Currency: mapper.AvaxCurrency,
					},
				},
			},
		}

		assert.Nil(t, err)
		assert.Equal(
			t,
			expected,
			resp,
		)
	})
}

func makeUtxoBytes(t *testing.T, backend *Backend, utxoIdStr string, amount uint64) []byte {
	utxoId, err := common.DecodeUTXOID(utxoIdStr)
	if err != nil {
		t.Fail()
		return nil
	}

	utxoBytes, err := backend.codec.Marshal(0, &avax.UTXO{
		UTXOID: *utxoId,
		Out:    &secp256k1fx.TransferOutput{Amt: amount},
	})
	if err != nil {
		t.Fail()
	}

	return utxoBytes
}
