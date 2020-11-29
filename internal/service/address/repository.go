package address

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Repository struct {
	elastic *elastic_cache.Index
}

var (
	ErrLatestHistoryNotFound = errors.New("Latest history not found")
)

func NewRepo(elastic *elastic_cache.Index) *Repository {
	return &Repository{elastic}
}

func (r *Repository) GetBestHeight() (uint64, error) {
	result, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Sort("height", false).
		Size(1).
		Do(context.Background())
	if err != nil {
		return 0, err
	}

	address, err := r.findOneAddress(result)
	if err != nil {
		return 0, err
	}

	return address.Height, nil
}

func (r *Repository) GetAddress(hash string) (*explorer.Address, error) {
	result, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findOneAddress(result)
}

func (r *Repository) GetAddresses(hashes []string) ([]*explorer.Address, error) {
	result, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewMatchQuery("hash", strings.Join(hashes, " "))).
		Size(len(hashes)).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findManyAddress(result)
}

func (r *Repository) GetAddressesHeightGt(height uint64) ([]*explorer.Address, error) {
	result, err := r.elastic.Client.
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("height").Gt(height)).
		Size(50000).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return r.findManyAddress(result)
}

func (r *Repository) GetOrCreateAddress(hash string) (*explorer.Address, error) {
	result, err := r.elastic.Client.
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	if result.TotalHits() == 0 {
		address := CreateAddress(hash)
		log.Debug("Persisted new address ", hash)
		r.elastic.Save(elastic_cache.AddressIndex, address)
		return address, nil
	}

	return r.findOneAddress(result)
}

func (r *Repository) GetLatestHistoryByHash(hash string) (*explorer.AddressHistory, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
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
	history.SetId(hit.Id)

	return history, err
}

func (r *Repository) findOneAddress(result *elastic.SearchResult) (*explorer.Address, error) {
	if result == nil || len(result.Hits.Hits) != 1 {
		return nil, errors.New("AddressRepository: findOneAddress - Invalid result")
	}

	var address *explorer.Address
	hit := result.Hits.Hits[0]
	if err := json.Unmarshal(hit.Source, &address); err != nil {
		return nil, err
	}
	address.SetId(hit.Id)

	return address, nil
}

func (r *Repository) findManyAddress(result *elastic.SearchResult) ([]*explorer.Address, error) {
	if result == nil {
		return nil, errors.New("AddressRepository: findManyAddress - Invalid result")
	}

	addresses := make([]*explorer.Address, 0)
	for _, hit := range result.Hits.Hits {
		var address *explorer.Address
		if err := json.Unmarshal(hit.Source, &address); err != nil {
			return nil, err
		}
		address.SetId(hit.Id)
		addresses = append(addresses, address)
	}

	return addresses, nil
}
