package c

import (
	"errors"
	"fmt"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service/chain/common"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"strconv"
)

func (c *CChainAtomicTxBackend) buildTx(opType string, matches []*parser.Match, metadata cBackendMetadata) (*evm.Tx, []*types.AccountIdentifier, error) {
	switch opType {
	case mapper.OpExport:
		return c.buildExportTx(matches, metadata)
	case mapper.OpImport:
		return c.buildImportTx(matches, metadata)
	default:
		return nil, nil, fmt.Errorf("unsupported atomic operation type %s", opType)
	}
}

func (c *CChainAtomicTxBackend) parseTx(tx evm.Tx, hrp string) ([]*types.Operation, error) {
	switch unsignedTx := tx.UnsignedAtomicTx.(type) {
	case *evm.UnsignedExportTx:
		return c.parseExportTx(unsignedTx, hrp)
	case *evm.UnsignedImportTx:
		return c.parseImportTx(unsignedTx), nil
	default:
		return nil, fmt.Errorf("unsupported tx type")
	}
}

func (c *CChainAtomicTxBackend) buildExportTx(matches []*parser.Match, metadata cBackendMetadata) (*evm.Tx, []*types.AccountIdentifier, error) {
	ins, signers, err := c.buildIns(matches, metadata)
	if err != nil {
		return nil, nil, err
	}

	exportedOutputs, err := c.buildExportedOutputs(matches)
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

func (c *CChainAtomicTxBackend) parseExportTx(exportTx *evm.UnsignedExportTx, hrp string) ([]*types.Operation, error) {
	var operations []*types.Operation
	ins := c.parseIns(0, mapper.OpExport, exportTx.Ins)
	operations = append(operations, ins...)
	outs, err := c.parseExportedOutputs(len(ins), mapper.OpExport, hrp, exportTx.ExportedOutputs)
	if err != nil {
		return nil, err
	}
	operations = append(operations, outs...)

	return operations, nil
}

func (c *CChainAtomicTxBackend) buildImportTx(matches []*parser.Match, metadata cBackendMetadata) (*evm.Tx, []*types.AccountIdentifier, error) {
	importedInputs, signers, err := c.buildImportedInputs(matches)
	if err != nil {
		return nil, nil, err
	}

	outs := c.buildOuts(matches)

	tx := &evm.Tx{UnsignedAtomicTx: &evm.UnsignedImportTx{
		NetworkID:      metadata.NetworkID,
		BlockchainID:   metadata.CChainID,
		SourceChain:    *metadata.SourceChainID,
		ImportedInputs: importedInputs,
		Outs:           outs,
	}}
	return tx, signers, nil
}

func (c *CChainAtomicTxBackend) parseImportTx(importTx *evm.UnsignedImportTx) []*types.Operation {
	var operations []*types.Operation
	ins := c.parseImportedInputs(0, mapper.OpImport, importTx.ImportedInputs)
	operations = append(operations, ins...)
	outs := c.parseOuts(len(ins), mapper.OpImport, importTx.Outs)
	operations = append(operations, outs...)

	return operations
}

func (c *CChainAtomicTxBackend) buildIns(matches []*parser.Match, metadata cBackendMetadata) ([]evm.EVMInput, []*types.AccountIdentifier, error) {
	inputMatch := matches[0]

	var ins []evm.EVMInput
	var signers []*types.AccountIdentifier
	for i, op := range inputMatch.Operations {
		ins = append(ins, evm.EVMInput{
			Address: ethcommon.HexToAddress(op.Account.Address),
			Amount:  inputMatch.Amounts[i].Uint64(),
			AssetID: c.avaxAssetId,
			Nonce:   metadata.Nonce,
		})
		signers = append(signers, op.Account)
	}
	return ins, signers, nil
}

func (c *CChainAtomicTxBackend) parseIns(startIdx int64, opType string, ins []evm.EVMInput) []*types.Operation {
	idx := startIdx
	var operations []*types.Operation
	for _, in := range ins {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: idx,
			},
			Type:    opType,
			Account: &types.AccountIdentifier{Address: in.Address.Hex()},
			Amount: &types.Amount{
				Value:    strconv.FormatInt(-int64(in.Amount), 10),
				Currency: mapper.AvaxCurrency,
			},
		})
		idx++
	}
	return operations
}

