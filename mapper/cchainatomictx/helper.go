package cchainatomictx

import (
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

func IsCChainAtomicRequest(req interface{}) bool {
	switch r := req.(type) {
	case *types.AccountBalanceRequest:
		return isCChainBech32(r.AccountIdentifier)
	case *types.AccountCoinsRequest:
		return isCChainBech32(r.AccountIdentifier)
	case *types.ConstructionDeriveRequest:
		return r.Metadata[mapper.MetaAddressFormat] == mapper.AddressFormatBech32
	case *types.ConstructionMetadataRequest:
		return r.Options[MetadataAtomicTxGas] != nil
	case *types.ConstructionPreprocessRequest:
		return isAtomicOpType(r.Operations[0].Type)
	case *types.ConstructionPayloadsRequest:
		return isAtomicOpType(r.Operations[0].Type)
	case *types.ConstructionParseRequest:
		return isEvmAtomicTx(r.Transaction)
	case *types.ConstructionCombineRequest:
		return isEvmAtomicTx(r.UnsignedTransaction)
	case *types.ConstructionHashRequest:
		return isEvmAtomicTx(r.SignedTransaction)
	case *types.ConstructionSubmitRequest:
		return isEvmAtomicTx(r.SignedTransaction)
	}

	return false
}

func isCChainBech32(accountIdentifier *types.AccountIdentifier) bool {
	if chainID, _, _, err := address.Parse(accountIdentifier.Address); err == nil {
		return chainID == mapper.CChainNetworkIdentifier
	}
	return false
}

func isAtomicOpType(t string) bool {
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

func isEvmAtomicTx(transaction string) bool {
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
