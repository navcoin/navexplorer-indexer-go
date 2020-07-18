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

func (i *Indexer) GenerateAddressHistory(height uint64, txs []*explorer.BlockTransaction) []*explorer.AddressHistory {
	addresses := getAddressesForTxs(txs)
	addressHistory := make([]*explorer.AddressHistory, 0)

	history, err := i.navcoin.GetAddressHistory(&height, &height, addresses...)
	if err != nil {
		log.WithError(err).Fatalf("Could not get address history for height: %d", height)
	}

	getTxsById := func(txid string) *explorer.BlockTransaction {
		for _, tx := range txs {
			if tx.Txid == txid {
				return tx
			}
		}
		log.Fatal("Could not match tx")
		return nil
	}

	for _, h := range history {
		addressHistory = append(addressHistory, CreateAddressHistory(h, getTxsById(h.TxId)))
	}

	return addressHistory
}

func (i *Indexer) Index(txs []*explorer.BlockTransaction, block *explorer.Block) {
	if len(txs) == 0 {
		return
	}

	for _, addressHistory := range i.GenerateAddressHistory(block.Height, txs) {
		i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)
	}
}

func getAddressesForTxs(txs []*explorer.BlockTransaction) []string {
	aMap := make(map[string]struct{})
	for _, tx := range txs {
		for _, a := range tx.GetAllAddresses() {
			aMap[a] = struct{}{}
		}
	}

	addresses := make([]string, 0)
	for k := range aMap {
		addresses = append(addresses, k)
	}

	return addresses
}
