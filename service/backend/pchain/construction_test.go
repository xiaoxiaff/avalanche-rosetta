package pchain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	ajson "github.com/ava-labs/avalanchego/utils/json"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	mocks "github.com/ava-labs/avalanche-rosetta/mocks/client"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
)

var (
	pChainNetworkIdentifier = &types.NetworkIdentifier{
		Blockchain: service.BlockchainName,
		Network:    mapper.FujiNetwork,
		SubNetworkIdentifier: &types.SubNetworkIdentifier{
			Network: mapper.PChainNetworkIdentifier,
		},
	}

	cAccountIdentifier = &types.AccountIdentifier{Address: "C-fuji123zu6qwhtd9qdd45ryu3j0qtr325gjgddys6u8"}
	pAccountIdentifier = &types.AccountIdentifier{Address: "P-fuji123zu6qwhtd9qdd45ryu3j0qtr325gjgddys6u8"}
	stakeRewardAccount = &types.AccountIdentifier{Address: "P-fuji1ea7dxk8zazpyf8tgc8yg3xyfatey0deqvg9pv2"}

	cChainID, _ = ids.FromString("yH8D7ThNJkxmtkuv2jgBa4P1Rn3Qpr4pPr7QYNfcdoS6k6HWp")
	pChainID    = ids.Empty

	nodeID = "NodeID-Bvsx89JttQqhqdgwtizAPoVSNW74Xcr2S"

	networkID = 5

	avaxAssetID, _ = ids.FromString("U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK")

	opTypeInput  = "INPUT"
	opTypeImport = "IMPORT"
	opTypeExport = "EXPORT"
	opTypeOutput = "OUTPUT"
	opTypeStake  = "STAKE"

	txFee = 1_000_000

	coinId1 = "2ryRVCwNSjEinTViuvDkzX41uQzx3g4babXxZMD46ZV1a9X4Eg:0"
)

func buildRosettaSignerJson(coinIdentifiers []string, signers []*types.AccountIdentifier) string {
	var importSigners []*common.Signer
	for i, s := range signers {
		importSigners = append(importSigners, &common.Signer{
			CoinIdentifier:    coinIdentifiers[i],
			AccountIdentifier: s,
		})
	}
	bytes, _ := json.Marshal(importSigners)
	return string(bytes)
}

