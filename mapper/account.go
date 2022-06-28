package mapper

import (
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/coinbase/rosetta-sdk-go/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func Account(address *ethcommon.Address) *types.AccountIdentifier {
	if address == nil {
		return nil
	}
	return &types.AccountIdentifier{
		Address: address.String(),
	}
}

func IsBech32(accountIdentifier *types.AccountIdentifier) bool {
	if _, _, _, err := address.Parse(accountIdentifier.Address); err == nil {
		return true
	}
	return false
}

func IsCChainBech32(accountIdentifier *types.AccountIdentifier) bool {
	if chainID, _, _, err := address.Parse(accountIdentifier.Address); err == nil {
		return chainID == CChainIDAlias
	}
	return false
}
