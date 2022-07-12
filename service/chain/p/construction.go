package p

import (
	"context"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	p "github.com/ava-labs/avalanche-rosetta/mapper/p"
	"github.com/ava-labs/avalanche-rosetta/service"

	"github.com/ava-labs/avalanche-rosetta/service/chain/common"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
)

var codecVersion uint16 = platformvm.CodecVersion

func (c *Backend) ConstructionDerive(
	ctx context.Context,
	req *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	return common.DeriveBech32Address(c.fac, mapper.PChainIDAlias, req)
}
func (c *Backend) ConstructionPreprocess(
	ctx context.Context,
	req *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	opType, err := common.ParseOpType(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	reqMetadata := req.Metadata
	reqMetadata["type"] = opType

	return &types.ConstructionPreprocessResponse{
		Options: reqMetadata,
	}, nil
}
func (c *Backend) ConstructionMetadata(
	ctx context.Context,
	req *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	opMetadata, err := parseOpMetadata(req.Options)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	networkID, err := c.pClient.GetNetworkID(context.Background())
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	// Getting Chain ID from Info APIs
	pChainID, err := c.pClient.GetBlockchainID(ctx, mapper.PChainNetworkIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	var metadataMap map[string]interface{}
	switch opMetadata.Type {
	case mapper.OpImportAvax, mapper.OpExportAvax:
		metadataMap, err = c.buildTxMetadata(ctx, opMetadata.Type, req.Options, networkID, pChainID)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
	case mapper.OpAddValidator, mapper.OpAddDelegator:
		metadataMap, err = c.buildStakingMetadata(ctx, opMetadata.Type, req.Options, networkID, pChainID)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
	default:
		return nil, service.WrapError(
			service.ErrInternalError,
			fmt.Errorf("invalid tx type for building metadata: %s", opMetadata.Type),
		)
	}

	return &types.ConstructionMetadataResponse{
		Metadata: metadataMap,
	}, nil
}

func (c *Backend) buildTxMetadata(
	ctx context.Context,
	txType string,
	options map[string]interface{},
	networkID uint32,
	pChainID ids.ID,
) (map[string]interface{}, error) {
	var preprocessOptions txOptions
	if err := mapper.UnmarshalJSONMap(options, &preprocessOptions); err != nil {
		return nil, err
	}

	txMetadata := &txMetadata{
		NetworkID:    networkID,
		BlockchainID: pChainID.String(),
	}

	switch txType {
	case mapper.OpImportAvax:
		sourceChainID, err := c.pClient.GetBlockchainID(ctx, preprocessOptions.SourceChain)
		if err != nil {
			return nil, err
		}
		txMetadata.SourceChainID = sourceChainID.String()
	case mapper.OpExportAvax:
		destinationChainID, err := c.pClient.GetBlockchainID(ctx, preprocessOptions.DestinationChain)
		if err != nil {
			return nil, err
		}
		txMetadata.DestinationChainID = destinationChainID.String()
	default:
		return nil, fmt.Errorf("invalid tx type for building tx metadata: %s", txType)
	}
	return mapper.MarshalJSONMap(txMetadata)
}

func (c *Backend) buildStakingMetadata(
	ctx context.Context,
	txType string,
	options map[string]interface{},
	networkID uint32,
	pChainID ids.ID,
) (map[string]interface{}, error) {
	var preprocessOptions stakingOptions
	if err := mapper.UnmarshalJSONMap(options, &preprocessOptions); err != nil {
		return nil, err
	}

	stakingMetadata := &stakingMetadata{
		NodeID:          preprocessOptions.NodeID,
		Start:           preprocessOptions.Start,
		End:             preprocessOptions.End,
		Wght:            preprocessOptions.Wght,
		Memo:            preprocessOptions.Memo,
		NetworkID:       networkID,
		BlockchainID:    pChainID.String(),
		Locktime:        preprocessOptions.Locktime,
		Threshold:       preprocessOptions.Threshold,
		RewardAddresses: preprocessOptions.RewardAddresses,
		Shares:          preprocessOptions.Shares,
	}

	return mapper.MarshalJSONMap(stakingMetadata)
}

func (c *Backend) ConstructionPayloads(
	ctx context.Context,
	req *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	opType, err := common.ParseOpType(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	matches, err := common.MatchOperations(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrBlockInvalidInput, err)
	}

	tx, signers, err := c.buildTransaction(ctx, opType, matches, req.Metadata)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	unsignedBytes, err := c.codec.Marshal(codecVersion, &tx.UnsignedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, fmt.Errorf("couldn't marshal UnsignedTx: %w", err))
	}

	hash := hashing.ComputeHash256(unsignedBytes)

	payloads := make([]*types.SigningPayload, len(signers))

	for i, signer := range signers {
		payloads[i] = &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{Address: signer},
			Bytes:             hash,
			SignatureType:     types.EcdsaRecovery,
		}
	}

	txBytes, err := c.codec.Marshal(codecVersion, tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: string(txBytes),
		Payloads:            payloads,
	}, nil
}

