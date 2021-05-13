package address

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"time"
)

type Repository interface {
	GetBestHeight() (uint64, error)
	GetAddress(hash string) (*explorer.Address, error)
	GetAllAddresses() ([]explorer.Address, error)
	GetAddresses(hashes []string) ([]explorer.Address, error)
	GetAddressesHeightGt(height uint64) ([]explorer.Address, error)
	GetAddressesHeightLt(height uint64, size int) ([]explorer.Address, error)
	CreateAddress(hash string, createdBlock uint64, createdTime time.Time) (*explorer.Address, error)
	GetOrCreateAddress(hash string) explorer.Address
	GetLatestHistoryByHash(hash string) (*explorer.AddressHistory, error)
	GetAddressBalanceAtHeight(height uint64) (int64, error)

	findOne(result *elastic.SearchResult) (*explorer.Address, error)
	findMany(result *elastic.SearchResult) ([]explorer.Address, error)
}

type repository struct {
	elastic elastic_cache.Index
}

var (
	ErrLatestHistoryNotFound = errors.New("Latest history not found")
)

func NewRepo(elastic elastic_cache.Index) Repository {
	return repository{elastic}
}

func (r repository) GetBestHeight() (uint64, error) {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Sort("height", false).
		Size(1).
		Do(context.Background())
	if err != nil {
		return 0, err
	}

	if len(result.Hits.Hits) == 0 {
		return 0, nil
	}

	address, err := r.findOne(result)
	if err != nil {
		return 0, err
	}

	return address.Height, nil
}

func (r repository) GetAddress(hash string) (*explorer.Address, error) {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findOne(result)
}

func (r repository) GetAllAddresses() ([]explorer.Address, error) {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Size(1000).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findMany(result)
}

func (r repository) GetAddresses(hashes []string) ([]explorer.Address, error) {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermsQuery("hash", hashes)).
		Size(len(hashes)).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findMany(result)
}

func (r repository) GetAddressesHeightGt(height uint64) ([]explorer.Address, error) {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("height").Gt(height)).
		Size(50000).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findMany(result)
}

func (r repository) GetAddressesHeightLt(height uint64, size int) ([]explorer.Address, error) {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("height").Lt(height)).
		Size(size).
		Sort("attempt", true).
		Sort("created_block", true).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findMany(result)
}

func (r repository) CreateAddress(hash string, createdBlock uint64, createdTime time.Time) (*explorer.Address, error) {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Fatal("Failed to create address: " + hash)
		return nil, err
	}

	if len(result.Hits.Hits) != 0 {
		address, _ := r.findOne(result)
		return address, errors.New("Address already exists: " + hash)
	}

	address := CreateAddress(hash)
	address.CreatedBlock = createdBlock
	address.CreatedTime = createdTime

	r.elastic.Save(elastic_cache.AddressIndex.Get(), address)

	return &address, nil
}

func (r repository) GetOrCreateAddress(hash string) explorer.Address {
	result, err := r.elastic.GetClient().
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Fatal("Failed to get or create address")
	}

	if result.TotalHits() == 0 {
		address := CreateAddress(hash)
		r.elastic.Save(elastic_cache.AddressIndex.Get(), address)
		return address
	}

	address, _ := r.findOne(result)
	return *address
}

func (r repository) GetLatestHistoryByHash(hash string) (*explorer.AddressHistory, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	results, err := r.elastic.GetClient().
		Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Sort("height", false).
		Size(1).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to find address")
		err = ErrLatestHistoryNotFound
		return nil, err
	}

	if results.TotalHits() == 0 {
		return nil, nil
	}

	var history *explorer.AddressHistory
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &history)
	if err != nil {
		return nil, err
	}

	return history, err
}

func (r repository) GetAddressBalanceAtHeight(height uint64) (int64, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("height").Lte(height))

	agg := elastic.NewNestedAggregation().Path("changes").
		SubAggregation("spendable", elastic.NewSumAggregation().Field("changes.spendable"))

	results, err := r.elastic.GetClient().
		Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Aggregation("changes", agg).
		Size(0).
		Do(context.Background())
	if err != nil || results.TotalHits() == 0 {
		err = ErrLatestHistoryNotFound
		return 0, err
	}

	var addressBalance int64
	if changes, found := results.Aggregations.Nested("changes"); found {
		if spendableSum, found := changes.Sum("spendable"); found {
			addressBalance += int64(*spendableSum.Value)
		}
	}

	return addressBalance, err
}

func (r repository) findOne(result *elastic.SearchResult) (*explorer.Address, error) {
	if result == nil || len(result.Hits.Hits) != 1 {
		return nil, errors.New("AddressRepository: findOneAddress - Invalid result")
	}

	var address *explorer.Address
	hit := result.Hits.Hits[0]
	if err := json.Unmarshal(hit.Source, &address); err != nil {
		return nil, err
	}

	return address, nil
}

func (r repository) findMany(result *elastic.SearchResult) ([]explorer.Address, error) {
	if result == nil {
		return nil, errors.New("AddressRepository: findManyAddress - Invalid result")
	}

	addresses := make([]explorer.Address, 0)
	for _, hit := range result.Hits.Hits {
		var address explorer.Address
		if err := json.Unmarshal(hit.Source, &address); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}
