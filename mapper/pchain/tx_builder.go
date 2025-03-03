package pchain

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

var errInvalidMetadata = errors.New("invalid metadata")

func BuildTx(
	opType string,
	matches []*parser.Match,
	payloadMetadata Metadata,
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []*types.AccountIdentifier, error) {
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
	metadata Metadata,
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []*types.AccountIdentifier, error) {
	blockchainID := metadata.BlockchainID
	sourceChainID := metadata.SourceChainID

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
			NetworkID:    metadata.NetworkID,
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
	metadata Metadata,
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []*types.AccountIdentifier, error) {
	if metadata.ExportMetadata == nil {
		return nil, nil, errInvalidMetadata
	}
	blockchainID := metadata.BlockchainID
	destinationChainID := metadata.DestinationChainID

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
			NetworkID:    metadata.NetworkID,
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
	sMetadata Metadata,
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []*types.AccountIdentifier, error) {
	if sMetadata.StakingMetadata == nil {
		return nil, nil, errInvalidMetadata
	}

	blockchainID := sMetadata.BlockchainID

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
			Wght:   sumOutputAmounts(stakeOutputs),
		},
		RewardsOwner: rewardsOwner,
		Shares:       sMetadata.Shares,
	}}

	return tx, signers, nil
}

func buildAddDelegatorTx(
	matches []*parser.Match,
	sMetadata Metadata,
	codec codec.Manager,
	avaxAssetId ids.ID,
) (*platformvm.Tx, []*types.AccountIdentifier, error) {
	if sMetadata.StakingMetadata == nil {
		return nil, nil, errInvalidMetadata
	}

	blockchainID := sMetadata.BlockchainID

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
			Wght:   sumOutputAmounts(stakeOutputs),
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
	signers []*types.AccountIdentifier,
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
		signers = append(signers, op.Account)
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

	// set threshold default to 1
	if operationMetadata.Threshold == 0 {
		operationMetadata.Threshold = 1
	}

	// set sig indices to a single signer if not provided
	if operationMetadata.SigIndices == nil {
		operationMetadata.SigIndices = []uint32{0}
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

		addrID, err := address.ParseToID(op.Account.Address)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse output address failed: %w", err)
		}

		outputOwners := &secp256k1fx.OutputOwners{
			Addrs:     []ids.ShortID{addrID},
			Locktime:  opMetadata.Locktime,
			Threshold: opMetadata.Threshold,
		}

		val, err := types.AmountValue(op.Amount)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse operation amount failed: %w", err)
		}

		out := &avax.TransferableOutput{
			Asset: avax.Asset{ID: avaxAssetId},
			Out: &secp256k1fx.TransferOutput{
				Amt:          val.Uint64(),
				OutputOwners: *outputOwners,
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

func sumOutputAmounts(stakeOutputs []*avax.TransferableOutput) uint64 {
	var stakeOutputAmountSum uint64
	for _, out := range stakeOutputs {
		stakeOutputAmountSum += out.Output().Amount()
	}
	return stakeOutputAmountSum
}
