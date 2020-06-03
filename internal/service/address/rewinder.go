package address

import (
	"context"
	"fmt"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
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
		log.Error("Failed to get addresses greater than ", height)
		return err
	}

	log.Infof("Rewinding %d addresses", len(addresses))
	err = r.elastic.DeleteHeightGT(height, elastic_cache.AddressTransactionIndex.Get())
	if err != nil {
		log.Error("Failed to delete address transactions greater than ", height)
		return err
	}

	for idx, hash := range addresses {
		Addresses.Delete(hash)

		log.Infof("Rewinding address %d: %s", idx+1, hash)
		address, err := r.repo.GetAddress(hash)
		if err != nil {
			log.Error("Failed to get address with hash ", hash)
			return err
		}

		err = r.RewindAddressToHeight(address, height)
		if err != nil {
			log.Errorf("Failed to rewind address %s to height %s", address, height)
			return err
		}
	}

	return nil
}

func (r *Rewinder) ResetAddressToHeight(address *explorer.Address, height uint64) error {
	log.Infof("Resetting address %s from %d to %d", address.Hash, 0, height)

	ResetAddress(address)
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
			log.Debugf("Setting address balance to %d", address.Balance)
		}
		address.Height = addressTx.Height
	}

	return nil
}

func (r *Rewinder) RewindAddressToHeight(address *explorer.Address, height uint64) error {
	log.WithField("address", address.Hash).Debugf("Rewind balance to height %d: %d", address.Height, address.Balance)

	err := r.ResetAddressToHeight(address, height)
	if err != nil {
		log.WithField("address", address.Hash).WithError(err).Error("Failed to rest the address")
		raven.CaptureError(err, nil)
		return err
	}

	_, err = r.elastic.Client.Index().
		Index(elastic_cache.AddressIndex.Get()).
		BodyJson(address).
		Id(fmt.Sprintf("%s-%s", config.Get().Network, address.Slug())).
		Do(context.Background())

	if err != nil {
		log.WithField("address", address.Hash).WithError(err).Error("Failed to persist the rewind")
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}
