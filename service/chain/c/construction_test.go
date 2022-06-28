package c

import (
	"encoding/hex"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	mocks "github.com/ava-labs/avalanche-rosetta/mocks/client"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/coinbase/rosetta-sdk-go/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/big"
	"testing"

	"context"
)

var (
	networkIdentifier = &types.NetworkIdentifier{
		Blockchain: service.BlockchainName,
		Network:    mapper.FujiNetwork,
	}

	cAccountIdentifier = &types.AccountIdentifier{Address: "0x3158e80abD5A1e1aa716003C9Db096792C379621"}
	pAccountIdentifier = &types.AccountIdentifier{Address: "P-fuji1wmd9dfrqpud6daq0cde47u0r7pkrr46ep60399"}

	// Export Tx
	unsignedExportTx = "0x000000000001000000057fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac1" +
		"0d50000000000000000000000000000000000000000000000000000000000000000000000013158e80abd5a1e1aa716003c9db096792" +
		"c37962100000000009896803d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa000000000000003000000" +
		"0013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa000000070000000000944dd200000000000000000" +
		"00000010000000176da56a4600f1ba6f40fc3735f71e3f06c31d7590000000024739402"
	unsignedExportTxHash, _    = hex.DecodeString("75afdcba5bf36457ba9edd65b07f40dcd3111d3c98a53550025af931b7500a7b")
	signedExportTxSignature, _ = hex.DecodeString("2acfc2cedd3c42978728518b13cc84a64f23784af591516e8dfe0cce544bc63" +
		"c370ca6d64b2550f12f56a800b8a73ff8573131bf54e584de38c91fc14dd7336801")
	signedExportTx = "0x000000000001000000057fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac1" +
		"0d50000000000000000000000000000000000000000000000000000000000000000000000013158e80abd5a1e1aa716003c9db096792" +
		"c37962100000000009896803d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa000000000000003000000" +
		"0013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa000000070000000000944dd200000000000000000" +
		"00000010000000176da56a4600f1ba6f40fc3735f71e3f06c31d7590000000100000009000000012acfc2cedd3c42978728518b13cc8" +
		"4a64f23784af591516e8dfe0cce544bc63c370ca6d64b2550f12f56a800b8a73ff8573131bf54e584de38c91fc14dd733680149056c11"

	// Import Tx
	unsignedImportTx = "0x000000000000000000057fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac1" +
		"0d500000000000000000000000000000000000000000000000000000000000000000000000288ae5dd070e6d74f16c26358cd4a8f437" +
		"46d4d338b5b75b668741c6d95816af5000000023d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000" +
		"0050000000000e4e1c00000000100000000b9a824340e1b94f27500cdfcbf8eaa9d4ee5e57b2823cb8b158de17689916c74000000013" +
		"d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000500000000004c4b40000000010000000000000" +
		"0013158e80abd5a1e1aa716003c9db096792c37962100000000012c7a123d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c9" +
		"6f7d28f61bbe2aa00000000c67b0534"
	unsignedImportTxHash, _    = hex.DecodeString("33f98143f7f061e262e0fabca57b7f0dc110a79073ed263fc900ebdd0c1fe6fc")
	signedImportTxSignature, _ = hex.DecodeString("a06d20d1d175b1e1d2b6e647ab5321717967de7e9367c28df8c0e20634ec782" +
		"7019fe38e8d4f123f8e5286f3236db8dbb419e264628e2f17330a6c8da60d342401")
	signedImportTx = "0x000000000000000000057fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac" +
		"10d500000000000000000000000000000000000000000000000000000000000000000000000288ae5dd070e6d74f16c26358cd4a8f4" +
		"3746d4d338b5b75b668741c6d95816af5000000023d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00" +
		"0000050000000000e4e1c00000000100000000b9a824340e1b94f27500cdfcbf8eaa9d4ee5e57b2823cb8b158de17689916c7400000" +
		"0013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000500000000004c4b400000000100000000" +
		"000000013158e80abd5a1e1aa716003c9db096792c37962100000000012c7a123d9bdac0ed1d761330cf680efdeb1a42159eb387d6d" +
		"2950c96f7d28f61bbe2aa000000020000000900000001a06d20d1d175b1e1d2b6e647ab5321717967de7e9367c28df8c0e20634ec78" +
		"27019fe38e8d4f123f8e5286f3236db8dbb419e264628e2f17330a6c8da60d3424010000000900000001a06d20d1d175b1e1d2b6e64" +
		"7ab5321717967de7e9367c28df8c0e20634ec7827019fe38e8d4f123f8e5286f3236db8dbb419e264628e2f17330a6c8da60d342401" +
		"dc68b1fc"
	signedImportTxHash = "2Rz6T1gteozqm5sCG52hDHk6m4iMY65R1LWfBCuPo3f595yrT7"

	coinId1 = "23CLURk1Czf1aLui1VdcuWSiDeFskfp3Sn8TQG7t6NKfeQRYDj:2"
	coinId2 = "2QmMXKS6rKQMnEh2XYZ4ZWCJmy8RpD3LyVZWxBG25t4N1JJqxY:1"

	cChainId, _ = ids.FromString("yH8D7ThNJkxmtkuv2jgBa4P1Rn3Qpr4pPr7QYNfcdoS6k6HWp")
	pChainId    = ids.Empty

	networkID = uint64(5)
	nonce     = uint64(48)

	avaxAssetID, _ = ids.FromString("U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK")
)

