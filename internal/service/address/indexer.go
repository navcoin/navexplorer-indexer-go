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

var (
	addresses = make(map[string]*explorer.Address)
)

func NewIndexer(elastic *elastic_cache.Index, repo *Repository) *Indexer {
	return &Indexer{elastic, repo}
}

func (i *Indexer) Index(txs []*explorer.BlockTransaction, block *explorer.Block) {
	if len(txs) == 0 {
		return
	}

	hashes := make([]string, 0)
	for _, tx := range txs {
		for _, addressTx := range CreateAddressTransaction(tx, block) {
			if addresses[addressTx.Hash] == nil {
				address, err := i.repo.GetOrCreateAddress(addressTx.Hash)
				if err != nil {
					raven.CaptureError(err, nil)
					log.WithError(err).Fatalf("Could not get or create address: %s", addressTx.Hash)
				}
				addresses[addressTx.Hash] = address
			}

			previousBalance := addresses[addressTx.Hash].Balance
			if addressTx.Cold == true {
				addressTx.Balance = uint64(int64(addresses[addressTx.Hash].ColdBalance) + addressTx.Total)
			} else {
				addressTx.Balance = uint64(int64(addresses[addressTx.Hash].Balance) + addressTx.Total)
			}

			log.WithField("tx", tx).Debug("CreateTransaction")
			log.WithField("address", addressTx.Hash).Debug("PreviousBalance: ", previousBalance)

			i.elastic.AddIndexRequest(
				elastic_cache.AddressTransactionIndex.Get(),
				addressTx.Slug(),
				addressTx,
			)

			ApplyTxToAddress(addresses[addressTx.Hash], addressTx)
			if addressTx.Cold == true {
				addresses[addressTx.Hash].ColdBalance += addressTx.Total
			} else {
				addresses[addressTx.Hash].Balance += addressTx.Total
			}

			i.elastic.AddUpdateRequest(
				elastic_cache.AddressIndex.Get(),
				addresses[addressTx.Hash].Slug(),
				addresses[addressTx.Hash],
			)
		}
	}

	if len(hashes) > 0 {
		if _, err := i.repo.GetAddresses(hashes); err != nil {
			raven.CaptureError(err, nil)
			log.WithError(err).Fatalf("Could not get addresses for txs at height %d", txs[0].Height)
			//addresses = i.createNewAddresses(hashes, addresses)
		}
	}
}

func (i *Indexer) createNewAddresses(hashes []string, addresses []*explorer.Address) []*explorer.Address {
	for _, hash := range hashes {
		found := func() bool {
			for _, address := range addresses {
				if hash == address.Hash {
					return true
				}
			}
			return false
		}()

		if !found {
			addresses = append(addresses, &explorer.Address{Hash: hash})
		}
	}

	return addresses
}
