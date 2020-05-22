package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type RawPaymentRequest struct {
}

type PaymentRequest struct {
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

func (p *PaymentRequest) Slug() string {
	return slug.Make(fmt.Sprintf("paymentrequest-%s", p.Hash))
}

func (p *PaymentRequest) GetHeight() uint64 {
	return p.Height
}
