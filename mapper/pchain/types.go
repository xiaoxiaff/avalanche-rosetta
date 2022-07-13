package pchain

const (
	OpImport      = "IMPORT"
	OpExport      = "EXPORT"
	OpInput       = "INPUT"
	OpOutput      = "OUTPUT"
	OpStakeOutput = "STAKE"
)

type OperationMetadata struct {
	Type         string   `json:"type"`
	SigIndices   []uint32 `json:"sig_indices"`
	OutputOwners string   `json:"output_owners"`
}
