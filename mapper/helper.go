package mapper

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// EqualFoldContains checks if the array contains the string regardless of casing
func EqualFoldContains(arr []string, str string) bool {
	for _, a := range arr {
		if strings.EqualFold(a, str) {
			return true
		}
	}
	return false
}

// IsPChain checks network identifier to make sure sub-network identifier set to "P"
func IsPChain(networkIdentifier *types.NetworkIdentifier) bool {
	if networkIdentifier != nil &&
		networkIdentifier.SubNetworkIdentifier != nil &&
		networkIdentifier.SubNetworkIdentifier.Network == PChainNetworkIdentifier {
		return true
	}

	return false
}

// IsCChain checks network identifier to make sure sub-network is not specified or identifier set to "C"
func IsCChain(networkIdentifier *types.NetworkIdentifier) bool {
	if networkIdentifier != nil &&
		(networkIdentifier.SubNetworkIdentifier == nil ||
			networkIdentifier.SubNetworkIdentifier.Network == CChainNetworkIdentifier) {
		return true
	}

	return false
}

// GetHRP fetches hrp for address formatting.
func GetHRP(networkIdentifier *types.NetworkIdentifier) (string, error) {
	var hrp string
	switch networkIdentifier.Network {
	case FujiNetwork:
		hrp = constants.GetHRP(constants.FujiID)
	case MainnetNetwork:
		hrp = constants.GetHRP(constants.MainnetID)
	default:
		return "", errors.New("can't recognize network")
	}

	return hrp, nil
}

// UnmarshalJSONMap converts map[string]interface{} into a interface{}.
func UnmarshalJSONMap(m map[string]interface{}, i interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, i)
}

// MarshalJSONMap converts an interface into a map[string]interface{}.
func MarshalJSONMap(i interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	return m, nil
}
