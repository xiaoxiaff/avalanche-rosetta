package mapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeUTXOID(t *testing.T) {
	testCases := map[string]struct {
		id        string
		errMsg    string
		expectErr bool
	}{
		"empty string": {
			id:        "",
			expectErr: true,
			errMsg:    "invalid utxo ID format",
		},
		"invalid id without index": {
			id:        "2KWdUnE6Qp4CbSj3Bb5ZVcLqdCYECy4AJuWUxFBG8ACxMBKtCx",
			expectErr: true,
			errMsg:    "invalid utxo ID format",
		},
		"invalid id without invalid index": {
			id:        "2KWdUnE6Qp4CbSj3Bb5ZVcLqdCYECy4AJuWUxFBG8ACxMBKtCx:a",
			expectErr: true,
			errMsg:    "invalid syntax",
		},
		"valid id": {
			id:        "2KWdUnE6Qp4CbSj3Bb5ZVcLqdCYECy4AJuWUxFBG8ACxMBKtCx:1",
			expectErr: false,
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			utxoID, err := DecodeUTXOID(tc.id)
			if tc.expectErr {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NotNil(t, utxoID)
				assert.Nil(t, err)
			}
		})
	}
}
