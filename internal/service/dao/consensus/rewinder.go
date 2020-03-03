package consensus

import (
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewRewinder(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository) *Rewinder {
	return &Rewinder{navcoin, elastic, repo}
}

func (r *Rewinder) Rewind() error {
	navcoindConsensusParameters, err := r.navcoin.GetConsensusParameters(true)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Error("Failed to get consensus parameters")
		return err
	}

	consensusParameters, err := r.repo.GetConsensusParameters()
	for _, navcoindConsensusParameter := range navcoindConsensusParameters {
		for _, consensusParameter := range consensusParameters {
			if navcoindConsensusParameter.Id == consensusParameter.Id {
				UpdateConsensus(navcoindConsensusParameter, consensusParameter)
				r.elastic.AddUpdateRequest(
					elastic_cache.ConsensusIndex.Get(),
					fmt.Sprintf("consensus_%d", consensusParameter.Id),
					consensusParameter,
					consensusParameter.MetaData.Id,
				)
			}
		}
	}

	return nil
}
