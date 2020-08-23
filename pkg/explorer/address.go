package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type Address struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`

	Spending int64 `json:"spending"`
	Staking  int64 `json:"staking"`
	Voting   int64 `json:"voting"`

	Position uint64 `json:"position,omitempty"`
}

func (a *Address) Slug() string {
	return slug.Make(fmt.Sprintf("address-%s", a.Hash))
}
