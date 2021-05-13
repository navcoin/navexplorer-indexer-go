package consensus

import (
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetConsensusParameters() explorer.ConsensusParameters
	GetConsensusParameter(parameter explorer.Parameter) *explorer.ConsensusParameter
	Update(parameters explorer.ConsensusParameters, persist bool)
	InitConsensusParameters()
	InitialState() explorer.ConsensusParameters
}

type service struct {
	network    string
	elastic    elastic_cache.Index
	cache      *cache.Cache
	repository Repository
}

var (
	cacheKey = "explorer.ConsensusParameters"
)

func NewService(network string, elastic elastic_cache.Index, cache *cache.Cache, repository Repository) Service {
	return service{network, elastic, cache, repository}
}

func (s service) GetConsensusParameters() explorer.ConsensusParameters {
	parameters, exists := s.cache.Get(cacheKey)
	if exists == false {
		return nil
	}

	return parameters.(explorer.ConsensusParameters)
}

func (s service) GetConsensusParameter(parameter explorer.Parameter) *explorer.ConsensusParameter {
	parameters := s.GetConsensusParameters()
	for idx := range parameters {
		if parameters[idx].Id == int(parameter) {
			return parameters[idx]
		}
	}

	return nil
}

func (s service) Update(parameters explorer.ConsensusParameters, persist bool) {
	s.cache.Set(cacheKey, parameters, cache.NoExpiration)

	for _, parameter := range parameters {
		if persist {
			s.elastic.Save(elastic_cache.ConsensusIndex.Get(), parameter)
		} else {
			s.elastic.AddUpdateRequest(elastic_cache.ConsensusIndex.Get(), parameter)
		}
	}
}

func (s service) InitConsensusParameters() {
	parameters, err := s.repository.GetConsensusParameters()
	if err != nil && err != elastic_cache.ErrRecordNotFound {
		log.WithError(err).Fatal("Failed to load consensus parameters")
		return
	}

	if len(parameters) == 0 {
		parameters = s.InitialState()
		for _, parameter := range parameters {
			parameter.UpdatedOnBlock = 0
		}
	}

	log.Info("Consensus parameters initialised")
	for idx := range parameters {
		log.WithField("slug", parameters[idx].Slug()).Infof("Consensus Parameter %s", parameters[idx].Description)
	}

	s.Update(parameters, true)
}

func (s service) InitialState() explorer.ConsensusParameters {
	parameters := make(explorer.ConsensusParameters, 0)
	var byteParams []byte
	if config.Get().SoftForkBlockCycle != 20160 {
		log.Info("Initialising Testnet Consensus parameters")
		byteParams = []byte(testnet)
	} else {
		log.Info("Initialising Mainnet Consensus parameters")
		byteParams = []byte(mainnet)
	}

	if err := json.Unmarshal(byteParams, &parameters); err != nil {
		log.WithError(err).Fatalf("Failed to load consensus parameters from JSON")
	}

	return parameters
}
