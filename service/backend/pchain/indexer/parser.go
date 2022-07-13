package indexer

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/genesis"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/stakeable"
	"github.com/ava-labs/avalanchego/vms/proposervm/block"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"

	"github.com/ava-labs/avalanche-rosetta/client"
	"github.com/ava-labs/avalanche-rosetta/mapper"
	"github.com/ava-labs/avalanche-rosetta/service/backend/common"
)

type Parser struct {
	networkID   uint32
	avaxAssetID ids.ID

	codec codec.Manager

	time ChainTime

	ctx *snow.Context

	pChainClient client.PChainClient
}

func NewParser(ctx context.Context, pChainClient client.PChainClient) (*Parser, error) {
	errs := wrappers.Errs{}

	networkID, err := pChainClient.GetNetworkID(ctx)
	errs.Add(err)

	aliaser := ids.NewAliaser()
	errs.Add(aliaser.Alias(constants.PlatformChainID, mapper.PChainNetworkIdentifier))

	return &Parser{
		networkID:    networkID,
		codec:        platformvm.Codec,
		pChainClient: pChainClient,
		ctx: &snow.Context{
			BCLookup:  aliaser,
			NetworkID: networkID,
		},
	}, errs.Err
}

func (p *Parser) GetPlatformHeight(ctx context.Context) (uint64, error) {
	return p.pChainClient.GetHeight(ctx)
}

func (p *Parser) readTime() int64 {
	return p.time.Read()
}

func (p *Parser) writeTime(t time.Time) {
	p.time.Write(t)
}

func (p *Parser) formatAddress(addr []byte) (string, error) {
	return address.Format("P", constants.GetHRP(p.networkID), addr)
}

func (p *Parser) extractCredData(creds []verify.Verifiable, bytes []byte) ([][]CredData, error) {
	credData := [][]CredData{}
	errs := wrappers.Errs{}
	ecdsaRecoveryFactory := crypto.FactorySECP256K1R{}

	for _, cred := range creds {
		switch castCred := cred.(type) {
		case *secp256k1fx.Credential:
			data := []CredData{}

			for _, sig := range castCred.Sigs {
				key, err := ecdsaRecoveryFactory.RecoverPublicKey(bytes, sig[:])
				errs.Add(err)
				addr, err := p.formatAddress(key.Address().Bytes())
				errs.Add(err)
				data = append(data, CredData{
					Address:   addr,
					PublicKey: base64.RawStdEncoding.EncodeToString(key.Bytes()),
					Signature: base64.RawStdEncoding.EncodeToString(sig[:]),
				})
			}

			credData = append(credData, data)
		default:
			errs.Add(fmt.Errorf("unexpected cred type found: %T", castCred))
		}
	}

	return credData, errs.Err
}

func (p *Parser) Initialize(ctx context.Context) (*ParsedGenesisBlock, error) {
	errs := wrappers.Errs{}

	bytes, avaxAssetID, err := genesis.FromConfig(genesis.GetConfig(p.networkID))
	errs.Add(err)
	p.avaxAssetID = avaxAssetID

	genesis := &platformvm.Genesis{}
	_, err = platformvm.GenesisCodec.Unmarshal(bytes, genesis)
	errs.Add(err)
	errs.Add(genesis.Initialize())
	p.writeTime(time.Unix(int64(genesis.Timestamp), 0))

	txs := make([]interface{}, len(genesis.Validators)+len(genesis.Chains))

	for i, tx := range append(genesis.Validators, genesis.Chains...) {
		parsedTx, err := p.parseTx(ctx, ids.Empty, *tx, true, 0)
		errs.Add(err)
		txs[i] = parsedTx
	}

	for _, utxo := range genesis.UTXOs {
		utxo.UTXO.Out.InitCtx(p.ctx)
	}

	return &ParsedGenesisBlock{
		ParsedBlock: ParsedBlock{
			ParentID:  ids.Empty,
			Height:    0,
			BlockID:   ids.Empty,
			BlockType: "GenesisBlock",
			Timestamp: p.readTime(),
			Txs:       txs,
			Proposer:  Proposer{},
		},
		GenesisBlockData: GenesisBlockData{
			Message:       genesis.Message,
			InitialSupply: genesis.InitialSupply,
			UTXOs:         genesis.UTXOs,
		},
	}, errs.Err
}

