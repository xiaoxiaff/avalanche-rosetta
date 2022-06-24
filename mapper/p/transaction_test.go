package mapper

import (
	"testing"

	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/stretchr/testify/assert"
)

func TestMapAddValidatorTx(t *testing.T) {

	pTx := platformvm.Tx{}
	_, err := platformvm.Codec.Unmarshal(addValidatortxBytes, &pTx)
	assert.Nil(t, err)

	err = pTx.Sign(platformvm.Codec, [][]*crypto.PrivateKeySECP256K1R{})
	assert.Nil(t, err)

	addvalidatorTx := pTx.UnsignedTx.(*platformvm.UnsignedAddValidatorTx)

	assert.Equal(t, 8, len(addvalidatorTx.Ins))
	assert.Equal(t, 1, len(addvalidatorTx.Outs))

	rosettaTransaction, err := Transaction(addvalidatorTx)
	assert.Nil(t, err)

	assert.Equal(t, 9, len(rosettaTransaction.Operations))

	// TODO: Add TxIns and TxOuts validation
	// addValTxIns := addvalidatorTx.Ins[0].In.(*secp256k1fx.TransferInput)
	// addValTxOuts := addvalidatorTx.Outs[0].Out.(*secp256k1fx.TransferOutput)
	// r, err := address.Format("P", "fuji", rewardOwnerAddrShort[:])
}
