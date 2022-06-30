package mapper

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

func outToOperation(txOut []*avax.TransferableOutput, startIndex int, opType string, metaType string) ([]*types.Operation, error) {

	outs := make([]*types.Operation, 0)
	for _, out := range txOut {
		outAddrID := out.Out.(*secp256k1fx.TransferOutput).Addrs[0]
		//TODO: [NM] use variables form somewhere
		outAddrFormat, err := address.Format("P", "fuji", outAddrID[:])
		if err != nil {
			return nil, err
		}

		metadata := &OperationMetadata{
			Type: metaType,
		}

		if transferOutput, ok := out.Out.(*secp256k1fx.TransferOutput); ok {
			outputOwnersBytes, err := platformvm.Codec.Marshal(platformvm.CodecVersion, transferOutput.OutputOwners)
			if err != nil {
				return nil, err
			}

			outputOwnersHex, err := formatting.EncodeWithChecksum(formatting.Hex, outputOwnersBytes)
			if err != nil {
				return nil, err
			}

			metadata.OutputOwners = outputOwnersHex
		}

		opMetadata, err := mapper.MarshalJSONMap(metadata)
		if err != nil {
			return nil, err
		}

		outOp := &types.Operation{
			Type: opType,
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(startIndex),
			},
			Account: &types.AccountIdentifier{Address: outAddrFormat, SubAccount: nil, Metadata: nil},
			Amount: &types.Amount{
				Value:    fmt.Sprint(out.Output().Amount()),
				Currency: mapper.AvaxCurrency,
			},
			Metadata: opMetadata,
		}
		outs = append(outs, outOp)
		startIndex++
	}

	return outs, nil
}

func inToOperation(txIns []*avax.TransferableInput, startIndex int, opType string, metaType string) ([]*types.Operation, error) {

	ins := make([]*types.Operation, 0)
	for _, in := range txIns {
		metadata := &OperationMetadata{
			Type: metaType,
		}

		if transferInput, ok := in.In.(*secp256k1fx.TransferInput); ok {
			metadata.SigIndices = transferInput.SigIndices
		}

		opMetadata, err := mapper.MarshalJSONMap(metadata)
		if err != nil {
			return nil, err
		}

		inOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(startIndex),
			},
			Type: string(opType),
			Amount: &types.Amount{
				Value:    fmt.Sprint(in.Input().Amount()),
				Currency: mapper.AvaxCurrency,
			},
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{
					Identifier: in.AssetID().String(),
				},
				CoinAction: types.CoinSpent,
			},
			Metadata: opMetadata,
		}

		ins = append(ins, inOp)
		startIndex++
	}
	return ins, nil
}

func baseTxToOperations(tx *platformvm.BaseTx, txType string) ([]*types.Operation, error) {

	insAndOuts := make([]*types.Operation, 0)
	ins, err := inToOperation(tx.Ins, 0, txType, OpInput)
	if err != nil {
		return nil, err
	}

	outs, err := outToOperation(tx.Outs, len(ins), txType, OpOutput)
	if err != nil {
		return nil, err
	}

	insAndOuts = append(insAndOuts, ins...)
	insAndOuts = append(insAndOuts, outs...)

	return insAndOuts, nil
}

func Transaction(tx interface{}) (*types.Transaction, error) {
	var (
		ops []*types.Operation
		err error
		id  ids.ID
	)

	switch v := tx.(type) {
	case nil:
		return nil, errors.New("tx unknown")
	case *platformvm.UnsignedExportTx:
		id = v.ID()
		ops, err = outToOperation(v.Outs, 0, mapper.OpExport, OpOutput)
		if err != nil {
			return nil, err
		}
	case *platformvm.UnsignedImportTx:
		id = v.ID()
		ops, err = baseTxToOperations(&v.BaseTx, mapper.OpImport)
		if err != nil {
			return nil, err
		}

		importedIns, err := inToOperation(v.ImportedInputs, len(ops), mapper.OpImport, OpImport)
		if err != nil {
			return nil, err
		}

		ops = append(ops, importedIns...)

	case *platformvm.UnsignedAddValidatorTx:
		id = v.ID()

		ops, err = baseTxToOperations(&v.BaseTx, mapper.OpAddValidator)
		if err != nil {
			return nil, err
		}

		stakeOuts, err := outToOperation(v.Stake, len(ops), mapper.OpAddValidator, OpStakeOutput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, stakeOuts...)

	case *platformvm.UnsignedAddDelegatorTx:
		id = v.ID()

		ops, err = baseTxToOperations(&v.BaseTx, mapper.OpAddDelegator)
		if err != nil {
			return nil, err
		}

		stakeOuts, err := outToOperation(v.Stake, len(ops), mapper.OpAddValidator, OpStakeOutput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, stakeOuts...)
	default:
		// unknown transaction ignore operation.
		ops = nil
	}

	blockIdHexWithChecksum, err := formatting.EncodeWithChecksum(formatting.Hex, []byte(id.String()))
	if err != nil {
		return nil, err
	}

	t := &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: blockIdHexWithChecksum,
		},
		Operations: ops,
	}

	return t, nil
}
