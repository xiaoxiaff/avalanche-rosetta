package pchain

const (
	OpImportAvax      = "IMPORT_AVAX"
	OpExportAvax      = "EXPORT_AVAX"
	OpAddValidator    = "ADD_VALIDATOR"
	OpAddDelegator    = "ADD_DELEGATOR"
	OpRewardValidator = "REWARD_VALIDATOR"

	OpTypeImport      = "IMPORT"
	OpTypeExport      = "EXPORT"
	OpTypeInput       = "INPUT"
	OpTypeOutput      = "OUTPUT"
	OpTypeStakeOutput = "STAKE"

	MetadataOpType      = "type"
	MetadataStakingTxId = "staking_tx"
)

var (
	OperationTypes = []string{
		OpImportAvax,
		OpExportAvax,
		OpAddValidator,
		OpAddDelegator,
		OpRewardValidator,
	}
	CallMethods = []string{}
)

type OperationMetadata struct {
	Type       string   `json:"type"`
	SigIndices []uint32 `json:"sig_indices"`
	Locktime   uint64   `json:"locktime"`
	Threshold  uint32   `json:"threshold"`
}

type ImportExportOptions struct {
	SourceChain      string `json:"source_chain"`
	DestinationChain string `json:"destination_chain"`
}

type StakingOptions struct {
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

type ImportExportMetadata struct {
	SourceChainID      string `json:"source_chain_id"`
	DestinationChainID string `json:"destination_chain_id"`
	NetworkID          uint32 `json:"network_id"`
	BlockchainID       string `json:"blockchain_id"`
}

type StakingMetadata struct {
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