func (p *Parser) ParseCurrentBlock(ctx context.Context) (*ParsedBlock, error) {
	height, err := p.GetPlatformHeight(ctx)
	if err != nil {
		return nil, err
	}

	return p.ParseBlockAtIndex(ctx, height)
}

func (p *Parser) ParseBlockAtIndex(ctx context.Context, index uint64) (*ParsedBlock, error) {
	container, err := p.pChainClient.GetContainerByIndex(ctx, index-1)
	if err != nil {
		return nil, err
	}

	return p.parseBlockBytes(ctx, container.Bytes)
}

func (p *Parser) parseBlockBytes(ctx context.Context, proposerBytes []byte) (*ParsedBlock, error) {
	errs := wrappers.Errs{}

	proposer, bytes, err := getProposerFromBytes(proposerBytes)
	if err != nil {
		return nil, fmt.Errorf("fetching proposer from block bytes errored with %w", err)
	}

	var blk platformvm.Block
	ver, err := p.codec.Unmarshal(bytes, &blk)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling block bytes errored with %w", err)
	}
	blkID := ids.ID(hashing.ComputeHash256Array(bytes))

	parsedBlock := ParsedBlock{
		Height:    blk.Height(),
		BlockID:   blkID,
		BlockType: fmt.Sprintf("%T", blk),
		Timestamp: p.readTime(),
		Proposer:  proposer,
	}

	switch castBlk := blk.(type) {
	case *platformvm.ProposalBlock:
		errs.Add(common.InitializeTx(ver, p.codec, castBlk.Tx))
		tx, err := p.parseTx(ctx, blkID, castBlk.Tx, false, castBlk.Hght)
		errs.Add(err)

		parsedBlock.ParentID = castBlk.PrntID
		parsedBlock.Txs = []interface{}{tx}
	case *platformvm.AtomicBlock:
		errs.Add(common.InitializeTx(ver, p.codec, castBlk.Tx))
		tx, err := p.parseTx(ctx, blkID, castBlk.Tx, false, castBlk.Hght)
		errs.Add(err)

		parsedBlock.ParentID = castBlk.PrntID
		parsedBlock.Txs = []interface{}{tx}
	case *platformvm.StandardBlock:
		var txs []interface{}

		for _, tx := range castBlk.Txs {
			errs.Add(common.InitializeTx(ver, p.codec, *tx))
			parsedTx, err := p.parseTx(ctx, blkID, *tx, false, castBlk.Hght)
			errs.Add(err)

			txs = append(txs, parsedTx)
		}

		parsedBlock.ParentID = castBlk.PrntID
		parsedBlock.Txs = txs
	case *platformvm.AbortBlock:
		p.time.RejectProposedWrite()

		parsedBlock.ParentID = castBlk.PrntID
		parsedBlock.Txs = []interface{}{}
	case *platformvm.CommitBlock:
		p.time.AcceptProposedWrite()

		parsedBlock.ParentID = castBlk.PrntID
		parsedBlock.Txs = []interface{}{}
	default:
		errs.Add(fmt.Errorf("no handler exists for block type %T", castBlk))
	}

	return &parsedBlock, errs.Err
}

func getProposerFromBytes(bytes []byte) (Proposer, []byte, error) {
	proposer, err := block.Parse(bytes)
	if err != nil || proposer == nil {
		return Proposer{}, bytes, nil
	}

	switch castBlock := proposer.(type) {
	case block.SignedBlock:
		return Proposer{
			ID:           castBlock.ID(),
			NodeID:       castBlock.Proposer(),
			PChainHeight: castBlock.PChainHeight(),
			Timestamp:    castBlock.Timestamp().Unix(),
			ParentID:     castBlock.ParentID(),
		}, castBlock.Block(), nil
	case block.Block:
		return Proposer{}, castBlock.Block(), nil
	default:
		return Proposer{}, bytes, fmt.Errorf("no handler exists for proposer block type %T", castBlock)
	}
}

func sumIns(utxos []*avax.TransferableInput) uint64 {
	amt := uint64(0)

	for _, utxo := range utxos {
		amt += utxo.In.Amount()
	}

	return amt
}

func sumOuts(utxos []*avax.TransferableOutput) uint64 {
	amt := uint64(0)

	for _, utxo := range utxos {
		amt += utxo.Out.Amount()
	}

	return amt
}

