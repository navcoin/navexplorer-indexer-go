package block

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const (
	logType = "BlockIndexer"
)

type Indexer interface {
	Index(height uint64, option IndexOption.IndexOption) (*explorer.Block, []explorer.BlockTransaction, *navcoind.BlockHeader, error)
}

type indexer struct {
	navcoin          *navcoind.Navcoind
	elastic          elastic_cache.Index
	orphanService    OrphanService
	repository       Repository
	service          Service
	consensusService consensus.Service
}

func NewIndexer(
	navcoin *navcoind.Navcoind,
	elastic elastic_cache.Index,
	orphanService OrphanService,
	repository Repository,
	service Service,
	consensusService consensus.Service,
) Indexer {
	return indexer{
		navcoin,
		elastic,
		orphanService,
		repository,
		service,
		consensusService,
	}
}

func (i indexer) Index(height uint64, option IndexOption.IndexOption) (*explorer.Block, []explorer.BlockTransaction, *navcoind.BlockHeader, error) {
	navBlock, err := i.getBlockAtHeight(height)
	if err != nil {
		zap.L().With(zap.Error(err), zap.Uint64("height", height)).Error("BlockIndexer: Failed to get block")
		return nil, nil, nil, err
	}
	header, err := i.navcoin.GetBlockheader(navBlock.Hash)
	if err != nil {
		zap.L().With(zap.Error(err), zap.Uint64("height", height)).Error("BlockIndexer: Failed to get block header")
		return nil, nil, nil, err
	}

	var cycleSize uint
	if exists, votingCycleParameter := i.consensusService.GetConsensusParameter(explorer.VOTING_CYCLE_LENGTH); exists {
		cycleSize = uint(votingCycleParameter.Value)
	} else {
		zap.L().Fatal("Consensus Parameter `VOTING_CYCLE_LENGTH` not found")
	}
	block := CreateBlock(navBlock, i.service.GetLastBlockIndexed(), cycleSize)

	available, err := strconv.ParseFloat(header.NcfSupply, 64)
	if err != nil {
		zap.L().With(
			zap.Error(err),
			zap.Uint64("height", height),
			zap.String("ncfSupply", header.NcfSupply),
		).Error("BlockIndexer: Failed to parse header.NcfSupply")
	}

	locked, err := strconv.ParseFloat(header.NcfLocked, 64)
	if err != nil {
		zap.L().With(
			zap.Error(err),
			zap.Uint64("height", height),
			zap.String("ncfLocked", header.NcfLocked),
		).Error("BlockIndexer: Failed to parse header.NcfLocked")
	}

	block.Cfund = explorer.Cfund{Available: available, Locked: locked}

	if option == IndexOption.SingleIndex {
		orphan, err := i.orphanService.IsOrphanBlock(block, i.service.GetLastBlockIndexed())
		if orphan == true || err != nil {
			zap.L().With(zap.Error(err), zap.Uint64("height", height)).Error("BlockIndexer: Orphan Block Found")
			i.service.ClearLastBlockIndexed()
			return nil, nil, nil, ErrOrphanBlockFound
		}
	}

	txs, err := i.createBlockTransactions(block)
	if err != nil {
		return nil, nil, nil, err
	}

	i.updateStakingFees(block, txs)
	i.updateSupply(block, txs)

	if option == IndexOption.SingleIndex {
		i.updateNextHashOfPreviousBlock(block)
	}

	i.service.SetLastBlockIndexed(block)
	i.elastic.AddIndexRequest(elastic_cache.BlockIndex.Get(), block)

	return block, txs, header, err
}

func (i indexer) indexPreviousTxData(tx explorer.BlockTransaction) explorer.BlockTransaction {
	if tx.IsCoinbase() {
		return tx
	}

	for vdx := range tx.Vin {
		if tx.Vin[vdx].Vout == nil || tx.Vin[vdx].Txid == nil {
			continue
		}

		prevTx, err := i.repository.GetTransactionByHash(*tx.Vin[vdx].Txid)
		if err != nil {
			zap.L().With(
				zap.Error(err),
				zap.Uint64("height", tx.Height),
				zap.String("txid", tx.Txid),
				zap.String("previousHash", *tx.Vin[vdx].Txid),
			).Error("BlockIndexer: Failed to get previous transaction from index")
			if err != nil {
				i.elastic.Persist()
				zap.L().Info("BlockIndexer: Retry get previous transaction in 5 seconds")
				time.Sleep(5 * time.Second)
				prevTx, err = i.repository.GetTransactionByHash(*tx.Vin[vdx].Txid)
				if err != nil {
					zap.L().With(
						zap.Error(err),
						zap.String("hash", *tx.Vin[vdx].Txid),
					).Fatal("BlockIndexer: Failed to get previous transaction from index")
				}
			}
		}
		tx.Vin[vdx].PreviousOutput = &explorer.PreviousOutput{
			Height: prevTx.Height,
		}

		previousOutput := prevTx.Vout[*tx.Vin[vdx].Vout]
		tx.Vin[vdx].Value = previousOutput.Value
		tx.Vin[vdx].ValueSat = previousOutput.ValueSat
		tx.Vin[vdx].Addresses = previousOutput.ScriptPubKey.Addresses
		if previousOutput.IsMultiSig() {
			tx.Vin[vdx].PreviousOutput.Type = explorer.VoutMultiSig
			tx.Vin[vdx].PreviousOutput.MultiSig = previousOutput.MultiSig
		} else {
			tx.Vin[vdx].PreviousOutput.Type = previousOutput.ScriptPubKey.Type
		}

		if previousOutput.Wrapped {
			tx.Vin[vdx].PreviousOutput.Wrapped = true
			tx.Wrapped = true
		}

		if previousOutput.Private {
			tx.Vin[vdx].PreviousOutput.Private = true
			tx.Private = true
		}

		prevTx.Vout[*tx.Vin[vdx].Vout].Redeemed = true
		prevTx.Vout[*tx.Vin[vdx].Vout].RedeemedIn = &explorer.RedeemedIn{
			Hash:   tx.Txid,
			Height: tx.Height,
			Index:  *tx.Vin[vdx].Vout,
		}

		i.elastic.AddUpdateRequest(elastic_cache.BlockTransactionIndex.Get(), *prevTx)
	}

	return tx
}

