package consultation

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"reflect"
)

func CreateConsultation(consultation navcoind.Consultation, tx *explorer.BlockTransaction) *explorer.Consultation {
	c := &explorer.Consultation{
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
		Answers:             createAnswers(consultation),
		Height:              tx.Height,
		UpdatedOnBlock:      tx.Height,
		ProposedBy:          tx.Vin.First().Addresses[0],
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

	return c
}

func createAnswers(c navcoind.Consultation) []*explorer.Answer {
	answers := make([]*explorer.Answer, 0)
	for _, a := range c.Answers {
		answers = append(answers, createAnswer(a))
	}

	return answers
}

func createAnswer(a *navcoind.Answer) *explorer.Answer {
	return &explorer.Answer{
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
	}
}

func UpdateConsultation(navC navcoind.Consultation, c *explorer.Consultation) bool {
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

	if updateAnswers(navC, c) {
		updated = true
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

func updateAnswers(navC navcoind.Consultation, c *explorer.Consultation) bool {
	updated := false
	for _, navA := range navC.Answers {
		a := getAnswer(c, navA.Hash)
		if a == nil {
			c.Answers = append(c.Answers, createAnswer(navA))
			updated = true
		} else {
			if a.Support != navA.Support {
				a.Support = navA.Support
				updated = true
			}
			if a.StateChangedOnBlock != navA.StateChangedOnBlock {
				a.StateChangedOnBlock = navA.StateChangedOnBlock
				updated = true
			}
			if a.State != navA.State {
				a.State = navA.State
				a.Status = explorer.GetAnswerStatusByState(uint(a.State)).Status
				updated = true
			}

			supported := a.Support >= AnswerSupportRequired()
			if a.FoundSupport != supported {
				a.FoundSupport = supported
				updated = true
			}
			if a.Votes != navA.Votes {
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
			return a
		}
	}

	return nil
}
