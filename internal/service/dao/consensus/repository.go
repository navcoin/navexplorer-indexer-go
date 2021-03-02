package consensus

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type Repository struct {
	Client *elastic.Client
}

func NewRepo(client *elastic.Client) *Repository {
	return &Repository{client}
}

func (r *Repository) GetConsensusParameters() ([]*explorer.ConsensusParameter, error) {
	results, err := r.Client.Search(elastic_cache.ConsensusIndex.Get()).
		Sort("id", true).
		Size(10000).
		Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	if len(results.Hits.Hits) == 0 {
		return nil, elastic_cache.ErrRecordNotFound
	}

	var consensusParameters []*explorer.ConsensusParameter
	for _, hit := range results.Hits.Hits {
		var consensusParameter *explorer.ConsensusParameter
		if err = json.Unmarshal(hit.Source, &consensusParameter); err != nil {
			return nil, err
		}
		consensusParameter.SetId(hit.Id)
		consensusParameters = append(consensusParameters, consensusParameter)
	}

	return consensusParameters, nil
}

func (r *Repository) DeleteAll() error {
	log.Info("Deleting all consensus records")
	_, err := elastic.NewDeleteByQueryService(r.Client).
		Index(elastic_cache.ConsensusIndex.Get()).
		Query(elastic.NewMatchAllQuery()).
		Do(context.Background())

	return err
}
