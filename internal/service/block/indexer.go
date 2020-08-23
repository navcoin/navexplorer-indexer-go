package block

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Indexer struct {
	navcoin       *navcoind.Navcoind
	elastic       *elastic_cache.Index
	orphanService *OrphanService
	repository    *Repository
	service       *Service
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, orphanService *OrphanService, repository *Repository, service *Service) *Indexer {
	return &Indexer{navcoin, elastic, orphanService, repository, service}
}

func (i *Indexer) Index(height uint64, option IndexOption.IndexOption) (*explorer.Block, []*explorer.BlockTransaction, *navcoind.BlockHeader, error) {
	start := time.Now()

	navBlock, err := i.getBlockAtHeight(height)
	if err != nil {
		log.Fatal("Failed to get block at height ", height)
	}
	header, err := i.navcoin.GetBlockheader(navBlock.Hash)
	if err != nil {
		log.Fatal("Failed to get block at height ", height)
	}

	block := CreateBlock(navBlock, i.service.GetLastBlockIndexed(), uint(consensus.Parameters.Get(consensus.VOTING_CYCLE_LENGTH).Value))

	available, err := strconv.ParseFloat(header.NcfSupply, 64)
	if err != nil {
		log.WithError(err).Errorf("Failed to parse header.NcfSupply: %s", header.NcfSupply)
	}

	locked, err := strconv.ParseFloat(header.NcfLocked, 64)
	if err != nil {
		log.WithError(err).Errorf("Failed to parse header.NcfLocked: %s", header.NcfLocked)
	}

	block.Cfund = &explorer.Cfund{Available: available, Locked: locked}

	LastBlockIndexed = block

	if option == IndexOption.SingleIndex {
		log.Info("Indexing in single block mode")
		orphan, err := i.orphanService.IsOrphanBlock(block)
		if orphan == true || err != nil {
			log.WithFields(log.Fields{"block": block, "orphan": orphan}).WithError(err).Info("Orphan Block Found")

			return nil, nil, nil, ErrOrphanBlockFound
		}
	}
	i.elastic.AddIndexRequest(elastic_cache.BlockIndex.Get(), block)
	elapsed := time.Since(start)
	log.WithField("time", elapsed).Infof("Indexed block     at height %d", height)

	start = time.Now()
	var txs = make([]*explorer.BlockTransaction, 0)

	var wg sync.WaitGroup
	wg.Add(len(block.Tx))

	for idx, txHash := range block.Tx {
		go func(index int, hash string) {
			defer wg.Done()
			rawTx, err := i.navcoin.GetRawTransaction(hash, true)
			if err != nil {
				log.WithFields(log.Fields{"hash": block.Hash, "txHash": hash, "height": height}).WithError(err).
					Fatal("Failed to GetRawTransaction")
			}

			tx := CreateBlockTransaction(rawTx.(navcoind.RawTransaction), uint(index))
			applyType(tx)
			applyStaking(tx, block)
			applySpend(tx, block)
			applyCFundPayout(tx, block)

			i.indexPreviousTxData(tx)
			i.elastic.AddIndexRequest(elastic_cache.BlockTransactionIndex.Get(), tx)

			txs = append(txs, tx)
		}(idx, txHash)
	}

	wg.Wait()

	elapsed = time.Since(start)
	log.WithField("time", elapsed).Infof("Indexed %d txs     at height %d", len(block.Tx), height)

	if option == IndexOption.SingleIndex {
		i.updateNextHashOfPreviousBlock(block)
	}

	return block, txs, header, err
}

func (i *Indexer) indexPreviousTxData(tx *explorer.BlockTransaction) {
	for vdx := range tx.Vin {
		if tx.Vin[vdx].Vout == nil || tx.Vin[vdx].Txid == nil {
			continue
		}

		prevTx, err := i.repository.GetTransactionByHash(*tx.Vin[vdx].Txid)
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithFields(log.Fields{"hash": *tx.Vin[vdx].Txid}).WithError(err).Fatal("Failed to get previous transaction from index")
		}

		previousOutput := prevTx.Vout[*tx.Vin[vdx].Vout]
		tx.Vin[vdx].Value = previousOutput.Value
		tx.Vin[vdx].ValueSat = previousOutput.ValueSat
		tx.Vin[vdx].Addresses = previousOutput.ScriptPubKey.Addresses
		tx.Vin[vdx].PreviousOutput.Type = previousOutput.ScriptPubKey.Type
		tx.Vin[vdx].PreviousOutput.Height = prevTx.Height

		prevTx.Vout[*tx.Vin[vdx].Vout].RedeemedIn = &explorer.RedeemedIn{
			Hash:   tx.Txid,
			Height: tx.Height,
		}

		i.elastic.AddUpdateRequest(elastic_cache.BlockTransactionIndex.Get(), prevTx)
	}
}

func (i *Indexer) getBlockAtHeight(height uint64) (*navcoind.Block, error) {
	hash, err := i.navcoin.GetBlockHash(height)
	if err != nil {
		raven.CaptureError(err, nil)
		if err.Error() != "-8: Block height out of range" {
			log.WithFields(log.Fields{"hash": hash, "height": height}).WithError(err).Error("Failed to GetBlockHash")
		}
		return nil, err
	}

	block, err := i.navcoin.GetBlock(hash)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithFields(log.Fields{"hash": hash, "height": height}).WithError(err).Error("Failed to GetBlock")
		return nil, err
	}

	return &block, nil
}

func (i *Indexer) updateNextHashOfPreviousBlock(block *explorer.Block) {
	if prevBlock, err := i.repository.GetBlockByHeight(block.Height - 1); err == nil {
		log.Debugf("Update NextHash of PreviousBlock: %s", block.Hash)
		prevBlock.Nextblockhash = block.Hash
		i.elastic.AddUpdateRequest(elastic_cache.BlockIndex.Get(), prevBlock)
	}
}
