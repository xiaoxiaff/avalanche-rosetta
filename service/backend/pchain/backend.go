package pchain

import (
	"context"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ava-labs/avalanche-rosetta/client"
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
	avaxAssetID            ids.ID
	genesisBlock           *indexer.ParsedGenesisBlock
	genesisBlockIdentifier *types.BlockIdentifier
	chainIDs               map[string]string
}

func (b *Backend) getGenesisBlock(ctx context.Context) (*indexer.ParsedGenesisBlock, error) {
	// Initializing parser gives parsed genesis block
	if b.genesisBlock != nil {
		return b.genesisBlock, nil
	}
	genesisBlock, err := b.indexerParser.Initialize(ctx)
	if err != nil {
		return nil, err
	}
	b.genesisBlock = genesisBlock
	b.genesisBlockIdentifier = b.buildGenesisBlockIdentifier(genesisBlock)

	return genesisBlock, nil
}

func (b *Backend) buildGenesisBlockIdentifier(genesisBlock *indexer.ParsedGenesisBlock) *types.BlockIdentifier {
	return &types.BlockIdentifier{
		Index: int64(genesisBlock.Height),
		Hash:  genesisBlock.BlockID.String(),
	}
}

func NewBackend(
	pClient client.PChainClient,
	indexerParser *indexer.Parser,
	assetID ids.ID,
	networkIdentifier *types.NetworkIdentifier,
) *Backend {
	return &Backend{
		fac:               &crypto.FactorySECP256K1R{},
		pClient:           pClient,
		networkIdentifier: networkIdentifier,
		getUTXOsPageSize:  1024,
		codec:             platformvm.Codec,
		codecVersion:      platformvm.CodecVersion,
		avaxAssetID:       assetID,
		indexerParser:     indexerParser,
		chainIDs:          nil,
	}
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
