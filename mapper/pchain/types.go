package pchain

import (
	"github.com/ava-labs/avalanchego/ids"
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

	OpTypeImport      = "IMPORT"
	OpTypeExport      = "EXPORT"
	OpTypeInput       = "INPUT"
	OpTypeOutput      = "OUTPUT"
	OpTypeStakeOutput = "STAKE"
	OpTypeReward      = "REWARD"
	OpTypeCreateChain = "CREATE_CHAIN"

	MetadataOpType      = "type"
	MetadataStakingTxID = "staking_tx_id"
	MetadataSubnetID    = "subnet_id"
	MetadataChainName   = "chain_name"
	MetadataVMID        = "vmid"
	MetadataMemo        = "memo"
	MetadataMessage     = "message"

	UTXOTypeSharedMemory = "shared_memory"
	IsAtomicUTXO         = "is_atomic"
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
	MultiSig    bool     `json:"multi_sig,omitempty"`
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
	Wght            uint64   `json:"weight"`
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
	Wght            uint64   `json:"weight"`
	Shares          uint32   `json:"shares"`
	Locktime        uint64   `json:"locktime"`
	Threshold       uint32   `json:"threshold"`
	Memo            string   `json:"memo"`
}

func TimestampForBlock(block int64) int64 {
	// url := "https://testnet.avascan.info/blockchain/p/block/" + strconv.FormatInt(block, 10)
	// resp, err := http.Get(url)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// //We Read the response body on the line below.
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// //Convert the body to type string
	// sb := string(body)
	// idx := strings.Index(sb, "data-timestamp=")
	// startIdx := idx + len("data-timestamp=") + 1
	// endIdx := strings.Index(sb, "data-timestamp-full=") - 2
	// timestampStr := sb[startIdx:endIdx]
	// timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(timestamp)
	return 1657819573000 - block
}
