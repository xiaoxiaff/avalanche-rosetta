package pchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

const (
	OpImportAvax         = "IMPORT_AVAX"
	OpExportAvax         = "EXPORT_AVAX"
	OpAddValidator       = "ADD_VALIDATOR"
	OpAddDelegator       = "ADD_DELEGATOR"
	OpRewardValidator    = "REWARD_VALIDATOR"
	OpCreateChain        = "CREATE_CHAIN"
	OpCreateSubnet       = "CREATE_SUBNET"
	OpAddSubnetValidator = "ADD_SUBNET_VALIDATOR"
	OpAdvanceTime        = "ADVANCE_TIME"

	OpTypeImport      = "IMPORT"
	OpTypeExport      = "EXPORT"
	OpTypeInput       = "INPUT"
	OpTypeOutput      = "OUTPUT"
	OpTypeStakeOutput = "STAKE"
	OpTypeReward      = "REWARD"
	OpTypeCreateChain = "CREATE_CHAIN"

	MetadataTxType      = "tx_type"
	MetadataSkippedOuts = "skipped_outs"
	MetadataOpType      = "type"
	MetadataStakingTxID = "staking_tx_id"
	MetadataSubnetID    = "subnet_id"
	MetadataChainName   = "chain_name"
	MetadataVMID        = "vmid"
	MetadataMemo        = "memo"
	MetadataMessage     = "message"

	SubAccountTypeSharedMemory       = "shared_memory"
	SubAccountTypeUnlocked           = "unlocked"
	SubaccounttypelockedStakeable    = "locked_stakeable"
	SubaccounttypelockedNotStakeable = "locked_not_stakeable"
	SubAccountTypeStaked             = "staked"
)

var (
	OperationTypes = []string{
		OpImportAvax,
		OpExportAvax,
		OpAddValidator,
		OpAddDelegator,
		OpRewardValidator,
		OpCreateChain,
		OpCreateSubnet,
		OpAddSubnetValidator,
	}
	CallMethods = []string{}
)

type OperationMetadata struct {
	Type        string   `json:"type"`
	SigIndices  []uint32 `json:"sig_indices,omitempty"`
	Locktime    uint64   `json:"locktime"`
	Threshold   uint32   `json:"threshold,omitempty"`
	StakingTxID string   `json:"staking_tx_id,omitempty"`
}

type ImportExportOptions struct {
	SourceChain      string `json:"source_chain"`
	DestinationChain string `json:"destination_chain"`
}

type StakingOptions struct {
	NodeID          string   `json:"node_id"`
	Start           uint64   `json:"start"`
	End             uint64   `json:"end"`
	Shares          uint32   `json:"shares"`
	Memo            string   `json:"memo"`
	Locktime        uint64   `json:"locktime"`
	Threshold       uint32   `json:"threshold"`
	RewardAddresses []string `json:"reward_addresses"`
}

type Metadata struct {
	NetworkID    uint32 `json:"network_id"`
	BlockchainID ids.ID `json:"blockchain_id"`
	*ImportMetadata
	*ExportMetadata
	*StakingMetadata
}

type ImportMetadata struct {
	SourceChainID ids.ID `json:"source_chain_id"`
}

type ExportMetadata struct {
	DestinationChain   string `json:"destination_chain"`
	DestinationChainID ids.ID `json:"destination_chain_id"`
}

type StakingMetadata struct {
	NodeID          string   `json:"node_id"`
	RewardAddresses []string `json:"reward_addresses"`
	Start           uint64   `json:"start"`
	End             uint64   `json:"end"`
	Shares          uint32   `json:"shares"`
	Locktime        uint64   `json:"locktime"`
	Threshold       uint32   `json:"threshold"`
	Memo            string   `json:"memo"`
}

type DependencyTx struct {
	ID          ids.ID
	Tx          *platformvm.Tx
	RewardUTXOs []*avax.UTXO
}
