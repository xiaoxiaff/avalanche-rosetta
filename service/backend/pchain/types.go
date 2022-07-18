package pchain

import (
	"encoding/json"

	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

func (t *Transaction) MarshalJSON() ([]byte, error) {
	bytes, err := platformvm.Codec.Marshal(platformvm.CodecVersion, t.Tx)
	if err != nil {
		return nil, err
	}

	str, err := mapper.EncodeBytes(bytes)
	if err != nil {
		return nil, err
	}

	txWire := &transactionWire{
		Tx:        str,
		InputData: t.AccountIdentifierSigners,
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

	tx := &platformvm.Tx{}
	_, err = platformvm.Codec.Unmarshal(bytes, tx)
	if err != nil {
		return err
	}

	t.Tx = tx
	t.AccountIdentifierSigners = txWire.InputData

	return nil
}

type AccountIdentifierSigners struct {
	OperationIdentifier *types.OperationIdentifier `json:"operation_identifier"`
	AccountIdentifier   *types.AccountIdentifier   `json:"account_identifier"`
}

type Transaction struct {
	// The body of this transaction
	Tx                       *platformvm.Tx
	AccountIdentifierSigners []AccountIdentifierSigners
}

type transactionWire struct {
	Tx        string                     `json:"tx"`
	InputData []AccountIdentifierSigners `json:"input_data"`
}
