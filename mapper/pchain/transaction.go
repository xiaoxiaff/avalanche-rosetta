package pchain

import (
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/parser"
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
	ins, err := inToOperation(tx.Ins, 0, txType, OpTypeInput)
	if err != nil {
		return nil, err
	}

	outs, err := outToOperation(tx.Outs, len(ins), txType, OpTypeOutput)
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
				MetadataStakingTxId: v.TxID.String(),
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

		exportedOuts, err := outToOperation(v.ExportedOutputs, len(ops), mapper.OpExport, OpTypeOutput)
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

		importedIns, err := inToOperation(v.ImportedInputs, len(ops), mapper.OpImport, OpTypeImport)
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

		stakeOuts, err := outToOperation(v.Stake, len(ops), OpAddValidator, OpTypeStakeOutput)
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

		stakeOuts, err := outToOperation(v.Stake, len(ops), OpAddDelegator, OpTypeStakeOutput)
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

func BuildTransaction(
	opType string,
	matches []*parser.Match,
	payloadMetadata map[string]interface{},
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []string, error) {
	switch opType {
	case OpImportAvax:
		return buildImportTx(matches, payloadMetadata, codec, avaxAssetId)
	case OpExportAvax:
		return buildExportTx(matches, payloadMetadata, codec, avaxAssetId)
	case OpAddValidator:
		return buildAddValidatorTx(matches, payloadMetadata, codec, avaxAssetId)
	case OpAddDelegator:
		return buildAddDelegatorTx(matches, payloadMetadata, codec, avaxAssetId)
	default:
		return nil, nil, fmt.Errorf("invalid tx type: %s", opType)
	}
}

func buildImportTx(
	matches []*parser.Match,
	metadata map[string]interface{},
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []string, error) {
	var txMetadata ImportExportMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &txMetadata); err != nil {
		return nil, nil, err
	}
	blockchainID, err := ids.FromString(txMetadata.BlockchainID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid blockchain id: %s", txMetadata.BlockchainID)
	}

	sourceChainID, err := ids.FromString(txMetadata.SourceChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid source chain id: %s", txMetadata.SourceChainID)
	}

	ins, imported, signers, err := buildInputs(matches[0].Operations, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, _, _, err := buildOutputs(matches[1].Operations, codec, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedImportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    txMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
		}},
		ImportedInputs: imported,
		SourceChain:    sourceChainID,
	}}

	return tx, signers, nil
}

func buildExportTx(
	matches []*parser.Match,
	metadata map[string]interface{},
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []string, error) {
	var txMetadata ImportExportMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &txMetadata); err != nil {
		return nil, nil, err
	}
	blockchainID, err := ids.FromString(txMetadata.BlockchainID)
	if err != nil {
		return nil, nil, err
	}

	destinationChainID, err := ids.FromString(txMetadata.DestinationChainID)
	if err != nil {
		return nil, nil, err
	}

	ins, _, signers, err := buildInputs(matches[0].Operations, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, _, exported, err := buildOutputs(matches[1].Operations, codec, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedExportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    txMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
		}},
		DestinationChain: destinationChainID,
		ExportedOutputs:  exported,
	}}

	return tx, signers, nil
}

func buildAddValidatorTx(
	matches []*parser.Match,
	metadata map[string]interface{},
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []string, error) {
	var sMetadata StakingMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &sMetadata); err != nil {
		return nil, nil, err
	}

	blockchainID, err := ids.FromString(sMetadata.BlockchainID)
	if err != nil {
		return nil, nil, err
	}
	nodeID, err := ids.NodeIDFromString(sMetadata.NodeID)
	if err != nil {
		return nil, nil, err
	}

	rewardsOwner, err := buildOutputOwner(
		sMetadata.RewardAddresses,
		sMetadata.Locktime,
		sMetadata.Threshold,
	)
	if err != nil {
		return nil, nil, err
	}

	ins, _, signers, err := buildInputs(matches[0].Operations, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, stakeOutputs, _, err := buildOutputs(matches[1].Operations, codec, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	memo, err := mapper.DecodeToBytes(sMetadata.Memo)
	if err != nil {
		return nil, nil, fmt.Errorf("parse memo failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedAddValidatorTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    sMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
			Memo:         memo,
		}},
		Stake: stakeOutputs,
		Validator: validator.Validator{
			NodeID: nodeID,
			Start:  sMetadata.Start,
			End:    sMetadata.End,
			Wght:   sMetadata.Wght,
		},
		RewardsOwner: rewardsOwner,
		Shares:       sMetadata.Shares,
	}}

	return tx, signers, nil
}

