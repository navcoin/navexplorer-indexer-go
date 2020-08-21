package address

import (
	"context"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewRewinder(elastic *elastic_cache.Index, repo *Repository) *Rewinder {
	return &Rewinder{elastic, repo}
}

func (r *Rewinder) Rewind(height uint64) error {
	log.Infof("Rewinding address index to height: %d", height)

	addresses, err := r.repo.GetAddressesHeightGt(height)
	if err != nil {
		log.Error("Failed to get addresses greater than ", height)
		return err
	}

	log.Infof("Rewinding %d addresses", len(addresses))
	err = r.elastic.DeleteHeightGT(height, elastic_cache.AddressHistoryIndex.Get())
	if err != nil {
		log.Error("Failed to delete address history greater than ", height)
		return err
	}

	for _, address := range addresses {
		log.Infof("Rewinding address: %s", address.Hash)

		latestHistory, err := r.repo.GetLatestHistoryByHash(address.Hash)
		if err != nil {
			log.Fatal(err.Error())
		}

		if latestHistory == nil {
			address.Height = 0
			address.Balance = explorer.AddressBalance{}
		} else {
			address.Height = latestHistory.Height
			address.Balance = latestHistory.Balance
		}

		_, err = r.elastic.Client.Index().
			Index(elastic_cache.AddressIndex.Get()).
			BodyJson(address).
			Id(address.Slug()).
			Do(context.Background())
	}

	return nil
}