func TestConstructionDerive(t *testing.T) {
	backend := NewAtomicTxBackend(&mocks.Client{}, ids.Empty)

	t.Run("c-chain address", func(t *testing.T) {
		src := "02e0d4392cfa224d4be19db416b3cf62e90fb2b7015e7b62a95c8cb490514943f6"
		b, _ := hex.DecodeString(src)

		resp, err := backend.ConstructionDerive(
			context.Background(),
			&types.ConstructionDeriveRequest{
				NetworkIdentifier: networkIdentifier,
				PublicKey: &types.PublicKey{
					Bytes:     b,
					CurveType: types.Secp256k1,
				},
			},
		)
		assert.Nil(t, err)
		assert.Equal(
			t,
			"C-fuji15f9g0h5xkr5cp47n6u3qxj6yjtzzzrdr23a3tl",
			resp.AccountIdentifier.Address,
		)
	})
}

func TestConstructionPreprocess(t *testing.T) {
	backend := NewAtomicTxBackend(&mocks.Client{}, ids.Empty)

	t.Run("C-chain export preprocess", func(t *testing.T) {
		req := &types.ConstructionPreprocessRequest{
			NetworkIdentifier: networkIdentifier,
			Operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					RelatedOperations:   nil,
					Type:                "EXPORT",
					Account:             cAccountIdentifier,
					Amount:              mapper.AvaxAmount(big.NewInt(-10_000_000)),
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 1},
					RelatedOperations: []*types.OperationIdentifier{
						{Index: 0},
					},
					Type:    "EXPORT",
					Account: pAccountIdentifier,
					Amount:  mapper.AvaxAmount(big.NewInt(10_000_000)),
				},
			},
		}

		resp, apiErr := backend.ConstructionPreprocess(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, 11230., resp.Options["atomic_tx_gas"])
		assert.Equal(t, cAccountIdentifier.Address, resp.Options["from"])
		assert.Equal(t, "P", resp.Options["destination_chain"])
	})

	t.Run("C-chain import preprocess", func(t *testing.T) {
		req := &types.ConstructionPreprocessRequest{
			NetworkIdentifier: networkIdentifier,
			Operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					RelatedOperations:   nil,
					Type:                "IMPORT",
					Account:             pAccountIdentifier,
					Amount:              mapper.AvaxAmount(big.NewInt(-15_000_000)),
					CoinChange: &types.CoinChange{
						CoinIdentifier: &types.CoinIdentifier{Identifier: coinId1},
						CoinAction:     types.CoinSpent,
					},
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 1},
					RelatedOperations:   nil,
					Type:                "IMPORT",
					Account:             pAccountIdentifier,
					Amount:              mapper.AvaxAmount(big.NewInt(-5_000_000)),
					CoinChange: &types.CoinChange{
						CoinIdentifier: &types.CoinIdentifier{Identifier: coinId2},
						CoinAction:     types.CoinSpent,
					},
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 2},
					RelatedOperations: []*types.OperationIdentifier{
						{Index: 0},
						{Index: 1},
					},
					Type:    "IMPORT",
					Account: cAccountIdentifier,
					Amount:  mapper.AvaxAmount(big.NewInt(20_000_000)),
				},
			},
		}

		resp, apiErr := backend.ConstructionPreprocess(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, 12318., resp.Options["atomic_tx_gas"])
		assert.Equal(t, "P", resp.Options["source_chain"])
	})
}