func buildAddDelegatorTx(
	matches []*parser.Match,
	metadata map[string]interface{},
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []string, error) {
	var sMetadata StakingMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &sMetadata); err != nil {
		return nil, nil, err
	}

	blockchainID, err := ids.FromString(sMetadata.BlockchainID)
	if err != nil {
		return nil, nil, err
	}
	nodeID, err := ids.NodeIDFromString(sMetadata.NodeID)
	if err != nil {
		return nil, nil, err
	}
	rewardsOwner, err := buildOutputOwner(sMetadata.RewardAddresses, sMetadata.Locktime, sMetadata.Threshold)
	if err != nil {
		return nil, nil, err
	}

	ins, _, signers, err := buildInputs(matches[0].Operations, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, stakeOutputs, _, err := buildOutputs(matches[1].Operations, codec, avaxAssetId)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	memo, err := mapper.DecodeToBytes(sMetadata.Memo)
	if err != nil {
		return nil, nil, fmt.Errorf("parse memo failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedAddDelegatorTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    sMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
			Memo:         memo,
		}},
		Stake: stakeOutputs,
		Validator: validator.Validator{
			NodeID: nodeID,
			Start:  sMetadata.Start,
			End:    sMetadata.End,
			Wght:   sMetadata.Wght,
		},
		RewardsOwner: rewardsOwner,
	}}

	return tx, signers, nil
}

func buildOutputOwner(
	addrs []string,
	locktime uint64,
	threshold uint32,
) (*secp256k1fx.OutputOwners, error) {

	rewardAddrs := make([]ids.ShortID, len(addrs))
	for i, addr := range addrs {

		addrID, err := address.ParseToID(addr)
		if err != nil {
			return nil, err
		}
		rewardAddrs[i] = addrID
	}

	ids.SortShortIDs(rewardAddrs)

	return &secp256k1fx.OutputOwners{
		Locktime:  locktime,
		Threshold: threshold,
		Addrs:     rewardAddrs,
	}, nil
}

func buildInputs(
	operations []*types.Operation,
	avaxAssetId ids.ID,
) (
	ins []*avax.TransferableInput,
	imported []*avax.TransferableInput,
	signers []string,
	err error,
) {
	for _, op := range operations {
		UTXOID, err := mapper.DecodeUTXOID(op.CoinChange.CoinIdentifier.Identifier)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode UTXO ID: %w", err)
		}

		addr, err := address.ParseToID(op.Account.Address)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse address: %w", err)
		}
		addrSet := ids.NewShortSet(1)
		addrSet.Add(addr)

		opMetadata, err := ParseOpMetadata(op.Metadata)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse input operation Metadata failed: %w", err)
		}

		val, err := types.AmountValue(op.Amount)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse operation amount failed: %w", err)
		}

		in := &avax.TransferableInput{
			UTXOID: *UTXOID,
			Asset:  avax.Asset{ID: avaxAssetId},
			In: &secp256k1fx.TransferInput{
				Amt: val.Uint64(),
				Input: secp256k1fx.Input{
					SigIndices: opMetadata.SigIndices,
				},
			}}

		switch opMetadata.Type {
		case OpTypeImport:
			imported = append(imported, in)
		case OpTypeInput:
			ins = append(ins, in)
		default:
			return nil, nil, nil, fmt.Errorf("invalid option type: %s", op.Type)
		}
		signers = append(signers, op.Account.Address)
	}

	avax.SortTransferableInputs(ins)
	avax.SortTransferableInputs(imported)

	return ins, imported, signers, nil
}

func ParseOpMetadata(metadata map[string]interface{}) (*OperationMetadata, error) {
	var operationMetadata OperationMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &operationMetadata); err != nil {
		return nil, err
	}

	return &operationMetadata, nil
}

func buildOutputs(
	operations []*types.Operation,
	codec codec.Manager,
	avaxAssetId ids.ID,
) (
	outs []*avax.TransferableOutput,
	stakeOutputs []*avax.TransferableOutput,
	exported []*avax.TransferableOutput,
	err error,
) {
	for _, op := range operations {
		opMetadata, err := ParseOpMetadata(op.Metadata)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse output operation Metadata failed: %w", err)
		}

		var outputOwners secp256k1fx.OutputOwners

		outputOwnerBytes, err := mapper.DecodeToBytes(opMetadata.OutputOwners)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode output owner hex failed: %w", err)
		}
		if _, err := codec.Unmarshal(outputOwnerBytes, &outputOwners); err != nil {
			return nil, nil, nil, fmt.Errorf("parse output owner failed: %w", err)
		}

		val, err := types.AmountValue(op.Amount)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse operation amount failed: %w", err)
		}

		out := &avax.TransferableOutput{
			Asset: avax.Asset{ID: avaxAssetId},
			Out: &secp256k1fx.TransferOutput{
				Amt:          val.Uint64(),
				OutputOwners: outputOwners,
			}}

		switch opMetadata.Type {
		case OpTypeOutput:
			outs = append(outs, out)
		case OpTypeStakeOutput:
			stakeOutputs = append(stakeOutputs, out)
		case OpTypeExport:
			exported = append(exported, out)
		default:
			return nil, nil, nil, fmt.Errorf("invalid option type: %s", op.Type)
		}
	}

	avax.SortTransferableOutputs(outs, codec)
	avax.SortTransferableOutputs(stakeOutputs, codec)
	avax.SortTransferableOutputs(exported, codec)

	return outs, stakeOutputs, exported, nil
}
