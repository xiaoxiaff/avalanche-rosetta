package pchain

import (
	"errors"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/types"
	"log"
	"math/big"

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
			Status:   types.String(mapper.StatusSuccess),
			Account:  &types.AccountIdentifier{Address: outAddrFormat, SubAccount: nil, Metadata: nil},
			Amount:   mapper.AvaxAmount(big.NewInt(int64(out.Out.Amount()))),
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
			Type:   opType,
			Status: types.String(mapper.StatusSuccess),
			Amount: mapper.AvaxAmount(big.NewInt(int64(in.In.Amount()))),
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

func rewardValidatorToOperation(v *platformvm.UnsignedRewardValidatorTx) []*types.Operation {
	return []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                OpRewardValidator,
			Status:              types.String(mapper.StatusSuccess),
			Metadata: map[string]interface{}{
				MetaStakingTxId: v.TxID.String(),
			},
		},
	}
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
		ops, err = baseTxToOperations(&v.BaseTx, mapper.OpExport)
		if err != nil {
			return nil, err
		}

		exportedOuts, err := outToOperation(v.ExportedOutputs, len(ops), mapper.OpExport, OpOutput)
		if err != nil {
			return nil, err
		}
		ops = append(ops, exportedOuts...)

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

		ops, err = baseTxToOperations(&v.BaseTx, OpAddValidator)
		if err != nil {
			return nil, err
		}

		stakeOuts, err := outToOperation(v.Stake, len(ops), OpAddValidator, OpStakeOutput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, stakeOuts...)

	case *platformvm.UnsignedAddDelegatorTx:
		id = v.ID()

		ops, err = baseTxToOperations(&v.BaseTx, OpAddDelegator)
		if err != nil {
			return nil, err
		}

		stakeOuts, err := outToOperation(v.Stake, len(ops), OpAddDelegator, OpStakeOutput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, stakeOuts...)
	case *platformvm.UnsignedRewardValidatorTx:
		id = v.ID()
		ops = rewardValidatorToOperation(v)
	case *platformvm.UnsignedAdvanceTimeTx:
		id = v.ID()
	case *platformvm.UnsignedCreateSubnetTx:
		id = v.ID()
	case *platformvm.UnsignedCreateChainTx:
		id = v.ID()
	case *platformvm.UnsignedAddSubnetValidatorTx:
		id = v.ID()
	default:
		// unknown transactions ignore operations
		ops = nil
		log.Printf("unknown type %T", v)
	}

	blockIdHexWithChecksum, err := formatting.EncodeWithChecksum(formatting.Hex, id[:])
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
