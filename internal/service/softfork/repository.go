package softfork

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type Repository struct {
	Client *elastic.Client
}

func NewRepo(client *elastic.Client) *Repository {
	return &Repository{client}
}

func (r *Repository) GetSoftForks() ([]*explorer.SoftFork, error) {
	var softForks []*explorer.SoftFork

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

func (r *Repository) GetSoftFork(name string) (*explorer.SoftFork, error) {
	var softfork *explorer.SoftFork

	results, err := r.Client.Search(elastic_cache.SoftForkIndex.Get()).
		Query(elastic.NewTermQuery("name", name)).
		Size(1).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	if results == nil || len(results.Hits.Hits) != 1 {
		return nil, errors.New("Invalid result found")
	}

	err = json.Unmarshal(results.Hits.Hits[0].Source, &softfork)
	if err != nil {
		return nil, err
	}

	return softfork, nil
}
