package payment_request

import (
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo}
}

func (s *Service) LoadVotingPaymentRequests(block *explorer.Block) {
	log.Info("Load Voting Payment Requests")

	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	paymentRequests, err := s.repo.GetPossibleVotingRequests(excludeOlderThan)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Error("Failed to load pending proposals")
	}

	PaymentRequests = paymentRequests
}
