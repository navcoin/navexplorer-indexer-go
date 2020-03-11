package dao

import (
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/consensus"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	elastic           *elastic_cache.Index
	consensusRewinder *consensus.Rewinder
}

func NewRewinder(elastic *elastic_cache.Index, consensusRewinder *consensus.Rewinder) *Rewinder {
	return &Rewinder{elastic, consensusRewinder}
}

func (r *Rewinder) Rewind(height uint64) error {
	log.Infof("Rewinding DAO index to height: %d", height)

	if err := r.consensusRewinder.Rewind(); err != nil {
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
