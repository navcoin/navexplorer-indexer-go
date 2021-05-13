package signal

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
)

type Repository interface {
	GetSignals(start uint64, end uint64) []explorer.Signal
}

type repository struct {
	elastic elastic_cache.Index
}

func NewRepo(elastic elastic_cache.Index) Repository {
	return repository{elastic}
}

func (r repository) GetSignals(start uint64, end uint64) []explorer.Signal {
	signals := make([]explorer.Signal, 0)

	results, err := r.elastic.GetClient().Search(elastic_cache.SignalIndex.Get()).
		Sort("height", true).
		Query(elastic.NewRangeQuery("height").Gte(start).Lte(end)).
		Size(int(end - start + 1)).
		Do(context.Background())

	if err == nil && results != nil && results.Hits != nil {
		for _, hit := range results.Hits.Hits {
			var signal explorer.Signal
			if err := json.Unmarshal(hit.Source, &signal); err == nil {
				signals = append(signals, signal)
			}
		}
	}

	return signals
}
