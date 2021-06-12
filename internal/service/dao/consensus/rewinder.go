package consensus

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"strconv"
)

type Rewinder interface {
	Rewind(consultations []*explorer.Consultation) error
}

type rewinder struct {
	navcoin    *navcoind.Navcoind
	elastic    elastic_cache.Index
	repository Repository
	service    Service
}

func NewRewinder(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, repository Repository, service Service) Rewinder {
	return rewinder{navcoin, elastic, repository, service}
}

func (r rewinder) Rewind(consultations []*explorer.Consultation) error {
	zap.L().Info("ConsensusRewinder: Rewind on initial state")

	parameters := r.service.InitialState()

	for _, c := range consultations {
		for _, p := range parameters.All() {
			if c.Min == p.Id {
				value, _ := strconv.Atoi(c.GetPassedAnswer().Answer)
				zap.L().With(
					zap.Int("old", p.Value),
					zap.Int("new", value),
					zap.String("name", p.Description),
				).Info("ConsensusRewinder: Update Consensus Parameter")
				p.Value = value
				p.UpdatedOnBlock = c.UpdatedOnBlock
			}
		}
	}

	r.service.Update(parameters, true)

	return nil
}
