package pchain

import (
	"errors"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	pmapper "github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
)

type pTx struct {
	Tx           *platformvm.Tx
	Codec        codec.Manager
	CodecVersion uint16
}

func (p *pTx) Marshal() ([]byte, error) {
	return p.Codec.Marshal(p.CodecVersion, p.Tx)
}

func (p *pTx) Unmarshal(bytes []byte) error {
	tx := platformvm.Tx{}
	_, err := p.Codec.Unmarshal(bytes, &tx)
	if err != nil {
		return err
	}
	p.Tx = &tx
	return nil
}

func (p *pTx) SigningPayload() ([]byte, error) {
	unsignedAtomicBytes, err := p.Codec.Marshal(p.CodecVersion, &p.Tx.UnsignedTx)
	if err != nil {
		return nil, err
	}

	hash := hashing.ComputeHash256(unsignedAtomicBytes)
	return hash, nil
}

func (p *pTx) Hash() ([]byte, error) {
	bytes, err := p.Codec.Marshal(p.CodecVersion, &p.Tx)
	if err != nil {
		return nil, err
	}

	hash := hashing.ComputeHash256(bytes)
	return hash, nil
}

type pTxParser struct {
	hrp string
}

func (p pTxParser) ParseTx(tx common.AvaxTx, isConstruction bool) ([]*types.Operation, error) {
	pTx, ok := tx.(*pTx)
	if !ok {
		return nil, errors.New("invalid transaction")
	}
	transactions, err := pmapper.ParseTx(pTx.Tx.UnsignedTx, isConstruction)
	if err != nil {
		return nil, err
	}

	return transactions.Operations, nil
}

type pTxBuilder struct {
	avaxAssetID  ids.ID
	codec        codec.Manager
	codecVersion uint16
}

func (p pTxBuilder) BuildTx(operations []*types.Operation, metadata map[string]interface{}) (common.AvaxTx, []*types.AccountIdentifier, *types.Error) {
	matches, err := common.MatchOperations(operations)
	if err != nil {
		return nil, nil, service.WrapError(service.ErrInvalidInput, err)
	}

	opType := matches[0].Operations[0].Type

	tx, signers, err := pmapper.BuildTx(opType, matches, metadata, p.codec, p.avaxAssetID)
	if err != nil {
		return nil, nil, service.WrapError(service.ErrInvalidInput, err)
	}

	return &pTx{
		Tx:           tx,
		Codec:        p.codec,
		CodecVersion: p.codecVersion,
	}, signers, nil
}