func (c *Backend) buildTransaction(
	ctx context.Context,
	opType string,
	matches []*parser.Match,
	payloadMetadata map[string]interface{},
) (*platformvm.Tx, []string, error) {
	switch opType {
	case mapper.OpImportAvax:
		return c.buildImportTx(ctx, matches, payloadMetadata)
	case mapper.OpExportAvax:
		return c.buildExportTx(ctx, matches, payloadMetadata)
	case mapper.OpAddValidator:
		return c.buildAddValidatorTx(ctx, matches, payloadMetadata)
	case mapper.OpAddDelegator:
		return c.buildAddDelegatorTx(ctx, matches, payloadMetadata)
	default:
		return nil, nil, fmt.Errorf("invalid tx type: %s", opType)
	}
}

func (c *Backend) buildImportTx(
	ctx context.Context,
	matches []*parser.Match,
	metadata map[string]interface{},
) (*platformvm.Tx, []string, error) {
	var tMetadata txMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &tMetadata); err != nil {
		return nil, nil, err
	}
	blockchainID, err := ids.FromString(tMetadata.BlockchainID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid blockchain id: %s", tMetadata.BlockchainID)
	}

	sourceChainID, err := ids.FromString(tMetadata.SourceChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid source chain id: %s", tMetadata.SourceChainID)
	}

	ins, imported, signers, err := c.buildInputs(matches[0].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, _, _, err := c.buildOutputs(matches[1].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedImportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    tMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
		}},
		ImportedInputs: imported,
		SourceChain:    sourceChainID,
	}}

	return tx, signers, nil
}

func (c *Backend) buildExportTx(
	ctx context.Context,
	matches []*parser.Match,
	metadata map[string]interface{},
) (*platformvm.Tx, []string, error) {
	var tMetadata txMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &tMetadata); err != nil {
		return nil, nil, err
	}
	blockchainID, err := ids.FromString(tMetadata.BlockchainID)
	if err != nil {
		return nil, nil, err
	}

	destinationChainID, err := ids.FromString(tMetadata.DestinationChainID)
	if err != nil {
		return nil, nil, err
	}

	ins, _, signers, err := c.buildInputs(matches[0].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, _, exported, err := c.buildOutputs(matches[1].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedExportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    tMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
		}},
		DestinationChain: destinationChainID,
		ExportedOutputs:  exported,
	}}

	return tx, signers, nil
}

func (c *Backend) buildAddValidatorTx(
	ctx context.Context,
	matches []*parser.Match,
	metadata map[string]interface{},
) (*platformvm.Tx, []string, error) {
	var sMetadata stakingMetadata
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

	ins, _, signers, err := c.buildInputs(matches[0].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, stakeOutputs, _, err := c.buildOutputs(matches[1].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	memo, err := common.DecodeToBytes(sMetadata.Memo)
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

func (c *Backend) buildAddDelegatorTx(
	ctx context.Context,
	matches []*parser.Match,
	metadata map[string]interface{},
) (*platformvm.Tx, []string, error) {
	var sMetadata stakingMetadata
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

	ins, _, signers, err := c.buildInputs(matches[0].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inputs failed: %w", err)
	}

	outs, stakeOutputs, _, err := c.buildOutputs(matches[1].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	memo, err := common.DecodeToBytes(sMetadata.Memo)
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

func (c *Backend) buildInputs(
	operations []*types.Operation,
) (
	ins []*avax.TransferableInput,
	imported []*avax.TransferableInput,
	signers []string,
	err error,
) {
	for _, op := range operations {
		UTXOID, err := common.DecodeUTXOID(op.CoinChange.CoinIdentifier.Identifier)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode UTXO ID: %w", err)
		}

		addr, err := address.ParseToID(op.Account.Address)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse address: %w", err)
		}
		addrSet := ids.NewShortSet(1)
		addrSet.Add(addr)

		opMetadata, err := parseOpMetadata(op.Metadata)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse input operation Metadata failed: %w", err)
		}

		val, err := types.AmountValue(op.Amount)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse operation amount failed: %w", err)
		}

		in := &avax.TransferableInput{
			UTXOID: *UTXOID,
			Asset:  avax.Asset{ID: c.assetID},
			In: &secp256k1fx.TransferInput{
				Amt: val.Uint64(),
				Input: secp256k1fx.Input{
					SigIndices: opMetadata.SigIndices,
				},
			}}

		switch opMetadata.Type {
		case p.OpImport:
			imported = append(imported, in)
		case p.OpInput:
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

func (c *Backend) buildOutputs(
	operations []*types.Operation,
) (
	outs []*avax.TransferableOutput,
	stakeOutputs []*avax.TransferableOutput,
	exported []*avax.TransferableOutput,
	err error,
) {
	for _, op := range operations {
		opMetadata, err := parseOpMetadata(op.Metadata)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse output operation Metadata failed: %w", err)
		}

		var outputOwners secp256k1fx.OutputOwners

		outputOwnerBytes, err := common.DecodeToBytes(opMetadata.OutputOwners)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("decode output owner hex failed: %w", err)
		}
		if _, err := c.codec.Unmarshal(outputOwnerBytes, &outputOwners); err != nil {
			return nil, nil, nil, fmt.Errorf("parse output owner failed: %w", err)
		}

		val, err := types.AmountValue(op.Amount)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse operation amount failed: %w", err)
		}

		out := &avax.TransferableOutput{
			Asset: avax.Asset{ID: c.assetID},
			Out: &secp256k1fx.TransferOutput{
				Amt:          val.Uint64(),
				OutputOwners: outputOwners,
			}}

		switch opMetadata.Type {
		case p.OpOutput:
			outs = append(outs, out)
		case p.OpStakeOutput:
			stakeOutputs = append(stakeOutputs, out)
		case p.OpExport:
			exported = append(exported, out)
		default:
			return nil, nil, nil, fmt.Errorf("invalid option type: %s", op.Type)
		}
	}

	avax.SortTransferableOutputs(outs, c.codec)
	avax.SortTransferableOutputs(stakeOutputs, c.codec)
	avax.SortTransferableOutputs(exported, c.codec)

	return outs, stakeOutputs, exported, nil
}

func parseOpMetadata(metadata map[string]interface{}) (*p.OperationMetadata, error) {
	var operationMetadata p.OperationMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &operationMetadata); err != nil {
		return nil, err
	}

	return &operationMetadata, nil
}

