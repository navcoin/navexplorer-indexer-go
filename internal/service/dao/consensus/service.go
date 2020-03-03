package consensus

import (
	"context"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewService(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository) *Service {
	return &Service{navcoin, elastic, repo}
}

func (s *Service) InitConsensusParameters() {
	_, err := s.repo.GetConsensusParameters()
	if err != nil && err != elastic_cache.ErrRecordNotFound {
		raven.CaptureError(err, nil)
		log.WithError(err).Fatal("Failed to load consensus parameters")
		return
	}

	navcoindConsensusParameters, err := s.navcoin.GetConsensusParameters(true)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Fatal("Failed to get consensus parameters")
		return
	}

	c := make([]*explorer.ConsensusParameter, 0)

	for _, navcoindConsensusParameter := range navcoindConsensusParameters {
		consensusParameter := CreateConsensusParameter(navcoindConsensusParameter)
		resp, err := s.elastic.Client.Index().Index(elastic_cache.ConsensusIndex.Get()).BodyJson(consensusParameter).Do(context.Background())
		if err != nil {
			log.WithError(err).Fatal("Failed to save new softfork")
		}

		log.Info("Saving new consensus parameter: ", consensusParameter.Description)
		consensusParameter.MetaData = explorer.NewMetaData(resp.Id, resp.Index)
		c = append(c, consensusParameter)
	}

	Parameters = c
}
