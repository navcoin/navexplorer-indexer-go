package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
	"time"
)

type Address struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`

	Spending int64 `json:"spending"`
	Staking  int64 `json:"staking"`
	Voting   int64 `json:"voting"`

	CreatedTime  time.Time `json:"created_time"`
	CreatedBlock uint64    `json:"created_block"`

	// Transient
	RichList RichList `json:"rich_list,omitempty"`
}

type RichList struct {
	Spending uint64 `json:"spending"`
	Staking  uint64 `json:"staking"`
	Voting   uint64 `json:"voting"`
}

func (a *Address) Slug() string {
	return slug.Make(fmt.Sprintf("address-%s", a.Hash))
}
