package address

import (
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewIndexer(elastic *elastic_cache.Index, repo *Repository) *Indexer {
	return &Indexer{elastic, repo}
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
				newBalance := Addresses[addressTx.Hash].ColdBalance + addressTx.Total
				if newBalance > 9223372036854775807 {
					newBalance = 0
				}
				addressTx.Balance = uint64(newBalance)
			} else {
				newBalance := Addresses[addressTx.Hash].Balance + addressTx.Total
				if newBalance > 9223372036854775807 {
					newBalance = 0
				}
				addressTx.Balance = uint64(newBalance)
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
