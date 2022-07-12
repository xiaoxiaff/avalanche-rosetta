package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func DeriveBech32Address(fac *crypto.FactorySECP256K1R, chainIdAlias string, req *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	pub, err := fac.ToPublicKey(req.PublicKey.Bytes)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	hrp, getErr := mapper.GetHRP(req.NetworkIdentifier)
	if getErr != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	addr, err := address.Format(chainIdAlias, hrp, pub.Address().Bytes())
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: addr,
		},
	}, nil
}

func HashTx(req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txByte, err := DecodeToBytes(req.SignedTransaction)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	txHash256 := hashing.ComputeHash256(txByte)
	atomicCHash, err := formatting.EncodeWithChecksum(formatting.CB58, txHash256)

	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}
	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: atomicCHash,
		},
	}, nil
}

type TransactionIssuer interface {
	IssueTx(ctx context.Context, txByte []byte) (ids.ID, error)
}

func SubmitTx(issuer TransactionIssuer, ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	txByte, err := DecodeToBytes(req.SignedTransaction)
	if err != nil {
		return nil, service.WrapError(service.ErrInvalidInput, err)
	}

	txID, err := issuer.IssueTx(ctx, txByte)
	if err != nil {
		return nil, service.WrapError(service.ErrClientError, err)
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txID.String(),
		},
	}, nil
}

func ParseOpType(operations []*types.Operation) (string, error) {
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

func MatchOperations(operations []*types.Operation) ([]*parser.Match, error) {
	if len(operations) == 0 {
		return nil, errors.New("no operations were passed to match")
	}
	opType := operations[0].Type

	var coinAction types.CoinAction
	var allowRepeatOutputs bool

	switch opType {
	case mapper.OpExport:
		coinAction = ""
		allowRepeatOutputs = false
	case mapper.OpImport:
		coinAction = types.CoinSpent
		allowRepeatOutputs = false
	default:
		coinAction = types.CoinSpent
		allowRepeatOutputs = true
	}

	descriptions := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type: opType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.NegativeAmountSign,
				},
				AllowRepeats: true,
				CoinAction:   coinAction,
			},
			{
				Type: opType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.PositiveAmountSign,
				},
				AllowRepeats: allowRepeatOutputs,
			},
		},
		ErrUnmatched: true,
	}

	return parser.MatchOperations(descriptions, operations)
}

func BuildPayloadsResponse(unsignedBytes, signingHash []byte, signers []*types.AccountIdentifier) (*types.ConstructionPayloadsResponse, *types.Error) {
	payloads := make([]*types.SigningPayload, len(signers))
	for i, signer := range signers {
		payloads[i] = &types.SigningPayload{
			AccountIdentifier: signer,
			Bytes:             signingHash,
			SignatureType:     types.EcdsaRecovery,
		}
	}

	txHex, err := EncodeBytes(unsignedBytes)
	if err != nil {
		return nil, service.WrapError(service.ErrInternalError, err)
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: txHex,
		Payloads:            payloads,
	}, nil
}

// Based on tx inputs, we can determine the number of signatures
// required by each input and put correct number of signatures to
// construct the signed tx.
// See https://github.com/ava-labs/avalanchego/blob/master/vms/platformvm/tx.go#L100
// for more details.
func BuildCredentialList(ins []*avax.TransferableInput, signatures []*types.Signature) ([]verify.Verifiable, error) {
	creds := make([]verify.Verifiable, len(ins))
	sigOffset := 0
	for i, transferInput := range ins {
		input, ok := transferInput.In.(*secp256k1fx.TransferInput)
		if !ok {
			return nil, errors.New("invalid input")
		}

		cred, err := buildCredential(len(input.SigIndices), &sigOffset, signatures)
		if err != nil {
			return nil, err
		}

		creds[i] = cred
	}

	if sigOffset != len(signatures) {
		return nil, errors.New("input signature length doesn't match credentials needed")
	}

	return creds, nil
}

func BuildSingletonCredentialList(signatures []*types.Signature) ([]verify.Verifiable, error) {
	offset := 0
	cred, err := buildCredential(1, &offset, signatures)
	if err != nil {
		return nil, err
	}

	return []verify.Verifiable{cred}, nil
}

func buildCredential(numSigs int, sigOffset *int, signatures []*types.Signature) (*secp256k1fx.Credential, error) {
	cred := &secp256k1fx.Credential{}
	cred.Sigs = make([][crypto.SECP256K1RSigLen]byte, numSigs)
	for j := 0; j < numSigs; j++ {
		if *sigOffset >= len(signatures) {
			return nil, errors.New("insufficient signatures")
		}

		if len(signatures[*sigOffset].Bytes) != crypto.SECP256K1RSigLen {
			return nil, errors.New("invalid signature length")
		}
		copy(cred.Sigs[j][:], signatures[*sigOffset].Bytes)
		*sigOffset++
	}
	return cred, nil
}

func EncodeBytes(bytes []byte) (string, error) {
	return formatting.EncodeWithChecksum(formatting.Hex, bytes)
}

func DecodeToBytes(binaryData string) ([]byte, error) {
	return formatting.Decode(formatting.Hex, binaryData)
}

func InitializeTx(version uint16, c codec.Manager, tx platformvm.Tx) error {
	errs := wrappers.Errs{}

	unsignedBytes, err := c.Marshal(version, &tx.UnsignedTx)
	errs.Add(err)

	signedBytes, err := c.Marshal(version, &tx)
	errs.Add(err)

	tx.Initialize(unsignedBytes, signedBytes)

	return errs.Err
}
