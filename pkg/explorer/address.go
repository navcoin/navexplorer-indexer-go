package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type Address struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`

	Received      int64 `json:"received"`
	ReceivedCount uint  `json:"receivedCount"`
	Sent          int64 `json:"sent"`
	SentCount     uint  `json:"sentCount"`
	Staked        int64 `json:"staked"`
	StakedCount   uint  `json:"stakedCount"`
	Balance       int64 `json:"balance"`

	ColdReceived      int64 `json:"coldReceived"`
	ColdReceivedCount uint  `json:"coldReceivedCount"`
	ColdSent          int64 `json:"coldSent"`
	ColdSentCount     uint  `json:"coldSentCount"`
	ColdStaked        int64 `json:"coldStaked"`
	ColdStakedCount   uint  `json:"coldStakedCount"`
	ColdBalance       int64 `json:"coldBalance"`

	Position uint `json:"position"`
}

func (a *Address) Slug() string {
	return slug.Make(fmt.Sprintf("address-%s", a.Hash))
}
