package address

import (
	"context"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"sort"
	"sync"
	"time"
)

type Indexer interface {
	bulkPersist(bulk *elastic.BulkService)
	Index(from, to uint64, txs []explorer.BlockTransaction, option IndexOption.IndexOption)
	ClearCache()
	indexAddresses(from, to uint64, option IndexOption.IndexOption)
	indexMultiSigs(txs []explorer.BlockTransaction, option IndexOption.IndexOption)
	getAddress(hash string, useCache bool) explorer.Address
	updateAddress(address explorer.Address, history explorer.AddressHistory) error
	getAndUpdateAddress(history explorer.AddressHistory, useCache bool) error
	generateAddressHistory(start, end uint64, addresses []string) ([]explorer.AddressHistory, error)
}

type indexer struct {
	navcoin           *navcoind.Navcoind
	elastic           elastic_cache.Index
	cache             *cache.Cache
	addressRepository Repository
	blockService      block.Service
	blockRepository   block.Repository
}

func NewIndexer(
	navcoin *navcoind.Navcoind,
	elastic elastic_cache.Index,
	cache *cache.Cache,
	addressRepository Repository,
	blockService block.Service,
	blockRepository block.Repository,
) Indexer {
	return indexer{
		navcoin,
		elastic,
		cache,
		addressRepository,
		blockService,
		blockRepository,
	}
}

func (i indexer) bulkPersist(bulk *elastic.BulkService) {
	response, err := bulk.Do(context.Background())
	if err != nil {
		zap.L().Fatal("AddressIndexer: Failed to persist requests")
	}

	if response != nil && response.Errors == true {
		for _, failed := range response.Failed() {
			zap.L().With(
				zap.Any("error", failed.Error),
				zap.String("index", failed.Index),
				zap.String("id", failed.Id),
			).Fatal("AddressIndexer: Failed to persist requests")
		}
	}
}

func (i indexer) Index(from, to uint64, txs []explorer.BlockTransaction, option IndexOption.IndexOption) {
	if len(txs) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		i.indexAddresses(from, to, option)
	}()

	go func() {
		defer wg.Done()
		i.indexMultiSigs(txs, option)
	}()

	wg.Wait()
}

func (i indexer) ClearCache() {
	i.cache.Flush()
}

func (i indexer) indexAddresses(from, to uint64, option IndexOption.IndexOption) {
	txs := make([]explorer.BlockTransaction, 0)
	for _, entity := range i.elastic.GetEntitiesByIndex(elastic_cache.BlockTransactionIndex.Get()) {
		txs = append(txs, entity.(explorer.BlockTransaction))
	}

	start := time.Now()
	addresses := getAddressesForTxs(txs)

	var wg sync.WaitGroup
	for _, chunk := range chunkAddresses(getAddressesForTxs(txs), 10) {
		wg.Add(1)

		go func(addresses []string) {
			defer wg.Done()
			addressHistorys, err := i.generateAddressHistory(from, to, addresses)
			if err != nil {
				zap.L().With(zap.Error(err)).Fatal("AddressIndexer: Failed to generate history")
			}

			sort.Slice(addressHistorys, func(i, j int) bool {
				return addressHistorys[i].Height < addressHistorys[j].Height
			})

			zap.L().With(
				zap.Int("addresses", len(addresses)),
				zap.Int("histories", len(addressHistorys)),
				zap.Duration("elapsed", time.Since(start)),
				zap.Uint64("from", from),
				zap.Uint64("to", to),
			).Info("AddressIndexer: Generate Address History")

			for _, history := range addressHistorys {
				i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), history)

				err := i.getAndUpdateAddress(history, option == IndexOption.BatchIndex)
				if err != nil {
					zap.L().With(
						zap.Error(err),
						zap.String("address", history.Hash),
					).Fatal("AddressIndexer: Could not update address")
				}
			}
		}(chunk)
	}

	wg.Wait()

	zap.L().With(
		zap.Int("addresses", len(addresses)),
		zap.Uint64("from", from),
		zap.Uint64("to", to),
		zap.Duration("elapsed", time.Since(start)),
	).Info("AddressIndexer: Address Index Complete")
}

