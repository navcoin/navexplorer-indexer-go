package address

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Repository struct {
	Client *elastic.Client
}

func NewRepo(Client *elastic.Client) *Repository {
	return &Repository{Client}
}

func (r *Repository) GetAddress(hash string) (*explorer.Address, error) {
	log.Debugf("GetAddress(hash:%s)", hash)

	results, err := r.Client.Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())
	if err != nil {
		raven.CaptureError(err, nil)
		return nil, err
	}

	if results == nil || len(results.Hits.Hits) != 1 {
		return nil, errors.New("Invalid result found")
	}

	var address *explorer.Address
	if err = json.Unmarshal(results.Hits.Hits[0].Source, &address); err != nil {
		return nil, err
	}

	return address, nil
}

func (r *Repository) GetAddresses(hashes []string) ([]*explorer.Address, error) {
	log.Debugf("GetAddresses([%s])", strings.Join(hashes, ","))

	results, err := r.Client.Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewMatchQuery("hash", strings.Join(hashes, " "))).
		Size(len(hashes)).
		Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	addresses := make([]*explorer.Address, 0)
	for _, hit := range results.Hits.Hits {
		var address *explorer.Address
		if err = json.Unmarshal(hit.Source, &address); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (r *Repository) GetAddressesHeightGt(height uint64) ([]*explorer.Address, error) {
	log.Debugf("GetAddressesHeightGt(height:%d)", height)

	results, err := r.Client.
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("height").Gt(height)).
		Size(50000).
		Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	addresses := make([]*explorer.Address, 0)
	for _, hit := range results.Hits.Hits {
		var address *explorer.Address
		if err = json.Unmarshal(hit.Source, &address); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (r *Repository) GetOrCreateAddress(hash string, block *explorer.Block) (*explorer.Address, error) {
	log.WithField("address", hash).Debug("GetOrCreateAddress")

	results, err := r.Client.
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	var address *explorer.Address
	if results.TotalHits() == 0 {
		address = CreateAddress(hash, block.Height, block.MedianTime)
		_, err := r.Client.
			Index().
			Index(elastic_cache.AddressIndex.Get()).
			Id(address.Slug()).
			BodyJson(address).
			Do(context.Background())
		if err != nil {
			return nil, err
		}

		return address, nil
	}

	if err = json.Unmarshal(results.Hits.Hits[0].Source, &address); err != nil {
		return nil, err
	}

	return address, nil
}

func (r *Repository) GetLatestHistoryByHash(hash string) (*explorer.AddressHistory, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	results, err := r.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Sort("height", false).
		Size(1).
		Do(context.Background())
	if err != nil || results == nil {
		log.WithError(err).Fatal("Failed to find address")
	}

	var history *explorer.AddressHistory
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &history)
	if err != nil {
		return nil, err
	}

	return history, err
}
