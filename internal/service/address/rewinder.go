package address

import (
	"context"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
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
	log.Infof("Address Rewinder: Rewinding address index to height: %d", height)

	addresses, err := r.repo.GetAddressesHeightGt(height)
	if err != nil {
		raven.CaptureError(err, nil)
		return err
	}
	log.Infof("Address Rewinder: Rewinding %d addresses", len(addresses))

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
		address = ResetAddress(address)
		address.Height = height

		log.Infof("Address Rewinder: Reindexing address index for %s", hash)
		addressTxs, _ := r.repo.GetTxsRangeForAddress(hash, 0, height)
		for _, addressTx := range addressTxs {
			ApplyTxToAddress(address, addressTx)
			address.Height = addressTx.Height
		}

		_, err = r.elastic.Client.Index().
			Index(elastic_cache.AddressIndex.Get()).
			BodyJson(address).
			Id(address.Slug()).
			Do(context.Background())
		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}
	}

	return err
}
