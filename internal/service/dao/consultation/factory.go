package consultation

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
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
		Height:              tx.Height,
		UpdatedOnBlock:      tx.Height,
		ProposedBy:          tx.Vin.First().Addresses[0],
	}

	answers := createAnswers(consultation)
	if len(answers) != 0 {
		c.Answers = answers
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
		MapState:            a.MapState,
	}
}

func UpdateConsultation(navC navcoind.Consultation, c *explorer.Consultation) bool {
	updated := false
	if navC.Support != c.Support {
		log.WithFields(log.Fields{"from": c.Support, "to": navC.Support}).Debug("Support changed")
		c.Support = navC.Support
		updated = true
	}

	if navC.VotingCyclesFromCreation != c.VotingCyclesFromCreation {
		log.WithFields(log.Fields{"from": c.VotingCyclesFromCreation, "to": navC.VotingCyclesFromCreation}).Debug("VotingCyclesFromCreation changed")
		c.VotingCyclesFromCreation = navC.VotingCyclesFromCreation
		updated = true
	}

	if navC.VotingCycleForState.Current != c.VotingCycleForState {
		log.WithFields(log.Fields{"from": c.VotingCycleForState, "to": navC.VotingCycleForState}).Debug("VotingCycleForState changed")
		c.VotingCycleForState = navC.VotingCycleForState.Current
		updated = true
	}

	if updateAnswers(navC, c) {
		log.Debug("Answers changed")
		updated = true
	}

	if navC.State != c.State {
		log.WithFields(log.Fields{"from": c.State, "to": navC.State}).Debug("State changed")
		c.State = navC.State
		c.Status = explorer.GetConsultationStatusByState(uint(c.State)).Status
		updated = true
	}

	if c.FoundSupport != c.HasAnswerWithSupport() {
		log.WithFields(log.Fields{"from": c.FoundSupport, "to": c.HasAnswerWithSupport()}).Debug("FoundSupport changed")
		c.FoundSupport = c.HasAnswerWithSupport()
		updated = true
	}

	if navC.StateChangedOnBlock != c.StateChangedOnBlock {
		log.WithFields(log.Fields{"from": c.StateChangedOnBlock, "to": navC.StateChangedOnBlock}).Debug("StateChangedOnBlock changed")
		c.StateChangedOnBlock = navC.StateChangedOnBlock
		updated = true
	}

	if reflect.DeepEqual(navC.MapState, c.MapState) {
		log.Debug("MapState changed")
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
				log.Info("UpdateAnswer Support")
				a.Support = navA.Support
				updated = true
			}
			if a.StateChangedOnBlock != navA.StateChangedOnBlock {
				log.Info("UpdateAnswer StateChangedOnBlock")
				a.StateChangedOnBlock = navA.StateChangedOnBlock
				updated = true
			}
			if a.State != navA.State {
				log.Info("UpdateAnswer State")
				a.State = navA.State
				a.Status = explorer.GetAnswerStatusByState(uint(a.State)).Status
				updated = true
			}

			supported := a.Support >= AnswerSupportRequired()
			if a.FoundSupport != supported {
				log.Info("UpdateAnswer AnswerSupportRequired")
				a.FoundSupport = supported
				updated = true
			}
			if a.Votes != navA.Votes {
				log.Info("UpdateAnswer Votes")
				a.Votes = navA.Votes
				updated = true
			}
			//if reflect.DeepEqual(navA.MapState, a.MapState) {
			//	log.Info("UpdateAnswer MapState")
			//	a.MapState = navA.MapState
			//	updated = true
			//}
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
