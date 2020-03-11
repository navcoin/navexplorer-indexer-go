package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type DaoVotes struct {
	//MetaData MetaData `json:"-"`

	Height  uint64 `json:"height"`
	Address string `json:"address"`
	Votes   []Vote `json:"votes"`
}

func (v *DaoVotes) Slug() string {
	return slug.Make(fmt.Sprintf("%d-%s", v.Height, v.Address))
}

type Vote struct {
	Type VoteType `json:"type"`
	Hash string   `json:"hash"`
	Vote int      `json:"vote"`
}

type VoteType string

var (
	ProposalVote       VoteType = "Proposal"
	PaymentRequestVote VoteType = "PaymentRequest"
	DaoSupport         VoteType = "DaoSupport"
	DaoVote            VoteType = "DaoVote"
)
