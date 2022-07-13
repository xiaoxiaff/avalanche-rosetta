package pchain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/validator"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

func buildImport() *platformvm.UnsignedImportTx {

	avaxAssetID, _ := ids.FromString("U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK")
	sourceChain, _ := ids.FromString("2JVSBoinj9C2J33VntvzYtVJNZdN2NKiwwKjcumHUWEb5DbBrm")
	outAddr1, _ := address.ParseToID("P-fuji1xm0r37l6gyf2mly4pmzc0tz6wnwqkugedh95fk")
	outAddr2, _ := address.ParseToID("P-fuji1fmragvegm5k26qzlt6vy0ghhdr508u6r4a5rxj")
	outAddr3, _ := address.ParseToID("P-fuji1j3sw805usytrsymfwxxrcwfqguyarumn45cllj")
	importedTxId, _ := ids.FromString("2DtYhzCvo9LRYMRJ6sCtYJ4aNPRpsibp46ETNyY6H5Cox1VLvX")
	impTx := &platformvm.UnsignedImportTx{
		BaseTx: platformvm.BaseTx{
			BaseTx: avax.BaseTx{
				Metadata:     avax.Metadata{},
				NetworkID:    uint32(5),
				BlockchainID: [32]byte{},
				Outs: []*avax.TransferableOutput{{
					Asset: avax.Asset{ID: avaxAssetID},
					FxID:  [32]byte{},
					Out: &secp256k1fx.TransferOutput{
						Amt: 8000000,
						OutputOwners: secp256k1fx.OutputOwners{
							Locktime:  0,
							Threshold: 2,
							Addrs:     []ids.ShortID{outAddr1, outAddr2, outAddr3},
						},
					},
				}},
				Ins:  nil,
				Memo: []byte{},
			},
		},
		SourceChain: sourceChain,
		ImportedInputs: []*avax.TransferableInput{{
			UTXOID: avax.UTXOID{
				TxID:        importedTxId,
				OutputIndex: 0,
				Symbol:      false,
			},
			Asset: avax.Asset{ID: avaxAssetID},
			FxID:  [32]byte{},
			In: &secp256k1fx.TransferInput{
				Amt: 9000000,
				Input: secp256k1fx.Input{
					SigIndices: []uint32{},
				},
			},
		}},
	}

	return impTx
}

var importTxString = `{
  "networkID": 5,
  "blockchainID": "11111111111111111111111111111111LpoYY",
  "outputs": [
    {
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "output": {
        "addresses": [
          "P-fuji1xm0r37l6gyf2mly4pmzc0tz6wnwqkugedh95fk",
          "P-fuji1fmragvegm5k26qzlt6vy0ghhdr508u6r4a5rxj", "P-fuji1j3sw805usytrsymfwxxrcwfqguyarumn45cllj"
        ],
        "amount": 8000000,
        "locktime": 0,
        "threshold": 2
      }
    }
  ],
  "inputs": [],
  "memo": "0x4d616e75616c6c79204578706f727420415641582066726f6d20582d436861696e20746f20502d436861696e",
  "sourceChain": "2JVSBoinj9C2J33VntvzYtVJNZdN2NKiwwKjcumHUWEb5DbBrm",
  "importedInputs": [
    {
      "txID": "2DtYhzCvo9LRYMRJ6sCtYJ4aNPRpsibp46ETNyY6H5Cox1VLvX",
      "outputIndex": 0,
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "input": {
        "amount": 9000000,
        "signatureIndices": [
          0
        ]
      }
    }
  ]
}`

func buildExport() *platformvm.UnsignedExportTx {
	avaxAssetID, _ := ids.FromString("U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK")
	outAddr, _ := address.ParseToID("P-fuji1wmd9dfrqpud6daq0cde47u0r7pkrr46ep60399")
	exportOutAddr, _ := address.ParseToID("P-fuji1wmd9dfrqpud6daq0cde47u0r7pkrr46ep60399")
	txId, _ := ids.FromString("27LaDkrUrMY1bhVf2i8RARCrRwFjeRw7vEu8ntLQXracgLzL1v")
	destinationId, _ := ids.FromString("yH8D7ThNJkxmtkuv2jgBa4P1Rn3Qpr4pPr7QYNfcdoS6k6HWp")
	exTx := &platformvm.UnsignedExportTx{
		BaseTx: platformvm.BaseTx{
			BaseTx: avax.BaseTx{
				Metadata:     avax.Metadata{},
				NetworkID:    uint32(5),
				BlockchainID: [32]byte{},
				Outs: []*avax.TransferableOutput{{
					Asset: avax.Asset{ID: avaxAssetID},
					FxID:  [32]byte{},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2910137500,
						OutputOwners: secp256k1fx.OutputOwners{
							Locktime:  0,
							Threshold: 1,
							Addrs:     []ids.ShortID{outAddr},
						},
					},
				}},
				Ins: []*avax.TransferableInput{{UTXOID: avax.UTXOID{TxID: txId, OutputIndex: 0, Symbol: false},
					Asset: avax.Asset{ID: avaxAssetID},
					FxID:  [32]byte{},
					In: &secp256k1fx.TransferInput{
						Amt:   2921137500,
						Input: secp256k1fx.Input{SigIndices: []uint32{}},
					},
				}},
				Memo: []byte{},
			},
		},
		DestinationChain: destinationId,
		ExportedOutputs: []*avax.TransferableOutput{{
			Asset: avax.Asset{ID: avaxAssetID},
			FxID:  [32]byte{},
			Out: &secp256k1fx.TransferOutput{
				Amt: 10000000,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{exportOutAddr},
				},
			},
		}},
	}

	return exTx
}

