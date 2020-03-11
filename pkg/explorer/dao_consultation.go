package explorer

import (
	"github.com/gosimple/slug"
)

type Consultation struct {
	Version             uint32 `json:"version"`
	Hash                string `json:"hash"`
	BlockHash           string `json:"blockhash"`
	Question            string `json:"question"`
	Support             int    `json:"support"`
	Min                 int    `json:"min"`
	Max                 int    `json:"max"`
	VotingCycle         int    `json:"votingCycle"`
	Status              string `json:"status"`
	State               int    `json:"state"`
	StateChangedOnBlock string `json:"stateChangedOnBlock"`
}

func (c *Consultation) Slug() string {
	return slug.Make(c.Hash)
}

type Answer struct {
	Abstain int `json:"abstain,omitempty"`

	Version   uint32 `json:"version,omitempty"`
	String    string `json:"string,omitempty"`
	Support   int    `json:"support,omitempty"`
	Votes     int    `json:"votes,omitempty"`
	Status    string `json:"status,omitempty"`
	State     int    `json:"state,omitempty"`
	Parent    string `json:"parent,omitempty"`
	Hash      string `json:"hash,omitempty"`
	BlockHash string `json:"blockhash"`
}
