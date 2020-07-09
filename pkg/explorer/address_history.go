package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
	"time"
)

type AddressHistory struct {
	Height  uint64                `json:"height"`
	TxIndex uint                  `json:"txindex"`
	Time    time.Time             `json:"time"`
	TxId    string                `json:"txid"`
	Address string                `json:"address"`
	Changes AddressHistoryChanges `json:"changes"`
	Result  AddressHistoryResult  `json:"result"`
}

type AddressHistoryChanges struct {
	Balance      int64 `json:"balance"`
	Stakable     int64 `json:"stakable"`
	VotingWeight int64 `json:"voting_weight"`
	Flags        uint  `json:"flags"`
}

type AddressHistoryResult struct {
	Balance      int64 `json:"balance"`
	Stakable     int64 `json:"stakable"`
	VotingWeight int64 `json:"voting_weight"`
}

func (a *AddressHistory) Slug() string {
	return slug.Make(fmt.Sprintf("addresshistory-%s-%s-%t", a.Address, a.TxId))
}
