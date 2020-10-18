package address

import (
	"context"
	"encoding/json"
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

	err = r.elastic.DeleteHeightGT(height, elastic_cache.AddressHistoryIndex.Get())
	if err != nil {
		log.Error("Failed to delete address history greater than ", height)
		return err
	}

	for _, address := range addresses {
		if err = r.ResetAddress(address); err != nil {
			return err
		}
	}

	return nil
}

func (r *Rewinder) ResetAddress(address *explorer.Address) error {
	log.Infof("Resetting address: %s", address.Hash)

	latestHistory, err := r.repo.GetLatestHistoryByHash(address.Hash)
	if err != nil && err != ErrLatestHistoryNotFound {
		return err
	}

	b, _ := json.Marshal(latestHistory)
	log.Info(string(b))

	if latestHistory == nil {
		log.Error("Failed to find latest history for ", address.Hash)
		address.Height = 0
		address.Spendable = 0
		address.Stakable = 0
		address.VotingWeight = 0
	} else {
		address.Height = latestHistory.Height
		address.Spendable = latestHistory.Balance.Spendable
		address.Stakable = latestHistory.Balance.Stakable
		address.VotingWeight = latestHistory.Balance.VotingWeight
	}

	_, err = r.elastic.Client.Index().
		Index(elastic_cache.AddressIndex.Get()).
		BodyJson(address).
		Id(address.Slug()).
		Do(context.Background())

	return nil
}
