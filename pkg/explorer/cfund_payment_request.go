package explorer

import "github.com/gosimple/slug"

type RawPaymentRequest struct {
	Version             uint32               `json:"version"`
	Hash                string               `json:"hash"`
	BlockHash           string               `json:"blockHash"`
	ProposalHash        string               `json:"proposalHash,omitempty"`
	Description         string               `json:"description"`
	RequestedAmount     float64              `json:"requestedAmount"`
	Status              PaymentRequestStatus `json:"status"`
	State               uint                 `json:"state"`
	StateChangedOnBlock string               `json:"stateChangedOnBlock,omitempty"`
}

type PaymentRequest struct {
	RawPaymentRequest
	Height         uint64 `json:"height"`
	UpdatedOnBlock uint64 `json:"updatedOnBlock"`
}

func (p *PaymentRequest) Slug() string {
	return slug.Make(p.Hash)
}

func (p *PaymentRequest) GetHeight() uint64 {
	return p.Height
}
