package pchain

import (
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

func TestShouldHandleRequest(t *testing.T) {
	pChainNetworkId := &types.NetworkIdentifier{
		Blockchain:           "Avalanche",
		Network:              "Mainnet",
		SubNetworkIdentifier: &types.SubNetworkIdentifier{Network: "P"},
	}
	cChainNetworkId := &types.NetworkIdentifier{
		Blockchain: "Avalanche",
		Network:    "Mainnet",
	}

	backend := &Backend{}

	t.Run("return true for p-chain", func(t *testing.T) {
		assert.True(t, backend.ShouldHandleRequest(&types.AccountBalanceRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.AccountCoinsRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.BlockRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.BlockTransactionRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionDeriveRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionMetadataRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionPreprocessRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionPayloadsRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionParseRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionCombineRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionHashRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.ConstructionSubmitRequest{NetworkIdentifier: pChainNetworkId}))
		assert.True(t, backend.ShouldHandleRequest(&types.NetworkRequest{NetworkIdentifier: pChainNetworkId}))
	})

	t.Run("return false for c-chain", func(t *testing.T) {
		assert.False(t, backend.ShouldHandleRequest(&types.AccountBalanceRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.AccountCoinsRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.BlockRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.BlockTransactionRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionDeriveRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionMetadataRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionPreprocessRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionPayloadsRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionParseRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionCombineRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionHashRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.ConstructionSubmitRequest{NetworkIdentifier: cChainNetworkId}))
		assert.False(t, backend.ShouldHandleRequest(&types.NetworkRequest{NetworkIdentifier: cChainNetworkId}))
	})

}
