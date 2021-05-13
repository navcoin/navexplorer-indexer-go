package proposal

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	LoadVotingProposals(block *explorer.Block)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return service{repository}
}

func (s service) LoadVotingProposals(block *explorer.Block) {
	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	proposals, _ := s.repository.GetPossibleVotingProposals(excludeOlderThan)
	log.Infof("Load Voting Proposals (%d)", len(proposals))

	Proposals = proposals
}
