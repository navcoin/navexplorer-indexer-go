package address

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository) *Indexer {
	return &Indexer{navcoin, elastic, repo}
}

func (i *Indexer) Index(txs []*explorer.BlockTransaction, block *explorer.Block) {
	if len(txs) == 0 {
		return
	}

	for _, addressHistory := range i.generateAddressHistory(block, txs) {
		i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

		err := i.updateAddress(addressHistory, block)
		if err != nil {
			log.WithError(err).Fatalf("Could not update address: %s", addressHistory.Hash)
		}
	}
}

func (i *Indexer) generateAddressHistory(block *explorer.Block, txs []*explorer.BlockTransaction) []*explorer.AddressHistory {
	addresses := getAddressesForTxs(txs)
	history, err := i.navcoin.GetAddressHistory(&block.Height, &block.Height, addresses...)
	if err != nil {
		log.WithError(err).Fatalf("Could not get address history for height: %d", block.Height)
	}

	addressHistory := make([]*explorer.AddressHistory, 0)
	for _, h := range history {
		addressHistory = append(addressHistory, CreateAddressHistory(h, getTxsById(h.TxId, txs), block))
	}

	return addressHistory
}

func (i *Indexer) updateAddress(history *explorer.AddressHistory, block *explorer.Block) error {
	address := Addresses.GetByHash(history.Hash)
	if address == nil {
		var err error
		address, err = i.repo.GetOrCreateAddress(history.Hash, block)
		if err != nil {
			return err
		}
	}

	address.Height = history.Height
	address.Spending = history.Balance.Spending
	address.Staking = history.Balance.Staking
	address.Voting = history.Balance.Voting

	Addresses[address.Hash] = address

	i.elastic.AddUpdateRequest(elastic_cache.AddressIndex.Get(), address)
	return nil
}

func getAddressesForTxs(txs []*explorer.BlockTransaction) []string {
	addresses := make([]string, 0)
	for _, tx := range txs {
		for _, address := range tx.GetAllAddresses() {
			addresses = append(addresses, address)
		}
	}

	return addresses
}

func getTxsById(txid string, txs []*explorer.BlockTransaction) *explorer.BlockTransaction {
	for _, tx := range txs {
		if tx.Txid == txid {
			return tx
		}
	}
	log.Fatal("Could not match tx")
	return nil
}
