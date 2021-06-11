package proposal

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
)

type Repository interface {
	GetPossibleVotingProposals(height uint64) ([]*explorer.Proposal, error)
}

type repository struct {
	elastic elastic_cache.Index
}

func NewRepo(elastic elastic_cache.Index) Repository {
	return repository{elastic}
}

func (r repository) GetPossibleVotingProposals(height uint64) ([]*explorer.Proposal, error) {
	var proposals []*explorer.Proposal

	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewMatchQuery("status", "pending accepted pending_voting_preq pending_funds"))
	query = query.Should(elastic.NewRangeQuery("updatedOnBlock").Gte(height))

	results, err := r.elastic.GetClient().Search(elastic_cache.ProposalIndex.Get()).
		Query(query).
		Size(9999).
		Sort("updatedOnBlock", false).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	if results != nil {
		for _, hit := range results.Hits.Hits {
			var proposal *explorer.Proposal
			if err := json.Unmarshal(hit.Source, &proposal); err != nil {
				zap.L().With(zap.Error(err)).Fatal("Failed to unmarshall proposal")
			}
			proposals = append(proposals, proposal)
		}
	}

	return proposals, nil
}