func (c *Backend) ConstructionParse(ctx context.Context, req *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	tx := platformvm.Tx{}

	_, err := c.codec.Unmarshal([]byte(req.Transaction), &tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	transactions, err := p.Transaction(tx.UnsignedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	signers := make([]string, 0)
	for _, v := range transactions.Operations {
		opMetadata, _ := parseOpMetadata(v.Metadata)
		switch opMetadata.Type {
		case p.OpImport, p.OpInput:
			signers = append(signers, v.Account.Address)
		}
	}
	accountIDSigners := make([]*types.AccountIdentifier, len(signers))

	for _, v := range signers {
		ai := &types.AccountIdentifier{Address: v}
		accountIDSigners = append(accountIDSigners, ai)

	}

	resp := &types.ConstructionParseResponse{
		Operations:               transactions.Operations,
		AccountIdentifierSigners: accountIDSigners,
		Metadata:                 nil,
	}

	return resp, nil
}

func (c *Backend) ConstructionCombine(ctx context.Context, req *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	tx := platformvm.Tx{}

	_, err := c.codec.Unmarshal([]byte(req.UnsignedTransaction), &tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	ins, err := getTxInputs(tx.UnsignedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	creds, err := common.BuildCredentialList(ins, req.Signatures)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	tx.Creds = creds

	signedBytes, err := platformvm.Codec.Marshal(platformvm.CodecVersion, tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	signedTx, err := common.EncodeBytes(signedBytes)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: signedTx,
	}, nil
}

// getTxInputs fetches tx inputs based on the tx type.
func getTxInputs(
	unsignedTx platformvm.UnsignedTx,
) ([]*avax.TransferableInput, error) {
	switch utx := unsignedTx.(type) {
	case *platformvm.UnsignedAddValidatorTx:
		return utx.Ins, nil
	case *platformvm.UnsignedAddSubnetValidatorTx:
		return utx.Ins, nil
	case *platformvm.UnsignedAddDelegatorTx:
		return utx.Ins, nil
	case *platformvm.UnsignedCreateChainTx:
		return utx.Ins, nil
	case *platformvm.UnsignedCreateSubnetTx:
		return utx.Ins, nil
	case *platformvm.UnsignedImportTx:
		return utx.ImportedInputs, nil
	case *platformvm.UnsignedExportTx:
		return utx.Ins, nil
	default:
		return nil, errors.New("unknown tx type")
	}
}

func (c *Backend) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return common.HashTx(req)
}

func (c *Backend) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return common.SubmitTx(c, ctx, req)
}

// Defining IssueTx here without rpc.Options... to be able to use it with common.SubmitTx
func (c *Backend) IssueTx(ctx context.Context, txByte []byte) (ids.ID, error) {
	return c.pClient.IssueTx(ctx, txByte)
}