func TestConstructionDerive(t *testing.T) {
	pChainMock := &mocks.PChainClient{}
	ctx := context.Background()
	pChainMock.Mock.On("GetNetworkID", ctx).Return(uint32(5), nil)
	backend := NewBackend(pChainMock, nil, ids.Empty, nil)

	t.Run("p-chain address", func(t *testing.T) {
		src := "02e0d4392cfa224d4be19db416b3cf62e90fb2b7015e7b62a95c8cb490514943f6"
		b, _ := hex.DecodeString(src)

		resp, err := backend.ConstructionDerive(
			context.Background(),
			&types.ConstructionDeriveRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
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

func TestExportTxConstruction(t *testing.T) {
	opExportAvax := "EXPORT_AVAX"

	exportOperations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			RelatedOperations:   nil,
			Type:                opExportAvax,
			Account:             pAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(-1_000_000_000)),
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{Identifier: coinId1},
				CoinAction:     types.CoinSpent,
			},
			Metadata: map[string]interface{}{
				"type":        opTypeInput,
				"sig_indices": []interface{}{0.0},
				"locktime":    0.0,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			Type:                opExportAvax,
			Account:             cAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(999_000_000)),
			Metadata: map[string]interface{}{
				"type":      opTypeExport,
				"threshold": 1.0,
				"locktime":  0.0,
			},
		},
	}

	preprocessMetadata := map[string]interface{}{
		"destination_chain": "C",
	}

	metadataOptions := map[string]interface{}{
		"destination_chain": "C",
		"type":              opExportAvax,
	}

	payloadsMetadata := map[string]interface{}{
		"network_id":           float64(networkID),
		"destination_chain":    "C",
		"destination_chain_id": cChainID.String(),
		"blockchain_id":        pChainID.String(),
	}

	signers := []*types.AccountIdentifier{pAccountIdentifier}
	exportSigners := buildRosettaSignerJson([]string{coinId1}, signers)

	unsignedExportTx := "0x0000000000120000000500000000000000000000000000000000000000000000000000000000000000000000000000000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000003b9aca000000000100000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d5000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b8b87c0000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d0000000065e8045f"
	unsignedExportTxHash, _ := hex.DecodeString("44d579f5cb3c83f4137223a0368721734b622ec392007760eed97f3f1a40c595")

	signingPayloads := []*types.SigningPayload{
		{
			AccountIdentifier: pAccountIdentifier,
			Bytes:             unsignedExportTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
	}

	signedExportTx := "0x0000000000120000000500000000000000000000000000000000000000000000000000000000000000000000000000000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000003b9aca000000000100000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d5000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b8b87c0000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d0000000100000009000000017403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b0137dc0dc4"
	signedExportTxSignature, _ := hex.DecodeString("7403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b01")
	signedExportTxHash := "bG7jzw16x495XSFdhEavHWR836Ya5teoB1YxRC1inN3HEtqbs"

	wrappedTxFormat := `{"tx":"%s","signers":%s,"destination_chain":"%s","destination_chain_id":"%s"}`
	wrappedUnsignedExportTx := fmt.Sprintf(wrappedTxFormat, unsignedExportTx, exportSigners, "C", cChainID.String())
	wrappedSignedExportTx := fmt.Sprintf(wrappedTxFormat, signedExportTx, exportSigners, "C", cChainID.String())

	signatures := []*types.Signature{{
		SigningPayload: &types.SigningPayload{
			AccountIdentifier: pAccountIdentifier,
			Bytes:             unsignedExportTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
		SignatureType: types.EcdsaRecovery,
		Bytes:         signedExportTxSignature,
	}}

	ctx := context.Background()
	clientMock := &mocks.PChainClient{}
	backend := NewBackend(clientMock, nil, avaxAssetID, pChainNetworkIdentifier)

	t.Run("preprocess endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        exportOperations,
				Metadata:          preprocessMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, metadataOptions, resp.Options)

		clientMock.AssertExpectations(t)
	})

	t.Run("metadata endpoint", func(t *testing.T) {
		clientMock.On("GetNetworkID", ctx).Return(uint32(networkID), nil)
		clientMock.On("GetTxFee", ctx).Return(&info.GetTxFeeResponse{TxFee: ajson.Uint64(txFee)}, nil)
		clientMock.On("GetBlockchainID", ctx, mapper.PChainNetworkIdentifier).Return(pChainID, nil)
		clientMock.On("GetBlockchainID", ctx, mapper.CChainNetworkIdentifier).Return(cChainID, nil)

		resp, err := backend.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           metadataOptions,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, payloadsMetadata, resp.Metadata)

		clientMock.AssertExpectations(t)
	})

	t.Run("payloads endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        exportOperations,
				Metadata:          payloadsMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, wrappedUnsignedExportTx, resp.UnsignedTransaction)
		assert.Equal(t, signingPayloads, resp.Payloads)

		clientMock.AssertExpectations(t)
	})

	t.Run("parse endpoint (unsigned)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedUnsignedExportTx,
				Signed:            false,
			},
		)
		assert.Nil(t, err)
		assert.Nil(t, resp.AccountIdentifierSigners)
		assert.Equal(t, exportOperations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("combine endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionCombine(
			ctx,
			&types.ConstructionCombineRequest{
				NetworkIdentifier:   pChainNetworkIdentifier,
				UnsignedTransaction: wrappedUnsignedExportTx,
				Signatures:          signatures,
			},
		)

		assert.Nil(t, err)
		assert.Equal(t, wrappedSignedExportTx, resp.SignedTransaction)
	})

	t.Run("parse endpoint (signed)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedSignedExportTx,
				Signed:            true,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, signers, resp.AccountIdentifierSigners)
		assert.Equal(t, exportOperations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("hash endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionHash(ctx, &types.ConstructionHashRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedExportTx,
		})
		assert.Nil(t, err)
		assert.Equal(t, signedExportTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})

	t.Run("submit endpoint", func(t *testing.T) {
		signedTxBytes, _ := formatting.Decode(formatting.Hex, signedExportTx)
		txId, _ := ids.FromString(signedExportTxHash)

		clientMock.On("IssueTx", ctx, signedTxBytes).Return(txId, nil)

		resp, apiErr := backend.ConstructionSubmit(ctx, &types.ConstructionSubmitRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedExportTx,
		})

		assert.Nil(t, apiErr)
		assert.Equal(t, signedExportTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})
}

