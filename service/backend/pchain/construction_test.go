package pchain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	mocks "github.com/ava-labs/avalanche-rosetta/mocks/client"
)

func TestConstructionDerive(t *testing.T) {
	pChainMock := &mocks.PChainClient{}
	ctx := context.Background()
	pChainMock.Mock.On("GetNetworkID", ctx).Return(uint32(5), nil)
	service, _ := NewBackend(ctx, pChainMock, ids.Empty, nil)

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
	pChainMock := &mocks.PChainClient{}
	ctx := context.Background()
	pChainMock.Mock.On("GetNetworkID", ctx).Return(uint32(5), nil)
	service, _ := NewBackend(ctx, pChainMock, ids.Empty, nil)

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

// https://explorer-xp.avax-test.network/tx/2boVqhWaZ7M1YmnCe6JscWJESK1LVpcGq5quGpoX4HtLdr1RHN
func TestConstructionCombine(t *testing.T) {
	pChainMock := &mocks.PChainClient{}
	ctx := context.Background()
	pChainMock.Mock.On("GetNetworkID", ctx).Return(uint32(5), nil)
	service, _ := NewBackend(ctx, pChainMock, ids.Empty, nil)

	pChainNetworkIdentifier := &types.NetworkIdentifier{
		Network:    "Fuji",
		Blockchain: "Avalanche",
		SubNetworkIdentifier: &types.SubNetworkIdentifier{
			Network: mapper.PChainNetworkIdentifier,
		},
	}

	t.Run("combine P chain tx", func(t *testing.T) {
		sig, _ := hex.DecodeString("72306e39e3ec145a43b40707040dc6cd169deafbb2629a350f9e4ae35cda4db16f7b1b84ebb3dc4983bb5fb1681c481ed130a6dec5cf0975b6c45ce58749913000")
		unsignedTx, _ := hex.DecodeString("00000000000c000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b8724b40000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf00000001614bb3fa8b0fd6f115b8bdff3e04975b1e33a323770b3e556373a2efbaa3bd34000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000007721eeb400000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062a11c640000000062a316a4000000003b9aca00000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b9aca000000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf0000000b0000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf000f424000000000")
		signedTx := "0x00000000000c000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b8724b40000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf00000001614bb3fa8b0fd6f115b8bdff3e04975b1e33a323770b3e556373a2efbaa3bd34000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000007721eeb400000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062a11c640000000062a316a4000000003b9aca00000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b9aca000000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf0000000b0000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf000f424000000001000000090000000172306e39e3ec145a43b40707040dc6cd169deafbb2629a350f9e4ae35cda4db16f7b1b84ebb3dc4983bb5fb1681c481ed130a6dec5cf0975b6c45ce587499130007a0f938a"

		resp, err := service.ConstructionCombine(
			context.Background(),
			&types.ConstructionCombineRequest{
				NetworkIdentifier:   pChainNetworkIdentifier,
				UnsignedTransaction: string(unsignedTx),
				Signatures: []*types.Signature{
					{Bytes: sig},
				},
			},
		)
		assert.Nil(t, err)
		assert.Equal(
			t,
			signedTx,
			resp.SignedTransaction,
		)
	})

	t.Run("combine P chain import tx", func(t *testing.T) {
		sig, _ := hex.DecodeString("292ca729ffbfca3ffe28bdea0f22fac34b1f5cd7d888e5432e72dd6a012b045f469a352f1b238f7c0700cafa8e238e2de2c1de62c1d86745c619bb20c3fabd1401")
		unsignedTx, _ := hex.DecodeString("000000000011000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000007590d12800000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d500000001952e0397dafcf7332370878c007ac07f3005b7faf6731d8523d6a124297dbc05000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa000000050000000075a01368000000010000000000000000")
		signedTx := "0x000000000011000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000007590d12800000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d500000001952e0397dafcf7332370878c007ac07f3005b7faf6731d8523d6a124297dbc05000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa000000050000000075a013680000000100000000000000010000000900000001292ca729ffbfca3ffe28bdea0f22fac34b1f5cd7d888e5432e72dd6a012b045f469a352f1b238f7c0700cafa8e238e2de2c1de62c1d86745c619bb20c3fabd1401129cc8ef"

		resp, err := service.ConstructionCombine(
			context.Background(),
			&types.ConstructionCombineRequest{
				NetworkIdentifier:   pChainNetworkIdentifier,
				UnsignedTransaction: string(unsignedTx),
				Signatures: []*types.Signature{
					{Bytes: sig},
				},
			},
		)
		assert.Nil(t, err)
		assert.Equal(
			t,
			signedTx,
			resp.SignedTransaction,
		)
	})

	t.Run("combine P chain export tx", func(t *testing.T) {
		sig, _ := hex.DecodeString("23740f4487b97b82c05f30f1ab6d78487315ffba0a6bcec5eb8c2a3a5a06ca96527296f0a2d40559933a7a08122ad043c3d8e8df212118751324127daf9d006300")
		unsignedTx, _ := hex.DecodeString("000000000012000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000000089544000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000001226fd389f04700af8651a50a631474419ffd71b4c1b03af23d69ab61cedc2a92000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000007721eeb40000000100000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d5000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000007689583400000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000000")
		signedTx := "0x000000000012000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000000089544000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000001226fd389f04700af8651a50a631474419ffd71b4c1b03af23d69ab61cedc2a92000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000007721eeb40000000100000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d5000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000007689583400000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000001000000090000000123740f4487b97b82c05f30f1ab6d78487315ffba0a6bcec5eb8c2a3a5a06ca96527296f0a2d40559933a7a08122ad043c3d8e8df212118751324127daf9d00630072306cf9"

		resp, err := service.ConstructionCombine(
			context.Background(),
			&types.ConstructionCombineRequest{
				NetworkIdentifier:   pChainNetworkIdentifier,
				UnsignedTransaction: string(unsignedTx),
				Signatures: []*types.Signature{
					{Bytes: sig},
				},
			},
		)
		assert.Nil(t, err)
		assert.Equal(
			t,
			signedTx,
			resp.SignedTransaction,
		)
	})
}

func TestConstructionTransaction(t *testing.T) {
	var (
		pc                      = &mocks.PChainClient{}
		ctx                     = context.Background()
		assetID, _              = ids.FromString("U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK")
		networkID               = uint32(5)
		pChainID                = ids.Empty
		cChainID, _             = ids.FromString("yH8D7ThNJkxmtkuv2jgBa4P1Rn3Qpr4pPr7QYNfcdoS6k6HWp")
		pChainNetworkIdentifier = &types.NetworkIdentifier{
			Network:    "Fuji",
			Blockchain: "Avalanche",
			SubNetworkIdentifier: &types.SubNetworkIdentifier{
				Network: mapper.PChainNetworkIdentifier,
			},
		}
	)
	pc.On("GetNetworkID", ctx).Return(networkID, nil)
	service, _ := NewBackend(context.Background(), pc, assetID, nil)

	pc.On("GetNetworkID", ctx).Return(networkID, nil)
	pc.On("GetBlockchainID", ctx, mapper.PChainNetworkIdentifier).Return(pChainID, nil)
	pc.On("GetBlockchainID", ctx, mapper.CChainNetworkIdentifier).Return(cChainID, nil)

	t.Run("construct p-chain import tx", func(t *testing.T) {
		intent := `[{"operation_identifier":{"index":0},"type":"IMPORT_AVAX","account":{"address":"C-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"-1999712500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"z8aoQdHbAgaj4uWToafsuMZLvKzCt6bSsXbN2Qtyte6GyGbvt:0"},"coin_action":"coin_spent"},"metadata":{"type":"IMPORT","sig_indices":[0]}},{"operation_identifier":{"index":1},"type":"IMPORT_AVAX","account":{"address":"C-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"-1973425000","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"28hbawmoHaWkmAKjgueWF18LrptCCCfprxaZeCf9QuBTCcLWEd:0"},"coin_action":"coin_spent"},"metadata":{"type":"IMPORT","sig_indices":[0]}},{"operation_identifier":{"index":2},"type":"IMPORT_AVAX","account":{"address":"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"3972137500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_action":"coin_created"},"metadata":{"type":"OUTPUT","output_owners":"0x000000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb818ac13004d"}}]`
		var ops []*types.Operation
		assert.NoError(t, json.Unmarshal([]byte(intent), &ops))

		preprocessResp, err := service.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata: map[string]interface{}{
					"source_chain": mapper.CChainNetworkIdentifier,
				},
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, preprocessResp)

		assert.Equal(t, "IMPORT_AVAX", preprocessResp.Options["type"])

		metadataResp, err := service.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           preprocessResp.Options,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, metadataResp)

		payloadResp, err := service.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata:          metadataResp.Metadata,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, payloadResp)
		assert.Equal(
			t,
			"000000000011000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000700000000ecc2021c00000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d50000000281b8ea7b7282685c79494712a633f9862d342c8dcb0431f88550b39ce4c46a40000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000500000000773130f40000000100000000952e0397dafcf7332370878c007ac07f3005b7faf6731d8523d6a124297dbc05000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa000000050000000075a01368000000010000000000000000",
			hex.EncodeToString([]byte(payloadResp.UnsignedTransaction)),
		)
	})

	t.Run("construct p-chain export tx", func(t *testing.T) {
		intent := `[{"operation_identifier":{"index":0},"type":"EXPORT_AVAX","account":{"address":"P-fuji1s8sg8392yl9sgcezhetkx0257k37pnd03662yv"},"amount":{"value":"-1998712500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"9EFAzbVcab16wRdf48pWXExqTwgPWfu36x3AoqJp2VD3ahrGU:0"},"coin_action":"coin_spent"},"metadata":{"type":"INPUT","sig_indices":[0]}},{"operation_identifier":{"index":1},"type":"EXPORT_AVAX","account":{"address":"P-fuji1s8sg8392yl9sgcezhetkx0257k37pnd03662yv"},"amount":{"value":"-998712500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"2boVqhWaZ7M1YmnCe6JscWJESK1LVpcGq5quGpoX4HtLdr1RHN:0"},"coin_action":"coin_spent"},"metadata":{"type":"INPUT","sig_indices":[0]}},{"operation_identifier":{"index":2},"type":"EXPORT_AVAX","account":{"address":"P-fuji1s8sg8392yl9sgcezhetkx0257k37pnd03662yv"},"amount":{"value":"-1000000000","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"2boVqhWaZ7M1YmnCe6JscWJESK1LVpcGq5quGpoX4HtLdr1RHN:1"},"coin_action":"coin_spent"},"metadata":{"type":"INPUT","sig_indices":[0]}},{"operation_identifier":{"index":3},"type":"EXPORT_AVAX","account":{"address":"P-fuji1s8sg8392yl9sgcezhetkx0257k37pnd03662yv"},"amount":{"value":"-391492","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"2boVqhWaZ7M1YmnCe6JscWJESK1LVpcGq5quGpoX4HtLdr1RHN:2"},"coin_action":"coin_spent"},"metadata":{"type":"INPUT","sig_indices":[0]}},{"operation_identifier":{"index":4},"type":"EXPORT_AVAX","account":{"address":"C-fuji1s8sg8392yl9sgcezhetkx0257k37pnd03662yv"},"amount":{"value":"996816492","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_action":"coin_created"},"metadata":{"type":"OUTPUT","output_owners":"0x00000000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf3c9becca"}},{"operation_identifier":{"index":5},"type":"EXPORT_AVAX","account":{"address":"C-fuji1s8sg8392yl9sgcezhetkx0257k37pnd03662yv"},"amount":{"value":"3000000000","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_action":"coin_created"},"metadata":{"type":"EXPORT","output_owners":"0x00000000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf3c9becca"}}]`

		var ops []*types.Operation
		assert.NoError(t, json.Unmarshal([]byte(intent), &ops))

		preprocessResp, err := service.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata: map[string]interface{}{
					"destination_chain": mapper.CChainNetworkIdentifier,
				},
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, preprocessResp)

		assert.Equal(t, "EXPORT_AVAX", preprocessResp.Options["type"])

		metadataResp, err := service.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           preprocessResp.Options,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, metadataResp)

		payloadResp, err := service.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata:          metadataResp.Metadata,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, payloadResp)
		assert.Equal(
			t,
			"000000000012000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b6a366c0000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf0000000412aef85d117564ab3410b1587a24afd497d93e7bf4e72dba094b7858f1b2ff67000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000007721eeb40000000100000000d2b7b1f46edf25528962c7d4115bb47972bccb674b51b16d3c058f5e7a0f938a000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000003b8724b40000000100000000d2b7b1f46edf25528962c7d4115bb47972bccb674b51b16d3c058f5e7a0f938a000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000003b9aca000000000100000000d2b7b1f46edf25528962c7d4115bb47972bccb674b51b16d3c058f5e7a0f938a000000023d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000000005f9440000000100000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d5000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000700000000b2d05e000000000000000000000000010000000181e083c4aa27cb046322be57633d54f5a3e0cdaf00000000",
			hex.EncodeToString([]byte(payloadResp.UnsignedTransaction)),
		)
	})

	t.Run("construct p-chain add validator tx", func(t *testing.T) {
		intent := `[{"operation_identifier":{"index":0},"type":"ADD_VALIDATOR","account":{"address":"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"-3972137500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"2MmPTid7Errf6MdDqgUPxhuhtoc9yhkn5uC4vwsqRXCJhVYt1h:0"},"coin_action":"coin_spent"},"metadata":{"type":"INPUT","sig_indices":[0]}},{"operation_identifier":{"index":1},"type":"ADD_VALIDATOR","account":{"address":"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"2972137500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_action":"coin_create"},"metadata":{"type":"OUTPUT","output_owners":"0x000000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb818ac13004d"}},{"operation_identifier":{"index":2},"type":"ADD_VALIDATOR","account":{"address":"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"1000000000","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_action":"coin_create"},"metadata":{"type":"STAKE","output_owners":"0x000000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb818ac13004d"}}]`
		var ops []*types.Operation
		assert.NoError(t, json.Unmarshal([]byte(intent), &ops))

		preprocessResp, err := service.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata: map[string]interface{}{
					"node_id":          "NodeID-Bvsx89JttQqhqdgwtizAPoVSNW74Xcr2S",
					"start":            1656460045,
					"end":              1656589645,
					"weight":           1000000000,
					"shares":           1000000,
					"locktime":         0,
					"threshold":        1,
					"reward_addresses": []string{"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},
				},
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, preprocessResp)

		assert.Equal(t, "ADD_VALIDATOR", preprocessResp.Options["type"])

		metadataResp, err := service.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           preprocessResp.Options,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, metadataResp)

		payloadResp, err := service.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata:          metadataResp.Metadata,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, payloadResp)
		assert.Equal(
			t,
			"00000000000c000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000700000000b127381c00000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000001b2d8a36998be5b19f468fbf573501cd0c93e9a7b5fb8edb2da54a473fa70ea64000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000500000000ecc2021c00000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062bb930d0000000062bd8d4d000000003b9aca00000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b9aca0000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb8180000000b00000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb818000f424000000000",
			hex.EncodeToString([]byte(payloadResp.UnsignedTransaction)),
		)
	})

	t.Run("construct p-chain add delegator tx", func(t *testing.T) {
		intent := `[{"operation_identifier":{"index":0},"type":"ADD_DELEGATOR","account":{"address":"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"-3972137500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_identifier":{"identifier":"2MmPTid7Errf6MdDqgUPxhuhtoc9yhkn5uC4vwsqRXCJhVYt1h:0"},"coin_action":"coin_spent"},"metadata":{"type":"INPUT","sig_indices":[0]}},{"operation_identifier":{"index":1},"type":"ADD_DELEGATOR","account":{"address":"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"2972137500","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_action":"coin_create"},"metadata":{"type":"OUTPUT","output_owners":"0x000000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb818ac13004d"}},{"operation_identifier":{"index":2},"type":"ADD_DELEGATOR","account":{"address":"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},"amount":{"value":"1000000000","currency":{"symbol":"AVAX","decimals":18}},"coin_change":{"coin_action":"coin_create"},"metadata":{"type":"STAKE","output_owners":"0x000000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb818ac13004d"}}]`
		var ops []*types.Operation
		assert.NoError(t, json.Unmarshal([]byte(intent), &ops))

		preprocessResp, err := service.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata: map[string]interface{}{
					"node_id":          "NodeID-Bvsx89JttQqhqdgwtizAPoVSNW74Xcr2S",
					"start":            1656460654,
					"end":              1656547054,
					"weight":           1000000000,
					"locktime":         0,
					"threshold":        1,
					"reward_addresses": []string{"P-fuji1qy8nsuzr9ee6fuuzsmmdwv67hrsuawqcz4cz89"},
				},
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, preprocessResp)

		assert.Equal(t, "ADD_DELEGATOR", preprocessResp.Options["type"])

		metadataResp, err := service.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           preprocessResp.Options,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, metadataResp)

		payloadResp, err := service.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        ops,
				Metadata:          metadataResp.Metadata,
			},
		)
		assert.Nil(t, err)
		assert.NotNil(t, payloadResp)
		assert.Equal(
			t,
			"00000000000e000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000700000000b127381c00000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000001b2d8a36998be5b19f468fbf573501cd0c93e9a7b5fb8edb2da54a473fa70ea64000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000500000000ecc2021c00000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062bb956e0000000062bce6ee000000003b9aca00000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b9aca0000000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb8180000000b00000000000000000000000100000001010f3870432e73a4f38286f6d7335eb8e1ceb81800000000",
			hex.EncodeToString([]byte(payloadResp.UnsignedTransaction)),
		)
	})
}
