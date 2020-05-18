package dao

import (
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/consultation"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	elastic           *elastic_cache.Index
	consensusRewinder *consensus.Rewinder
	consultationRepo  *consultation.Repository
}

func NewRewinder(elastic *elastic_cache.Index, consensusRewinder *consensus.Rewinder, consultationRepo *consultation.Repository) *Rewinder {
	return &Rewinder{elastic, consensusRewinder, consultationRepo}
}

func (r *Rewinder) Rewind(height uint64) error {
	log.Infof("Rewinding DAO index to height: %d", height)

	passedConsultations, err := r.consultationRepo.GetPassedConsultations(height)
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
