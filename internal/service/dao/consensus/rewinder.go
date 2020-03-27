package consensus

import (
	"context"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
	service *Service
}

func NewRewinder(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository, service *Service) *Rewinder {
	return &Rewinder{navcoin, elastic, repo, service}
}

func (r *Rewinder) Rewind() error {
	log.Debug("Rewind consensus")
	initialParameters, _ := r.service.InitialState()

	consensusParameters, err := r.repo.GetConsensusParameters()
	if err != nil {
		log.WithError(err).Fatal("Failed to get consensus parameters from repo")
	}

	for _, initialParameter := range initialParameters {
		for _, consensusParameter := range consensusParameters {
			if initialParameter.Id == consensusParameter.Id {
				_, err = r.elastic.Client.Index().
					Index(elastic_cache.AddressIndex.Get()).
					BodyJson(initialParameter).
					Id(consensusParameter.Slug()).
					Do(context.Background())

				if err != nil {
					log.WithField("consensusParameter", consensusParameter).WithError(err).Error("Failed to persist the rewind")
					raven.CaptureError(err, nil)
					return err
				}
			}
		}
	}

	log.Debug("Rewind consensus success")

	return nil
}
