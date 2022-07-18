package pchain

import (
	"testing"

	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/avalanche-rosetta/mapper"
)

func TestMapInOperation(t *testing.T) {
	addValidatorTx := buildValidatorTx()

	assert.Equal(t, 1, len(addValidatorTx.Ins))
	assert.Equal(t, 0, len(addValidatorTx.Outs))

	avaxIn := addValidatorTx.Ins[0]

	rosettaInOp, err := inToOperation([]*avax.TransferableInput{avaxIn}, 9, OpAddValidator, OpTypeInput, false)
	assert.Nil(t, err)

	assert.Equal(t, int64(9), rosettaInOp[0].OperationIdentifier.Index)
	assert.Equal(t, OpAddValidator, rosettaInOp[0].Type)
	assert.Equal(t, avaxIn.UTXOID.String(), rosettaInOp[0].CoinChange.CoinIdentifier.Identifier)
	assert.Equal(t, types.CoinSpent, rosettaInOp[0].CoinChange.CoinAction)
	assert.Equal(t, OpTypeInput, rosettaInOp[0].Metadata["type"])
	assert.Equal(t, OpAddValidator, rosettaInOp[0].Type)
	assert.Equal(t, types.String(mapper.StatusSuccess), rosettaInOp[0].Status)

	assert.Nil(t, rosettaInOp[0].Metadata["threshold"])
	assert.NotNil(t, rosettaInOp[0].Metadata["sig_indices"])
}

func TestMapOutOperation(t *testing.T) {
	addDelegatorTx := buildAddDelegator()

	assert.Equal(t, 1, len(addDelegatorTx.Ins))
	assert.Equal(t, 1, len(addDelegatorTx.Outs))

	avaxOut := addDelegatorTx.Outs[0]

	rosettaOutOp, err := outToOperation([]*avax.TransferableOutput{avaxOut}, 0, OpAddDelegator, OpTypeOutput, true)
	assert.Nil(t, err)

	assert.Equal(t, int64(0), rosettaOutOp[0].OperationIdentifier.Index)
	assert.Equal(t, "P-fuji1gdkq8g208e3j4epyjmx65jglsw7vauh86l47ac", rosettaOutOp[0].Account.Address)
	assert.Equal(t, mapper.AvaxCurrency, rosettaOutOp[0].Amount.Currency)
	assert.Equal(t, "996649063", rosettaOutOp[0].Amount.Value)
	assert.Equal(t, OpTypeOutput, rosettaOutOp[0].Metadata["type"])
	assert.Nil(t, rosettaOutOp[0].Status)
	assert.Equal(t, OpAddDelegator, rosettaOutOp[0].Type)

	assert.NotNil(t, rosettaOutOp[0].Metadata["threshold"])
	assert.NotNil(t, rosettaOutOp[0].Metadata["locktime"])
	assert.Nil(t, rosettaOutOp[0].Metadata["sig_indices"])

}

func TestMapAddValidatorTx(t *testing.T) {

	addvalidatorTx := buildValidatorTx()

	assert.Equal(t, 1, len(addvalidatorTx.Ins))
	assert.Equal(t, 0, len(addvalidatorTx.Outs))

	rosettaTransaction, err := Transaction(addvalidatorTx, true)
	assert.Nil(t, err)

	total := len(addvalidatorTx.Ins) + len(addvalidatorTx.Outs) + len(addvalidatorTx.Stake)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	nbTxType, nbInputMeta, nbOutputMeta, nbMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpAddValidator, OpTypeStakeOutput)

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

	rosettaTransaction, err := Transaction(addDelegatorTx, true)
	assert.Nil(t, err)

	total := len(addDelegatorTx.Ins) + len(addDelegatorTx.Outs) + len(addDelegatorTx.Stake)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	nbTxType, nbInputMeta, nbOutputMeta, nbMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpAddDelegator, OpTypeStakeOutput)

	assert.Equal(t, 3, nbTxType)
	assert.Equal(t, 1, nbInputMeta)
	assert.Equal(t, 1, nbOutputMeta)
	assert.Equal(t, 1, nbMetaType)

	assert.Equal(t, types.CoinSpent, rosettaTransaction.Operations[0].CoinChange.CoinAction)
	assert.Equal(t, types.CoinCreated, rosettaTransaction.Operations[1].CoinChange.CoinAction)
	assert.Equal(t, types.CoinCreated, rosettaTransaction.Operations[2].CoinChange.CoinAction)

	assert.Equal(t, addDelegatorTx.Ins[0].UTXOID.String(), rosettaTransaction.Operations[0].CoinChange.CoinIdentifier.Identifier)
	assert.Equal(t, (*types.CoinIdentifier)(nil), rosettaTransaction.Operations[1].CoinChange.CoinIdentifier)
	assert.Equal(t, (*types.CoinIdentifier)(nil), rosettaTransaction.Operations[2].CoinChange.CoinIdentifier)

	assert.Equal(t, int64(0), rosettaTransaction.Operations[0].OperationIdentifier.Index)
	assert.Equal(t, int64(1), rosettaTransaction.Operations[1].OperationIdentifier.Index)
	assert.Equal(t, int64(2), rosettaTransaction.Operations[2].OperationIdentifier.Index)

	assert.Equal(t, OpAddDelegator, rosettaTransaction.Operations[0].Type)
	assert.Equal(t, OpAddDelegator, rosettaTransaction.Operations[1].Type)
	assert.Equal(t, OpAddDelegator, rosettaTransaction.Operations[2].Type)

	assert.Equal(t, OpTypeInput, rosettaTransaction.Operations[0].Metadata["type"])
	assert.Equal(t, OpTypeOutput, rosettaTransaction.Operations[1].Metadata["type"])
	assert.Equal(t, OpTypeStakeOutput, rosettaTransaction.Operations[2].Metadata["type"])
}
func TestMapImportTx(t *testing.T) {

	importTx := buildImport()

	assert.Equal(t, 0, len(importTx.Ins))
	assert.Equal(t, 1, len(importTx.Outs))
	assert.Equal(t, 1, len(importTx.ImportedInputs))

	rosettaTransaction, err := Transaction(importTx, true)
	assert.Nil(t, err)

	total := len(importTx.Ins) + len(importTx.Outs) + len(importTx.ImportedInputs)
	assert.Equal(t, total, len(rosettaTransaction.Operations))

	nbTxType, nbInputMeta, nbOutputMeta, nbMetaType := verifyRosettaTransaction(rosettaTransaction.Operations, OpImportAvax, OpTypeImport)

	assert.Equal(t, 2, nbTxType)
	assert.Equal(t, 0, nbInputMeta)
	assert.Equal(t, 1, nbOutputMeta)
	assert.Equal(t, 1, nbMetaType)

	assert.Equal(t, types.CoinSpent, rosettaTransaction.Operations[0].CoinChange.CoinAction)
	assert.Equal(t, types.CoinCreated, rosettaTransaction.Operations[1].CoinChange.CoinAction)
}

func TestMapExportTx(t *testing.T) {

	exportTx := buildExport()

	assert.Equal(t, 1, len(exportTx.Ins))
	assert.Equal(t, 1, len(exportTx.Outs))
	assert.Equal(t, 1, len(exportTx.ExportedOutputs))

	rosettaTransaction, err := Transaction(exportTx, true)
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

		if meta.Type == OpTypeInput {
			nbOpInputMeta++
			continue
		}
		if meta.Type == OpTypeOutput {
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
