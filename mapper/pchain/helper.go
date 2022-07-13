package pchain

import (
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func IsPChainRequest(req interface{}) bool {
	switch r := req.(type) {
	case *types.AccountBalanceRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.AccountCoinsRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.BlockRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.BlockTransactionRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionDeriveRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionMetadataRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionPreprocessRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionPayloadsRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionParseRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionCombineRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionHashRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.ConstructionSubmitRequest:
		return isPChain(r.NetworkIdentifier)
	case *types.NetworkRequest:
		return isPChain(r.NetworkIdentifier)
	}

	return false
}

// IsPChain checks network identifier to make sure sub-network identifier set to "P"
func isPChain(networkIdentifier *types.NetworkIdentifier) bool {
	if networkIdentifier != nil &&
		networkIdentifier.SubNetworkIdentifier != nil &&
		networkIdentifier.SubNetworkIdentifier.Network == mapper.PChainNetworkIdentifier {
		return true
	}

	return false
}
