package pchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	pmapper "github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
)

var (
	errUnknownTxType = errors.New("unknown tx type")
	errUndecodableTx = errors.New("undecodable transaction")
	errNoTxGiven     = errors.New("no transaction was given")
)

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
	matches, err := common.MatchOperations(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	reqMetadata := req.Metadata
	reqMetadata[pmapper.MetadataOpType] = matches[0].Operations[0].Type

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
		Metadata:     metadataMap,
		SuggestedFee: nil, // TODO: return tx fee based on type
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
	builder := pTxBuilder{
		avaxAssetID:  b.avaxAssetID,
		codec:        b.codec,
		codecVersion: b.codecVersion,
	}
	return common.BuildPayloads(builder, req)
}

func (b *Backend) ConstructionParse(
	ctx context.Context,
	req *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	rosettaTx, err := b.parsePayloadTxFromString(req.Transaction)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	hrp, err := mapper.GetHRP(req.NetworkIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "incorrect network identifier")
	}
	txParser := pTxParser{hrp: hrp}

	return common.Parse(txParser, rosettaTx, req.Signed)
}

func (b *Backend) ConstructionCombine(
	ctx context.Context,
	req *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	rosettaTx, err := b.parsePayloadTxFromString(req.UnsignedTransaction)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return common.Combine(b, rosettaTx, req.Signatures)
}

func (b *Backend) CombineTx(tx common.AvaxTx, signatures []*types.Signature) (common.AvaxTx, *types.Error) {
	pTx, ok := tx.(*pTx)
	if !ok {
		return nil, service.WrapError(service.ErrInvalidInput, "invalid transaction")
	}

	ins, err := getTxInputs(pTx.Tx.UnsignedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	creds, err := common.BuildCredentialList(ins, signatures)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	unsignedBytes, err := pTx.Marshal()
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	pTx.Tx.Creds = creds

	signedBytes, err := pTx.Marshal()
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	pTx.Tx.Initialize(unsignedBytes, signedBytes)

	return pTx, nil
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
		return nil, errUnknownTxType
	}
}

func (b *Backend) ConstructionHash(
	ctx context.Context,
	req *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	rosettaTx, err := b.parsePayloadTxFromString(req.SignedTransaction)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return common.HashTx(rosettaTx)
}

func (b *Backend) ConstructionSubmit(
	ctx context.Context,
	req *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	rosettaTx, err := b.parsePayloadTxFromString(req.SignedTransaction)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return common.SubmitTx(ctx, b, rosettaTx)
}

// Defining IssueTx here without rpc.Options... to be able to use it with common.SubmitTx
func (b *Backend) IssueTx(ctx context.Context, txByte []byte) (ids.ID, error) {
	return b.pClient.IssueTx(ctx, txByte)
}

func (b *Backend) parsePayloadTxFromString(transaction string) (*common.RosettaTx, error) {
	// Unmarshal input transaction
	payloadsTx := &common.RosettaTx{
		Tx: &pTx{
			Codec:        b.codec,
			CodecVersion: b.codecVersion,
		},
	}

	err := json.Unmarshal([]byte(transaction), payloadsTx)
	if err != nil {
		return nil, errUndecodableTx
	}

	if payloadsTx.Tx == nil {
		return nil, errNoTxGiven
	}

	return payloadsTx, nil
}
