package address

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
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

func (i *Indexer) GenerateAddressHistory(height uint64, addresses []string) []*explorer.AddressHistory {
	addressHistory := make([]*explorer.AddressHistory, 0)

	history, err := i.navcoin.GetAddressHistory(&height, &height, addresses...)
	if err != nil {
		log.WithError(err).Fatalf("Could not get address history for height: %d", height)
	}
	for _, h := range history {
		addressHistory = append(addressHistory, CreateAddressHistory(h))
	}

	return addressHistory
}

func (i *Indexer) GenerateAddressTransactions(address *explorer.Address, tx *explorer.BlockTransaction, block *explorer.Block) []*explorer.AddressTransaction {
	txs := make([]*explorer.AddressTransaction, 0)
	for _, tx := range CreateAddressTransaction(tx, block) {
		if tx.Hash != address.Hash {
			continue
		}
		if tx.Cold == true {
			address.ColdBalance += tx.Total
			tx.Balance = uint64(address.ColdBalance)
		} else {
			address.Balance += tx.Total
			tx.Balance = uint64(address.Balance)
		}
		txs = append(txs, tx)
	}
	return txs
}

func (i *Indexer) Index(txs []*explorer.BlockTransaction, block *explorer.Block) {
	if len(txs) == 0 {
		return
	}

	for _, addressHistory := range i.GenerateAddressHistory(block.Height, getAddressesForTxs(txs)) {
		i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)
	}

	hashes := make([]string, 0)
	for _, tx := range txs {
		for _, addressTx := range CreateAddressTransaction(tx, block) {
			if Addresses.GetByHash(addressTx.Hash) == nil {
				address, err := i.repo.GetOrCreateAddress(addressTx.Hash)
				if err != nil {
					raven.CaptureError(err, nil)
					log.WithError(err).Fatalf("Could not get or create address: %s", addressTx.Hash)
				}
				Addresses[addressTx.Hash] = address
			}

			if addressTx.Cold == true {
				addressTx.Balance = uint64(Addresses[addressTx.Hash].ColdBalance + addressTx.Total)
			} else {
				addressTx.Balance = uint64(Addresses[addressTx.Hash].Balance + addressTx.Total)
			}

			i.elastic.AddIndexRequest(elastic_cache.AddressTransactionIndex.Get(), addressTx)

			ApplyTxToAddress(Addresses[addressTx.Hash], addressTx)
			if addressTx.Cold == true {
				Addresses[addressTx.Hash].ColdBalance += addressTx.Total
			} else {
				Addresses[addressTx.Hash].Balance += addressTx.Total
			}

			i.elastic.AddUpdateRequest(elastic_cache.AddressIndex.Get(), Addresses[addressTx.Hash])
		}
	}

	if len(hashes) > 0 {
		if _, err := i.repo.GetAddresses(hashes); err != nil {
			raven.CaptureError(err, nil)
			log.WithError(err).Fatalf("Could not get addresses for txs at height %d", txs[0].Height)
		}
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
