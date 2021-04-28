package address

import (
	"context"
	"errors"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Indexer struct {
	navcoin           *navcoind.Navcoind
	elastic           *elastic_cache.Index
	cache             *cache.Cache
	addressRepository *Repository
	blockService      *block.Service
	blockRepository   *block.Repository
}

func NewIndexer(
	navcoin *navcoind.Navcoind,
	elastic *elastic_cache.Index,
	cache *cache.Cache,
	addressRepository *Repository,
	blockService *block.Service,
	blockRepository *block.Repository,
) *Indexer {
	return &Indexer{
		navcoin,
		elastic,
		cache,
		addressRepository,
		blockService,
		blockRepository,
	}
}

func (i *Indexer) BulkIndex(target uint64) error {
	log.Infof("AddressIndexer: Bulk index addresses to %d", target)
	return i.bulkIndex(target)
}

func (i *Indexer) bulkIndex(target uint64) error {
	addresses, err := i.addressRepository.GetAddressesHeightLt(1, 100)
	if err != nil {
		log.WithError(err).Fatal("AddressIndexer: GetAddressesHeightLt")
		return err
	}

	if len(addresses) == 0 {
		bb, err := i.blockRepository.GetBestBlock()
		if err != nil {
			log.WithError(err).Errorf("AddressIndexer: Best block not found")
			time.Sleep(2 * time.Second)
		} else if bb.Height >= target {
			log.Errorf("AddressIndexer: All addresses indexed to height %d", target)
			return nil
		} else {
			log.Info("AddressIndexer: Paused Address indexing for 2 seconds")
			time.Sleep(2 * time.Second)
		}
	}

	log.Infof("AddressIndexer: Found %d addresses to index", len(addresses))
	for _, address := range addresses {
		if address.MultiSig != nil {
			continue
		}

		from := address.CreatedBlock
		if from == 0 {
			from = 1
		}
		txids, err := i.navcoin.GetAddressHistory(&from, &target, address.Hash)
		if err != nil || len(txids) == 0 {
			address.Attempt++
			i.elastic.Save(elastic_cache.AddressIndex.Get(), address)
			log.WithError(err).Errorf("AddressIndexer: Failed to get txids for address %s from %d, to %d", address.Hash, from, target)
			continue
		}

		lastTxid := txids[len(txids)-1]

		tx, _ := i.blockRepository.GetTransactionByHash(lastTxid.TxId)
		if tx == nil {
			address.Attempt++
			i.elastic.Save(elastic_cache.AddressIndex.Get(), address)
			log.Debugf("AddressIndexer: Could not find latest tx for %s", address.Hash)
			continue
		}

		addressHistorys, err := i.generateAddressHistory(address.CreatedBlock, tx.Height, address.Hash, txids)
		if err != nil {
			address.Attempt++
			i.elastic.Save(elastic_cache.AddressIndex.Get(), address)
			log.WithError(err).Errorf("AddressIndexer: Failed to generate history for address %s", address.Hash)
			continue
		}

		bulk := i.elastic.Client.Bulk()
		log.Infof("AddressIndexer: Index address history for address: %s", address.Hash)
		for _, addressHistory := range addressHistorys {
			bulk.Add(elastic.NewBulkIndexRequest().Index(elastic_cache.AddressHistoryIndex.Get()).Id(addressHistory.Slug()).Doc(addressHistory))

			address.Height = addressHistory.Height
			address.Spendable = addressHistory.Balance.Spendable
			address.Stakable = addressHistory.Balance.Stakable
			address.VotingWeight = addressHistory.Balance.VotingWeight

			actions := bulk.NumberOfActions()
			if actions >= 500 {
				log.Infof("Persisting %d address actions", actions)
				i.bulkPersist(bulk)
				bulk = i.elastic.Client.Bulk()
			}
		}

		bulk.Add(elastic.NewBulkUpdateRequest().Index(elastic_cache.AddressIndex.Get()).Id(address.Slug()).Doc(address))

		i.bulkPersist(bulk)
		i.cache.Delete(address.Hash)
	}

	log.Info("AddressIndexer: Paused Address indexing for 1 seconds")
	time.Sleep(1 * time.Second)

	return i.bulkIndex(target)
}

func (i *Indexer) bulkPersist(bulk *elastic.BulkService) {
	response, err := bulk.Do(context.Background())
	if err != nil {
		log.WithError(err).Error("AddressIndexer: Failed to persist requests")
		for {
			switch {
			}
		}
	}

	if response.Errors == true {
		for _, failed := range response.Failed() {
			log.WithFields(log.Fields{
				"error": failed.Error,
				"index": failed.Index,
				"id":    failed.Id,
			}).Error("AddressIndexer: Failed to persist to ES")
		}
		for {
			switch {
			}
		}
	}
}

func (i *Indexer) Index(block *explorer.Block, txs []explorer.BlockTransaction, includeHistory bool) {
	if len(txs) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		i.indexAddresses(block.Height, txs, includeHistory)
	}()

	go func() {
		defer wg.Done()
		i.indexMultiSigs(txs)
	}()

	wg.Wait()
}