func chunkAddresses(addresses []string, chunks int) [][]string {
	var chunked [][]string

	chunkSize := (len(addresses) + chunks - 1) / chunks

	for i := 0; i < len(addresses); i += chunkSize {
		end := i + chunkSize
		if end > len(addresses) {
			end = len(addresses)
		}

		chunked = append(chunked, addresses[i:end])
	}

	return chunked
}

func (i indexer) indexMultiSigs(txs []explorer.BlockTransaction, option IndexOption.IndexOption) {
	for _, tx := range txs {
		for _, multiSig := range tx.GetAllMultiSigs() {
			address := i.getAddress(multiSig.Key(), option == IndexOption.BatchIndex)
			address.MultiSig = &multiSig
			if len(tx.GetAllMultiSigs()) == 0 {
				continue
			}

			addressHistory := CreateMultiSigAddressHistory(tx, address.MultiSig, address)
			i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

			err := i.updateAddress(address, addressHistory)
			if err != nil {
				zap.S().With(zap.Error(err), zap.String("hash", addressHistory.Hash)).
					Fatalf("AddressIndexer: Could not update address")
			}
		}
	}
}

func (i indexer) getAddress(hash string, useCache bool) explorer.Address {
	if !useCache {
		address := i.addressRepository.GetOrCreateAddress(hash)
		return address
	}

	address, exists := i.cache.Get(hash)
	if exists == false {
		address := i.addressRepository.GetOrCreateAddress(hash)
		return address
	}

	return address.(explorer.Address)
}

func (i indexer) generateAddressHistory(start, end uint64, addresses []string) ([]explorer.AddressHistory, error) {
	var err error
	startTime := time.Now()

	historys, err := i.navcoin.GetAddressHistory(&start, &end, addresses...)
	addressHistorys := make([]explorer.AddressHistory, len(historys))

	zap.L().With(
		zap.Duration("elapsed", time.Since(startTime)),
		zap.Int("count", len(historys)),
	).Info("AddressIndexer: Get address histories")

	if err != nil {
		zap.L().With(
			zap.Error(err),
			zap.Uint64("from", start),
			zap.Uint64("to", end),
		).Fatal("AddressIndexer: Could not get address history")
		return nil, err
	}

	startTime = time.Now()
	var wg sync.WaitGroup
	wg.Add(len(historys))
	for idx, history := range historys {
		go func(idx int, history *navcoind.AddressHistory) {
			defer wg.Done()

			tx, err := i.blockRepository.GetTransactionByHash(history.TxId)
			if err != nil {
				zap.L().With(
					zap.Error(err),
					zap.String("address", history.Address),
					zap.String("txid", history.TxId),
				).Fatal("AddressIndexer: TX related to address history is not available")
			}

			addressHistorys[idx] = CreateAddressHistory(uint(idx), history, tx)
		}(idx, history)
	}
	wg.Wait()

	zap.L().With(
		zap.Duration("elapsed", time.Since(startTime)),
		zap.Int("count", len(addressHistorys)),
	).Info("AddressIndexer: Create Address Histories")

	return addressHistorys, nil
}

func (i indexer) getAndUpdateAddress(history explorer.AddressHistory, useCache bool) error {
	start := time.Now()
	err := i.updateAddress(i.getAddress(history.Hash, useCache), history)

	zap.L().With(
		zap.Duration("elapsed", time.Since(start)),
		zap.Uint64("height", history.Height),
		zap.String("address", history.Hash),
	).Debug("AddressIndexer: Get and Update Address")

	return err
}

func (i indexer) updateAddress(address explorer.Address, history explorer.AddressHistory) error {
	if address.CreatedBlock == 0 {
		address.CreatedBlock = history.Height
		address.CreatedTime = history.Time
	}
	address.Height = history.Height
	address.Spendable = history.Balance.Spendable
	address.Stakable = history.Balance.Stakable
	address.VotingWeight = history.Balance.VotingWeight

	i.cache.Set(address.Hash, address, cache.NoExpiration)
	i.elastic.AddUpdateRequest(elastic_cache.AddressIndex.Get(), address)

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