func TestConstructionMetadata(t *testing.T) {
	clientMock := &mocks.Client{}
	backend := NewAtomicTxBackend(clientMock, ids.Empty)

	t.Run("C-chain export metadata", func(t *testing.T) {
		req := &types.ConstructionMetadataRequest{
			NetworkIdentifier: networkIdentifier,
			Options: map[string]interface{}{
				"atomic_tx_gas":     11230.,
				"from":              cAccountIdentifier.Address,
				"destination_chain": "P",
			},
		}

		clientMock.On("GetNetworkID", mock.Anything).Return(uint32(networkID), nil)
		clientMock.On("GetBlockchainID", mock.Anything, "C").Return(cChainId, nil)
		clientMock.On("GetBlockchainID", mock.Anything, "P").Return(pChainId, nil)
		clientMock.
			On("NonceAt", mock.Anything, ethcommon.HexToAddress(cAccountIdentifier.Address), (*big.Int)(nil)).
			Return(nonce, nil)
		clientMock.On("EstimateBaseFee", mock.Anything).Return(big.NewInt(25_000_000_000), nil)

		resp, apiErr := backend.ConstructionMetadata(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, float64(networkID), resp.Metadata["network_id"].(float64))
		assert.Equal(t, cChainId.String(), resp.Metadata["c_chain_id"])
		assert.Equal(t, pChainId.String(), resp.Metadata["destination_chain_id"])
		assert.Equal(t, float64(nonce), resp.Metadata["nonce"].(float64))
		assert.Equal(t, "280750", resp.SuggestedFee[0].Value)

		clientMock.AssertExpectations(t)
	})

	t.Run("C-chain import metadata", func(t *testing.T) {
		req := &types.ConstructionMetadataRequest{
			NetworkIdentifier: networkIdentifier,
			Options: map[string]interface{}{
				"atomic_tx_gas": 12318.,
				"source_chain":  "P",
			},
		}

		clientMock.On("GetNetworkID", mock.Anything).Return(networkID, nil)
		clientMock.On("GetBlockchainID", mock.Anything, "C").Return(cChainId, nil)
		clientMock.On("GetBlockchainID", mock.Anything, "P").Return(pChainId, nil)
		clientMock.On("EstimateBaseFee", mock.Anything).Return(big.NewInt(25_000_000_000), nil)

		resp, apiErr := backend.ConstructionMetadata(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, float64(networkID), resp.Metadata["network_id"].(float64))
		assert.Equal(t, cChainId.String(), resp.Metadata["c_chain_id"])
		assert.Equal(t, pChainId.String(), resp.Metadata["source_chain_id"])
		assert.Equal(t, "307950", resp.SuggestedFee[0].Value)

		clientMock.AssertExpectations(t)
	})
}

