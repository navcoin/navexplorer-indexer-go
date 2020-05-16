package consultation

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type Repository struct {
	Client *elastic.Client
}

func NewRepo(client *elastic.Client) *Repository {
	return &Repository{client}
}

func (r *Repository) GetOpenConsultations(height uint64) ([]*explorer.Consultation, error) {
	var consultations []*explorer.Consultation

	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewMatchQuery("state", explorer.ConsultationPending.State))
	query = query.Should(elastic.NewRangeQuery("updatedOnBlock").Gte(height))

	results, err := r.Client.Search(elastic_cache.DaoConsultationIndex.Get()).
		Query(query).
		Size(9999).
		Sort("updatedOnBlock", false).
		Do(context.Background())
	if err != nil {
		raven.CaptureError(err, nil)
		return nil, err
	}

	if results != nil {
		for _, hit := range results.Hits.Hits {
			var consultation *explorer.Consultation
			if err := json.Unmarshal(hit.Source, &consultation); err != nil {
				log.WithError(err).Fatal("Failed to unmarshall consultation")
			}
			consultations = append(consultations, consultation)
		}
	}

	return consultations, nil
}

func (r *Repository) GetPassedConsultations(maxHeight uint64) ([]*explorer.Consultation, error) {
	var consultations []*explorer.Consultation

	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewMatchQuery("state", explorer.ConsultationPassed.State))
	query = query.Should(elastic.NewRangeQuery("updatedOnBlock").Lte(maxHeight))

	results, err := r.Client.Search(elastic_cache.DaoConsultationIndex.Get()).
		Query(query).
		Size(9999).
		Sort("updatedOnBlock", true).
		Do(context.Background())
	if err != nil {
		raven.CaptureError(err, nil)
		return nil, err
	}

	if results != nil {
		for _, hit := range results.Hits.Hits {
			var consultation *explorer.Consultation
			if err := json.Unmarshal(hit.Source, &consultation); err != nil {
				log.WithError(err).Fatal("Failed to unmarshall consultation")
			}
			if consultation.HasPassedAnswer() {
				consultations = append(consultations, consultation)
			}
		}
	}

	return consultations, nil
}
