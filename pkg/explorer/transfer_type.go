package explorer

type TransferType string

var (
	TransferSend                TransferType = "send"
	TransferReceive             TransferType = "receive"
	TransferStake               TransferType = "stake"
	TransferColdStake           TransferType = "cold_stake"
	TransferDelegateStake       TransferType = "delegate_stake"
	TransferColdDelegateStake   TransferType = "cold_delegate_stake"
	TransferPoolStake           TransferType = "pool_stake"
	TransferPoolFee             TransferType = "pool_fee"
	TransferCommunityFund       TransferType = "community_fund"
	TransferCommunityFundPayout TransferType = "community_fund_payout"
)

func IsStake(tt TransferType) bool {
	return tt == TransferStake || tt == TransferDelegateStake || tt == TransferPoolStake
}

func IsColdStake(tt TransferType) bool {
	return tt == TransferColdStake || tt == TransferColdDelegateStake
}