func TestConstructionPayload(t *testing.T) {
	backend := NewAtomicTxBackend(&mocks.Client{}, avaxAssetID)

	t.Run("C-chain export payloads", func(t *testing.T) {
		req := &types.ConstructionPayloadsRequest{
			NetworkIdentifier: networkIdentifier,
			Metadata: map[string]interface{}{
				"network_id":           networkID,
				"c_chain_id":           cChainId.String(),
				"destination_chain_id": pChainId.String(),
				"nonce":                nonce,
			},
			Operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					RelatedOperations:   nil,
					Type:                "EXPORT",
					Account:             cAccountIdentifier,
					Amount:              mapper.AvaxAmount(big.NewInt(-10_000_000)),
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 1},
					RelatedOperations: []*types.OperationIdentifier{
						{Index: 0},
					},
					Type:    "EXPORT",
					Account: pAccountIdentifier,
					Amount:  mapper.AvaxAmount(big.NewInt(9_719_250)),
				},
			},
		}

		resp, apiErr := backend.ConstructionPayloads(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, unsignedExportTx, resp.UnsignedTransaction)
		assert.Equal(t, 1, len(resp.Payloads))
		assert.Equal(t, cAccountIdentifier, resp.Payloads[0].AccountIdentifier)
		assert.Equal(t, types.EcdsaRecovery, resp.Payloads[0].SignatureType)
		assert.Equal(t, unsignedExportTxHash, resp.Payloads[0].Bytes)
	})

	t.Run("C-chain import payloads", func(t *testing.T) {
		req := &types.ConstructionPayloadsRequest{
			NetworkIdentifier: networkIdentifier,
			Metadata: map[string]interface{}{
				"network_id":      networkID,
				"c_chain_id":      cChainId.String(),
				"source_chain_id": pChainId.String(),
			},
			Operations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					RelatedOperations:   nil,
					Type:                "IMPORT",
					Account:             pAccountIdentifier,
					Amount:              mapper.AvaxAmount(big.NewInt(-15_000_000)),
					CoinChange: &types.CoinChange{
						CoinIdentifier: &types.CoinIdentifier{Identifier: coinId1},
						CoinAction:     types.CoinSpent,
					},
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 1},
					RelatedOperations:   nil,
					Type:                "IMPORT",
					Account:             pAccountIdentifier,
					Amount:              mapper.AvaxAmount(big.NewInt(-5_000_000)),
					CoinChange: &types.CoinChange{
						CoinIdentifier: &types.CoinIdentifier{Identifier: coinId2},
						CoinAction:     types.CoinSpent,
					},
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 2},
					RelatedOperations: []*types.OperationIdentifier{
						{Index: 0},
						{Index: 1},
					},
					Type:    "IMPORT",
					Account: cAccountIdentifier,
					Amount:  mapper.AvaxAmount(big.NewInt(19_692_050)),
				},
			},
		}

		resp, apiErr := backend.ConstructionPayloads(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, unsignedImportTx, resp.UnsignedTransaction)
		assert.Equal(t, 2, len(resp.Payloads))
		assert.Equal(t, pAccountIdentifier, resp.Payloads[0].AccountIdentifier)
		assert.Equal(t, types.EcdsaRecovery, resp.Payloads[0].SignatureType)
		assert.Equal(t, unsignedImportTxHash, resp.Payloads[0].Bytes)
		assert.Equal(t, pAccountIdentifier, resp.Payloads[1].AccountIdentifier)
		assert.Equal(t, types.EcdsaRecovery, resp.Payloads[1].SignatureType)
		assert.Equal(t, unsignedImportTxHash, resp.Payloads[1].Bytes)
	})
}

