package cchainatomictx

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

var (
	errMissingCoinIdentifier = errors.New("input operation does not have coin identifier")
)

func BuildTx(opType string, matches []*parser.Match, metadata Metadata, codec codec.Manager, avaxAssetId ids.ID) (*evm.Tx, []*types.AccountIdentifier, error) {
	switch opType {
	case mapper.OpExport:
		return buildExportTx(matches, metadata, codec, avaxAssetId)
	case mapper.OpImport:
		return buildImportTx(matches, metadata, avaxAssetId)
	default:
		return nil, nil, fmt.Errorf("unsupported atomic operation type %s", opType)
	}
}

func buildExportTx(
	matches []*parser.Match,
	metadata Metadata,
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*evm.Tx, []*types.AccountIdentifier, error) {
	ins, signers, err := buildIns(matches, metadata, avaxAssetId)
	if err != nil {
		return nil, nil, err
	}

	exportedOutputs, err := buildExportedOutputs(matches, codec, avaxAssetId)
	if err != nil {
		return nil, nil, err
	}

	tx := &evm.Tx{UnsignedAtomicTx: &evm.UnsignedExportTx{
		NetworkID:        metadata.NetworkID,
		BlockchainID:     metadata.CChainID,
		DestinationChain: *metadata.DestinationChainId,
		Ins:              ins,
		ExportedOutputs:  exportedOutputs,
	}}
	return tx, signers, nil
}

func buildImportTx(matches []*parser.Match, metadata Metadata, avaxAssetId ids.ID) (*evm.Tx, []*types.AccountIdentifier, error) {
	importedInputs, signers, err := buildImportedInputs(matches, avaxAssetId)
	if err != nil {
		return nil, nil, err
	}

	outs := buildOuts(matches, avaxAssetId)

	tx := &evm.Tx{UnsignedAtomicTx: &evm.UnsignedImportTx{
		NetworkID:      metadata.NetworkID,
		BlockchainID:   metadata.CChainID,
		SourceChain:    *metadata.SourceChainID,
		ImportedInputs: importedInputs,
		Outs:           outs,
	}}
	return tx, signers, nil
}

func buildIns(matches []*parser.Match, metadata Metadata, avaxAssetId ids.ID) ([]evm.EVMInput, []*types.AccountIdentifier, error) {
	inputMatch := matches[0]

	var ins []evm.EVMInput
	var signers []*types.AccountIdentifier
	for i, op := range inputMatch.Operations {
		ins = append(ins, evm.EVMInput{
			Address: ethcommon.HexToAddress(op.Account.Address),
			Amount:  inputMatch.Amounts[i].Uint64(),
			AssetID: avaxAssetId,
			Nonce:   metadata.Nonce,
		})
		signers = append(signers, op.Account)
	}
	return ins, signers, nil
}

func buildImportedInputs(matches []*parser.Match, avaxAssetId ids.ID) ([]*avax.TransferableInput, []*types.AccountIdentifier, error) {
	inputMatch := matches[0]

	var importedInputs []*avax.TransferableInput
	var signers []*types.AccountIdentifier
	for i, op := range inputMatch.Operations {
		if op.CoinChange == nil || op.CoinChange.CoinIdentifier == nil {
			return nil, nil, errMissingCoinIdentifier
		}
		utxoId, err := mapper.DecodeUTXOID(op.CoinChange.CoinIdentifier.Identifier)
		if err != nil {
			return nil, nil, err
		}

		importedInputs = append(importedInputs, &avax.TransferableInput{
			UTXOID: *utxoId,
			Asset:  avax.Asset{ID: avaxAssetId},
			In: &secp256k1fx.TransferInput{
				Amt: inputMatch.Amounts[i].Uint64(),
				Input: secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
		})
		signers = append(signers, op.Account)
	}
	avax.SortTransferableInputs(importedInputs)

	return importedInputs, signers, nil
}

func buildOuts(matches []*parser.Match, avaxAssetId ids.ID) []evm.EVMOutput {
	outputMatch := matches[1]

	var outs []evm.EVMOutput
	for i, op := range outputMatch.Operations {
		outs = append(outs, evm.EVMOutput{
			Address: ethcommon.HexToAddress(op.Account.Address),
			Amount:  outputMatch.Amounts[i].Uint64(),
			AssetID: avaxAssetId,
		})
	}

	return outs
}

func buildExportedOutputs(matches []*parser.Match, codec codec.Manager, avaxAssetId ids.ID) ([]*avax.TransferableOutput, error) {
	outputMatch := matches[1]

	var outs []*avax.TransferableOutput
	for i, op := range outputMatch.Operations {
		destinationAddress, err := address.ParseToID(op.Account.Address)
		if err != nil {
			return nil, err
		}

		outs = append(outs, &avax.TransferableOutput{
			Asset: avax.Asset{ID: avaxAssetId},
			Out: &secp256k1fx.TransferOutput{
				Amt: outputMatch.Amounts[i].Uint64(),
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{destinationAddress},
				},
			},
		})
	}

	avax.SortTransferableOutputs(outs, codec)

	return outs, nil
}

func buildRelatedOperations(idx int) []*types.OperationIdentifier {
	var identifiers []*types.OperationIdentifier
	for i := 0; i < idx; i++ {
		identifiers = append(identifiers, &types.OperationIdentifier{
			Index: int64(i),
		})
	}
	return identifiers
}
