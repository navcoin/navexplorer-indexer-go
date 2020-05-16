package explorer

import (
	"github.com/gosimple/slug"
)

type Consultation struct {
	Version                  uint32         `json:"version"`
	Hash                     string         `json:"hash"`
	BlockHash                string         `json:"blockHash"`
	Question                 string         `json:"question"`
	Support                  int            `json:"support"`
	Abstain                  int            `json:"abstain,omitempty"`
	Answers                  []*Answer      `json:"answers"`
	Min                      int            `json:"min"`
	Max                      int            `json:"max"`
	VotingCyclesFromCreation int            `json:"votingCyclesFromCreation"`
	VotingCycleForState      int            `json:"votingCycleForState"`
	State                    int            `json:"state"`
	Status                   string         `json:"status"`
	FoundSupport             bool           `json:"foundSupport,omitempty"`
	StateChangedOnBlock      string         `json:"stateChangedOnBlock"`
	Height                   uint64         `json:"height"`
	UpdatedOnBlock           uint64         `json:"updatedOnBlock"`
	ProposedBy               string         `json:"proposedBy"`
	MapState                 map[string]int `json:"mapState"`

	AnswerIsARange     bool `json:"answerIsARange"`
	MoreAnswers        bool `json:"moreAnswers"`
	ConsensusParameter bool `json:"consensusParameter"`
}

func (c *Consultation) Slug() string {
	return slug.Make(c.Hash)
}

func (c *Consultation) HasAnswerWithSupport() bool {
	for _, a := range c.Answers {
		if a.FoundSupport == true {
			return true
		}
	}

	return false
}

func (c *Consultation) HasPassedAnswer() bool {
	if uint(c.State) != ConsultationPassed.State {
		return false
	}
	for _, a := range c.Answers {
		if uint(a.State) == AnswerPassed.State {
			return true
		}
	}

	return false
}

func (c *Consultation) GetPassedAnswer() *Answer {
	if uint(c.State) != ConsultationPassed.State {
		return nil
	}
	for _, a := range c.Answers {
		if uint(a.State) == AnswerPassed.State {
			return a
		}
	}

	return nil
}

type Answer struct {
	Version             uint32         `json:"version"`
	Answer              string         `json:"answer"`
	Support             int            `json:"support"`
	Votes               int            `json:"votes"`
	State               int            `json:"state"`
	Status              string         `json:"status"`
	FoundSupport        bool           `json:"foundSupport"`
	StateChangedOnBlock string         `json:"stateChangedOnBlock"`
	TxBlockHash         string         `json:"txblockhash"`
	Parent              string         `json:"parent"`
	Hash                string         `json:"hash"`
	MapState            map[string]int `json:"mapState"`
}
