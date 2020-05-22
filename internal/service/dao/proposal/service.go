package proposal

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

func (s *Service) LoadVotingProposals(block *explorer.Block) {
	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	proposals, err := s.repo.GetPossibleVotingProposals(excludeOlderThan)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Error("Failed to load pending proposals")
	}
	log.Infof("Load Voting Proposals (%d)", len(proposals))

	Proposals = proposals
}
