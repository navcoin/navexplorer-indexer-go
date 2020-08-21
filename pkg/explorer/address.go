package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type Address struct {
	Hash    string         `json:"hash"`
	Height  uint64         `json:"height"`
	Balance AddressBalance `json:"balance"`
}

func (a *Address) Slug() string {
	return slug.Make(fmt.Sprintf("address-%s", a.Hash))
}
