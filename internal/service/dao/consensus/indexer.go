package consensus

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
	service *Service
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository, service *Service) *Indexer {
	return &Indexer{navcoin, elastic, repo, service}
}

func (i *Indexer) Index() error {
	initialParameters, _ := i.service.InitialState()

	c := make([]*explorer.ConsensusParameter, 0)

	consensusParameters, err := i.repo.GetConsensusParameters()
	if err != nil {
		log.WithError(err).Fatal("Failed to get consensus parameters from repo")
	}

	for _, initialParameter := range initialParameters {
		for _, consensusParameter := range consensusParameters {
			if initialParameter.Uid == consensusParameter.Uid {
				i.elastic.AddUpdateRequest(elastic_cache.ConsensusIndex.Get(), initialParameter)
				c = append(c, consensusParameter)
			}
		}
	}

	Parameters = c

	return nil
}

func (i *Indexer) Update(block *explorer.Block) {
	for _, p := range Parameters {
		if p.UpdatedOnBlock != block.Height {
			continue
		}

		i.elastic.Save(elastic_cache.ConsensusIndex, p)
	}
}
