package p

type txOptions struct {
	SourceChain      string `json:"source_chain"`
	DestinationChain string `json:"destination_chain"`
}

type stakingOptions struct {
	NodeID          string   `json:"node_id"`
	Start           uint64   `json:"start"`
	End             uint64   `json:"end"`
	Wght            uint64   `json:"weight"`
	Shares          uint32   `json:"shares"`
	Memo            string   `json:"memo"`
	Locktime        uint64   `json:"locktime"`
	Threshold       uint32   `json:"threshold"`
	RewardAddresses []string `json:"reward_addresses"`
}

type txMetadata struct {
	SourceChainID      string `json:"source_chain_id"`
	DestinationChainID string `json:"destination_chain_id"`
	NetworkID          uint32 `json:"network_id"`
	BlockchainID       string `json:"blockchain_id"`
}

type stakingMetadata struct {
	NodeID          string   `json:"node_id"`
	Start           uint64   `json:"start"`
	End             uint64   `json:"end"`
	Wght            uint64   `json:"weight"`
	Shares          uint32   `json:"shares"`
	Memo            string   `json:"memo"`
	NetworkID       uint32   `json:"network_id"`
	BlockchainID    string   `json:"blockchain_id"`
	Locktime        uint64   `json:"locktime"`
	Threshold       uint32   `json:"threshold"`
	RewardAddresses []string `json:"reward_addresses"`
}

type sigIndicesMetadata struct {
	Type string `json:"type"`
}
