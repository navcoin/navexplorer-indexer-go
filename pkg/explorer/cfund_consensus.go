package explorer

type Consensus struct {
	BlocksPerVotingCycle                uint    `json:"blocksPerVotingCycle"`
	MinSumVotesPerVotingCycle           uint    `json:"minSumVotesPerVotingCycle"`
	MaxCountVotingCycleProposals        uint    `json:"maxCountVotingCycleProposals"`
	MaxCountVotingCyclePaymentRequests  uint    `json:"maxCountVotingCyclePaymentRequests"`
	VotesAcceptProposalPercentage       uint    `json:"votesAcceptProposalPercentage"`
	VotesRejectProposalPercentage       uint    `json:"votesRejectProposalPercentage"`
	VotesAcceptPaymentRequestPercentage uint    `json:"votesAcceptPaymentRequestPercentage"`
	VotesRejectPaymentRequestPercentage uint    `json:"votesRejectPaymentRequestPercentage"`
	ProposalMinimalFee                  float64 `json:"proposalMinimalFee"`
}