var exportTxString = `{
  "networkID": 5,
  "blockchainID": "11111111111111111111111111111111LpoYY",
  "outputs": [
    {
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "output": {
        "addresses": [
          "P-fuji1wmd9dfrqpud6daq0cde47u0r7pkrr46ep60399"
        ],
        "amount": 2910137500,
        "locktime": 0,
        "threshold": 1
      }
    }
  ],
  "inputs": [
    {
      "txID": "27LaDkrUrMY1bhVf2i8RARCrRwFjeRw7vEu8ntLQXracgLzL1v",
      "outputIndex": 0,
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "input": {
        "amount": 2921137500,
        "signatureIndices": [
          0
        ]
      }
    }
  ],
  "memo": "0x",
  "destinationChain": "yH8D7ThNJkxmtkuv2jgBa4P1Rn3Qpr4pPr7QYNfcdoS6k6HWp",
  "exportedOutputs": [
    {
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "output": {
        "addresses": [
          "P-fuji1wmd9dfrqpud6daq0cde47u0r7pkrr46ep60399"
        ],
        "amount": 10000000,
        "locktime": 0,
        "threshold": 1
      }
    }
  ]
}`

func buildAddDelegator() *platformvm.UnsignedAddDelegatorTx {
	avaxAssetID, _ := ids.FromString("U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK")
	txId, _ := ids.FromString("2JQGX1MBdszAaeV6eApCZm7CBpc917qWiyQ2cygFRJ6WteDkre")
	outAddr, _ := address.ParseToID("P-fuji1gdkq8g208e3j4epyjmx65jglsw7vauh86l47ac")
	validatorId, _ := ids.NodeIDFromString("NodeID-BFa1padLXBj7VHa2JYvYGzcTBPQGjPhUy")
	stakeAddr, _ := address.ParseToID("P-fuji1l022sue7g2kzvrcuxughl30xkss2cj0az3e5r2")
	rewardAddr, _ := address.ParseToID("P-fuji1l022sue7g2kzvrcuxughl30xkss2cj0az3e5r2")
	tx := &platformvm.UnsignedAddDelegatorTx{
		BaseTx: platformvm.BaseTx{
			BaseTx: avax.BaseTx{
				Metadata:     avax.Metadata{},
				NetworkID:    uint32(5),
				BlockchainID: [32]byte{},
				Outs: []*avax.TransferableOutput{{
					Asset: avax.Asset{ID: avaxAssetID},
					FxID:  [32]byte{},
					Out: &secp256k1fx.TransferOutput{
						Amt: 996649063,
						OutputOwners: secp256k1fx.OutputOwners{
							Locktime:  0,
							Threshold: 1,
							Addrs:     []ids.ShortID{outAddr},
						},
					},
				}},
				Ins: []*avax.TransferableInput{{UTXOID: avax.UTXOID{TxID: txId, OutputIndex: 0, Symbol: false},
					Asset: avax.Asset{ID: avaxAssetID},
					FxID:  [32]byte{},
					In: &secp256k1fx.TransferInput{
						Amt:   1996649063,
						Input: secp256k1fx.Input{SigIndices: []uint32{}},
					},
				}},
				Memo: []byte{},
			},
		},
		Validator: validator.Validator{
			NodeID: validatorId,
			Start:  1656058022,
			End:    1657872569,
			Wght:   1000000000,
		},
		Stake: []*avax.TransferableOutput{{
			Asset: avax.Asset{ID: avaxAssetID},
			FxID:  [32]byte{},
			Out: &secp256k1fx.TransferOutput{
				Amt: 1000000000,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{stakeAddr},
				},
			},
		}},
		RewardsOwner: &secp256k1fx.OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs:     []ids.ShortID{rewardAddr},
		},
	}

	return tx

}

