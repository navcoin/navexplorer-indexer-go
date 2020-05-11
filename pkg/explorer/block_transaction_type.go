package explorer

type BlockTransactionType string

var (
	TxCoinbase      BlockTransactionType = "coinbase"
	TxStaking       BlockTransactionType = "staking"
	TxColdStaking   BlockTransactionType = "cold_staking"
	TxColdStakingV2 BlockTransactionType = "cold_staking_v2"
	TxPoolStaking   BlockTransactionType = "pool_staking"
	TxSpend         BlockTransactionType = "spend"
)