func (i *Indexer) ClearCache() {
	i.cache.Flush()
}

func (i *Indexer) indexAddresses(height uint64, txs []explorer.BlockTransaction, includeHistory bool) {
	addresses := getAddressesForTxs(txs)

	if includeHistory {
		for _, address := range addresses {
			addressHistorys, err := i.generateAddressHistory(height, height, address, nil)
			if err != nil {
				log.Fatal(err.Error())
			}
			for _, addressHistory := range addressHistorys {
				i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

				err := i.getAndUpdateAddress(addressHistory)
				if err != nil {
					log.WithError(err).Fatalf("AddressIndexer: Could not update address: %s", addressHistory.Hash)
				}
			}
		}
	} else {
		for _, hash := range addresses {
			if _, exists := i.cache.Get(hash); exists == false {
				address, err := i.addressRepository.CreateAddress(hash, txs[0].Height, txs[0].Time)
				if err != nil {
					log.WithError(err).Errorf("AddressIndexer: Failed to create address %s ", hash)
				} else {
					log.Infof("Address created at height %d: %s ", txs[0].Height, hash)
				}
				i.cache.Set(hash, address, cache.NoExpiration)
			}
		}
	}
}

func (i *Indexer) indexMultiSigs(txs []explorer.BlockTransaction) {
	for _, tx := range txs {
		for _, multiSig := range tx.GetAllMultiSigs() {
			address := i.getAddress(multiSig.Key())
			address.MultiSig = &multiSig
			if tx.Height == 11981 {
				log.Info("AddressIndexer: Indexing tx index", tx.Index)
			}
			if len(tx.GetAllMultiSigs()) == 0 {
				continue
			}

			addressHistory := CreateMultiSigAddressHistory(tx, address.MultiSig, address)
			i.elastic.Save(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

			err := i.updateAddress(address, addressHistory)
			if err != nil {
				log.WithError(err).Fatalf("AddressIndexer: Could not update address: %s", addressHistory.Hash)
			}
		}
	}
}

func (i *Indexer) getAddress(hash string) explorer.Address {
	address, exists := i.cache.Get(hash)
	if exists == false {
		address := i.addressRepository.GetOrCreateAddress(hash)
		return address
	}

	return address.(explorer.Address)
}

func (i *Indexer) generateAddressHistory(start, end uint64, address string, history []*navcoind.AddressHistory) ([]explorer.AddressHistory, error) {
	addressHistorys := make([]explorer.AddressHistory, 0)

	if history == nil {
		var err error
		history, err = i.navcoin.GetAddressHistory(&start, &end, address)
		if err != nil {
			log.WithError(err).Fatalf("AddressIndexer: Could not get address history for height: %d-%d", start, end)
			return nil, err
		}
	}

	for idx, h := range history {
		tx, err := i.blockRepository.GetTransactionByHash(h.TxId)
		if err != nil {
			return nil, errors.New("TX related to address history is not available")
		}
		addressHistory := CreateAddressHistory(address, uint(idx), h, tx)
		addressHistorys = append(addressHistorys, addressHistory)
	}

	return addressHistorys, nil
}

func (i *Indexer) getAndUpdateAddress(history explorer.AddressHistory) error {
	return i.updateAddress(i.getAddress(history.Hash), history)
}

func (i *Indexer) updateAddress(address explorer.Address, history explorer.AddressHistory) error {
	if address.CreatedBlock == 0 {
		address.CreatedBlock = history.Height
		address.CreatedTime = history.Time
	}
	address.Height = history.Height
	address.Spendable = history.Balance.Spendable
	address.Stakable = history.Balance.Stakable
	address.VotingWeight = history.Balance.VotingWeight

	i.cache.Set(address.Hash, address, cache.NoExpiration)
	i.elastic.Save(elastic_cache.AddressIndex.Get(), address)

	return nil
}

func getAddressesForTxs(txs []explorer.BlockTransaction) []string {
	addresses := make([]string, 0)
	for _, tx := range txs {
		for _, address := range tx.GetAllAddresses() {
			addresses = append(addresses, address)
		}
	}

	return unique(addresses)
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
