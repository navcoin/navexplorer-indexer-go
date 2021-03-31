package explorer

type VoutType string

var (
	VoutNonstandard                VoutType = "nonstandard"
	VoutPubkey                     VoutType = "pubkey"
	VoutPubkeyhash                 VoutType = "pubkeyhash"
	VoutScripthash                 VoutType = "scripthash"
	VoutMultiSig                   VoutType = "address_multisig"
	VoutNulldata                   VoutType = "nulldata"
	VoutCfundContribution          VoutType = "cfund_contribution"
	VoutProposalYesVote            VoutType = "proposal_yes_vote"
	VoutPaymentRequestYesVote      VoutType = "payment_request_yes_vote"
	VoutProposalNoVote             VoutType = "proposal_no_vote"
	VoutPaymentRequestNoVote       VoutType = "payment_request_no_vote"
	VoutProposalAbstainVote        VoutType = "proposal_abstain_vote"
	VoutProposalRemoveVote         VoutType = "proposal_remove_vote"
	VoutPaymentRequestAbstainVote  VoutType = "payment_request_abstain_vote"
	VoutPaymentRequestRemoveVote   VoutType = "payment_request_remove_vote"
	VoutConsultationVote           VoutType = "consultation_vote"
	VoutConsultationVoteRemove     VoutType = "consultation_vote_remove"
	VoutConsultationVoteAbstention VoutType = "consultation_vote_abstention"
	VoutDaoSupport                 VoutType = "dao_support"
	VoutDaoSupportRemove           VoutType = "dao_support_remove"
	VoutColdStaking                VoutType = "cold_staking"
	VoutColdStakingV2              VoutType = "cold_staking_v2"
	VoutPoolStaking                VoutType = "pool_staking"
)

var VoutTypes = map[string]VoutType{
	"nonstandard":                  VoutNonstandard,
	"pubkey":                       VoutPubkey,
	"pubkeyhash":                   VoutPubkeyhash,
	"scripthash":                   VoutScripthash,
	"address_multisig":             VoutMultiSig,
	"nulldata":                     VoutNulldata,
	"cfund_contribution":           VoutCfundContribution,
	"proposal_yes_vote":            VoutProposalYesVote,
	"payment_request_yes_vote":     VoutPaymentRequestYesVote,
	"proposal_no_vote":             VoutProposalNoVote,
	"payment_request_no_vote":      VoutPaymentRequestNoVote,
	"proposal_abstain_vote":        VoutProposalAbstainVote,
	"proposal_remove_vote":         VoutProposalRemoveVote,
	"payment_request_abstain_vote": VoutPaymentRequestAbstainVote,
	"payment_request_remove_vote":  VoutPaymentRequestRemoveVote,
	"consultation_vote":            VoutConsultationVote,
	"consultation_vote_remove":     VoutConsultationVoteRemove,
	"consultation_vote_abstention": VoutConsultationVoteAbstention,
	"dao_support":                  VoutDaoSupport,
	"dao_support_remove":           VoutDaoSupportRemove,
	"cold_staking":                 VoutColdStaking,
	"cold_staking_v2":              VoutColdStakingV2,
	"pool_staking":                 VoutPoolStaking,
}
