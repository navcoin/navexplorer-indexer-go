package address

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
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

func (i *Indexer) Index(block *explorer.Block, txs []*explorer.BlockTransaction) {
	if len(txs) == 0 {
		return
	}

	for _, addressHistory := range i.generateAddressHistory(&block.Height, &block.Height, txs) {
		i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

		err := i.updateAddress(addressHistory, block)
		if err != nil {
			log.WithError(err).Fatalf("Could not update address: %s", addressHistory.Hash)
		}
	}
}

func (i *Indexer) generateAddressHistory(start, end *uint64, txs []*explorer.BlockTransaction) []*explorer.AddressHistory {
	addressHistory := make([]*explorer.AddressHistory, 0)

	addresses := getAddressesForTxs(txs)
	history, err := i.navcoin.GetAddressHistory(start, end, addresses...)
	if err != nil {
		log.WithError(err).Errorf("Could not get address history for height: %d-%d", start, end)
		return addressHistory
	}

	for _, h := range history {
		addressHistory = append(addressHistory, CreateAddressHistory(h, getTxById(h.TxId, txs)))
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
	address.Spendable = history.Balance.Spendable
	address.Stakable = history.Balance.Stakable
	address.VotingWeight = history.Balance.VotingWeight

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

func getTxById(id string, txs []*explorer.BlockTransaction) *explorer.BlockTransaction {
	for _, tx := range txs {
		if tx.Txid == id {
			return tx
		}
	}
	log.Fatal("Could not match tx: ", id)
	return nil
}
