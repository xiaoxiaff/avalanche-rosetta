# Transactions
1. [BaseTx](#UnsignedBaseTx)
2. [AddValidator](#UnsignedAddValidatorTx)
3. [AddSubnetValidator](#UnsignedAddSubnetValidatorTx)
4. [AddDelegator](#UnsignedAddDelegatorTx)
5. [CreateChain](#UnsignedCreateChainTx)
6. [CreateSubnet](#UnsignedCreateSubnetTx)
7. [Import](#UnsignedImportTx)
8. [Export](#UnsignedExportTx)

## Transactions format
### UnsignedBaseTx
| Field      | Description |
| ----------- | ----------- |
| TypeID      | is the ID for this type. It is 0x00000000.       |
| NetworkID    | is an int that defines which network this transaction is meant to be issued to. This value is meant to support transaction routing and is not designed for replay attack prevention.        |
| BlockchainID     | is a 32-byte array that defines which blockchain this transaction was issued to. This is used for replay attack prevention for transactions that could potentially be valid across network or blockchain.|
|[Outputs](#SECP256K1TransferOutput) |is an array of transferable output objects. Outputs must be sorted lexicographically by their serialized representation. The total quantity of the assets created in these outputs must be less than or equal to the total quantity of each asset consumed in the inputs minus the transaction fee.
|[Inputs](#SECP256K1TransferInput) |is an array of transferable input objects. Inputs must be sorted and unique. Inputs are sorted first lexicographically by their TxID and then by the UTXOIndex from low to high. If there are inputs that have the same TxID and UTXOIndex, then the transaction is invalid as this would result in a double spend.
|Memo | Memo field contains arbitrary bytes, up to 256 bytes.


### UnsignedAddValidatorTx
| Field      | Description |
| ----------- | ----------- |
| [BaseTx](#unsignedbasetx)| 
|Validator |has a NodeID, StartTime, EndTime, and Weight
|Validator.NodeID |is 20 bytes which is the node ID of the validator.
|Validator.StartTime  |is a long which is the Unix time when the validator starts validating.
|Validator.EndTime |is a long which is the Unix time when the validator stops validating.
|Validator.Header|is a long which is the amount the validator stakes
|Stake |Stake has LockedOuts
|[Stake.LockedOuts](#secp256k1transferoutput)|LockedOuts An array of Transferable Outputs that are locked for the duration of the staking period. At the end of the staking period, these outputs are refunded to their respective addresses.
|RewardsOwner |A SECP256K1OutputOwners
|Shares |10,000 times percentage of reward taken from delegators


### UnsignedAddSubnetValidatorTx
| Field      | Description |
| ----------- | ----------- |
| [BaseTx](#unsignedbasetx)| 
|Validator |Validator has a NodeID, StartTime, EndTime, and Weight
|Validator.NodeID |is 20 bytes which is the node ID of the validator.
|Validator.StartTime |is a long which is the Unix time when the validator starts validating.
|Validator.EndTime |is a long which is the Unix time when the validator stops validating.
|Validator.Weight |is a long which is the amount the validator stakes
|SubnetID |a 32 byte subnet id
|Header|contains SigIndices and has a type id of 0x0000000a. SigIndices is a list of unique ints that define the addresses signing the control signature to add a validator to a subnet. The array must be sorted low to high.

### UnsignedAddDelegatorTx
| Field      | Description |
| ----------- | ----------- |
|BaseTx|
|Validator |Validator has a NodeID, StartTime, EndTime, and Weight
|Validator.NodeID |is 20 bytes which is the node ID of the validator.
|Validator.StartTime |is a long which is the Unix time when the validator starts validating.
|Validator.EndTime |is a long which is the Unix time when the validator stops validating.
|Validator.Weight |is a long which is the amount the validator stakes
|Stake |Stake has LockedOuts
|[Stake.LockedOuts](#secp256k1transferoutput)|LockedOuts An array of Transferable Outputs that are locked for the duration of the staking period. At the end of the staking period, these outputs are refunded to their respective addresses.
|RewardsOwner| An SECP256K1OutputOwners


### UnsignedCreateChainTx
| Field      | Description |
| ----------- | ----------- |
| [BaseTx](#unsignedbasetx)| 
|SubnetID |a 32 byte subnet id
|ChainName|A human readable name for the chain; need not be unique
| VMID|ID of the VM running on the new chain
| FxIDs|IDs of the feature extensions running on the new chain
| GenesisData|Byte representation of genesis state of the new chain
| SubnetAuth|Authorizes this blockchain to be added to this subnet


### UnsignedCreateSubnetTx
| Field      | Description |
| ----------- | ----------- |
| [BaseTx](#unsignedbasetx)| 
| RewardsOwner|A SECP256K1OutputOwners


### UnsignedImportTx
| Field      | Description |
| ----------- | ----------- |
| [BaseTx](#unsignedbasetx)| 
| SourceChain|is a 32-byte source blockchain ID.
| [Inputs](#SECP256K1TransferInput)|is a variable length array of Transferable Inputs.


### UnsignedExportTx
| Field      | Description |
| ----------- | ----------- |
| DestinationChain|is the 32 byte ID of the chain where the funds are being exported to.
| [Outputs](#SECP256K1TransferOutput)|is a variable length array of Transferable Outputs.


## SECP256K1TransferOutput
| Field      | Description |
| ----------- | ----------- |
|TypeID| is the ID for this output type. It is 0x00000007.
|Amount| is a long that specifies the quantity of the asset that this output owns. Must be positive.
|Locktime| is a long that contains the unix timestamp that this output can be spent after. The unix timestamp is specific to the second.
|Threshold| is an int that names the number of unique signatures required to spend the output. Must be less than or equal to the length of Addresses. If Addresses is empty, must be 0.
|Addresses| is a list of unique addresses that correspond to the private keys that can be used to spend this output. Addresses must be sorted lexicographically.


## SECP256K1TransferInput
| Field      | Description |
| ----------- | ----------- |
|TypeID| is the ID for this output type. It is 0x00000005.
|Amount| is a long that specifies the quantity that this input should be consuming from the UTXO. Must be positive. Must be equal to the amount specified in the UTXO.
|AddressIndices| is a list of unique ints that define the private keys are being used to spend the UTXO. Each UTXO has an array of addresses that can spend the UTXO. Each int represents the index in this address array that will sign this transaction. The array must be sorted low to high.
