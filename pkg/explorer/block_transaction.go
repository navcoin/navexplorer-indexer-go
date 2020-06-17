package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
	"time"
)

type RawBlockTransaction struct {
	Hex             string    `json:"hex"`
	Txid            string    `json:"txid"`
	Hash            string    `json:"hash"`
	Size            uint64    `json:"size"`
	VSize           uint64    `json:"vsize"`
	Version         uint32    `json:"version"`
	LockTime        uint32    `json:"locktime"`
	Strdzeel        string    `json:"strdzeel"`
	AnonDestination string    `json:"anon-destination"`
	BlockHash       string    `json:"blockhash, omitempty"`
	Height          uint64    `json:"height"`
	Confirmations   uint64    `json:"confirmations, omitempty"`
	Time            time.Time `json:"time, omitempty"`
	BlockTime       time.Time `json:"blocktime, omitempty"`

	Vin  RawVins  `json:"vin"`
	Vout RawVouts `json:"vout"`
}

type BlockTransaction struct {
	RawBlockTransaction
	Index uint  `json:"index"`
	Vin   Vins  `json:"vin"`
	Vout  Vouts `json:"vout"`

	Type  BlockTransactionType `json:"type"`
	Stake uint64               `json:"stake"`
	Spend uint64               `json:"spend"`
	Fees  uint64               `json:"fees"`
}

func (tx *BlockTransaction) Slug() string {
	return CreateBlockTxSlug(tx.Hash)
}

func CreateBlockTxSlug(hash string) string {
	return slug.Make(fmt.Sprintf("blocktx-%s", hash))
}

type BlockTransactions []*BlockTransaction

func (blockTransactions *BlockTransactions) GetCoinbase() *BlockTransaction {
	for _, tx := range *blockTransactions {
		if tx.IsCoinbase() {
			return tx
		}
	}

	return nil
}

func (tx *BlockTransaction) GetAllAddresses() []string {
	addresses := make([]string, 0)

	exists := func(address string, addresses []string) bool {
		for i := range addresses {
			if addresses[i] == address {
				return true
			}
		}
		return false
	}

	for _, vin := range tx.Vin {
		for _, address := range vin.Addresses {
			if !exists(address, addresses) {
				addresses = append(addresses, address)
			}
		}
	}

	for _, vout := range tx.Vout {
		for _, address := range vout.ScriptPubKey.Addresses {
			if !exists(address, addresses) {
				addresses = append(addresses, address)
			}
		}
	}

	return addresses
}

func (tx *BlockTransaction) IsCoinbase() bool {
	return !tx.Vin.Empty() && tx.Vin[0].IsCoinbase()
}

func (tx *BlockTransaction) IsSpend() bool {
	return tx.Type == TxSpend
}

func (tx *BlockTransaction) IsAnyStaking() bool {
	return tx.Type == TxColdStaking || tx.Type == TxColdStakingV2 || tx.Type == TxStaking || tx.Type == TxPoolStaking
}

func (tx *BlockTransaction) IsStaking() bool {
	return tx.Type == TxStaking
}

func (tx *BlockTransaction) IsColdStaking() bool {
	return tx.Type == TxColdStaking || tx.Type == TxColdStakingV2
}

func (tx *BlockTransaction) IsPoolStaking() bool {
	return tx.Type == TxPoolStaking
}

func (tx *BlockTransaction) HasColdInput(address string) bool {
	for _, i := range tx.Vin {
		if (i.PreviousOutput.Type == VoutColdStaking || i.PreviousOutput.Type == VoutColdStakingV2) && i.Addresses[0] == address {
			return true
		}
	}
	return false
}

func (tx *BlockTransaction) HasColdStakeStake(address string) bool {
	return len(tx.Vout) > 1 && isColdStaking(tx.Vout[1]) && tx.Vout[1].ScriptPubKey.Addresses[0] == address
}

func (tx *BlockTransaction) HasColdStakeSpend(address string) bool {
	for _, o := range tx.Vout {
		if isColdStaking(o) && o.ScriptPubKey.Addresses[1] == address {
			return true
		}
	}
	return false
}

func (tx *BlockTransaction) HasColdStakeReceive(address string) bool {
	if tx.IsSpend() {
		for _, o := range tx.Vout {
			if isColdStaking(o) && o.ScriptPubKey.Addresses[0] == address {
				return true
			}
		}
	}
	return false
}

func isColdStaking(o Vout) bool {
	return o.ScriptPubKey.Type == VoutColdStaking || o.ScriptPubKey.Type == VoutColdStakingV2
}
