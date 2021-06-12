package consensus

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

type Indexer interface {
	Update(block *explorer.Block)
}

type indexer struct {
	navcoin    *navcoind.Navcoind
	elastic    elastic_cache.Index
	repository Repository
	service    Service
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, repository Repository, service Service) Indexer {
	return indexer{navcoin, elastic, repository, service}
}

func (i indexer) Update(block *explorer.Block) {
	parameters := i.service.GetConsensusParameters()
	for _, p := range parameters.All() {
		if p.UpdatedOnBlock != block.Height {
			continue
		}

		i.elastic.Save(elastic_cache.ConsensusIndex.Get(), p)
	}
}