func (c *CChainAtomicTxBackend) buildImportedInputs(matches []*parser.Match) ([]*avax.TransferableInput, []*types.AccountIdentifier, error) {
	inputMatch := matches[0]

	var importedInputs []*avax.TransferableInput
	var signers []*types.AccountIdentifier
	for i, op := range inputMatch.Operations {
		if op.CoinChange == nil || op.CoinChange.CoinIdentifier == nil {
			return nil, nil, errors.New("input operation does not have coin identifier")
		}
		utxoId, err := common.DecodeUTXOID(op.CoinChange.CoinIdentifier.Identifier)
		if err != nil {
			return nil, nil, err
		}

		importedInputs = append(importedInputs, &avax.TransferableInput{
			UTXOID: *utxoId,
			Asset:  avax.Asset{ID: c.avaxAssetId},
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

func (c *CChainAtomicTxBackend) parseImportedInputs(startIdx int64, opType string, ins []*avax.TransferableInput) []*types.Operation {
	idx := startIdx
	var operations []*types.Operation
	for _, in := range ins {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: idx,
			},
			Type: opType,
			// We are unable to get account information from UTXOs offline
			// therefore Account field is omitted for imported inputs
			Amount: &types.Amount{
				Value:    strconv.FormatInt(-int64(in.In.Amount()), 10),
				Currency: mapper.AvaxCurrency,
			},
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{Identifier: in.UTXOID.String()},
				CoinAction:     types.CoinSpent,
			},
		})
		idx++
	}
	return operations
}

func (c *CChainAtomicTxBackend) buildOuts(matches []*parser.Match) []evm.EVMOutput {
	outputMatch := matches[1]

	var outs []evm.EVMOutput
	for i, op := range outputMatch.Operations {
		outs = append(outs, evm.EVMOutput{
			Address: ethcommon.HexToAddress(op.Account.Address),
			Amount:  outputMatch.Amounts[i].Uint64(),
			AssetID: c.avaxAssetId,
		})
	}

	return outs
}

func (c *CChainAtomicTxBackend) parseOuts(startIdx int, opType string, outs []evm.EVMOutput) []*types.Operation {
	idx := startIdx
	var operations []*types.Operation
	for _, out := range outs {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(idx),
			},
			Account:           &types.AccountIdentifier{Address: out.Address.Hex()},
			RelatedOperations: buildRelatedOperations(startIdx),
			Type:              opType,
			Amount: &types.Amount{
				Value:    strconv.FormatUint(out.Amount, 10),
				Currency: mapper.AvaxCurrency,
			},
		})
		idx++
	}
	return operations
}

func (c *CChainAtomicTxBackend) buildExportedOutputs(matches []*parser.Match) ([]*avax.TransferableOutput, error) {
	outputMatch := matches[1]

	var outs []*avax.TransferableOutput
	for i, op := range outputMatch.Operations {
		destinationAddress, err := address.ParseToID(op.Account.Address)
		if err != nil {
			return nil, err
		}

		outs = append(outs, &avax.TransferableOutput{
			Asset: avax.Asset{ID: c.avaxAssetId},
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

	avax.SortTransferableOutputs(outs, c.codec)

	return outs, nil
}

func (c *CChainAtomicTxBackend) parseExportedOutputs(startIdx int, opType string, hrp string, outs []*avax.TransferableOutput) ([]*types.Operation, error) {
	idx := startIdx
	var operations []*types.Operation
	for _, out := range outs {
		var addr string
		transferOutput := out.Output().(*secp256k1fx.TransferOutput)
		if transferOutput != nil && len(transferOutput.Addrs) > 0 {
			var err error
			// TODO: chain alias is hardcoded for now, need to figure out how to fetch it from tx
			addr, err = address.Format(mapper.PChainIDAlias, hrp, transferOutput.Addrs[0][:])
			if err != nil {
				return nil, err
			}
		}

		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(idx),
			},
			Account:           &types.AccountIdentifier{Address: addr},
			RelatedOperations: buildRelatedOperations(startIdx),
			Type:              opType,
			Amount: &types.Amount{
				Value:    strconv.FormatUint(out.Out.Amount(), 10),
				Currency: mapper.AvaxCurrency,
			},
		})
		idx++
	}
	return operations, nil
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
