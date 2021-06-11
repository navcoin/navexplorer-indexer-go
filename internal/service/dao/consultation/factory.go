package consultation

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"reflect"
)

func CreateConsultation(consultation navcoind.Consultation, tx *explorer.BlockTransaction) explorer.Consultation {
	c := explorer.Consultation{
		Version:             consultation.Version,
		Hash:                consultation.Hash,
		BlockHash:           consultation.BlockHash,
		Question:            consultation.Question,
		Support:             consultation.Support,
		Min:                 consultation.Min,
		Max:                 consultation.Max,
		State:               consultation.State,
		Status:              explorer.GetConsultationStatusByState(uint(consultation.State)).Status,
		FoundSupport:        false,
		StateChangedOnBlock: consultation.StateChangedOnBlock,
	}
	if tx != nil {
		c.Height = tx.Height
		c.UpdatedOnBlock = tx.Height
		c.ProposedBy = tx.Vin.First().Addresses[0]
	}

	if consultation.Version>>1&1 == 1 {
		c.AnswerIsARange = true
	}

	if consultation.Version>>2&1 == 1 {
		c.MoreAnswers = true
	}

	if consultation.Version>>3&1 == 1 {
		c.ConsensusParameter = true
	}

	createAnswers(consultation, &c)

	return c
}

func createAnswers(navC navcoind.Consultation, c *explorer.Consultation) {
	if c.AnswerIsARange {
		c.RangeAnswers = navC.RangeAnswers
	} else {
		answers := make([]explorer.Answer, 0)
		for _, a := range navC.Answers {
			answers = append(answers, createAnswer(a))
		}
		c.Answers = answers
	}
}

func createAnswer(a *navcoind.Answer) explorer.Answer {
	return explorer.Answer{
		Version:             a.Version,
		Answer:              a.Answer,
		Support:             a.Support,
		Votes:               a.Votes,
		State:               a.State,
		Status:              explorer.GetAnswerStatusByState(uint(a.State)).Status,
		StateChangedOnBlock: a.StateChangedOnBlock,
		FoundSupport:        false,
		TxBlockHash:         a.TxBlockHash,
		Parent:              a.Parent,
		Hash:                a.Hash,
		MapState:            a.MapState,
	}
}

func UpdateConsultation(navC navcoind.Consultation, c *explorer.Consultation, parameters explorer.ConsensusParameters) bool {
	updated := false
	if navC.Support != c.Support {
		c.Support = navC.Support
		updated = true
	}

	if navC.VotingCyclesFromCreation != c.VotingCyclesFromCreation {
		c.VotingCyclesFromCreation = navC.VotingCyclesFromCreation
		updated = true
	}

	if navC.VotingCycleForState.Current != c.VotingCycleForState {
		c.VotingCycleForState = navC.VotingCycleForState.Current
		updated = true
	}

	if c.AnswerIsARange {
		updated = updateRangeAnswers(navC, c)
	} else {
		updated = updateAnswers(navC, c, parameters)
	}

	if navC.State != c.State {
		c.State = navC.State
		c.Status = explorer.GetConsultationStatusByState(uint(c.State)).Status
		updated = true
	}

	if c.FoundSupport != c.HasAnswerWithSupport() {
		c.FoundSupport = c.HasAnswerWithSupport()
		updated = true
	}

	if navC.StateChangedOnBlock != c.StateChangedOnBlock {
		c.StateChangedOnBlock = navC.StateChangedOnBlock
		updated = true
	}

	if reflect.DeepEqual(navC.MapState, c.MapState) {
		c.MapState = navC.MapState
		updated = true
	}

	return updated
}

func updateRangeAnswers(navC navcoind.Consultation, c *explorer.Consultation) bool {
	updated := false

	c.RangeAnswers = make(map[string]int)
	for k, v := range navC.RangeAnswers {
		c.RangeAnswers[k] = v
		updated = true
	}
	c.Answers = nil

	return updated
}

func updateAnswers(navC navcoind.Consultation, c *explorer.Consultation, parameters explorer.ConsensusParameters) bool {
	updated := false
	for _, navA := range navC.Answers {
		a := getAnswer(c, navA.Hash)
		if a == nil {
			c.Answers = append(c.Answers, createAnswer(navA))
			updated = true
		} else {
			if a.Support != navA.Support {
				zap.L().Info("UpdateAnswer Support")
				a.Support = navA.Support
				updated = true
			}
			if a.StateChangedOnBlock != navA.StateChangedOnBlock {
				zap.L().Info("UpdateAnswer StateChangedOnBlock")
				a.StateChangedOnBlock = navA.StateChangedOnBlock
				updated = true
			}
			if a.State != navA.State {
				zap.L().Info("UpdateAnswer State")
				a.State = navA.State
				a.Status = explorer.GetAnswerStatusByState(uint(a.State)).Status
				updated = true
			}

			supported := a.Support >= AnswerSupportRequired(
				parameters.GetConsensusParameter(explorer.CONSULTATION_ANSWER_MIN_SUPPORT),
				parameters.GetConsensusParameter(explorer.VOTING_CYCLE_LENGTH))
			if a.FoundSupport != supported {
				zap.L().With(zap.String("consultation", navC.Hash), zap.Bool("supported", supported)).Debug("UpdateAnswer: Supported")
				a.FoundSupport = supported
				updated = true
			}
			if a.Votes != navA.Votes {
				zap.L().With(zap.String("consultation", navC.Hash), zap.Int("voted", navA.Votes)).Debug("UpdateAnswer: Votes")
				a.Votes = navA.Votes
				updated = true
			}
		}
	}

	return updated
}

func getAnswer(c *explorer.Consultation, hash string) *explorer.Answer {
	for _, a := range c.Answers {
		if a.Hash == hash {
			return &a
		}
	}

	return nil
}
