package p

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	p "github.com/ava-labs/avalanche-rosetta/mapper/p"
	"github.com/ava-labs/avalanche-rosetta/service"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
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
	pub, err := c.fac.ToPublicKey(req.PublicKey.Bytes)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	chainIDAlias, hrp, getErr := mapper.GetAliasAndHRP(req.NetworkIdentifier)
	if getErr != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	addr, err := address.Format(chainIDAlias, hrp, pub.Address().Bytes())
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: addr,
		},
	}, nil
}
func (c *Backend) ConstructionPreprocess(
	ctx context.Context,
	req *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	opType, err := parseOpType(req.Operations)
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
	case mapper.OpImport, mapper.OpExport:
		metadataMap, err = c.buildTxMetadata(ctx, opMetadata.Type, req.Options, networkID, pChainID)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
	case mapper.OpAddValidator, mapper.OpAddDelegator:
		metadataMap, err = c.buildStakingMetadata(ctx, opMetadata.Type, req.Options, networkID, pChainID)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
	}

	txFee, err := c.pClient.GetTxFee(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return &types.ConstructionMetadataResponse{
		Metadata: metadataMap,
		SuggestedFee: []*types.Amount{
			mapper.AvaxAmount(big.NewInt(int64(txFee.TxFee))),
		},
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
	case mapper.OpImport:
		sourceChainID, err := c.pClient.GetBlockchainID(ctx, preprocessOptions.SourceChain)
		if err != nil {
			return nil, err
		}
		txMetadata.SourceChainID = sourceChainID.String()
	case mapper.OpExport:
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
	opType, err := parseOpType(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	matches, err := parser.MatchOperations(createOperationDescription(opType), req.Operations)
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

	unsignedTxJSON, err := json.Marshal(tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: string(unsignedTxJSON),
		Payloads:            payloads,
	}, nil
}

func parseOpType(operations []*types.Operation) (string, error) {
	if len(operations) == 0 {
		return "", fmt.Errorf("operation is empty")
	}

	opType := operations[0].Type
	for _, op := range operations {
		if op.Type != opType {
			return "", fmt.Errorf("multiple operation types found")
		}
	}

	return opType, nil
}

func createOperationDescription(txType string) *parser.Descriptions {
	return &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type: txType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.NegativeAmountSign,
				},
				AllowRepeats: true,
				CoinAction:   types.CoinSpent,
			},
			{
				Type: txType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.PositiveAmountSign,
				},
				AllowRepeats: true,
			},
		},
		ErrUnmatched: true,
	}
}

