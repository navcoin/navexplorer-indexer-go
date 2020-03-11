package explorer

import (
	"github.com/gosimple/slug"
)

type Address struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`

	Received      uint64 `json:"received"`
	ReceivedCount uint   `json:"receivedCount"`
	Sent          uint64 `json:"sent"`
	SentCount     uint   `json:"sentCount"`
	Staked        uint64 `json:"staked"`
	StakedCount   uint   `json:"stakedCount"`
	Balance       uint64 `json:"balance"`

	ColdReceived      uint64 `json:"coldReceived"`
	ColdReceivedCount uint   `json:"coldReceivedCount"`
	ColdSent          uint64 `json:"coldSent"`
	ColdSentCount     uint   `json:"coldSentCount"`
	ColdStaked        uint64 `json:"coldStaked"`
	ColdStakedCount   uint   `json:"coldStakedCount"`
	ColdBalance       uint64 `json:"coldBalance"`

	Position uint `json:"position"`
}

func (a *Address) Slug() string {
	return slug.Make(a.Hash)
}