func TestConstructionParse(t *testing.T) {
	backend := NewAtomicTxBackend(&mocks.Client{}, ids.Empty)

	exportOperations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			RelatedOperations:   nil,
			Type:                "EXPORT",
			Account:             cAccountIdentifier,
			Amount:              mapper.AvaxAmount(big.NewInt(-10_000_000)),
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			RelatedOperations: []*types.OperationIdentifier{
				{Index: 0},
			},
			Type:    "EXPORT",
			Account: pAccountIdentifier,
			Amount:  mapper.AvaxAmount(big.NewInt(9_719_250)),
		},
	}

	importOperations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                "IMPORT",
			Amount:              mapper.AvaxAmount(big.NewInt(-15_000_000)),
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{Identifier: coinId1},
				CoinAction:     types.CoinSpent,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			Type:                "IMPORT",
			Amount:              mapper.AvaxAmount(big.NewInt(-5_000_000)),
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{Identifier: coinId2},
				CoinAction:     types.CoinSpent,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 2},
			RelatedOperations: []*types.OperationIdentifier{
				{Index: 0},
				{Index: 1},
			},
			Type:    "IMPORT",
			Account: cAccountIdentifier,
			Amount:  mapper.AvaxAmount(big.NewInt(19_692_050)),
		},
	}

	t.Run("C-chain export unsigned parse", func(t *testing.T) {
		req := &types.ConstructionParseRequest{
			NetworkIdentifier: networkIdentifier,
			Transaction:       unsignedExportTx,
			Signed:            false,
		}

		resp, apiErr := backend.ConstructionParse(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Nil(t, resp.AccountIdentifierSigners)
		assert.Equal(t, exportOperations, resp.Operations)
	})

	t.Run("C-chain export signed parse", func(t *testing.T) {
		req := &types.ConstructionParseRequest{
			NetworkIdentifier: networkIdentifier,
			Transaction:       signedExportTx,
			Signed:            true,
		}

		resp, apiErr := backend.ConstructionParse(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, []*types.AccountIdentifier{cAccountIdentifier}, resp.AccountIdentifierSigners)
		assert.Equal(t, exportOperations, resp.Operations)
	})

	t.Run("C-chain import unsigned parse", func(t *testing.T) {
		req := &types.ConstructionParseRequest{
			NetworkIdentifier: networkIdentifier,
			Transaction:       unsignedImportTx,
			Signed:            false,
		}

		resp, apiErr := backend.ConstructionParse(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Nil(t, resp.AccountIdentifierSigners)
		assert.Equal(t, importOperations, resp.Operations)
	})

	t.Run("C-chain import signed parse", func(t *testing.T) {
		req := &types.ConstructionParseRequest{
			NetworkIdentifier: networkIdentifier,
			Transaction:       signedImportTx,
			Signed:            true,
		}

		resp, apiErr := backend.ConstructionParse(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Nil(t, resp.AccountIdentifierSigners)
		assert.Equal(t, importOperations, resp.Operations)
	})
}

func TestConstructionCombine(t *testing.T) {
	backend := NewAtomicTxBackend(&mocks.Client{}, ids.Empty)

	t.Run("C-chain export combine", func(t *testing.T) {
		req := &types.ConstructionCombineRequest{
			NetworkIdentifier:   networkIdentifier,
			UnsignedTransaction: unsignedExportTx,
			Signatures: []*types.Signature{
				{
					SigningPayload: &types.SigningPayload{
						AccountIdentifier: cAccountIdentifier,
						Bytes:             unsignedExportTxHash,
						SignatureType:     types.EcdsaRecovery,
					},
					SignatureType: types.EcdsaRecovery,
					Bytes:         signedExportTxSignature,
				},
			},
		}

		resp, apiErr := backend.ConstructionCombine(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, signedExportTx, resp.SignedTransaction)
	})

	t.Run("C-chain import combine", func(t *testing.T) {
		signature := &types.Signature{
			SigningPayload: &types.SigningPayload{
				AccountIdentifier: pAccountIdentifier,
				Bytes:             unsignedImportTxHash,
				SignatureType:     types.EcdsaRecovery,
			},
			SignatureType: types.EcdsaRecovery,
			Bytes:         signedImportTxSignature,
		}

		req := &types.ConstructionCombineRequest{
			NetworkIdentifier:   networkIdentifier,
			UnsignedTransaction: unsignedImportTx,
			// two signatures, one for each input
			Signatures: []*types.Signature{signature, signature},
		}

		resp, apiErr := backend.ConstructionCombine(context.Background(), req)

		assert.Nil(t, apiErr)
		assert.Equal(t, signedImportTx, resp.SignedTransaction)
	})
}

func TestConstructionHash(t *testing.T) {
	backend := NewAtomicTxBackend(&mocks.Client{}, ids.Empty)

	t.Run("C-chain valid atomic export transaction", func(t *testing.T) {
		resp, err := backend.ConstructionHash(context.Background(), &types.ConstructionHashRequest{
			NetworkIdentifier: networkIdentifier,
			SignedTransaction: signedImportTx,
		})
		assert.Nil(t, err)
		assert.Equal(t, signedImportTxHash, resp.TransactionIdentifier.Hash)
	})
}

func TestConstructionSubmit(t *testing.T) {
	clientMock := &mocks.Client{}
	backend := NewAtomicTxBackend(clientMock, ids.Empty)

	t.Run("C-chain valid atomic export transaction", func(t *testing.T) {
		signedTxBytes, _ := formatting.Decode(formatting.Hex, signedImportTx)
		txId, _ := ids.FromString(signedImportTxHash)

		clientMock.On("IssueTx", mock.Anything, signedTxBytes).Return(txId, nil)

		resp, apiErr := backend.ConstructionSubmit(context.Background(), &types.ConstructionSubmitRequest{
			NetworkIdentifier: networkIdentifier,
			SignedTransaction: signedImportTx,
		})

		assert.Nil(t, apiErr)
		assert.Equal(t, signedImportTxHash, resp.TransactionIdentifier.Hash)
		clientMock.AssertExpectations(t)
	})
}
