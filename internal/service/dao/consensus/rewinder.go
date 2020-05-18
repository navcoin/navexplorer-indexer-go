package consensus

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"strconv"
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

func (r *Rewinder) Rewind(consultations []*explorer.Consultation) error {
	log.Debug("Rewind consensus")
	initialParameters, _ := r.service.InitialState()

	consensusParameters, err := r.repo.GetConsensusParameters()
	if err != nil {
		log.WithError(err).Fatal("Failed to get consensus parameters from repo")
	}

	for _, initialParameter := range initialParameters {
		for _, consensusParameter := range consensusParameters {
			if initialParameter.Id == consensusParameter.Id {
				r.elastic.AddUpdateRequest(elastic_cache.ConsensusIndex.Get(), consensusParameter)
			}
		}
	}

	for _, c := range consultations {
		for _, p := range consensusParameters {
			if c.Min == p.Id {
				value, _ := strconv.Atoi(c.GetPassedAnswer().Answer)
				p.Value = value
				r.elastic.AddUpdateRequest(elastic_cache.ConsensusIndex.Get(), p)
				break
			}
		}
	}

	log.Debug("Rewind consensus success")

	return nil
}
