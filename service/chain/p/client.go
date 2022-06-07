package p

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

type Client struct {
	fac crypto.FactorySECP256K1R
}

func NewClient() *Client {
	return &Client{
		fac: crypto.FactorySECP256K1R{},
	}
}

func (c *Client) DeriveAddress(
	ctx context.Context,
	req *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, error) {
	pub, err := c.fac.ToPublicKey(req.PublicKey.Bytes)
	if err != nil {
		return nil, err
	}

	chainIDAlias, hrp, getErr := mapper.GetAliasAndHRP(req.NetworkIdentifier)
	if getErr != nil {
		return nil, getErr
	}

	addr, err := address.Format(chainIDAlias, hrp, pub.Address().Bytes())
	if err != nil {
		return nil, err
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: addr,
		},
	}, nil
}

func (c *Client) CombinePChainTx(
	ctx context.Context,
	req *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, error) {
	manager := codec.NewDefaultManager()
	tx := platformvm.Tx{}

	_, err := manager.Unmarshal([]byte(req.UnsignedTransaction), tx.UnsignedTx)
	if err != nil {
		return nil, err
	}

	tx.Creds = make([]verify.Verifiable, len(req.Signatures))

	ins, err := getTxInputs(tx.UnsignedTx)
	if err != nil {
		return nil, err
	}

	creds, err := getCredentialList(ins, req.Signatures)
	if err != nil {
		return nil, err
	}

	tx.Creds = creds

	signedBytes, err := platformvm.Codec.Marshal(platformvm.CodecVersion, tx)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal tx: %w", err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: hex.EncodeToString(signedBytes),
	}, nil
}

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
		return utx.Ins, nil
	case *platformvm.UnsignedExportTx:
		return utx.Ins, nil
	default:
		return nil, errors.New("unknown tx type")
	}
}

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
