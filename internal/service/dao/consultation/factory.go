package consultation

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

func CreateConsultation(consultation navcoind.Consultation) *explorer.Consultation {
	c := &explorer.Consultation{
		Version:             consultation.Version,
		Hash:                consultation.Hash,
		BlockHash:           consultation.BlockHash,
		Question:            consultation.Question,
		Support:             consultation.Support,
		Min:                 consultation.Min,
		Max:                 consultation.Max,
		VotingCycle:         consultation.VotingCycle,
		State:               consultation.State,
		Status:              explorer.GetConsultationStatusByState(uint(consultation.State)).Status,
		StateChangedOnBlock: consultation.StateChangedOnBlock,
		Answers:             createAnswers(consultation),
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
		Status:              explorer.GetConsultationStatusByState(uint(a.State)).Status,
		StateChangedOnBlock: a.StateChangedOnBlock,
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

	if navC.VotingCycle != c.VotingCycle {
		c.VotingCycle = navC.VotingCycle
		updated = true
	}

	if navC.Status != c.Status {
		c.Status = navC.Status
		updated = true
	}

	if navC.State != c.State {
		c.State = navC.State
		updated = true
	}

	if navC.StateChangedOnBlock != c.StateChangedOnBlock {
		c.StateChangedOnBlock = navC.StateChangedOnBlock
		updated = true
	}

	if updateAnswers(navC, c) {
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
				updated = true
			}
			if a.Status != navA.Status {
				a.Status = navA.Status
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
