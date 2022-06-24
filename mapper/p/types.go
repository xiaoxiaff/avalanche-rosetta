package mapper

const (
	OpImport      = "IMPORT"
	OpInput       = "INPUT"
	OpOutput      = "OUTPUT"
	OpStakeOutput = "STAKE"
)

type OperationMetadata struct {
	Type         string   `json:"type"`
	SigIndices   []uint32 `json:"sig_indices"`
	OutputOwners []byte   `json:"output_owners"`
}
