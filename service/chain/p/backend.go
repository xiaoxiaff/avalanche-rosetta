package p

import (
	"context"
	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service/chain"
	"github.com/ava-labs/avalanche-rosetta/service/chain/p/indexer"
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"
)

type Backend struct {
	chain.ConstructionBackend
	chain.NetworkBackend

	networkIdentifier      *types.NetworkIdentifier
	fac                    *crypto.FactorySECP256K1R
	pClient                client.PChainClient
	indexerParser          *indexer.Parser
	getUTXOsPageSize       uint32
	codec                  codec.Manager
	codecVersion           uint16
	assetID                ids.ID
	genesisBlock           *indexer.ParsedGenesisBlock
	genesisBlockIdentifier *types.BlockIdentifier
}

func (c *Backend) makeGenesisBlock() *types.BlockResponse {
	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       c.genesisBlockIdentifier,
			ParentBlockIdentifier: c.genesisBlockIdentifier,
			Transactions:          []*types.Transaction{},
			Timestamp:             mapper.UnixToUnixMilli(c.genesisBlock.Timestamp),
		},
	}
}

func NewBackend(
	ctx context.Context,
	pClient client.PChainClient,
	assetID ids.ID,
	networkIdentifier *types.NetworkIdentifier,
) (*Backend, error) {
	indexerParser, err := indexer.NewParser(ctx, pClient)
	if err != nil {
		return nil, err
	}

	// Initializing parser gives parsed genesis block
	genesisBlock, err := indexerParser.Initialize(ctx)
	if err != nil {
		return nil, err
	}

	return &Backend{
		fac:               &crypto.FactorySECP256K1R{},
		pClient:           pClient,
		networkIdentifier: networkIdentifier,
		getUTXOsPageSize:  1024,
		codec:             platformvm.Codec,
		codecVersion:      platformvm.CodecVersion,
		assetID:           assetID,
		indexerParser:     indexerParser,
		genesisBlock:      genesisBlock,
		genesisBlockIdentifier: &types.BlockIdentifier{
			Index: int64(genesisBlock.Height),
			Hash:  genesisBlock.BlockID.String(),
		},
	}, nil
}
