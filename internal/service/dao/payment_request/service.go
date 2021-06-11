package payment_request

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
)

type Service interface {
	LoadVotingPaymentRequests(block *explorer.Block)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return service{repository}
}

func (s service) LoadVotingPaymentRequests(block *explorer.Block) {
	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	paymentRequests, _ := s.repository.GetPossibleVotingRequests(excludeOlderThan)
	zap.S().Infof("Load Voting Payment Requests (%d)", len(paymentRequests))

	PaymentRequests = paymentRequests
}
