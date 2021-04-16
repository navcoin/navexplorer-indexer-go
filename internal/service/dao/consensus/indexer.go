package consensus

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
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

func (i *Indexer) Update(block *explorer.Block) {
	parameters := i.service.GetConsensusParameters()
	for _, p := range parameters {
		if p.UpdatedOnBlock != block.Height {
			continue
		}

		i.elastic.Save(elastic_cache.ConsensusIndex.Get(), p)
	}
}
