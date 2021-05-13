package dao

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consultation"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Rewinder interface {
	Rewind(height uint64) error
}

type rewinder struct {
	elastic           elastic_cache.Index
	consensusRewinder consensus.Rewinder
	repository        consultation.Repository
}

func NewRewinder(elastic elastic_cache.Index, consensusRewinder consensus.Rewinder, repository consultation.Repository) Rewinder {
	return rewinder{elastic, consensusRewinder, repository}
}

func (r rewinder) Rewind(height uint64) error {
	log.Infof("Rewinding DAO index to height: %d", height)

	passedConsultations, err := r.repository.GetPassedConsultations(height)
	if err != nil {
		return err
	}

	if err := r.consensusRewinder.Rewind(passedConsultations); err != nil {
		raven.CaptureError(err, nil)
		return err
	}

	return r.elastic.DeleteHeightGT(height,
		elastic_cache.DaoConsultationIndex.Get(),
		elastic_cache.ProposalIndex.Get(),
		elastic_cache.PaymentRequestIndex.Get(),
		elastic_cache.DaoVoteIndex.Get(),
	)
}
