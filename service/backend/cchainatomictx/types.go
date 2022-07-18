package cchainatomictx

import (
	"encoding/json"

	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

type Transaction struct {
	Tx                       *evm.Tx
	AccountIdentifierSigners []Signer
}

type Signer struct {
	OperationIdentifier *types.OperationIdentifier
	AccountIdentifier   *types.AccountIdentifier
}

type transactionWire struct {
	Tx                       string   `json:"tx"`
	AccountIdentifierSigners []Signer `json:"signers"`
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	bytes, err := evm.Codec.Marshal(0, t.Tx)
	if err != nil {
		return nil, err
	}

	str, err := mapper.EncodeBytes(bytes)
	if err != nil {
		return nil, err
	}

	txWire := &transactionWire{
		Tx:                       str,
		AccountIdentifierSigners: t.AccountIdentifierSigners,
	}
	return json.Marshal(txWire)
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	txWire := &transactionWire{}
	err := json.Unmarshal(data, txWire)
	if err != nil {
		return err
	}

	bytes, err := mapper.DecodeToBytes(txWire.Tx)
	if err != nil {
		return err
	}

	tx := evm.Tx{}
	_, err = evm.Codec.Unmarshal(bytes, &tx)
	if err != nil {
		return err
	}

	t.Tx = &tx
	t.AccountIdentifierSigners = txWire.AccountIdentifierSigners

	return nil
}
