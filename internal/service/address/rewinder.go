package address

import (
	"context"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
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
		raven.CaptureError(err, nil)
		return err
	}
	log.Infof("Rewinding %d addresses", len(addresses))

	err = r.elastic.DeleteHeightGT(height, elastic_cache.AddressTransactionIndex.Get())
	if err != nil {
		raven.CaptureError(err, nil)
		return err
	}

	for _, hash := range addresses {
		address, err := r.repo.GetOrCreateAddress(hash)
		if err != nil {
			return err
		}

		err = r.RewindAddressToHeight(address, height)
		if err != nil {
			return err
		}
	}

	return err
}

func (r *Rewinder) RewindAddressToHeight(address *explorer.Address, height uint64) error {
	log.WithField("address", address.Hash).Debug("Reindexing address index")

	address = ResetAddress(address)
	address.Height = height

	addressTxs, err := r.repo.GetTxsRangeForAddress(address.Hash, 0, height)
	if err != nil {
		log.WithField("address", address.Hash).WithError(err).Error("Failed to get tx range")
		raven.CaptureError(err, nil)
		return err
	}

	for _, addressTx := range addressTxs {
		ApplyTxToAddress(address, addressTx)
		if addressTx.Cold == true {
			address.ColdBalance = int64(addressTx.Balance)
		} else {
			address.Balance = int64(addressTx.Balance)
		}
		address.Height = addressTx.Height
	}

	log.WithField("address", address.Hash).Debugf("Rewind balance to height %d: %d", address.Height, address.Balance)

	_, err = r.elastic.Client.Index().
		Index(elastic_cache.AddressIndex.Get()).
		BodyJson(address).
		Id(address.Slug()).
		Do(context.Background())

	if err != nil {
		log.WithField("address", address.Hash).WithError(err).Error("Failed to persist the rewind")
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}
