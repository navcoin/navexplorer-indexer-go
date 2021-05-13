package softfork

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetSoftForks() (explorer.SoftForks, error)
	GetSoftFork(name string) (*explorer.SoftFork, error)
}

type repository struct {
	elastic elastic_cache.Index
}

func NewRepo(elastic elastic_cache.Index) Repository {
	return repository{elastic}
}

func (r repository) GetSoftForks() (explorer.SoftForks, error) {
	var softForks []*explorer.SoftFork

	results, err := r.elastic.GetClient().Search(elastic_cache.SoftForkIndex.Get()).
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

func (r repository) GetSoftFork(name string) (*explorer.SoftFork, error) {
	var softfork *explorer.SoftFork

	results, err := r.elastic.GetClient().Search(elastic_cache.SoftForkIndex.Get()).
		Query(elastic.NewTermQuery("name", name)).
		Size(1).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	if results == nil || len(results.Hits.Hits) != 1 {
		return nil, errors.New("Invalid result found")
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &softfork)
	if err != nil {
		return nil, err
	}

	return softfork, nil
}
