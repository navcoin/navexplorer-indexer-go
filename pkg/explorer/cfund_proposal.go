package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type Proposal struct {
	Version             uint32  `json:"version"`
	Hash                string  `json:"hash"`
	BlockHash           string  `json:"blockHash"`
	Description         string  `json:"description"`
	RequestedAmount     float64 `json:"requestedAmount"`
	NotPaidYet          float64 `json:"notPaidYet"`
	NotRequestedYet     float64 `json:"notRequestedYet"`
	UserPaidFee         float64 `json:"userPaidFee"`
	PaymentAddress      string  `json:"paymentAddress"`
	ProposalDuration    uint64  `json:"proposalDuration"`
	ExpiresOn           uint64  `json:"expiresOn,omitempty"`
	State               uint    `json:"state"`
	Status              string  `json:"status"`
	StateChangedOnBlock string  `json:"stateChangedOnBlock,omitempty"`
	Height              uint64  `json:"height"`
	UpdatedOnBlock      uint64  `json:"updatedOnBlock"`

	VotesYes    uint `json:"votesYes"`
	VotesAbs    uint `json:"votesAbs"`
	VotesNo     uint `json:"votesNo"`
	VotingCycle uint `json:"votingCycle"`
}

func (p *Proposal) Slug() string {
	return slug.Make(fmt.Sprintf("proposal-%s", p.Hash))
}

func (p *Proposal) GetHeight() uint64 {
	return p.Height
}
