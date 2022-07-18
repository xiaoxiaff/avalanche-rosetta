package cchainatomictx

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/coinbase/rosetta-sdk-go/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	cmapper "github.com/ava-labs/avalanche-rosetta/mapper/cchainatomictx"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
)

func (b *Backend) ConstructionDerive(
	ctx context.Context,
	req *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	return common.DeriveBech32Address(b.fac, mapper.CChainNetworkIdentifier, req)
}

func (b *Backend) ConstructionPreprocess(ctx context.Context, req *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	matches, err := common.MatchOperations(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	firstIn, _ := matches[0].First()
	firstOut, _ := matches[1].First()

	if firstIn == nil || firstOut == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "both input and output operations must be specified")
	}

	var preprocessOptions cmapper.Options

	switch firstIn.Type {
	case mapper.OpImport:
		chain, _, _, err := address.Parse(firstIn.Account.Address)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}

		preprocessOptions = cmapper.Options{
			SourceChain: chain,
		}
	case mapper.OpExport:
		chain, _, _, err := address.Parse(firstOut.Account.Address)
		if err != nil {
			return nil, service.WrapError(service.ErrInternalError, err)
		}

		preprocessOptions = cmapper.Options{
			From:             firstIn.Account.Address,
			DestinationChain: chain,
		}

		if v, ok := req.Metadata[cmapper.MetadataNonce]; ok {
			stringObj, ok := v.(string)
			if !ok {
				return nil, service.WrapError(service.ErrInvalidInput, fmt.Errorf("%s is not a valid nonce string", v))
			}
			bigObj, ok := new(big.Int).SetString(stringObj, 10)
			if !ok {
				return nil, service.WrapError(service.ErrInvalidInput, fmt.Errorf("%s is not a valid nonce", v))
			}
			preprocessOptions.Nonce = bigObj
		}

	}

	tx, _, err := cmapper.BuildTx(firstIn.Type, matches, cmapper.Metadata{
		SourceChainID:      &ids.Empty,
		DestinationChainId: &ids.Empty,
	}, b.codec, b.avaxAssetId)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	err = tx.Sign(b.codec, [][]*crypto.PrivateKeySECP256K1R{})
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	gasUsed, err := tx.GasUsed(true)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	preprocessOptions.AtomicTxGas = big.NewInt(int64(gasUsed))

	optionsMap, err := mapper.MarshalJSONMap(preprocessOptions)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionPreprocessResponse{
		Options: optionsMap,
	}, nil
}

