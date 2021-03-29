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
		orphan, err := i.orphanService.IsOrphanBlock(block, i.service.GetLastBlockIndexed())
		if orphan == true || err != nil {
			log.WithFields(log.Fields{"block": block.Hash}).WithError(err).Info("Orphan Block Found")
			i.service.setLastBlockIndexed(nil)
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

	i.service.setLastBlockIndexed(block)
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
		if previousOutput.IsMultiSig() {
			tx.Vin[vdx].PreviousOutput.Type = explorer.VoutMultiSig
			tx.Vin[vdx].PreviousOutput.MultiSig = previousOutput.MultiSig
		} else {
			tx.Vin[vdx].PreviousOutput.Type = previousOutput.ScriptPubKey.Type
		}
		tx.Vin[vdx].PreviousOutput.Height = prevTx.Height

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
	i.service.GetLastBlockIndexed().Nextblockhash = block.Hash
	i.elastic.AddUpdateRequest(elastic_cache.BlockIndex.Get(), i.service.GetLastBlockIndexed())
}

func (i *Indexer) createBlockTransactions(block *explorer.Block) ([]*explorer.BlockTransaction, error) {
	var txs = make([]*explorer.BlockTransaction, 0)
	for idx, txHash := range block.Tx {
		rawTx, err := i.navcoin.GetRawTransaction(txHash, true)
		if err != nil {
			return nil, err
		}

		tx := CreateBlockTransaction(rawTx.(navcoind.RawTransaction), uint(idx), block)
		i.indexPreviousTxData(tx)

		i.elastic.AddIndexRequest(elastic_cache.BlockTransactionIndex.Get(), tx)

		txs = append(txs, tx)
	}

	return txs, nil
}

func (i *Indexer) updateStakingFees(block *explorer.Block, txs []*explorer.BlockTransaction) {
	for _, tx := range txs {
		if tx.IsAnyStaking() {
			tx.Fees = block.Fees
		}
	}
}

func (i *Indexer) updateSupply(block *explorer.Block, txs []*explorer.BlockTransaction) {
	log.Debugf("Updating Supply for block %d", block.Height)

	for _, tx := range txs {
		log.Debugf("Updating Supply for tx %s", tx.Hash)
		for idx, vin := range tx.Vin {
			log.Debugf("Updating Supply for vin %d", idx)
			if !vin.PreviousOutput.Private && !vin.PreviousOutput.Wrapped {
				block.SupplyBalance.Public -= vin.ValueSat
				log.Debugf("Public decrease by %d", vin.ValueSat)
			}
			if tx.Private {
				block.SupplyBalance.Private += vin.ValueSat
				log.Debugf("Private increase by %d", vin.ValueSat)
			}
			if tx.Wrapped && vin.PreviousOutput.Wrapped {
				block.SupplyBalance.Wrapped -= vin.ValueSat
				log.Debugf("Wrapped decrease by %d", vin.ValueSat)
			}
		}
		for idx, vout := range tx.Vout {
			log.Debugf("Updating Supply for vout %d - %d", idx, vout.N)
			if !vout.Private && !vout.Wrapped {
				block.SupplyBalance.Public += vout.ValueSat
				log.Debugf("Public increase by %d", vout.ValueSat)
			}
			if tx.Private {
				block.SupplyBalance.Private -= vout.ValueSat
				log.Debugf("Private decrease by %d", vout.ValueSat)
				if vout.IsPrivateFee() {
					block.SupplyBalance.Public -= vout.ValueSat
				}
			}
			if tx.Wrapped && vout.Wrapped {
				log.Debugf("Wrapped increase by %d", vout.ValueSat)
				block.SupplyBalance.Wrapped += vout.ValueSat
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
