package payment_request

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetPossibleVotingRequests(height uint64) ([]*explorer.PaymentRequest, error)
}

type repository struct {
	elastic elastic_cache.Index
}

func NewRepo(elastic elastic_cache.Index) Repository {
	return repository{elastic}
}

func (r repository) GetPossibleVotingRequests(height uint64) ([]*explorer.PaymentRequest, error) {
	var paymentRequests []*explorer.PaymentRequest

	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewMatchQuery("status", "pending accepted"))
	query = query.Should(elastic.NewRangeQuery("updatedOnBlock").Gte(height))

	results, err := r.elastic.GetClient().Search(elastic_cache.PaymentRequestIndex.Get()).
		Query(query).
		Size(9999).
		Sort("updatedOnBlock", false).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	if results != nil {
		for _, hit := range results.Hits.Hits {
			var paymentRequest *explorer.PaymentRequest
			if err := json.Unmarshal(hit.Source, &paymentRequest); err != nil {
				log.WithError(err).Fatal("Failed to unmarshall payment request")
			}
			paymentRequests = append(paymentRequests, paymentRequest)
		}
	}

	return paymentRequests, nil
}
