package softfork

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

func (r *Repository) getSoftForks() (explorer.SoftForks, error) {
	softForks := make(explorer.SoftForks, 0)

	results, err := r.Client.Search(elastic_cache.SoftForkIndex.Get()).
		Size(9999).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	if results == nil {
		return nil, elastic_cache.ErrResultsNotFound
	}

	for _, hit := range results.Hits.Hits {
		var softFork *explorer.SoftFork
		if err := json.Unmarshal(hit.Source, &softFork); err != nil {
			log.WithError(err).Fatal("Failed to unmarshall soft fork")
		}
		softForks = append(softForks, softFork)
	}

	return softForks, nil
}
