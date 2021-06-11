package payment_request

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"strconv"
)

func CreatePaymentRequest(paymentRequest navcoind.PaymentRequest, height uint64) *explorer.PaymentRequest {
	return &explorer.PaymentRequest{
		Version:             paymentRequest.Version,
		Hash:                paymentRequest.Hash,
		BlockHash:           paymentRequest.BlockHash,
		ProposalHash:        paymentRequest.ProposalHash,
		Description:         paymentRequest.Description,
		RequestedAmount:     convertStringToFloat(paymentRequest.RequestedAmount),
		State:               paymentRequest.State,
		Status:              explorer.GetPaymentRequestStatusByState(paymentRequest.State).Status,
		StateChangedOnBlock: paymentRequest.StateChangedOnBlock,

		Height:         height,
		UpdatedOnBlock: height,

		VotesYes:    paymentRequest.VotesYes,
		VotesAbs:    paymentRequest.VotesAbs,
		VotesNo:     paymentRequest.VotesNo,
		VotingCycle: paymentRequest.VotingCycle,
	}
}

func UpdatePaymentRequest(paymentRequest navcoind.PaymentRequest, height uint64, p *explorer.PaymentRequest) {
	if p.State != paymentRequest.State {
		p.State = paymentRequest.State
		p.Status = explorer.GetPaymentRequestStatusByState(p.State).Status
		p.UpdatedOnBlock = height
	}

	if p.StateChangedOnBlock != paymentRequest.StateChangedOnBlock {
		p.StateChangedOnBlock = paymentRequest.StateChangedOnBlock
		p.UpdatedOnBlock = height
	}

	if p.VotesYes != paymentRequest.VotesYes {
		p.VotesYes = paymentRequest.VotesYes
		p.UpdatedOnBlock = height
	}

	if p.VotesAbs != paymentRequest.VotesAbs {
		p.VotesAbs = paymentRequest.VotesAbs
		p.UpdatedOnBlock = height
	}

	if p.VotesNo != paymentRequest.VotesNo {
		p.VotesNo = paymentRequest.VotesNo
		p.UpdatedOnBlock = height
	}

	if p.VotingCycle != paymentRequest.VotingCycle {
		p.VotingCycle = paymentRequest.VotingCycle
		p.UpdatedOnBlock = height
	}
}

func convertStringToFloat(input string) float64 {
	output, err := strconv.ParseFloat(input, 64)
	if err != nil {
		zap.S().With(zap.Error(err)).Errorf("Unable to convert %s to uint64", input)
		return 0
	}

	return output
}
