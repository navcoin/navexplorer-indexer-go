package block

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/asaskevich/EventBus"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Indexer struct {
	navcoin       *navcoind.Navcoind
	elastic       *elastic_cache.Index
	event         EventBus.Bus
	orphanService *OrphanService
	repository    *Repository
	service       *Service
}

func NewIndexer(
	navcoin *navcoind.Navcoind,
	elastic *elastic_cache.Index,
	event EventBus.Bus,
	orphanService *OrphanService,
	repository *Repository,
	service *Service,
) *Indexer {
	i := &Indexer{navcoin, elastic, event, orphanService, repository, service}
	if err := event.Subscribe("block.indexed", i.OnIndexed); err != nil {
		log.WithError(err).Fatal("Failed to subscribe to block.indexed event")
	}

	return i
}

func (i *Indexer) Index(height uint64, option IndexOption.IndexOption) (*explorer.Block, []*explorer.BlockTransaction, *navcoind.BlockHeader, error) {
	navBlock, err := i.getBlockAtHeight(height)
	if err != nil {
		log.Error("Failed to get block at height ", height)
		return nil, nil, nil, err
	}
	header, err := i.navcoin.GetBlockheader(navBlock.Hash)
	if err != nil {
		log.Error("Failed to get header at height ", height)
		return nil, nil, nil, err
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

	if option == IndexOption.SingleIndex {
		orphan, err := i.orphanService.IsOrphanBlock(block, LastBlockIndexed)
		if orphan == true || err != nil {
			log.WithFields(log.Fields{"block": block.Hash}).WithError(err).Info("Orphan Block Found")
			LastBlockIndexed = nil
			return nil, nil, nil, ErrOrphanBlockFound
		}
	}
	LastBlockIndexed = block

	var txs = make([]*explorer.BlockTransaction, 0)
	for idx, txHash := range block.Tx {
		rawTx, err := i.navcoin.GetRawTransaction(txHash, true)
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithFields(log.Fields{"hash": block.Hash, "txHash": txHash, "height": height}).WithError(err).Error("Failed to GetRawTransaction")
			return nil, nil, nil, err
		}
		tx := CreateBlockTransaction(rawTx.(navcoind.RawTransaction), uint(idx))
		applyType(tx)
		applyWrappedAndPrivateStatus(tx)
		applyStaking(tx, block)
		applySpend(tx, block)
		applyCFundPayout(tx, block)
		applyFees(tx, block)

		i.indexPreviousTxData(tx)
		i.elastic.AddIndexRequest(elastic_cache.BlockTransactionIndex.Get(), tx)

		txs = append(txs, tx)
	}
	for _, tx := range txs {
		if tx.IsAnyStaking() {
			tx.Fees = block.Fees
		}
	}

	if option == IndexOption.SingleIndex {
		i.updateNextHashOfPreviousBlock(block)
	}

	i.elastic.AddIndexRequest(elastic_cache.BlockIndex.Get(), block)

	if option == IndexOption.BatchIndex {
		i.event.Publish("block.indexed", block, txs, header)
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
			log.WithFields(log.Fields{"hash": *tx.Vin[vdx].Txid}).WithError(err).Fatal("Failed to get previous transaction from index")
		}

		previousOutput := prevTx.Vout[*tx.Vin[vdx].Vout]
		tx.Vin[vdx].Value = previousOutput.Value
		tx.Vin[vdx].ValueSat = previousOutput.ValueSat
		tx.Vin[vdx].Addresses = previousOutput.ScriptPubKey.Addresses
		tx.Vin[vdx].PreviousOutput.Type = previousOutput.ScriptPubKey.Type
		tx.Vin[vdx].PreviousOutput.Height = prevTx.Height
		tx.Vin[vdx].Wrapped = previousOutput.IsWrapped()
		if tx.Wrapped == false {
			tx.Wrapped = previousOutput.IsWrapped()
		}
		tx.Vin[vdx].Private = previousOutput.IsPrivate()
		if tx.Private == false {
			tx.Private = previousOutput.IsPrivate()
		}

		prevTx.Vout[*tx.Vin[vdx].Vout].SpentHeight = tx.Height
		prevTx.Vout[*tx.Vin[vdx].Vout].SpentIndex = *tx.Vin[vdx].Vout
		prevTx.Vout[*tx.Vin[vdx].Vout].SpentTxId = tx.Txid

		prevTx.Vout[*tx.Vin[vdx].Vout].RedeemedIn = &explorer.RedeemedIn{
			Hash:   tx.Txid,
			Height: tx.Height,
			Index:  *tx.Vin[vdx].Vout,
		}

		i.elastic.AddUpdateRequest(elastic_cache.BlockTransactionIndex.Get(), prevTx)
	}
}

func (i *Indexer) getBlockAtHeight(height uint64) (*navcoind.Block, error) {
	hash, err := i.navcoin.GetBlockHash(height)
	if err != nil {
		log.WithFields(log.Fields{"hash": hash, "height": height}).WithError(err).Error("Failed to GetBlockHash")
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
