package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type PaymentRequest struct {
	id string

	Version             uint32  `json:"version"`
	Hash                string  `json:"hash"`
	BlockHash           string  `json:"blockHash"`
	ProposalHash        string  `json:"proposalHash,omitempty"`
	Description         string  `json:"description"`
	RequestedAmount     float64 `json:"requestedAmount"`
	Status              string  `json:"status"`
	State               uint    `json:"state"`
	StateChangedOnBlock string  `json:"stateChangedOnBlock,omitempty"`

	Height         uint64 `json:"height"`
	UpdatedOnBlock uint64 `json:"updatedOnBlock"`

	VotesYes    uint `json:"votesYes"`
	VotesAbs    uint `json:"votesAbs"`
	VotesNo     uint `json:"votesNo"`
	VotingCycle uint `json:"votingCycle"`
}

func (p *PaymentRequest) Id() string {
	return p.id
}

func (p *PaymentRequest) SetId(id string) {
	p.id = id
}

func (p *PaymentRequest) Slug() string {
	return slug.Make(fmt.Sprintf("paymentrequest-%s", p.Hash))
}

func (p *PaymentRequest) GetHeight() uint64 {
	return p.Height
}