func (c *Backend) buildTransaction(
	ctx context.Context,
	opType string,
	matches []*parser.Match,
	payloadMetadata map[string]interface{},
) (*platformvm.Tx, []string, error) {
	switch opType {
	case mapper.OpImport:
		return c.buildImportTx(ctx, matches, payloadMetadata)
	case mapper.OpExport:
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

	outs, _, err := c.buildOutputs(matches[1].Operations)
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

	outs, _, err := c.buildOutputs(matches[1].Operations)
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

	outs, stakeOutputs, err := c.buildOutputs(matches[1].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedAddValidatorTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    sMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
			Memo:         sMetadata.Memo,
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

	outs, stakeOutputs, err := c.buildOutputs(matches[1].Operations)
	if err != nil {
		return nil, nil, fmt.Errorf("parse outputs failed: %w", err)
	}

	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedAddDelegatorTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    sMetadata.NetworkID,
			BlockchainID: blockchainID,
			Outs:         outs,
			Ins:          ins,
			Memo:         sMetadata.Memo,
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
		UTXOID, err := decodeUTXOID(op.CoinChange.CoinIdentifier.Identifier)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode UTXO ID: %w", err)
		}

		addr, err := address.ParseToID(op.Account.Address)
		if err != nil {
			return nil, nil, nil, err
		}
		addrSet := ids.NewShortSet(1)
		addrSet.Add(addr)

		opMetadata, err := parseOpMetadata(op.Metadata)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parse input operation Metadata failed: %w", err)
		}

		in := &avax.TransferableInput{
			UTXOID: *UTXOID,
			Asset:  avax.Asset{ID: c.assetID},
			In: &secp256k1fx.TransferInput{
				Amt: uint64(op.Amount.Currency.Decimals),
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
	err error,
) {
	for _, op := range operations {
		opMetadata, err := parseOpMetadata(op.Metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("parse output operation Metadata failed: %w", err)
		}

		var outputOwners secp256k1fx.OutputOwners
		if _, err := c.codec.Unmarshal(opMetadata.OutputOwners, &outputOwners); err != nil {
			return nil, nil, err
		}

		out := &avax.TransferableOutput{
			Asset: avax.Asset{ID: c.assetID},
			Out: &secp256k1fx.TransferOutput{
				Amt:          uint64(op.Amount.Currency.Decimals),
				OutputOwners: outputOwners,
			}}

		switch opMetadata.Type {
		case p.OpOutput:
			outs = append(outs, out)
		case p.OpStakeOutput:
			stakeOutputs = append(outs, out)

		default:
			return nil, nil, fmt.Errorf("invalid option type: %s", op.Type)
		}
	}

	avax.SortTransferableOutputs(outs, c.codec)
	avax.SortTransferableOutputs(stakeOutputs, c.codec)

	return outs, stakeOutputs, nil
}

func parseOpMetadata(metadata map[string]interface{}) (*p.OperationMetadata, error) {
	var operationMetadata p.OperationMetadata
	if err := mapper.UnmarshalJSONMap(metadata, &operationMetadata); err != nil {
		return nil, err
	}

	return &operationMetadata, nil
}

func decodeUTXOID(s string) (*avax.UTXOID, error) {
	split := strings.Split(s, ":")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid utxo ID format")
	}

	txID, err := ids.FromString(split[0])
	if err != nil {
		return nil, fmt.Errorf("invalid tx ID: %w", err)
	}

	outputIdx, err := strconv.ParseUint(split[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid output index: %w", err)
	}

	return &avax.UTXOID{
		TxID:        txID,
		OutputIndex: uint32(outputIdx),
	}, nil
}

func (c *Backend) ConstructionParse(ctx context.Context, req *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	return nil, service.ErrNotImplemented
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

	creds, err := getCredentialList(ins, req.Signatures)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	tx.Creds = creds

	signedBytes, err := platformvm.Codec.Marshal(platformvm.CodecVersion, tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: hex.EncodeToString(signedBytes),
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

// Based on tx inputs, we can determine the number of signatures
// required by each input and put correct number of signatures to
// construct the signed tx.
// See https://github.com/ava-labs/avalanchego/blob/master/vms/platformvm/tx.go#L100
// for more details.
func getCredentialList(ins []*avax.TransferableInput, signatures []*types.Signature) ([]verify.Verifiable, error) {
	creds := make([]verify.Verifiable, len(ins))
	sigOffset := 0
	for i, transferInput := range ins {
		input, ok := transferInput.In.(*secp256k1fx.TransferInput)
		if !ok {
			return nil, errors.New("invalid input")
		}
		cred := &secp256k1fx.Credential{}
		cred.Sigs = make([][crypto.SECP256K1RSigLen]byte, len(input.SigIndices))
		for j := 0; j < len(input.SigIndices); j++ {
			if sigOffset >= len(signatures) {
				return nil, errors.New("insufficient signatures")
			}

			if len(signatures[sigOffset].Bytes) != crypto.SECP256K1RSigLen {
				return nil, errors.New("invalid signature length")
			}
			copy(cred.Sigs[j][:], signatures[sigOffset].Bytes)
			sigOffset++
		}

		creds[i] = cred
	}

	if sigOffset != len(signatures) {
		return nil, errors.New("input signature length doesn't match credentials needed")
	}

	return creds, nil
}

func (c *Backend) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txHex := req.SignedTransaction
	txByte, err := formatting.Decode(formatting.Hex, txHex)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}
	txHash256 := hashing.ComputeHash256(txByte)
	pHash, err := formatting.EncodeWithChecksum(formatting.CB58, txHash256)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}
	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: pHash,
		},
	}, nil
}

func (c *Backend) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txHex := req.SignedTransaction
	txByte, err := formatting.Decode(formatting.Hex, txHex)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	txID, err := c.pClient.IssueTx(ctx, txByte)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txID.String(),
		},
	}, nil
}
