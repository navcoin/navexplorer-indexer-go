package proposal

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

func (s *Service) LoadVotingProposals(block *explorer.Block) {
	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	proposals, _ := s.repo.GetPossibleVotingProposals(excludeOlderThan)
	log.Infof("Load Voting Proposals (%d)", len(proposals))

	Proposals = proposals
}
