package consultation

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/getsentry/raven-go"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetOpenConsultations(height uint64) ([]explorer.Consultation, error)
	GetPassedConsultations(maxHeight uint64) ([]*explorer.Consultation, error)
}

type repository struct {
	elastic elastic_cache.Index
}

func NewRepo(elastic elastic_cache.Index) Repository {
	return repository{elastic}
}

func (r repository) GetOpenConsultations(height uint64) ([]explorer.Consultation, error) {
	var consultations []explorer.Consultation

	openStatuses := make([]interface{}, 4)
	openStatuses[0] = explorer.ConsultationPending.Status
	openStatuses[1] = explorer.ConsultationFoundSupport.Status
	openStatuses[2] = explorer.ConsultationReflection.Status
	openStatuses[3] = explorer.ConsultationVotingStarted.Status

	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewTermsQuery("status.keyword", openStatuses...))
	query = query.Should(elastic.NewRangeQuery("updatedOnBlock").Gte(height))

	results, err := r.elastic.GetClient().Search(elastic_cache.DaoConsultationIndex.Get()).
		Query(query).
		Size(9999).
		Sort("updatedOnBlock", false).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	if results != nil {
		for _, hit := range results.Hits.Hits {
			var consultation explorer.Consultation
			if err := json.Unmarshal(hit.Source, &consultation); err != nil {
				log.WithError(err).Fatal("Failed to unmarshall consultation")
			}
			consultations = append(consultations, consultation)
		}
	}

	return consultations, nil
}

func (r repository) GetPassedConsultations(maxHeight uint64) ([]*explorer.Consultation, error) {
	var consultations []*explorer.Consultation

	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewMatchQuery("state", explorer.ConsultationPassed.State))
	query = query.Should(elastic.NewRangeQuery("updatedOnBlock").Lte(maxHeight))

	results, err := r.elastic.GetClient().Search(elastic_cache.DaoConsultationIndex.Get()).
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
