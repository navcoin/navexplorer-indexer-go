package block

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	navcoin       *navcoind.Navcoind
	elastic       *elastic_cache.Index
	orphanService *OrphanService
	repository    *Repository
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, orphanService *OrphanService, repository *Repository) *Indexer {
	return &Indexer{navcoin, elastic, orphanService, repository}
}

func (i *Indexer) Index(height uint64, option IndexOption.IndexOption) (*explorer.Block, []*explorer.BlockTransaction, error) {
	navBlock, err := i.getBlockAtHeight(height)
	if err != nil {
		if err.Error() != "-8: Block height out of range" {
			raven.CaptureError(err, nil)
			log.WithFields(log.Fields{"height": height}).WithError(err).Error("Failed to GetBlockHash")
		}
		return nil, nil, err
	}

	block := CreateBlock(navBlock)
	if option == IndexOption.SingleIndex {
		orphan, err := i.orphanService.IsOrphanBlock(block)
		if orphan == true || err != nil {
			return nil, nil, ErrOrphanBlockFound
		}
	}

	var txs = make([]*explorer.BlockTransaction, 0)
	for idx, txHash := range block.Tx {
		rawTx, err := i.navcoin.GetRawTransaction(txHash, true)
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithFields(log.Fields{"hash": block.Hash, "txHash": txHash, "height": height}).WithError(err).Error("Failed to GetRawTransaction")
			return nil, nil, err
		}
		tx := CreateBlockTransaction(rawTx.(navcoind.RawTransaction), uint(idx))
		applyType(tx)
		applyStaking(tx, block)
		applySpend(tx, block)
		applyCFundPayout(tx, block)
		i.indexPreviousTxData(*tx)

		if option == IndexOption.SingleIndex {
			i.updateNextHashOfPreviousBlock(block)
		}

		txs = append(txs, tx)
		i.elastic.AddIndexRequest(elastic_cache.BlockTransactionIndex.Get(), tx.Hash, tx)
	}

	i.elastic.AddIndexRequest(elastic_cache.BlockIndex.Get(), block.Hash, block)

	return block, txs, err
}

func (i *Indexer) indexPreviousTxData(tx explorer.BlockTransaction) {
	vin := tx.Vin

	for vdx := range vin {
		if vin[vdx].Vout == nil || vin[vdx].Txid == nil {
			continue
		}

		rawTx, err := i.navcoin.GetRawTransaction(*vin[vdx].Txid, true)
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithFields(log.Fields{"hash": *vin[vdx].Txid}).WithError(err).Fatal("Failed to get previous transaction")
		}

		prevTx := CreateBlockTransaction(rawTx.(navcoind.RawTransaction), 0)
		if len(prevTx.Vout) <= *vin[vdx].Vout {
			log.WithFields(log.Fields{"index": vdx, "tx": prevTx.Hash}).Fatal("Vout does not exist")
		}

		previousOutput := prevTx.Vout[*vin[vdx].Vout]
		vin[vdx].Value = previousOutput.Value
		vin[vdx].ValueSat = previousOutput.ValueSat
		vin[vdx].Addresses = previousOutput.ScriptPubKey.Addresses
		vin[vdx].PreviousOutput.Type = previousOutput.ScriptPubKey.Type
		vin[vdx].PreviousOutput.Height = prevTx.Height
	}

	i.elastic.AddIndexRequest(elastic_cache.BlockTransactionIndex.Get(), tx.Hash, tx)
}

func (i *Indexer) getBlockAtHeight(height uint64) (*navcoind.Block, error) {
	hash, err := i.navcoin.GetBlockHash(height)
	if err != nil {
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
		prevBlock.Nextblockhash = block.Hash
		i.elastic.AddUpdateRequest(elastic_cache.BlockIndex.Get(), prevBlock.Hash, prevBlock, prevBlock.MetaData.Id)
	}
}
