package pchain

import (
	"errors"
	"log"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/stakeable"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

var (
	errUnknownDestinationChain = errors.New("unknown destination chain")
)

type TxParser struct {
	isConstruction bool
	hrp            string
	chainIDs       map[string]string
}

func NewTxParser(isConstruction bool, hrp string, chainIDs map[string]string) *TxParser {
	return &TxParser{isConstruction: isConstruction, hrp: hrp, chainIDs: chainIDs}
}

func (t *TxParser) Parse(tx platformvm.UnsignedTx) (*types.Transaction, error) {
	var (
		ops []*types.Operation
		err error
		id  ids.ID
	)

	switch v := tx.(type) {
	case *platformvm.UnsignedExportTx:
		id = v.ID()
		ins, outs, err := t.baseTxToOperations(&v.BaseTx, OpExportAvax)
		if err != nil {
			return nil, err
		}
		ops = append(ops, ins...)
		ops = append(ops, outs...)

		chainID := v.DestinationChain.String()
		chainIDAlias, ok := t.chainIDs[chainID]
		if !ok {
			return nil, errUnknownDestinationChain
		}

		exportedOuts, err := t.outsToOperations(len(ops), OpExportAvax, v.ExportedOutputs, OpTypeExport, chainIDAlias)
		if err != nil {
			return nil, err
		}

		ops = append(ops, exportedOuts...)

	case *platformvm.UnsignedImportTx:
		id = v.ID()
		ins, err := t.insToOperations(0, OpImportAvax, v.Ins, OpTypeInput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, ins...)
		importedIns, err := t.insToOperations(len(ops), OpImportAvax, v.ImportedInputs, OpTypeImport)
		if err != nil {
			return nil, err
		}

		ops = append(ops, importedIns...)
		outs, err := t.outsToOperations(len(ops), OpImportAvax, v.Outs, OpTypeOutput, mapper.PChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}

		ops = append(ops, outs...)

	case *platformvm.UnsignedAddValidatorTx:
		id = v.ID()

		ins, outs, err := t.baseTxToOperations(&v.BaseTx, OpAddValidator)
		if err != nil {
			return nil, err
		}
		ops = append(ops, ins...)
		ops = append(ops, outs...)

		stakeOuts, err := t.outsToOperations(len(ops), OpAddValidator, v.Stake, OpTypeStakeOutput, mapper.PChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}

		ops = append(ops, stakeOuts...)

	case *platformvm.UnsignedAddDelegatorTx:
		id = v.ID()

		ins, outs, err := t.baseTxToOperations(&v.BaseTx, OpAddDelegator)
		if err != nil {
			return nil, err
		}
		ops = append(ops, ins...)
		ops = append(ops, outs...)

		stakeOuts, err := t.outsToOperations(len(ops), OpAddDelegator, v.Stake, OpTypeStakeOutput, mapper.PChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}

		ops = append(ops, stakeOuts...)
	case *platformvm.UnsignedRewardValidatorTx:
		id = v.ID()
		ops = t.rewardValidatorToOperation(v)
	case *platformvm.UnsignedAdvanceTimeTx:
		id = v.ID()
	case *platformvm.UnsignedCreateSubnetTx:
		id = v.ID()
		ins, err := t.insToOperations(0, OpCreateSubnet, v.Ins, OpTypeInput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, ins...)

		outs, err := t.outsToOperations(len(ops), OpCreateSubnet, v.Outs, OpTypeOutput, mapper.PChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}

		ops = append(ops, outs...)
	case *platformvm.UnsignedCreateChainTx:
		id = v.ID()
		ops = t.createChainToOperation(v)
		ins, err := t.insToOperations(0, OpCreateChain, v.Ins, OpTypeInput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, ins...)

		outs, err := t.outsToOperations(len(ops), OpCreateChain, v.Outs, OpTypeOutput, mapper.PChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}

		ops = append(ops, outs...)
	case *platformvm.UnsignedAddSubnetValidatorTx:
		id = v.ID()
		ins, err := t.insToOperations(0, OpAddSubnetValidator, v.Ins, OpTypeInput)
		if err != nil {
			return nil, err
		}

		ops = append(ops, ins...)

		outs, err := t.outsToOperations(len(ops), OpAddSubnetValidator, v.Outs, OpTypeOutput, mapper.PChainNetworkIdentifier)
		if err != nil {
			return nil, err
		}

		ops = append(ops, outs...)
	default:
		// unknown transactions ignore operations
		ops = nil
		log.Printf("unknown type %T", v)
	}

	blockIdHexWithChecksum, err := formatting.EncodeWithChecksum(formatting.Hex, id[:])
	if err != nil {
		return nil, err
	}

	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: blockIdHexWithChecksum,
		},
		Operations: ops,
	}, nil
}

