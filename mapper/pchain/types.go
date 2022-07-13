package pchain

const (
	OpImportAvax      = "IMPORT_AVAX"
	OpExportAvax      = "EXPORT_AVAX"
	OpAddValidator    = "ADD_VALIDATOR"
	OpAddDelegator    = "ADD_DELEGATOR"
	OpRewardValidator = "REWARD_VALIDATOR"

	OpImport      = "IMPORT"
	OpExport      = "EXPORT"
	OpInput       = "INPUT"
	OpOutput      = "OUTPUT"
	OpStakeOutput = "STAKE"

	MetaStakingTxId = "staking_tx"
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
	Type         string   `json:"type"`
	SigIndices   []uint32 `json:"sig_indices"`
	OutputOwners string   `json:"output_owners"`
}
