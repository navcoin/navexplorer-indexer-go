package consultation

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

func CreateConsultation(consultation navcoind.Consultation) *explorer.Consultation {
	return &explorer.Consultation{
		Version:             consultation.Version,
		Hash:                consultation.Hash,
		BlockHash:           consultation.BlockHash,
		Question:            consultation.Question,
		Support:             consultation.Support,
		Min:                 0,
		Max:                 0,
		VotingCycle:         0,
		Status:              "",
		State:               0,
		StateChangedOnBlock: "",
	}
}

func UpdateConsultation(c navcoind.Consultation, consultation *explorer.PaymentRequest) {

}

func CreateAnswers(c navcoind.Consultation, initial []string) []*explorer.Answer {
	answers := make([]*explorer.Answer, 0)
	for _, a := range c.Answers {
		if !isAnswerInitial(a.String, initial) {
			continue
		}

		answer := &explorer.Answer{
			Version: a.Version,
			String:  a.String,
			Status:  "waiting for support",
			State:   1,
			Parent:  a.Parent,
			Hash:    a.Hash,
		}

		answers = append(answers, answer)
	}

	return answers
}

func isAnswerInitial(a string, initial []string) bool {
	for _, i := range initial {
		if i == a {
			return true
		}
	}

	return false
}
