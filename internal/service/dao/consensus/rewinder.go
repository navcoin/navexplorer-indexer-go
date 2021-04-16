package consensus

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
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

	parameters := r.service.InitialState()

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

	r.service.Update(parameters, true)

	log.Info("Rewind consensus success")

	return nil
}
