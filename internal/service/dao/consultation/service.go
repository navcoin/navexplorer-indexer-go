package consultation

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"math"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo}
}

func (s *Service) LoadOpenConsultations(block *explorer.Block) {
	log.Info("Load Open Consultations")

	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	consultations, err := s.repo.GetOpenConsultations(excludeOlderThan)
	if err != nil {
		log.WithError(err).Error("Failed to load consultations")
	}

	for _, c := range consultations {
		log.WithField("hash", c.Hash).Info("Loaded consultation")
		Consultations[c.Hash] = c
	}
}

func ConsultationSupportRequired() int {
	//minSupport := float64(consensus.Parameters.Get(consensus.CONSULTATION_MIN_SUPPORT).Value)
	//cycleSize := float64(consensus.Parameters.Get(consensus.VOTING_CYCLE_LENGTH).Value)
	//
	//return int(math.Ceil((minSupport / 10000) * cycleSize))
	return 0
}

func AnswerSupportRequired(minSupport *explorer.ConsensusParameter, votingCycleLength *explorer.ConsensusParameter) int {
	return int(math.Ceil((float64(minSupport.Value) / 10000) * float64(votingCycleLength.Value)))
}