func TestImportTxConstruction(t *testing.T) {
	opImportAvax := "IMPORT_AVAX"

	importOperations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			RelatedOperations:   nil,
			Type:                opImportAvax,
			Account:             cAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(-1_000_000_000)),
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{Identifier: coinId1},
				CoinAction:     types.CoinSpent,
			},
			Metadata: map[string]interface{}{
				"type":        opTypeImport,
				"sig_indices": []interface{}{0.0},
				"locktime":    0.0,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			Type:                opImportAvax,
			Account:             pAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(999_000_000)),
			Metadata: map[string]interface{}{
				"type":      opTypeOutput,
				"threshold": 1.0,
				"locktime":  0.0,
			},
		},
	}

	preprocessMetadata := map[string]interface{}{
		"source_chain": "C",
	}

	metadataOptions := map[string]interface{}{
		"source_chain": "C",
		"type":         opImportAvax,
	}

	payloadsMetadata := map[string]interface{}{
		"network_id":      float64(networkID),
		"source_chain_id": cChainID.String(),
		"blockchain_id":   pChainID.String(),
	}

	signers := []*types.AccountIdentifier{cAccountIdentifier}
	importSigners := buildRosettaSignerJson([]string{coinId1}, signers)

	unsignedImportTx := "0x000000000011000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b8b87c0000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d00000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d500000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000003b9aca000000000100000000000000004ce8b27d"
	unsignedImportTxHash, _ := hex.DecodeString("e9114ae12065d1f8631bc40729c806a3a4793de714001bfee66482f520dc1865")
	wrappedUnsignedImportTx := `{"tx":"` + unsignedImportTx + `","signers":` + importSigners + `}`

	signingPayloads := []*types.SigningPayload{
		{
			AccountIdentifier: cAccountIdentifier,
			Bytes:             unsignedImportTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
	}

	signedImportTx := "0x000000000011000000050000000000000000000000000000000000000000000000000000000000000000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000000003b8b87c0000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d00000000000000007fc93d85c6d62c5b2ac0b519c87010ea5294012d1e407030d6acd0021cac10d500000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000000003b9aca0000000001000000000000000100000009000000017403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b018ac25b4e"
	signedImportTxSignature, _ := hex.DecodeString("7403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b01")
	signedImportTxHash := "byyEVU6RL7PQNSVT8qEnybWGV5BbBfJwFV6bEDV5mkymXRz62"

	wrappedSignedImportTx := `{"tx":"` + signedImportTx + `","signers":` + importSigners + `}`

	signatures := []*types.Signature{{
		SigningPayload: &types.SigningPayload{
			AccountIdentifier: cAccountIdentifier,
			Bytes:             unsignedImportTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
		SignatureType: types.EcdsaRecovery,
		Bytes:         signedImportTxSignature,
	}}

	ctx := context.Background()
	clientMock := &mocks.PChainClient{}
	backend := NewBackend(clientMock, nil, avaxAssetID, pChainNetworkIdentifier)

	t.Run("preprocess endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        importOperations,
				Metadata:          preprocessMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, metadataOptions, resp.Options)

		clientMock.AssertExpectations(t)
	})

	t.Run("metadata endpoint", func(t *testing.T) {
		clientMock.On("GetNetworkID", ctx).Return(uint32(networkID), nil)
		clientMock.On("GetTxFee", ctx).Return(&info.GetTxFeeResponse{TxFee: ajson.Uint64(txFee)}, nil)
		clientMock.On("GetBlockchainID", ctx, mapper.PChainNetworkIdentifier).Return(pChainID, nil)
		clientMock.On("GetBlockchainID", ctx, mapper.CChainNetworkIdentifier).Return(cChainID, nil)

		resp, err := backend.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           metadataOptions,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, payloadsMetadata, resp.Metadata)

		clientMock.AssertExpectations(t)
	})

	t.Run("payloads endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        importOperations,
				Metadata:          payloadsMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, wrappedUnsignedImportTx, resp.UnsignedTransaction)
		assert.Equal(t, signingPayloads, resp.Payloads)

		clientMock.AssertExpectations(t)
	})

	t.Run("parse endpoint (unsigned)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedUnsignedImportTx,
				Signed:            false,
			},
		)
		assert.Nil(t, err)
		assert.Nil(t, resp.AccountIdentifierSigners)
		assert.Equal(t, importOperations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("combine endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionCombine(
			ctx,
			&types.ConstructionCombineRequest{
				NetworkIdentifier:   pChainNetworkIdentifier,
				UnsignedTransaction: wrappedUnsignedImportTx,
				Signatures:          signatures,
			},
		)

		assert.Nil(t, err)
		assert.Equal(t, wrappedSignedImportTx, resp.SignedTransaction)
	})

	t.Run("parse endpoint (signed)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedSignedImportTx,
				Signed:            true,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, signers, resp.AccountIdentifierSigners)
		assert.Equal(t, importOperations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("hash endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionHash(ctx, &types.ConstructionHashRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedImportTx,
		})
		assert.Nil(t, err)
		assert.Equal(t, signedImportTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})

	t.Run("submit endpoint", func(t *testing.T) {
		signedTxBytes, _ := formatting.Decode(formatting.Hex, signedImportTx)
		txId, _ := ids.FromString(signedImportTxHash)

		clientMock.On("IssueTx", ctx, signedTxBytes).Return(txId, nil)

		resp, apiErr := backend.ConstructionSubmit(ctx, &types.ConstructionSubmitRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedImportTx,
		})

		assert.Nil(t, apiErr)
		assert.Equal(t, signedImportTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})
}

