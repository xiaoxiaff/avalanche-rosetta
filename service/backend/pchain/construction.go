package pchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	pmapper "github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	"github.com/ava-labs/avalanche-rosetta/service"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
)

var codecVersion uint16 = platformvm.CodecVersion

func (b *Backend) ConstructionDerive(
	ctx context.Context,
	req *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	return common.DeriveBech32Address(b.fac, mapper.PChainNetworkIdentifier, req)
}
func (b *Backend) ConstructionPreprocess(
	ctx context.Context,
	req *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	opType, err := common.ParseOpType(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	reqMetadata := req.Metadata
	reqMetadata[pmapper.MetadataOpType] = opType

	return &types.ConstructionPreprocessResponse{
		Options: reqMetadata,
	}, nil
}
func (b *Backend) ConstructionMetadata(
	ctx context.Context,
	req *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	opMetadata, err := pmapper.ParseOpMetadata(req.Options)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	networkID, err := b.pClient.GetNetworkID(context.Background())
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	// Getting Chain ID from Info APIs
	pChainID, err := b.pClient.GetBlockchainID(ctx, mapper.PChainNetworkIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	var metadataMap map[string]interface{}
	switch opMetadata.Type {
	case pmapper.OpImportAvax, pmapper.OpExportAvax:
		metadataMap, err = b.buildImportExportMetadata(ctx, opMetadata.Type, req.Options, networkID, pChainID)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}
	case pmapper.OpAddValidator, pmapper.OpAddDelegator:
		metadataMap, err = b.buildStakingMetadata(req.Options, networkID, pChainID)
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

func (b *Backend) buildImportExportMetadata(
	ctx context.Context,
	txType string,
	options map[string]interface{},
	networkID uint32,
	pChainID ids.ID,
) (map[string]interface{}, error) {
	var preprocessOptions pmapper.ImportExportOptions
	if err := mapper.UnmarshalJSONMap(options, &preprocessOptions); err != nil {
		return nil, err
	}

	txMetadata := &pmapper.ImportExportMetadata{
		NetworkID:    networkID,
		BlockchainID: pChainID.String(),
	}

	switch txType {
	case pmapper.OpImportAvax:
		sourceChainID, err := b.pClient.GetBlockchainID(ctx, preprocessOptions.SourceChain)
		if err != nil {
			return nil, err
		}
		txMetadata.SourceChainID = sourceChainID.String()
	case pmapper.OpExportAvax:
		destinationChainID, err := b.pClient.GetBlockchainID(ctx, preprocessOptions.DestinationChain)
		if err != nil {
			return nil, err
		}
		txMetadata.DestinationChainID = destinationChainID.String()
	default:
		return nil, fmt.Errorf("invalid tx type for building tx metadata: %s", txType)
	}
	return mapper.MarshalJSONMap(txMetadata)
}

func (b *Backend) buildStakingMetadata(
	options map[string]interface{},
	networkID uint32,
	pChainID ids.ID,
) (map[string]interface{}, error) {
	var preprocessOptions pmapper.StakingOptions
	if err := mapper.UnmarshalJSONMap(options, &preprocessOptions); err != nil {
		return nil, err
	}

	stakingMetadata := &pmapper.StakingMetadata{
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

func (b *Backend) ConstructionPayloads(
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

	tx, signers, err := pmapper.BuildTransaction(opType, matches, req.Metadata, b.codec, b.assetID)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	unsignedBytes, err := b.codec.Marshal(codecVersion, &tx.UnsignedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, fmt.Errorf("couldn't marshal UnsignedTx: %w", err))
	}

	input := make([]AccountIdentifierSigners, len(req.Operations))
	for i, v := range req.Operations {
		input[i] = AccountIdentifierSigners{
			OperationIdentifier: v.OperationIdentifier,
			AccountIdentifier:   v.Account,
		}
	}

	wrappedTx := &Transaction{Tx: tx, AccountIdentifierSigners: input}
	hash := hashing.ComputeHash256(unsignedBytes)
	payloads := make([]*types.SigningPayload, len(signers))

	for i, signer := range signers {
		payloads[i] = &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{Address: signer},
			Bytes:             hash,
			SignatureType:     types.EcdsaRecovery,
		}
	}

	txBytes, err := json.Marshal(wrappedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: string(txBytes),
		Payloads:            payloads,
	}, nil
}

func (b *Backend) ConstructionParse(ctx context.Context, req *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	tx := Transaction{}

	err := json.Unmarshal([]byte(req.Transaction), &tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	transactions, err := pmapper.Transaction(tx.Tx.UnsignedTx, true)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	inputDataMap := make(map[int64]*types.AccountIdentifier)
	for _, v := range tx.AccountIdentifierSigners {
		inputDataMap[v.OperationIdentifier.Index] = v.AccountIdentifier
	}

	accountIDSigners := make([]*types.AccountIdentifier, 0, len(tx.AccountIdentifierSigners))
	for _, v := range transactions.Operations {
		v.Account = inputDataMap[v.OperationIdentifier.Index]
		if req.Signed {
			if v.Metadata["type"] == pmapper.OpTypeImport || v.Metadata["type"] == pmapper.OpTypeInput {
				accountIDSigners = append(accountIDSigners, inputDataMap[v.OperationIdentifier.Index])
			}
		}
	}

	resp := &types.ConstructionParseResponse{
		Operations:               transactions.Operations,
		AccountIdentifierSigners: accountIDSigners,
		Metadata:                 nil,
	}

	return resp, nil
}

func (b *Backend) ConstructionCombine(ctx context.Context, req *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	tx := Transaction{}

	err := json.Unmarshal([]byte(req.UnsignedTransaction), &tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	ins, err := getTxInputs(tx.Tx.UnsignedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	creds, err := common.BuildCredentialList(ins, req.Signatures)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	unsignedBytes, err := b.codec.Marshal(b.codecVersion, tx.Tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	tx.Tx.Creds = creds
	signedBytes, err := b.codec.Marshal(b.codecVersion, tx.Tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}
	tx.Tx.Initialize(unsignedBytes, signedBytes)

	wrappedTx := &Transaction{Tx: tx.Tx, AccountIdentifierSigners: tx.AccountIdentifierSigners}

	signedTx, err := json.Marshal(wrappedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: string(signedTx),
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

func (b *Backend) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return common.HashTx(req)
}

func (b *Backend) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return common.SubmitTx(b, ctx, req)
}

// Defining IssueTx here without rpc.Options... to be able to use it with common.SubmitTx
func (b *Backend) IssueTx(ctx context.Context, txByte []byte) (ids.ID, error) {
	return b.pClient.IssueTx(ctx, txByte)
}
