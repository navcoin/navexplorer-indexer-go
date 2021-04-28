package address

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	elastic *elastic_cache.Index
	repo    *Repository
	indexer *Indexer
}

func NewRewinder(elastic *elastic_cache.Index, repo *Repository, indexer *Indexer) *Rewinder {
	return &Rewinder{elastic, repo, indexer}
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

	r.indexer.ClearCache()

	return nil
}

func (r *Rewinder) ResetAddress(address *explorer.Address) error {
	log.Infof("Resetting address: %s", address.Hash)

	latestHistory, err := r.repo.GetLatestHistoryByHash(address.Hash)
	if err != nil && err != ErrLatestHistoryNotFound {
		return err
	}

	if latestHistory == nil {
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

	r.elastic.Save(elastic_cache.AddressIndex.Get(), address)

	return err
}