func standardizeIns(utxos []*avax.TransferableInput) error {
	for _, u := range utxos {
		switch castIn := u.In.(type) {
		case *stakeable.LockIn:
		case *secp256k1fx.TransferInput:
			u.In = &stakeable.LockIn{
				Locktime:       0,
				TransferableIn: castIn,
			}
		default:
			return fmt.Errorf("no handler exists for utxo out type %T", castIn)
		}
	}

	return nil
}

func standardizeOuts(utxos []*avax.TransferableOutput) error {
	for _, u := range utxos {
		switch castOut := u.Out.(type) {
		case *stakeable.LockOut:
		case *secp256k1fx.TransferOutput:
			u.Out = &stakeable.LockOut{
				Locktime:        0,
				TransferableOut: castOut,
			}
		default:
			return fmt.Errorf("no handler exists for utxo out type %T", castOut)
		}
	}

	return nil
}

func (p *Parser) parseTx(ctx context.Context, blkID ids.ID, tx platformvm.Tx, genesis bool, height uint64) (interface{}, error) {
	errs := wrappers.Errs{}

	tx.InitCtx(p.ctx)
	tx.UnsignedTx.InitCtx(p.ctx)

	creds, err := p.extractCredData(tx.Creds, tx.UnsignedBytes())
	errs.Add(err)

	parsedTx := ParsedTx{
		TxType:    fmt.Sprintf("%T", tx.UnsignedTx),
		BlockID:   blkID,
		Timestamp: p.readTime(),
		Creds:     creds,
		TxID:      tx.ID(),
	}

	switch castTx := tx.UnsignedTx.(type) {
	case *platformvm.UnsignedAddValidatorTx:
		if !genesis {
			parsedTx.Fee = sumIns(castTx.Ins) - sumOuts(append(castTx.Outs, castTx.Stake...))
		}
		castTx.BaseTx.InitCtx(p.ctx)

		errs.Add(standardizeIns(castTx.Ins))
		errs.Add(standardizeOuts(castTx.Outs))
		errs.Add(standardizeOuts(castTx.Stake))

		return &ParsedAddValidatorTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				Outs:   castTx.UTXOs(),
				BaseTx: castTx.BaseTx.BaseTx,
			},
			AddValidatorData: AddValidatorData{
				AddDelegatorData: AddDelegatorData{
					Validator:    castTx.Validator,
					RewardsOwner: castTx.RewardsOwner,
					Stake:        castTx.Stake,
				},
				Shares: castTx.Shares,
			},
		}, errs.Err
	case *platformvm.UnsignedAddDelegatorTx:
		parsedTx.Fee = sumIns(castTx.Ins) - sumOuts(append(castTx.Outs, castTx.Stake...))
		castTx.BaseTx.InitCtx(p.ctx)

		errs.Add(standardizeIns(castTx.Ins))
		errs.Add(standardizeOuts(castTx.Outs))
		errs.Add(standardizeOuts(castTx.Stake))

		return &ParsedAddDelegatorTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				Outs:   castTx.UTXOs(),
				BaseTx: castTx.BaseTx.BaseTx,
			},
			AddDelegatorData: AddDelegatorData{
				Validator:    castTx.Validator,
				RewardsOwner: castTx.RewardsOwner,
				Stake:        castTx.Stake,
			},
		}, errs.Err
	case *platformvm.UnsignedAdvanceTimeTx:
		p.time.ProposeWrite(castTx.Timestamp())
		parsedTx.Fee = uint64(0)

		return &ParsedAdvanceTimeTx{
			ParsedTx: parsedTx,
			AdvanceTimeData: AdvanceTimeData{
				Time: castTx.Timestamp().Unix(),
			},
		}, nil
	case *platformvm.UnsignedImportTx:
		parsedTx.Fee = sumIns(append(castTx.Ins, castTx.ImportedInputs...)) - sumOuts(castTx.Outs)
		castTx.BaseTx.InitCtx(p.ctx)

		errs.Add(standardizeIns(castTx.Ins))
		errs.Add(standardizeIns(castTx.ImportedInputs))
		errs.Add(standardizeOuts(castTx.Outs))

		return &ParsedImportTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				Outs:   castTx.UTXOs(),
				BaseTx: castTx.BaseTx.BaseTx,
			},
			ImportData: ImportData{
				SourceChain:    castTx.SourceChain,
				ImportedInputs: castTx.ImportedInputs,
			},
		}, errs.Err
	case *platformvm.UnsignedExportTx:
		parsedTx.Fee = sumIns(castTx.Ins) - sumOuts(append(castTx.Outs, castTx.ExportedOutputs...))
		castTx.BaseTx.InitCtx(p.ctx)

		errs.Add(standardizeIns(castTx.Ins))
		errs.Add(standardizeOuts(castTx.ExportedOutputs))
		errs.Add(standardizeOuts(castTx.Outs))

		return &ParsedExportTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				Outs:   castTx.UTXOs(),
				BaseTx: castTx.BaseTx.BaseTx,
			},
			ExportData: ExportData{
				DestinationChain: castTx.DestinationChain,
				ExportedOutputs:  castTx.ExportedOutputs,
			},
		}, errs.Err
	case *platformvm.UnsignedCreateSubnetTx:
		parsedTx.Fee = sumIns(castTx.Ins) - sumOuts(castTx.Outs)
		castTx.BaseTx.InitCtx(p.ctx)

		errs.Add(standardizeIns(castTx.Ins))
		errs.Add(standardizeOuts(castTx.Outs))

		return &ParsedCreateSubnetTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				Outs:   castTx.UTXOs(),
				BaseTx: castTx.BaseTx.BaseTx,
			},
			CreateSubnetData: CreateSubnetData{
				Owner: castTx.Owner,
			},
		}, errs.Err
	case *platformvm.UnsignedRewardValidatorTx:
		outs := []*avax.UTXO{}

		pChainHeight, err := p.GetPlatformHeight(ctx)
		if err != nil {
			errs.Add(err)
			return nil, errs.Err
		}

		if height+1 > pChainHeight {
			errs.Add(fmt.Errorf("reward UTXOs are not ready yet, waiting for height %d", height+1))
			return nil, errs.Err
		}

		rewardUTXOs, err := p.pChainClient.GetRewardUTXOs(ctx, &api.GetTxArgs{
			TxID:     castTx.TxID,
			Encoding: formatting.Hex,
		})
		errs.Add(err)

		for _, bytes := range rewardUTXOs {
			var utxo avax.UTXO
			_, err = p.codec.Unmarshal(bytes, &utxo)
			errs.Add(err)
			utxo.Out.InitCtx(p.ctx)
			outs = append(outs, &utxo)
		}

		parsedTx.Fee = uint64(0)

		return &ParsedRewardValidatorTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				BaseTx: avax.BaseTx{
					// [RewardValidatorTxs] do not have any inputs
					Ins: []*avax.TransferableInput{},
				},
				Outs: outs,
			},
			RewardValidatorData: RewardValidatorData{
				TxID: castTx.TxID,
			},
		}, errs.Err
	case *platformvm.UnsignedCreateChainTx:
		parsedTx.Fee = sumIns(castTx.Ins) - sumOuts(castTx.Outs)
		castTx.BaseTx.InitCtx(p.ctx)

		errs.Add(standardizeIns(castTx.Ins))
		errs.Add(standardizeOuts(castTx.Outs))

		return &ParsedCreateChainTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				Outs:   castTx.UTXOs(),
				BaseTx: castTx.BaseTx.BaseTx,
			},
			CreateChainData: CreateChainData{
				SubnetID:    castTx.SubnetID,
				ChainName:   castTx.ChainName,
				VMID:        castTx.VMID,
				FxIDs:       castTx.FxIDs,
				GenesisData: castTx.GenesisData,
				SubnetAuth:  castTx.SubnetAuth,
			},
		}, errs.Err
	case *platformvm.UnsignedAddSubnetValidatorTx:
		parsedTx.Fee = sumIns(castTx.Ins) - sumOuts(castTx.Outs)
		castTx.BaseTx.InitCtx(p.ctx)

		errs.Add(standardizeIns(castTx.Ins))
		errs.Add(standardizeOuts(castTx.Outs))

		return &ParsedAddSubnetValidatorTx{
			ParsedTx: parsedTx,
			BaseTxData: BaseTxData{
				Outs:   castTx.UTXOs(),
				BaseTx: castTx.BaseTx.BaseTx,
			},
			AddSubnetValidatorData: AddSubnetValidatorData{
				Validator:  castTx.Validator,
				SubnetAuth: castTx.SubnetAuth,
			},
		}, errs.Err
	default:
		errs.Add(fmt.Errorf("no handler exists for tx type %T", castTx))
	}

	return nil, errs.Err
}