func TestAddValidatorTxConstruction(t *testing.T) {
	opAddValidator := "ADD_VALIDATOR"
	startTime := uint64(1659592163)
	endTime := startTime + 14*86400
	shares := uint32(200000)

	operations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			RelatedOperations:   nil,
			Type:                opAddValidator,
			Account:             pAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(-2_000_000_000_000)),
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{Identifier: coinId1},
				CoinAction:     "coin_spent",
			},
			Metadata: map[string]interface{}{
				"type":        opTypeInput,
				"sig_indices": []interface{}{0.0},
				"locktime":    0.0,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			Type:                opAddValidator,
			Account:             pAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(2_000_000_000_000)),
			Metadata: map[string]interface{}{
				"type":      opTypeStake,
				"locktime":  0.0,
				"threshold": 1.0,
			},
		},
	}

	preprocessMetadata := map[string]interface{}{
		"node_id":          nodeID,
		"start":            startTime,
		"end":              endTime,
		"shares":           shares,
		"reward_addresses": []string{stakeRewardAccount.Address},
	}

	metadataOptions := map[string]interface{}{
		"type":             opAddValidator,
		"node_id":          nodeID,
		"start":            startTime,
		"end":              endTime,
		"shares":           shares,
		"reward_addresses": []string{stakeRewardAccount.Address},
	}

	payloadsMetadata := map[string]interface{}{
		"network_id":       float64(networkID),
		"blockchain_id":    pChainID.String(),
		"node_id":          nodeID,
		"start":            float64(startTime),
		"end":              float64(endTime),
		"shares":           float64(shares),
		"locktime":         0.0,
		"threshold":        0.0,
		"memo":             "",
		"reward_addresses": []interface{}{stakeRewardAccount.Address},
	}

	signers := []*types.AccountIdentifier{pAccountIdentifier}
	stakeSigners := buildRosettaSignerJson([]string{coinId1}, signers)

	unsignedTx := "0x00000000000c0000000500000000000000000000000000000000000000000000000000000000000000000000000000000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000001d1a94a200000000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062eb5de30000000062fdd2e3000001d1a94a2000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000001d1a94a2000000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d0000000b00000000000000000000000000000001cf7cd358e2e882449d68c1c8889889eaf247b72000030d4000000000cfc0bbdf"
	unsignedTxHash, _ := hex.DecodeString("36c4742139528891d2eeab82cf6d261e33ef374a0230c40f42bdcde7a875677c")
	wrappedUnsignedTx := `{"tx":"` + unsignedTx + `","signers":` + stakeSigners + `}`

	signingPayloads := []*types.SigningPayload{
		{
			AccountIdentifier: pAccountIdentifier,
			Bytes:             unsignedTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
	}

	signedTx := "0x00000000000c0000000500000000000000000000000000000000000000000000000000000000000000000000000000000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000005000001d1a94a200000000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062eb5de30000000062fdd2e3000001d1a94a2000000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa00000007000001d1a94a2000000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d0000000b00000000000000000000000000000001cf7cd358e2e882449d68c1c8889889eaf247b72000030d400000000100000009000000017403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b01f45d31dd"
	signedTxSignature, _ := hex.DecodeString("7403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b01")
	signedTxHash := "dhHkN5nDupSTDqudpbJXef1d4F1Zy3TUSbYqmWJBH6qzggFBw"

	wrappedSignedTx := `{"tx":"` + signedTx + `","signers":` + stakeSigners + `}`

	signatures := []*types.Signature{{
		SigningPayload: &types.SigningPayload{
			AccountIdentifier: cAccountIdentifier,
			Bytes:             unsignedTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
		SignatureType: types.EcdsaRecovery,
		Bytes:         signedTxSignature,
	}}

	ctx := context.Background()
	clientMock := &mocks.PChainClient{}
	backend := NewBackend(clientMock, nil, avaxAssetID, pChainNetworkIdentifier)

	t.Run("preprocess endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        operations,
				Metadata:          preprocessMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, metadataOptions, resp.Options)

		clientMock.AssertExpectations(t)
	})

	t.Run("metadata endpoint", func(t *testing.T) {
		clientMock.On("GetNetworkID", ctx).Return(uint32(networkID), nil)
		clientMock.On("GetBlockchainID", ctx, mapper.PChainNetworkIdentifier).Return(pChainID, nil)

		resp, err := backend.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           metadataOptions,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, payloadsMetadata, resp.Metadata)

		clientMock.AssertExpectations(t)
	})

	t.Run("payloads endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        operations,
				Metadata:          payloadsMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, wrappedUnsignedTx, resp.UnsignedTransaction)
		assert.Equal(t, signingPayloads, resp.Payloads)

		clientMock.AssertExpectations(t)
	})

	t.Run("parse endpoint (unsigned)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedUnsignedTx,
				Signed:            false,
			},
		)
		assert.Nil(t, err)
		assert.Nil(t, resp.AccountIdentifierSigners)
		assert.Equal(t, operations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("combine endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionCombine(
			ctx,
			&types.ConstructionCombineRequest{
				NetworkIdentifier:   pChainNetworkIdentifier,
				UnsignedTransaction: wrappedUnsignedTx,
				Signatures:          signatures,
			},
		)

		assert.Nil(t, err)
		assert.Equal(t, wrappedSignedTx, resp.SignedTransaction)
	})

	t.Run("parse endpoint (signed)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedSignedTx,
				Signed:            true,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, signers, resp.AccountIdentifierSigners)
		assert.Equal(t, operations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("hash endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionHash(ctx, &types.ConstructionHashRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedTx,
		})
		assert.Nil(t, err)
		assert.Equal(t, signedTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})

	t.Run("submit endpoint", func(t *testing.T) {
		signedTxBytes, _ := formatting.Decode(formatting.Hex, signedTx)
		txId, _ := ids.FromString(signedTxHash)

		clientMock.On("IssueTx", ctx, signedTxBytes).Return(txId, nil)

		resp, apiErr := backend.ConstructionSubmit(ctx, &types.ConstructionSubmitRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedTx,
		})

		assert.Nil(t, apiErr)
		assert.Equal(t, signedTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})
}

