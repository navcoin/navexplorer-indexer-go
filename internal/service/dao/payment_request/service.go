package payment_request

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo}
}

func (s *Service) LoadVotingPaymentRequests(block *explorer.Block) {
	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	paymentRequests, _ := s.repo.GetPossibleVotingRequests(excludeOlderThan)
	log.Infof("Load Voting Payment Requests (%d)", len(paymentRequests))

	PaymentRequests = paymentRequests
}
