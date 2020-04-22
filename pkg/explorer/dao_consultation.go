package explorer

import (
	"github.com/gosimple/slug"
)

type Consultation struct {
	Version             uint32    `json:"version"`
	Hash                string    `json:"hash"`
	BlockHash           string    `json:"blockHash"`
	Question            string    `json:"question"`
	Support             int       `json:"support"`
	Abstain             int       `json:"abstain,omitempty"`
	Answers             []*Answer `json:"answers"`
	Min                 int       `json:"min"`
	Max                 int       `json:"max"`
	VotingCycle         int       `json:"votingCycle"`
	State               int       `json:"state"`
	Status              string    `json:"status"`
	StateChangedOnBlock string    `json:"stateChangedOnBlock"`
	Height              uint64    `json:"height"`
	UpdatedOnBlock      uint64    `json:"updatedOnBlock"`
}

func (c *Consultation) Slug() string {
	return slug.Make(c.Hash)
}

type Answer struct {
	Version             uint32 `json:"version,omitempty"`
	Answer              string `json:"answer,omitempty"`
	Support             int    `json:"support,omitempty"`
	Votes               int    `json:"votes,omitempty"`
	State               int    `json:"state,omitempty"`
	Status              string `json:"status,omitempty"`
	StateChangedOnBlock string `json:"stateChangedOnBlock"`
	TxBlockHash         string `json:"txblockhash"`
	Parent              string `json:"parent,omitempty"`
	Hash                string `json:"hash,omitempty"`
}
