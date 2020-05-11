package consultation

import (
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
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
	log.Debug("Load Open Consultations")

	excludeOlderThan := block.Height - (uint64(block.BlockCycle.Size * 2))
	if excludeOlderThan < 0 {
		excludeOlderThan = 0
	}

	consultations, err := s.repo.GetOpenConsultations(excludeOlderThan)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Error("Failed to load consultations")
	}

	Consultations = consultations
}

func ConsultationSupportRequired() int {
	minSupport := float64(consensus.Parameters.Get(consensus.CONSULTATION_MIN_SUPPORT).Value)
	cycleSize := float64(consensus.Parameters.Get(consensus.VOTING_CYCLE_LENGTH).Value)

	return int(math.Ceil((minSupport / 10000) * cycleSize))
}

func AnswerSupportRequired() int {
	minSupport := float64(consensus.Parameters.Get(consensus.CONSULTATION_ANSWER_MIN_SUPPORT).Value)
	cycleSize := float64(consensus.Parameters.Get(consensus.VOTING_CYCLE_LENGTH).Value)

	return int(math.Ceil((minSupport / 10000) * cycleSize))
}