func (b *Backend) ConstructionMetadata(ctx context.Context, req *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	var input cmapper.Options
	err := mapper.UnmarshalJSONMap(req.Options, &input)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	networkId, err := b.cClient.GetNetworkID(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	cChainId, err := b.cClient.GetBlockchainID(ctx, mapper.CChainNetworkIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	metadata := cmapper.Metadata{
		NetworkID: networkId,
		CChainID:  cChainId,
	}

	if input.SourceChain != "" {
		id, err := b.cClient.GetBlockchainID(ctx, input.SourceChain)
		if err != nil {
			return nil, service.WrapError(service.ErrClientError, err)
		}
		metadata.SourceChainID = &id
	}

	if input.DestinationChain != "" {
		id, err := b.cClient.GetBlockchainID(ctx, input.DestinationChain)
		if err != nil {
			return nil, service.WrapError(service.ErrClientError, err)
		}
		metadata.DestinationChainId = &id

	}

	if input.From != "" {
		var nonce uint64
		if input.Nonce == nil {
			nonce, err = b.cClient.NonceAt(ctx, ethcommon.HexToAddress(input.From), nil)
			if err != nil {
				return nil, service.WrapError(service.ErrClientError, err)
			}
		} else {
			nonce = input.Nonce.Uint64()
		}
		metadata.Nonce = nonce
	}

	baseFee, err := b.cClient.EstimateBaseFee(ctx)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	suggestedFeeEth := new(big.Int).Mul(input.AtomicTxGas, baseFee)
	suggestedFee := new(big.Int).Div(suggestedFeeEth, mapper.X2crate)

	metadataMap, err := mapper.MarshalJSONMap(metadata)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionMetadataResponse{
		Metadata: metadataMap,
		SuggestedFee: []*types.Amount{
			{
				Value:    suggestedFee.String(),
				Currency: mapper.AvaxCurrency,
			},
		},
	}, nil
}

func (b *Backend) ConstructionPayloads(ctx context.Context, req *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	matches, err := common.MatchOperations(req.Operations)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	metadata := cmapper.Metadata{}
	err = mapper.UnmarshalJSONMap(req.Metadata, &metadata)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	tx, signers, err := cmapper.BuildTx(req.Operations[0].Type, matches, metadata, b.codec, b.avaxAssetId)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	unsignedAtomicBytes, err := b.codec.Marshal(b.codecVersion, &tx.UnsignedAtomicTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	hash := hashing.ComputeHash256(unsignedAtomicBytes)

	payloads := make([]*types.SigningPayload, len(signers))
	for i, signer := range signers {
		payloads[i] = &types.SigningPayload{
			AccountIdentifier: signer,
			Bytes:             hash,
			SignatureType:     types.EcdsaRecovery,
		}
	}

	accountIdentifierSigners := make([]Signer, 0, len(req.Operations))
	for _, o := range req.Operations {
		accountIdentifierSigners = append(accountIdentifierSigners, Signer{
			OperationIdentifier: o.OperationIdentifier,
			AccountIdentifier:   o.Account,
		})
	}

	txJson, err := json.Marshal(&Transaction{
		Tx:                       tx,
		AccountIdentifierSigners: accountIdentifierSigners,
	})
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: string(txJson),
		Payloads:            payloads,
	}, nil
}

func (b *Backend) ConstructionParse(ctx context.Context, req *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	// Unmarshal input transaction
	wrappedTx := Transaction{}

	err := json.Unmarshal([]byte(req.Transaction), &wrappedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "undecodable transaction")
	}

	if wrappedTx.Tx == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "no transaction was given")
	}
	tx := *wrappedTx.Tx

	// Convert input tx into operations
	hrp, err := mapper.GetHRP(req.NetworkIdentifier)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "incorrect network identifier")
	}

	operations, err := cmapper.ParseTx(tx, hrp)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "incorrect transaction input")
	}

	// Generate AccountIdentifierSigners if request is signed
	var signers []*types.AccountIdentifier
	if req.Signed {
		operationToAccountMap := make(map[int64]*types.AccountIdentifier)
		for _, data := range wrappedTx.AccountIdentifierSigners {
			operationToAccountMap[data.OperationIdentifier.Index] = data.AccountIdentifier
		}

		for _, op := range operations {
			signer := operationToAccountMap[op.OperationIdentifier.Index]
			if signer == nil {
				return nil, service.WrapError(service.ErrInvalidInput, "not all operations have signers")
			}
			signers = append(signers, signer)
		}
	}

	return &types.ConstructionParseResponse{
		Operations:               operations,
		AccountIdentifierSigners: signers,
	}, nil
}

func (b *Backend) ConstructionCombine(ctx context.Context, req *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	var wrappedTx Transaction
	err := json.Unmarshal([]byte(req.UnsignedTransaction), &wrappedTx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	if wrappedTx.Tx == nil {
		return nil, service.WrapError(service.ErrInvalidInput, "no transaction was given")
	}
	tx := wrappedTx.Tx

	var creds []verify.Verifiable
	switch uat := tx.UnsignedAtomicTx.(type) {
	case *evm.UnsignedImportTx:
		creds, err = common.BuildCredentialList(uat.ImportedInputs, req.Signatures)
	case *evm.UnsignedExportTx:
		creds, err = common.BuildSingletonCredentialList(req.Signatures)
	}
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable attach signatures to transaction")
	}

	unsignedBytes, err := b.codec.Marshal(b.codecVersion, tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, "unable to encode unsigned transaction")
	}

	tx.Creds = creds

	signedBytes, err := b.codec.Marshal(b.codecVersion, tx)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, "unable to marshal signed transaction")
	}

	tx.Initialize(unsignedBytes, signedBytes)

	signedTransaction, err := json.Marshal(&Transaction{
		Tx:                       tx,
		AccountIdentifierSigners: wrappedTx.AccountIdentifierSigners,
	})
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, "unable to encode signed transaction")
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: string(signedTransaction),
	}, nil
}

func (b *Backend) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return common.HashTx(req)
}

func (b *Backend) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return common.SubmitTx(b.cClient, ctx, req)
}
