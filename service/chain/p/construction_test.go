package p

import (
	"encoding/hex"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	mocks "github.com/ava-labs/avalanche-rosetta/mocks/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
	"testing"

	"context"
)

func TestConstructionDerive(t *testing.T) {
	service := NewBackend(&mocks.PChainClient{}, nil)

	t.Run("p-chain address", func(t *testing.T) {
		src := "02e0d4392cfa224d4be19db416b3cf62e90fb2b7015e7b62a95c8cb490514943f6"
		b, _ := hex.DecodeString(src)

		resp, err := service.ConstructionDerive(
			context.Background(),
			&types.ConstructionDeriveRequest{
				NetworkIdentifier: &types.NetworkIdentifier{
					Network: mapper.FujiNetwork,
					SubNetworkIdentifier: &types.SubNetworkIdentifier{
						Network: mapper.PChainNetworkIdentifier,
					},
				},
				PublicKey: &types.PublicKey{
					Bytes:     b,
					CurveType: types.Secp256k1,
				},
			},
		)
		assert.Nil(t, err)
		assert.Equal(
			t,
			"P-fuji15f9g0h5xkr5cp47n6u3qxj6yjtzzzrdr23a3tl",
			resp.AccountIdentifier.Address,
		)
	})
}

func TestConstructionHash(t *testing.T) {
	service := NewBackend(&mocks.PChainClient{})

	t.Run("P-chain valid transaction", func(t *testing.T) {
		signed := "0x00000000000e000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b8724b400000000000000000000000100000001790b9fc4f62b8eb2d2cf0177bda1ecc882a2d19e000000018be2098b614618321c855b6c7ca1cce33006902727d2a05f3ae7d5b18c14e24f000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000007721eeb4000000010000000000000000d325c150d0fec89b706ab5fd75ae7506a9912a9e00000000629a465500000000629b97d5000000003b9aca00000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b9aca0000000000000000000000000100000001790b9fc4f62b8eb2d2cf0177bda1ecc882a2d19e0000000b00000000000000000000000100000001e35e8550c1f09e1d3f6b97292eed8a1a76dcdd8a000000010000000900000001ebd189ad5e808ac24b69d8548980759067ce3b8b8caf9ece3ce3d032c5ec433d59e3767ffbbb2f9940894dd2eb96e6f93942b5535137a46097d124571b8dcf5700f323bc66"

		resp, err := service.ConstructionHash(context.Background(), &types.ConstructionHashRequest{
			NetworkIdentifier: &types.NetworkIdentifier{
				SubNetworkIdentifier: &types.SubNetworkIdentifier{
					Network: mapper.PChainNetworkIdentifier,
				},
			},
			SignedTransaction: signed,
		})
		assert.Nil(t, err)
		assert.Equal(
			t,
			"etWqwTN1YwhakxLnMDp7q6yaf4m7VJu4uB4vC4fEtNrFe9sDy",
			resp.TransactionIdentifier.Hash,
		)
	})
}
