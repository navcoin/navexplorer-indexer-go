package consensus

import (
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type Service interface {
	GetConsensusParameters() explorer.ConsensusParameters
	GetConsensusParameter(parameter explorer.Parameter) (bool, explorer.ConsensusParameter)
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
		return explorer.ConsensusParameters{}
	}

	return parameters.(explorer.ConsensusParameters)
}

func (s service) GetConsensusParameter(parameter explorer.Parameter) (bool, explorer.ConsensusParameter) {
	parameters := s.GetConsensusParameters()
	for _, p := range parameters.All() {
		zap.L().
			With(zap.String("desc", p.Description), zap.Int("id", p.Id), zap.Int("id", p.Value)).
			Debug("Consensus Parameter found")
		if p.Id == int(parameter) {
			return true, p
		}
	}

	return false, explorer.ConsensusParameter{}
}

func (s service) Update(parameters explorer.ConsensusParameters, persist bool) {
	s.cache.Set(cacheKey, parameters, cache.NoExpiration)

	for _, parameter := range parameters.All() {
		if persist {
			s.elastic.Save(elastic_cache.ConsensusIndex.Get(), parameter)
		} else {
			s.elastic.AddUpdateRequest(elastic_cache.ConsensusIndex.Get(), parameter)
		}
	}
}

func (s service) InitConsensusParameters() {
	parameters, err := s.repository.GetConsensusParameters()
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("ConsensusService: Failed to load consensus parameters")
		return
	}

	if len(parameters.All()) == 0 {
		parameters = s.InitialState()
		for _, parameter := range parameters.All() {
			parameter.UpdatedOnBlock = 0
		}
	}

	for _, p := range parameters.All() {
		zap.L().With(
			zap.String("name", p.Description),
			zap.Int("value", p.Value),
		).Info("ConsensusService: Parameter initialised")
	}

	s.Update(parameters, true)
}

func (s service) InitialState() explorer.ConsensusParameters {
	var byteParams []byte
	if config.Get().SoftForkBlockCycle != 20160 {
		zap.L().Info("ConsensusService: Initialising Testnet Consensus parameters")
		byteParams = []byte(testnet)
	} else {
		zap.L().Info("ConsensusService: Initialising Mainnet Consensus parameters")
		byteParams = []byte(mainnet)
	}

	parameterSlice := make([]explorer.ConsensusParameter, 0)
	if err := json.Unmarshal(byteParams, &parameterSlice); err != nil {
		zap.L().With(zap.Error(err)).Fatal("ConsensusService: Failed to load consensus parameters from JSON")
	}

	parameters := explorer.ConsensusParameters{}
	for idx := range parameterSlice {
		parameters.Add(parameterSlice[idx])
	}

	return parameters
}
