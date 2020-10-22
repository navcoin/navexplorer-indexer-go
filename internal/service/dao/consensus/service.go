package consensus

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
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
	parameters, err := s.repo.GetConsensusParameters()
	if err != nil && err != elastic_cache.ErrRecordNotFound {
		raven.CaptureError(err, nil)
		log.WithError(err).Fatal("Failed to load consensus parameters")
		return
	}

	if len(parameters) != 0 {
		log.Info("Consensus parameters initialised")
		Parameters = parameters
		return
	}

	initialParams, _ := s.InitialState()
	for _, initialParam := range initialParams {
		result, err := s.elastic.Client.Index().
			Index(elastic_cache.ConsensusIndex.Get()).
			BodyJson(initialParam).
			Do(context.Background())
		if err != nil {
			log.WithError(err).Fatal("Failed to save new softfork")
		}
		initialParam.SetId(result.Id)
		initialParam.UpdatedOnBlock = 0

		log.Info("Saving new consensus parameter: ", initialParam.Description)
	}

	Parameters = initialParams
}

func (s *Service) InitialState() ([]*explorer.ConsensusParameter, error) {
	parameters := make([]*explorer.ConsensusParameter, 0)
	var byteParams []byte
	if config.Get().SoftForkBlockCycle != 20160 {
		log.Info("Initialising Testnet Consensus parameters")
		byteParams = []byte(testnet)
	} else {
		log.Info("Initialising Mainnet Consensus parameters")
		byteParams = []byte(mainnet)
	}

	err := json.Unmarshal(byteParams, &parameters)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Fatalf("Failed to load consensus parameters from JSON")
		return nil, err
	}

	return parameters, nil
}
