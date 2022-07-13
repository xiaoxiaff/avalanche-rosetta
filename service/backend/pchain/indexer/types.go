package indexer

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/fx"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
)

type ParsedTx struct {
	TxID      ids.ID       `json:"id"`
	TxType    string       `json:"type"`
	BlockID   ids.ID       `json:"block"`
	Timestamp int64        `json:"timestamp"`
	Creds     [][]CredData `json:"credentials"`
	Fee       uint64       `json:"fee"`
}

type CredData struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
}

type BaseTxData struct {
	Outs []*avax.UTXO `json:"outputs"`

	avax.BaseTx
}

type AddValidatorData struct {
	Shares uint32 `json:"shares"`
	AddDelegatorData
}

type ParsedAddValidatorTx struct {
	ParsedTx
	BaseTxData
	AddValidatorData `json:"data"`
}

type AddDelegatorData struct {
	Validator    validator.Validator        `json:"validator"`
	RewardsOwner fx.Owner                   `json:"rewardOwner"`
	Stake        []*avax.TransferableOutput `json:"stake"`
}

type ParsedAddDelegatorTx struct {
	ParsedTx
	BaseTxData
	AddDelegatorData `json:"data"`
}

type AdvanceTimeData struct {
	Time int64 `json:"time"`
}

type ParsedAdvanceTimeTx struct {
	ParsedTx
	AdvanceTimeData `json:"data"`
}

type ImportData struct {
	SourceChain    ids.ID                    `json:"sourceChain"`
	ImportedInputs []*avax.TransferableInput `json:"importedInputs"`
}

type ParsedImportTx struct {
	ParsedTx
	BaseTxData
	ImportData `json:"data"`
}

type ExportData struct {
	DestinationChain ids.ID                     `json:"destinationChain"`
	ExportedOutputs  []*avax.TransferableOutput `json:"exportedOutputs"`
}

type ParsedExportTx struct {
	ParsedTx
	BaseTxData
	ExportData `json:"data"`
}

type ParsedCreateSubnetTx struct {
	ParsedTx
	BaseTxData
	CreateSubnetData `json:"data"`
}

type CreateSubnetData struct {
	Owner fx.Owner `json:"owner"`
}

type CreateChainData struct {
	SubnetID    ids.ID            `json:"subnetID"`
	ChainName   string            `json:"chainName"`
	VMID        ids.ID            `json:"vmID"`
	FxIDs       []ids.ID          `json:"fxIDs"`
	GenesisData []byte            `json:"genesisData"`
	SubnetAuth  verify.Verifiable `json:"subnetAuthorization"`
}

type ParsedCreateChainTx struct {
	ParsedTx
	BaseTxData
	CreateChainData `json:"data"`
}

type RewardValidatorData struct {
	TxID ids.ID `json:"txID"`
}

type ParsedRewardValidatorTx struct {
	ParsedTx
	BaseTxData
	RewardValidatorData `json:"data"`
}

type ParsedAddSubnetValidatorTx struct {
	ParsedTx
	BaseTxData
	AddSubnetValidatorData `json:"data"`
}

type AddSubnetValidatorData struct {
	Validator  validator.SubnetValidator `json:"validator"`
	SubnetAuth verify.Verifiable         `json:"subnetAuthorization"`
}

type ParsedBlock struct {
	BlockID   ids.ID        `json:"id"`
	BlockType string        `json:"type"`
	ParentID  ids.ID        `json:"parent"`
	Timestamp int64         `json:"timestamp"`
	Height    uint64        `json:"height"`
	Txs       []interface{} `json:"transactions"`
	Proposer  `json:"proposer"`
}

type GenesisBlockData struct {
	Message       string                    `json:"message"`
	InitialSupply uint64                    `json:"initialSupply"`
	UTXOs         []*platformvm.GenesisUTXO `json:"utxos"`
}

type ParsedGenesisBlock struct {
	ParsedBlock
	GenesisBlockData `json:"data"`
}

type Proposer struct {
	ID           ids.ID     `json:"id"`
	ParentID     ids.ID     `json:"parent"`
	NodeID       ids.NodeID `json:"nodeID"`
	PChainHeight uint64     `json:"pChainHeight"`
	Timestamp    int64      `json:"timestamp"`
}
