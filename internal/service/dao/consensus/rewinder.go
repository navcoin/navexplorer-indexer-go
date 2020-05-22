package consensus

import (
	"context"
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
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

	//parameters, err := r.repo.GetConsensusParameters()
	parameters, err := r.service.InitialState()
	if err != nil {
		log.WithError(err).Fatal("Failed to get consensus parameters from repo")
	}

	for _, c := range consultations {
		for _, p := range parameters {
			if c.Min == p.Id {
				value, _ := strconv.Atoi(c.GetPassedAnswer().Answer)
				log.WithFields(log.Fields{"old": p.Value, "new": value, "desc": p.Description}).Info("Update consensus parameter")
				p.Value = value
				p.UpdatedOnBlock = c.UpdatedOnBlock
			}
		}
	}

	for _, p := range parameters {
		_, err = r.elastic.Client.Index().
			Index(elastic_cache.ConsensusIndex.Get()).
			Id(fmt.Sprintf("%s-%s", config.Get().Network, p.Slug())).
			BodyJson(p).
			Do(context.Background())
		if err != nil {
			log.WithError(err).Fatal("Failed to get consensus parameters from repo")
		}
	}

	Parameters = parameters

	log.Info("Rewind consensus success")

	return nil
}
