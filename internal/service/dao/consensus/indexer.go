package consensus

import (
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository) *Indexer {
	return &Indexer{navcoin, elastic, repo}
}

func (i *Indexer) Index() error {
	navcoindConsensusParameters, err := i.navcoin.GetConsensusParameters(true)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Error("Failed to get consensus parameters")
		return err
	}

	c := make([]*explorer.ConsensusParameter, 0)

	consensusParameters, err := i.repo.GetConsensusParameters()
	for _, navcoindConsensusParameter := range navcoindConsensusParameters {
		for _, consensusParameter := range consensusParameters {
			if navcoindConsensusParameter.Id == consensusParameter.Id {
				UpdateConsensus(navcoindConsensusParameter, consensusParameter)
				i.elastic.AddUpdateRequest(
					elastic_cache.ConsensusIndex.Get(),
					fmt.Sprintf("consensus_%d", consensusParameter.Id),
					consensusParameter,
					consensusParameter.MetaData.Id,
				)
				c = append(c, consensusParameter)
			}
		}
	}

	Parameters = c

	return nil
}
