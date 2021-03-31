package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
	"time"
)

type Address struct {
	id string

	Hash   string `json:"hash"`
	Height uint64 `json:"height"`

	Spendable    int64 `json:"spendable"`
	Stakable     int64 `json:"stakable"`
	VotingWeight int64 `json:"voting_weight"`

	CreatedTime  time.Time `json:"created_time"`
	CreatedBlock uint64    `json:"created_block"`

	MultiSig *MultiSig `json:"multisig,omitempty"`

	// Transient
	RichList RichList `json:"rich_list,omitempty"`
}

type RichList struct {
	Spendable    uint64 `json:"spendable"`
	Stakable     uint64 `json:"stakable"`
	VotingWeight uint64 `json:"voting_weight"`
}

func (a *Address) Id() string {
	return a.id
}

func (a *Address) SetId(id string) {
	a.id = id
}

func (a *Address) Slug() string {
	return slug.Make(fmt.Sprintf("address-%s", a.Hash))
}
