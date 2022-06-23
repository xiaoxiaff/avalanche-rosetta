package p

const (
	OpImportAvax   = "IMPORT_AVAX"
	OpExportAvax   = "EXPORT_AVAX"
	OpAddValidator = "ADD_VALIDATOR"
	OpAddDelegator = "ADD_DELEGATOR"
)

var (
	PChainOperationTypes = []string{
		OpImportAvax,
		OpExportAvax,
		OpAddValidator,
		OpAddDelegator,
	}
)
