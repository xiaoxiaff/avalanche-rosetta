package pchain

import (
	"context"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	pmapper "github.com/ava-labs/avalanche-rosetta/mapper/pchain"
	"github.com/ava-labs/avalanche-rosetta/service"
	"github.com/ava-labs/avalanche-rosetta/service/backend/pchain/indexer"
)

type Backend struct {
	service.ConstructionBackend
	service.NetworkBackend

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

func (b *Backend) makeGenesisBlock() *types.BlockResponse {
	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       b.genesisBlockIdentifier,
			ParentBlockIdentifier: b.genesisBlockIdentifier,
			Transactions:          []*types.Transaction{},
			Timestamp:             mapper.UnixToUnixMilli(b.genesisBlock.Timestamp),
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

func (*Backend) ShouldHandleRequest(req interface{}) bool {
	switch r := req.(type) {
	case *types.AccountBalanceRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.AccountCoinsRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.BlockRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.BlockTransactionRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionDeriveRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionMetadataRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionPreprocessRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionPayloadsRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionParseRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionCombineRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionHashRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.ConstructionSubmitRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	case *types.NetworkRequest:
		return pmapper.IsPChain(r.NetworkIdentifier)
	}

	return false
}
