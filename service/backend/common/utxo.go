package common

import (
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"strconv"
	"strings"
)

func DecodeUTXOID(s string) (*avax.UTXOID, error) {
	split := strings.Split(s, ":")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid utxo ID format")
	}

	txID, err := ids.FromString(split[0])
	if err != nil {
		return nil, fmt.Errorf("invalid tx ID: %w", err)
	}

	outputIdx, err := strconv.ParseUint(split[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid output index: %w", err)
	}

	return &avax.UTXOID{
		TxID:        txID,
		OutputIndex: uint32(outputIdx),
	}, nil
}
