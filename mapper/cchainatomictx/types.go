package cchainatomictx

import (
	"github.com/ava-labs/avalanchego/ids"
	"math/big"
)

const (
	MetaAtomicTxGas = "atomic_tx_gas"
)

type Metadata struct {
	NetworkID          uint32  `json:"network_id,omitempty"`
	CChainID           ids.ID  `json:"c_chain_id,omitempty"`
	SourceChainID      *ids.ID `json:"source_chain_id,omitempty"`
	DestinationChainId *ids.ID `json:"destination_chain_id,omitempty"`
	Nonce              uint64  `json:"nonce"`
}

type Options struct {
	AtomicTxGas      *big.Int `json:"atomic_tx_gas"`
	From             string   `json:"from,omitempty"`
	SourceChain      string   `json:"source_chain,omitempty"`
	DestinationChain string   `json:"destination_chain,omitempty"`
	Nonce            *big.Int `json:"nonce,omitempty"`
}