func (i indexer) getBlockAtHeight(height uint64) (*navcoind.Block, error) {
	hash, err := i.navcoin.GetBlockHash(height)
	if err != nil {
		zap.L().With(zap.Error(err), zap.Uint64("height", height)).
			Error("BlockIndexer: Failed to get block hash")
		return nil, err
	}

	block, err := i.navcoin.GetBlock(hash)
	if err != nil {
		zap.L().With(zap.Error(err), zap.Uint64("height", height), zap.String("hash", hash)).
			Error("BlockIndexer: Failed to get block")
		return nil, err
	}

	return &block, nil
}

func (i indexer) updateNextHashOfPreviousBlock(block *explorer.Block) {
	i.service.GetLastBlockIndexed().Nextblockhash = block.Hash
	i.elastic.AddUpdateRequest(elastic_cache.BlockIndex.Get(), i.service.GetLastBlockIndexed())
}

func (i indexer) createBlockTransactions(block *explorer.Block) ([]explorer.BlockTransaction, error) {
	var txs = make([]explorer.BlockTransaction, 0)
	for idx, txHash := range block.Tx {

		start := time.Now()
		rawTx, err := i.navcoin.GetRawTransaction(txHash, true)
		if err != nil {
			return nil, err
		}

		tx := CreateBlockTransaction(rawTx.(navcoind.RawTransaction), uint(idx), block)
		tx = i.indexPreviousTxData(tx)

		i.elastic.AddIndexRequest(elastic_cache.BlockTransactionIndex.Get(), tx)

		zap.L().With(
			zap.Duration("elapsed", time.Since(start)),
			zap.String("txid", tx.Txid),
		).Debug("Index block tx")

		txs = append(txs, tx)
	}

	return txs, nil
}

func (i indexer) updateStakingFees(block *explorer.Block, txs []explorer.BlockTransaction) {
	for _, tx := range txs {
		if tx.IsAnyStaking() {
			tx.Fees = block.Fees
		}
	}
}

func (i indexer) updateSupply(block *explorer.Block, txs []explorer.BlockTransaction) {
	zap.L().With(zap.Uint64("height", block.Height)).Debug("BlockIndexer: Updating Supply")

	if block.Height == 1 {
		for _, tx := range txs {
			for _, vout := range tx.Vout {
				block.SupplyBalance.Public += vout.ValueSat
			}
		}
	} else {
		for _, tx := range txs {
			zap.L().With(zap.Uint64("height", block.Height), zap.String("hash", tx.Hash)).
				Debug("BlockIndexer: Updating Supply for TX")
			for _, vin := range tx.Vin {
				if vin.IsCoinbase() {
					continue
				}

				if !vin.PreviousOutput.Private && !vin.PreviousOutput.Wrapped {
					block.SupplyBalance.Public -= vin.ValueSat
				}
				if tx.Private {
					block.SupplyBalance.Private += vin.ValueSat
				}
				if tx.Wrapped && vin.PreviousOutput.Wrapped {
					block.SupplyBalance.Wrapped -= vin.ValueSat
				}
			}
			for _, vout := range tx.Vout {
				if vout.ScriptPubKey.Asm == "OP_RETURN OP_CFUND" {
					continue
				}
				if !vout.Private && !vout.Wrapped {
					block.SupplyBalance.Public += vout.ValueSat
				}
				if tx.Private {
					block.SupplyBalance.Private -= vout.ValueSat
					if vout.IsPrivateFee() {
						block.SupplyBalance.Public -= vout.ValueSat
					}
				}
				if tx.Wrapped && vout.Wrapped {
					block.SupplyBalance.Wrapped += vout.ValueSat
				}
			}
		}
	}

	if lastBlockIndexed := i.service.GetLastBlockIndexed(); lastBlockIndexed != nil {
		block.SupplyChange.Public = int64(block.SupplyBalance.Public) - int64(lastBlockIndexed.SupplyBalance.Public)
		block.SupplyChange.Private = int64(block.SupplyBalance.Private) - int64(lastBlockIndexed.SupplyBalance.Private)
		block.SupplyChange.Wrapped = int64(block.SupplyBalance.Wrapped) - int64(lastBlockIndexed.SupplyBalance.Wrapped)
	} else {
		block.SupplyChange.Public = int64(block.SupplyBalance.Public)
		block.SupplyChange.Private = int64(block.SupplyBalance.Private)
		block.SupplyChange.Wrapped = int64(block.SupplyBalance.Wrapped)
	}
}
