package cchainatomictx

import (
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

func IsCChainBech32Address(accountIdentifier *types.AccountIdentifier) bool {
	if chainID, _, _, err := address.Parse(accountIdentifier.Address); err == nil {
		return chainID == mapper.CChainNetworkIdentifier
	}
	return false
}

func IsAtomicOpType(t string) bool {
	atomicTypes := []string{
		mapper.OpExport,
		mapper.OpImport,
	}

	for _, atomicType := range atomicTypes {
		if atomicType == t {
			return true
		}
	}

	return false
}

func IsEvmAtomicTx(transaction string) bool {
	txBytes, err := formatting.Decode(formatting.Hex, transaction)
	if err != nil {
		return false
	}

	var tx evm.Tx
	_, err = evm.Codec.Unmarshal(txBytes, &tx)
	if err != nil {
		return false
	}

	return true
}
