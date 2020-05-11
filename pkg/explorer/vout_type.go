package explorer

type VoutType string

var (
	VoutNonstandard               VoutType = "nonstandard"
	VoutNulldata                  VoutType = "nulldata"
	VoutPubkeyhash                VoutType = "pubkeyhash"
	VoutPubkey                    VoutType = "pubkey"
	VoutScripthash                VoutType = "scripthash"
	VoutColdStaking               VoutType = "cold_staking"
	VoutColdStakingV2             VoutType = "cold_staking_v2"
	VoutCfundContribution         VoutType = "cfund_contribution"
	VoutProposalNoVote            VoutType = "proposal_no_vote"
	VoutProposalYesVote           VoutType = "proposal_yes_vote"
	VoutProposalAbstainVote       VoutType = "proposal_abstain_vote"
	VoutPaymentRequestNoVote      VoutType = "payment_request_no_vote"
	VoutPaymentRequestYesVote     VoutType = "payment_request_yes_vote"
	VoutPaymentRequestAbstainVote VoutType = "payment_request_abstain_vote"
	VoutPoolStaking               VoutType = "pool_staking"
)

var VoutTypes = map[string]VoutType{
	"nonstandard":                  VoutNonstandard,
	"nulldata":                     VoutNulldata,
	"pubkeyhash":                   VoutPubkeyhash,
	"pubkey":                       VoutPubkey,
	"scripthash":                   VoutScripthash,
	"cold_staking":                 VoutColdStaking,
	"cold_staking_v2":              VoutColdStakingV2,
	"cfund_contribution":           VoutCfundContribution,
	"proposal_no_vote":             VoutProposalNoVote,
	"proposal_yes_vote":            VoutProposalYesVote,
	"proposal_abstain_vote":        VoutProposalAbstainVote,
	"payment_request_no_vote":      VoutPaymentRequestNoVote,
	"payment_request_yes_vote":     VoutPaymentRequestYesVote,
	"payment_request_abstain_vote": VoutPaymentRequestAbstainVote,
	"pool_staking":                 VoutPoolStaking,
}