func (t *TxParser) outsToOperations(
	startIndex int,
	opType string,
	txOut []*avax.TransferableOutput,
	metaType string,
	chainIDAlias string,
) ([]*types.Operation, error) {
	status := types.String(mapper.StatusSuccess)
	if t.isConstruction {
		status = nil
	}

	outs := make([]*types.Operation, 0)
	for _, out := range txOut {
		transferOut := out.Out

		if lockOut, ok := transferOut.(*stakeable.LockOut); ok {
			transferOut = lockOut.TransferableOut
		}

		transferOutput, ok := transferOut.(*secp256k1fx.TransferOutput)
		if !ok {
			return nil, errors.New("output type assertion failed")
		}

		outOp, err := t.BuildOutputOperation(transferOutput, status, startIndex, opType, metaType, chainIDAlias)
		if err != nil {
			return nil, err
		}
		outs = append(outs, outOp)
		startIndex++
	}

	return outs, nil
}

func (t *TxParser) BuildOutputOperation(
	out *secp256k1fx.TransferOutput,
	status *string,
	startIndex int,
	opType,
	metaType string,
	chainIDAlias string,
) (*types.Operation, error) {
	if len(out.Addrs) == 0 {
		return nil, errors.New("empty output addresses")
	}

	outAddrID := out.Addrs[0]
	outAddrFormat, err := address.Format(chainIDAlias, t.hrp, outAddrID[:])
	if err != nil {
		return nil, err
	}

	metadata := &OperationMetadata{
		Type:      metaType,
		Threshold: out.OutputOwners.Threshold,
		Locktime:  out.OutputOwners.Locktime,
		MultiSig:  len(out.Addrs) > 1,
	}

	opMetadata, err := mapper.MarshalJSONMap(metadata)
	if err != nil {
		return nil, err
	}

	return &types.Operation{
		Type: opType,
		OperationIdentifier: &types.OperationIdentifier{
			Index: int64(startIndex),
		},
		Status:   status,
		Account:  &types.AccountIdentifier{Address: outAddrFormat, SubAccount: nil, Metadata: nil},
		Amount:   mapper.AvaxAmount(big.NewInt(int64(out.Amount()))),
		Metadata: opMetadata,
	}, nil
}

func (t *TxParser) insToOperations(
	startIndex int,
	opType string,
	txIns []*avax.TransferableInput,
	metaType string,
) ([]*types.Operation, error) {
	status := types.String(mapper.StatusSuccess)
	if t.isConstruction {
		status = nil
	}

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

		inputAmount := new(big.Int).SetUint64(in.In.Amount())
		inOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(startIndex),
			},
			Type:   opType,
			Status: status,
			// Negating input amount
			Amount: mapper.AvaxAmount(new(big.Int).Neg(inputAmount)),
			CoinChange: &types.CoinChange{
				CoinIdentifier: &types.CoinIdentifier{
					Identifier: in.UTXOID.String(),
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

func (t *TxParser) baseTxToOperations(tx *platformvm.BaseTx, txType string) ([]*types.Operation, []*types.Operation, error) {

	ins, err := t.insToOperations(0, txType, tx.Ins, OpTypeInput)
	if err != nil {
		return nil, nil, err
	}

	outs, err := t.outsToOperations(len(ins), txType, tx.Outs, OpTypeOutput, mapper.PChainNetworkIdentifier)
	if err != nil {
		return nil, nil, err
	}

	return ins, outs, nil
}

func (*TxParser) rewardValidatorToOperation(v *platformvm.UnsignedRewardValidatorTx) []*types.Operation {
	return []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                OpRewardValidator,
			Status:              types.String(mapper.StatusSuccess),
			Metadata: map[string]interface{}{
				MetadataStakingTxID: v.TxID.String(),
				MetadataOpType:      OpTypeReward,
			},
		},
	}
}

func (*TxParser) createChainToOperation(v *platformvm.UnsignedCreateChainTx) []*types.Operation {
	return []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                OpCreateChain,
			Status:              types.String(mapper.StatusSuccess),
			Metadata: map[string]interface{}{
				MetadataSubnetID:  v.SubnetID.String(),
				MetadataChainName: v.ChainName,
				MetadataVMID:      v.VMID,
				MetadataMemo:      v.Memo,
				MetadataOpType:    OpTypeCreateChain,
			},
		},
	}
}
