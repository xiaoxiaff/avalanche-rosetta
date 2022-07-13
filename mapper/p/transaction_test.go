package mapper

import (
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"testing"

	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

func TestMapInOperation(t *testing.T) {

	addvalidatorTx := buildValidatorTx()
	assert.Equal(t, 1, len(addvalidatorTx.Ins))
	assert.Equal(t, 0, len(addvalidatorTx.Outs))

	avaxIn := addvalidatorTx.Ins[0]

	rosettaInOp, err := inToOperation([]*avax.TransferableInput{avaxIn}, 9, mapper.OpAddValidator, OpInput)
	assert.Nil(t, err)

	assert.Equal(t, int64(9), rosettaInOp[0].OperationIdentifier.Index)
	assert.Equal(t, mapper.OpAddValidator, rosettaInOp[0].Type)
	assert.Equal(t, avaxIn.AssetID().String(), rosettaInOp[0].CoinChange.CoinIdentifier.Identifier)
	assert.Equal(t, types.CoinSpent, rosettaInOp[0].CoinChange.CoinAction)
	assert.Equal(t, OpInput, rosettaInOp[0].Metadata["type"])
}

func TestMapOutOperation(t *testing.T) {
	addDelegatorTx := buildAddDelegator()

	assert.Equal(t, 1, len(addDelegatorTx.Ins))
	assert.Equal(t, 1, len(addDelegatorTx.Outs))

	avaxOut := addDelegatorTx.Outs[0]

	rosettaInOp, err := outToOperation([]*avax.TransferableOutput{avaxOut}, 9, mapper.OpAddValidator, OpOutput)
	assert.Nil(t, err)

	assert.Equal(t, int64(9), rosettaInOp[0].OperationIdentifier.Index)
	assert.Equal(t, mapper.OpAddValidator, rosettaInOp[0].Type)
	assert.Equal(t, "P-fuji1gdkq8g208e3j4epyjmx65jglsw7vauh86l47ac", rosettaInOp[0].Account.Address)
	assert.Equal(t, mapper.AvaxCurrency, rosettaInOp[0].Amount.Currency)
	assert.Equal(t, "996649063", rosettaInOp[0].Amount.Value)
	assert.Equal(t, OpOutput, rosettaInOp[0].Metadata["type"])
}

func TestMapAddValidatorTx(t *testing.T) {

	addvalidatorTx := buildValidatorTx()

	assert.Equal(t, 1, len(addvalidatorTx.Ins))
	assert.Equal(t, 0, len(addvalidatorTx.Outs))

	rosettaTransaction, err := Transaction(addvalidatorTx)
	assert.Nil(t, err)

	total := len(addvalidatorTx.Ins) + len(addvalidatorTx.Outs) + len(addvalidatorTx.Stake)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	nbTxType, nbInputMeta, nbOutputMeta, nbMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, mapper.OpAddValidator, OpStakeOutput)

	assert.Equal(t, 2, nbTxType)
	assert.Equal(t, 1, nbInputMeta)
	assert.Equal(t, 0, nbOutputMeta)
	assert.Equal(t, 1, nbMetaType)

}

func TestMapAddDelegatorTx(t *testing.T) {

	addDelegatorTx := buildAddDelegator()

	assert.Equal(t, 1, len(addDelegatorTx.Ins))
	assert.Equal(t, 1, len(addDelegatorTx.Outs))
	assert.Equal(t, 1, len(addDelegatorTx.Stake))

	rosettaTransaction, err := Transaction(addDelegatorTx)
	assert.Nil(t, err)

	total := len(addDelegatorTx.Ins) + len(addDelegatorTx.Outs) + len(addDelegatorTx.Stake)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	nbTxType, nbInputMeta, nbOutputMeta, nbMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, mapper.OpAddDelegator, OpStakeOutput)

	assert.Equal(t, 3, nbTxType)
	assert.Equal(t, 1, nbInputMeta)
	assert.Equal(t, 1, nbOutputMeta)
	assert.Equal(t, 1, nbMetaType)

}
func TestMapImportTx(t *testing.T) {

	importTx := buildImport()

	assert.Equal(t, 0, len(importTx.Ins))
	assert.Equal(t, 1, len(importTx.Outs))
	assert.Equal(t, 1, len(importTx.ImportedInputs))

	rosettaTransaction, err := Transaction(importTx)
	assert.Nil(t, err)

	total := len(importTx.Ins) + len(importTx.Outs) + len(importTx.ImportedInputs)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	nbTxType, nbInputMeta, nbOutputMeta, nbMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, mapper.OpImport, OpImport)

	assert.Equal(t, 2, nbTxType)
	assert.Equal(t, 0, nbInputMeta)
	assert.Equal(t, 1, nbOutputMeta)
	assert.Equal(t, 1, nbMetaType)

}

func TestMapExportTx(t *testing.T) {

	exportTx := buildExport()

	assert.Equal(t, 1, len(exportTx.Ins))
	assert.Equal(t, 1, len(exportTx.Outs))
	assert.Equal(t, 1, len(exportTx.ExportedOutputs))

	rosettaTransaction, err := Transaction(exportTx)
	assert.Nil(t, err)

	total := len(exportTx.Ins) + len(exportTx.Outs) + len(exportTx.ExportedOutputs)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	nbTxType, nbInputMeta, nbOutputMeta, _ := verifyRosettaTransaction(rosettaTransaction.Operations, mapper.OpExport, "")

	assert.Equal(t, 3, nbTxType)
	assert.Equal(t, 1, nbInputMeta)
	assert.Equal(t, 2, nbOutputMeta)
}

func verifyRosettaTransaction(operations []*types.Operation, txType string, metaType string) (int, int, int, int) {

	nbOpInputMeta := 0
	nbOpOutputMeta := 0
	nbMetaType := 0
	nbTxType := 0

	for _, v := range operations {
		if v.Type == txType {
			nbTxType++
		}

		meta := &OperationMetadata{}
		_ = mapper.UnmarshalJSONMap(v.Metadata, meta)

		if meta.Type == OpInput {
			nbOpInputMeta++
			continue
		}
		if meta.Type == OpOutput {
			nbOpOutputMeta++
			continue
		}
		if meta.Type == metaType {
			nbMetaType++
			continue
		}
	}

	return nbTxType, nbOpInputMeta, nbOpOutputMeta, nbMetaType
}