func TestAddDelegatorTxConstruction(t *testing.T) {
	opAddDelegator := "ADD_DELEGATOR"
	startTime := uint64(1659592163)
	endTime := startTime + 14*86400

	operations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			RelatedOperations:   nil,
			Type:                opAddDelegator,
			Account:             pAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(-25_000_000_000)),
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{Identifier: coinId1},
				CoinAction:     "coin_spent",
			},
			Metadata: map[string]interface{}{
				"type":        opTypeInput,
				"sig_indices": []interface{}{0.0},
				"locktime":    0.0,
			},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			Type:                opAddDelegator,
			Account:             pAccountIdentifier,
			Amount:              mapper.AtomicAvaxAmount(big.NewInt(25_000_000_000)),
			Metadata: map[string]interface{}{
				"type":      opTypeStake,
				"locktime":  0.0,
				"threshold": 1.0,
			},
		},
	}

	preprocessMetadata := map[string]interface{}{
		"node_id":          nodeID,
		"start":            startTime,
		"end":              endTime,
		"reward_addresses": []string{stakeRewardAccount.Address},
	}

	metadataOptions := map[string]interface{}{
		"type":             opAddDelegator,
		"node_id":          nodeID,
		"start":            startTime,
		"end":              endTime,
		"reward_addresses": []string{stakeRewardAccount.Address},
	}

	payloadsMetadata := map[string]interface{}{
		"network_id":       float64(networkID),
		"blockchain_id":    pChainID.String(),
		"node_id":          nodeID,
		"start":            float64(startTime),
		"end":              float64(endTime),
		"shares":           0.0,
		"locktime":         0.0,
		"threshold":        0.0,
		"memo":             "",
		"reward_addresses": []interface{}{stakeRewardAccount.Address},
	}

	signers := []*types.AccountIdentifier{pAccountIdentifier}
	stakeSigners := buildRosettaSignerJson([]string{coinId1}, signers)

	unsignedTx := "0x00000000000e0000000500000000000000000000000000000000000000000000000000000000000000000000000000000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000500000005d21dba0000000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062eb5de30000000062fdd2e300000005d21dba00000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000700000005d21dba00000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d0000000b00000000000000000000000000000001cf7cd358e2e882449d68c1c8889889eaf247b72000000000d345d2b2"
	unsignedTxHash, _ := hex.DecodeString("768a18893738542fdb3c0bb03065c09249e502d909644990b4f2e7cbd77ff37a")
	wrappedUnsignedTx := `{"tx":"` + unsignedTx + `","signers":` + stakeSigners + `}`

	signingPayloads := []*types.SigningPayload{
		{
			AccountIdentifier: pAccountIdentifier,
			Bytes:             unsignedTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
	}

	signedTx := "0x00000000000e0000000500000000000000000000000000000000000000000000000000000000000000000000000000000001f52a5a6dd8f1b3fe05204bdab4f6bcb5a7059f88d0443c636f6c158f838dd1a8000000003d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000500000005d21dba0000000001000000000000000077e1d5c6c289c49976f744749d54369d2129d7500000000062eb5de30000000062fdd2e300000005d21dba00000000013d9bdac0ed1d761330cf680efdeb1a42159eb387d6d2950c96f7d28f61bbe2aa0000000700000005d21dba00000000000000000000000001000000015445cd01d75b4a06b6b41939193c0b1c5544490d0000000b00000000000000000000000000000001cf7cd358e2e882449d68c1c8889889eaf247b7200000000100000009000000017403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b011ea948a2"
	signedTxSignature, _ := hex.DecodeString("7403e32bb967e71902a988b7da635b4bca2475eedbfd23176610a88162f3a92f20b61f2185825b04b7f8ee8c76427c8dc80eb6091f9e594ef259a59856e5401b01")
	signedTxHash := "R8k9kbffu2AQHqaMKBEFxFnzUuS4pSG1i8XEFQX8Sn8ogrP7j"

	wrappedSignedTx := `{"tx":"` + signedTx + `","signers":` + stakeSigners + `}`

	signatures := []*types.Signature{{
		SigningPayload: &types.SigningPayload{
			AccountIdentifier: cAccountIdentifier,
			Bytes:             unsignedTxHash,
			SignatureType:     types.EcdsaRecovery,
		},
		SignatureType: types.EcdsaRecovery,
		Bytes:         signedTxSignature,
	}}

	ctx := context.Background()
	clientMock := &mocks.PChainClient{}
	backend := NewBackend(clientMock, nil, avaxAssetID, pChainNetworkIdentifier)

	t.Run("preprocess endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPreprocess(
			ctx,
			&types.ConstructionPreprocessRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        operations,
				Metadata:          preprocessMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, metadataOptions, resp.Options)

		clientMock.AssertExpectations(t)
	})

	t.Run("metadata endpoint", func(t *testing.T) {
		clientMock.On("GetNetworkID", ctx).Return(uint32(networkID), nil)
		clientMock.On("GetBlockchainID", ctx, mapper.PChainNetworkIdentifier).Return(pChainID, nil)

		resp, err := backend.ConstructionMetadata(
			ctx,
			&types.ConstructionMetadataRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Options:           metadataOptions,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, payloadsMetadata, resp.Metadata)

		clientMock.AssertExpectations(t)
	})

	t.Run("payloads endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionPayloads(
			ctx,
			&types.ConstructionPayloadsRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Operations:        operations,
				Metadata:          payloadsMetadata,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, wrappedUnsignedTx, resp.UnsignedTransaction)
		assert.Equal(t, signingPayloads, resp.Payloads)

		clientMock.AssertExpectations(t)
	})

	t.Run("parse endpoint (unsigned)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedUnsignedTx,
				Signed:            false,
			},
		)
		assert.Nil(t, err)
		assert.Nil(t, resp.AccountIdentifierSigners)
		assert.Equal(t, operations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("combine endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionCombine(
			ctx,
			&types.ConstructionCombineRequest{
				NetworkIdentifier:   pChainNetworkIdentifier,
				UnsignedTransaction: wrappedUnsignedTx,
				Signatures:          signatures,
			},
		)

		assert.Nil(t, err)
		assert.Equal(t, wrappedSignedTx, resp.SignedTransaction)
	})

	t.Run("parse endpoint (signed)", func(t *testing.T) {
		resp, err := backend.ConstructionParse(
			ctx,
			&types.ConstructionParseRequest{
				NetworkIdentifier: pChainNetworkIdentifier,
				Transaction:       wrappedSignedTx,
				Signed:            true,
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, signers, resp.AccountIdentifierSigners)
		assert.Equal(t, operations, resp.Operations)

		clientMock.AssertExpectations(t)
	})

	t.Run("hash endpoint", func(t *testing.T) {
		resp, err := backend.ConstructionHash(ctx, &types.ConstructionHashRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedTx,
		})
		assert.Nil(t, err)
		assert.Equal(t, signedTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})

	t.Run("submit endpoint", func(t *testing.T) {
		signedTxBytes, _ := formatting.Decode(formatting.Hex, signedTx)
		txId, _ := ids.FromString(signedTxHash)

		clientMock.On("IssueTx", ctx, signedTxBytes).Return(txId, nil)

		resp, apiErr := backend.ConstructionSubmit(ctx, &types.ConstructionSubmitRequest{
			NetworkIdentifier: pChainNetworkIdentifier,
			SignedTransaction: wrappedSignedTx,
		})

		assert.Nil(t, apiErr)
		assert.Equal(t, signedTxHash, resp.TransactionIdentifier.Hash)

		clientMock.AssertExpectations(t)
	})
}
