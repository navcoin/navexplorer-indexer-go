package address

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
)

type Rewinder interface {
	Rewind(height uint64) error
	ResetAddress(address explorer.Address) error
}

type rewinder struct {
	elastic    elastic_cache.Index
	repository Repository
	indexer    Indexer
}

func NewRewinder(elastic elastic_cache.Index, repository Repository, indexer Indexer) Rewinder {
	return rewinder{elastic, repository, indexer}
}

func (r rewinder) Rewind(height uint64) error {
	zap.L().With(zap.Uint64("height", height)).Info("AddressRewinder: Rewinding address index")

	addresses, err := r.repository.GetAddressesHeightGt(height)
	if err != nil {
		zap.S().With(zap.Error(err)).Errorf("AddressRewinder: Failed to get addresses greater than %d", height)
		return err
	}

	err = r.elastic.DeleteHeightGT(height, elastic_cache.AddressHistoryIndex.Get())
	if err != nil {
		zap.S().With(zap.Error(err)).Errorf("AddressRewinder: Failed to delete addresses greater than %d", height)
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

func (r rewinder) ResetAddress(address explorer.Address) error {
	zap.L().With(zap.String("hash", address.Hash)).Info("AddressRewinder: Resetting address")

	latestHistory, err := r.repository.GetLatestHistoryByHash(address.Hash)
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

	return nil
}
