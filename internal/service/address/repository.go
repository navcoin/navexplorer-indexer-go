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
	"io"
	"strings"
)

type Repository struct {
	Client *elastic.Client
}

func NewRepo(Client *elastic.Client) *Repository {
	return &Repository{Client}
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

func (r *Repository) GetAddressesHeightGt(height uint64) ([]string, error) {
	log.Debugf("GetAddressesHeightGt(height:%d)", height)

	addresses := make(map[string]struct{}, 0)
	getAddresses := func(addresses map[string]struct{}, results *elastic.SearchResult) map[string]struct{} {
		if agg, found := results.Aggregations.Terms("hash"); found {
			for _, bucket := range agg.Buckets {
				addresses[bucket.Key.(string)] = struct{}{}
			}
		}
		return addresses
	}
	getAddressesHeightGtByIndex := func(height uint64, index elastic_cache.Indices) (*elastic.SearchResult, error) {
		return r.Client.
			Search(index.Get()).
			Query(elastic.NewRangeQuery("height").Gt(height)).
			Aggregation("hash", elastic.NewTermsAggregation().Field("hash.keyword").Size(50000000)).
			Size(0).
			Do(context.Background())
	}

	results, err := getAddressesHeightGtByIndex(height, elastic_cache.AddressIndex)
	if err != nil || results == nil {
		raven.CaptureError(err, nil)
		return nil, err
	}
	addresses = getAddresses(addresses, results)

	results, err = getAddressesHeightGtByIndex(height, elastic_cache.AddressTransactionIndex)
	if err != nil || results == nil {
		raven.CaptureError(err, nil)
		return nil, err
	}
	addresses = getAddresses(addresses, results)

	addressesSlice := make([]string, len(addresses))
	i := 0
	for f, _ := range addresses {
		addressesSlice[i] = f
		i++
	}

	return addressesSlice, nil
}

func (r *Repository) getAddressesHeightGtByIndex(height uint64, index elastic_cache.Indices) (*elastic.SearchResult, error) {
	return r.Client.
		Search(index.Get()).
		Query(elastic.NewRangeQuery("height").Gt(height)).
		Aggregation("hash", elastic.NewTermsAggregation().Field("hash.keyword").Size(50000000)).
		Size(0).
		Do(context.Background())
}

func (r *Repository) GetOrCreateAddress(hash string) (*explorer.Address, error) {
	log.WithField("address", hash).Debug("GetOrCreateAddress")

	results, err := r.Client.
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())
	if err != nil || results == nil {
		raven.CaptureError(err, nil)
		return nil, err
	}

	var address *explorer.Address
	if len(results.Hits.Hits) == 0 {
		address = CreateAddress(hash)
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

func (r *Repository) GetAddress(hash string) (*explorer.Address, error) {
	log.Debugf("GetAddress(hash:%s)", hash)

	results, err := r.Client.
		Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash", hash)).
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

func (r *Repository) GetTxsRangeForAddress(hash string, from uint64, to uint64) ([]*explorer.AddressTransaction, error) {
	log.WithField("address", hash).Debugf("GetTxsRangeForAddress: from:%d to:%d", from, to)

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("hash", hash))
	query = query.Must(elastic.NewRangeQuery("height").Gt(from).Lte(to))

	service := r.Client.Scroll(elastic_cache.AddressTransactionIndex.Get()).Query(query).Size(10000).Sort("height", true)
	txs := make([]*explorer.AddressTransaction, 0)

	for {
		results, err := service.Do(context.Background())
		if err == io.EOF {
			break
		}
		if err != nil || results == nil {
			log.Fatal(err)
		}

		for _, hit := range results.Hits.Hits {
			var tx *explorer.AddressTransaction
			if err = json.Unmarshal(hit.Source, &tx); err != nil {
				raven.CaptureError(err, nil)
				return nil, err
			}
			txs = append(txs, tx)
		}
	}

	return txs, nil
}

func (r *Repository) GetTxsForAddress(hash string) ([]*explorer.AddressTransaction, error) {
	log.Debugf("GetTxsForAddress(hash:%s)", hash)

	results, err := r.Client.
		Search(elastic_cache.AddressTransactionIndex.Get()).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(50000000).
		Sort("height", true).
		Do(context.Background())
	if err != nil {
		raven.CaptureError(err, nil)
		return nil, err
	}

	txs := make([]*explorer.AddressTransaction, 0)
	for _, hit := range results.Hits.Hits {
		var tx *explorer.AddressTransaction
		if err = json.Unmarshal(hit.Source, &tx); err != nil {
			raven.CaptureError(err, nil)
			return nil, err
		}
		txs = append(txs, tx)
	}

	return txs, nil
}