var addDelegatorTxString = `{
  "networkID": 5,
  "blockchainID": "11111111111111111111111111111111LpoYY",
  "outputs": [
    {
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "output": {
        "addresses": [
          "P-fuji1gdkq8g208e3j4epyjmx65jglsw7vauh86l47ac"
        ],
        "amount": 996649063,
        "locktime": 0,
        "threshold": 1
      }
    }
  ],
  "inputs": [
    {
      "txID": "2JQGX1MBdszAaeV6eApCZm7CBpc917qWiyQ2cygFRJ6WteDkre",
      "outputIndex": 0,
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "input": {
        "amount": 1996649063,
        "signatureIndices": [
          0
        ]
      }
    }
  ],
  "memo": "0x",
  "validator": {
    "nodeID": "NodeID-BFa1padLXBj7VHa2JYvYGzcTBPQGjPhUy",
    "start": 1656058022,
    "end": 1657872569,
    "weight": 1000000000
  },
  "stake": [
    {
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "output": {
        "addresses": [
          "P-fuji1l022sue7g2kzvrcuxughl30xkss2cj0az3e5r2"
        ],
        "amount": 1000000000,
        "locktime": 0,
        "threshold": 1
      }
    }
  ],
  "rewardsOwner": {
    "addresses": [
      "P-fuji1l022sue7g2kzvrcuxughl30xkss2cj0az3e5r2"
    ],
    "locktime": 0,
    "threshold": 1
  }
}`

func buildValidatorTx() *platformvm.UnsignedAddValidatorTx {
	avaxAssetID, _ := ids.FromString("U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK")

	txId, _ := ids.FromString("88tfp1Pkw9vyKrRtVNiMrghFBrre6Q6CzqPW1t7StDNX9PJEo")
	stakeAddr, _ := address.ParseToID("P-fuji1ljdzyey6vu3hgn3cwg4j5lpy0svd6arlxpj6je")
	rewardAddr, _ := address.ParseToID("P-fuji1ljdzyey6vu3hgn3cwg4j5lpy0svd6arlxpj6je")
	validatorId, _ := ids.NodeIDFromString("NodeID-CCecHmRK3ANe92VyvASxkNav26W4vAVpX")
	addvalidator := &platformvm.UnsignedAddValidatorTx{
		BaseTx: platformvm.BaseTx{
			BaseTx: avax.BaseTx{
				Metadata:     avax.Metadata{},
				NetworkID:    uint32(5),
				BlockchainID: [32]byte{},
				Outs:         nil,
				Ins: []*avax.TransferableInput{{
					UTXOID: avax.UTXOID{TxID: txId, OutputIndex: 0, Symbol: false},
					Asset:  avax.Asset{ID: avaxAssetID},
					FxID:   [32]byte{},
					In: &secp256k1fx.TransferInput{
						Amt:   2000000000,
						Input: secp256k1fx.Input{SigIndices: []uint32{}},
					},
				}},
				Memo: []byte{},
			},
		},
		Validator: validator.Validator{
			NodeID: validatorId,
			Start:  1656084079,
			End:    1687620079,
			Wght:   2000000000,
		},
		Stake: []*avax.TransferableOutput{{
			Asset: avax.Asset{ID: avaxAssetID},
			FxID:  [32]byte{},
			Out: &secp256k1fx.TransferOutput{
				Amt: 2000000000,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{stakeAddr},
				},
			},
		}},
		RewardsOwner: &secp256k1fx.OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs:     []ids.ShortID{rewardAddr},
		},
		Shares: 20000}

	return addvalidator
}

var addValidatortxString = `{
  "networkID": 5,
  "blockchainID": "11111111111111111111111111111111LpoYY",
  "outputs": [],
  "inputs": [
    {
      "txID": "88tfp1Pkw9vyKrRtVNiMrghFBrre6Q6CzqPW1t7StDNX9PJEo",
      "outputIndex": 0,
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "input": {
        "amount": 2000000000,
        "signatureIndices": [
          0
        ]
      }
    }
  ],
  "memo": "0x",
  "validator": {
    "nodeID": "NodeID-CCecHmRK3ANe92VyvASxkNav26W4vAVpX",
    "start": 1656084079,
    "end": 1687620079,
    "weight": 2000000000
  },
  "stake": [
    {
      "assetID": "U8iRqJoiJm8xZHAacmvYyZVwqQx6uDNtQeP3CQ6fcgQk3JqnK",
      "fxID": "spdxUxVJQbX85MGxMHbKw1sHxMnSqJ3QBzDyDYEP3h6TLuxqQ",
      "output": {
        "addresses": [
          "P-fuji1ljdzyey6vu3hgn3cwg4j5lpy0svd6arlxpj6je"
        ],
        "amount": 2000000000,
        "locktime": 0,
        "threshold": 1
      }
    }
  ],
  "rewardsOwner": {
    "addresses": [
      "P-fuji1ljdzyey6vu3hgn3cwg4j5lpy0svd6arlxpj6je"
    ],
    "locktime": 0,
    "threshold": 1
  },
  "shares": 20000
}`
