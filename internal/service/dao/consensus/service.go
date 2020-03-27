package consensus

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	network string
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewService(network string, elastic *elastic_cache.Index, repo *Repository) *Service {
	return &Service{network, elastic, repo}
}

func (s *Service) InitConsensusParameters() {
	p, err := s.repo.GetConsensusParameters()
	if err != nil && err != elastic_cache.ErrRecordNotFound {
		raven.CaptureError(err, nil)
		log.WithError(err).Fatal("Failed to load consensus parameters")
		return
	}

	if len(p) != 0 {
		log.Info("Consensus parameters already initialised")
		Parameters = p
		return
	}

	parameters, _ := s.InitialState()
	for _, parameter := range parameters {
		_, err := s.elastic.Client.Index().
			Index(elastic_cache.ConsensusIndex.Get()).
			Id(parameter.Slug()).
			BodyJson(parameter).
			Do(context.Background())
		if err != nil {
			log.WithError(err).Fatal("Failed to save new softfork")
		}

		log.Info("Saving new consensus parameter: ", parameter.Description)
	}

	Parameters = parameters
}

func (s *Service) InitialState() ([]*explorer.ConsensusParameter, error) {
	parameters := make([]*explorer.ConsensusParameter, 0)
	byteParams := []byte(mainnet)
	if config.Get().Network != "mainnet" {
		byteParams = []byte(testnet)
	}

	err := json.Unmarshal(byteParams, &parameters)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Fatalf("Failed to load consensus parameters from JSON")
		return nil, err
	}

	return parameters, nil
}
