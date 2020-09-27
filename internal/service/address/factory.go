package address

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"time"
)

func CreateAddress(hash string, height uint64, time time.Time) *explorer.Address {
	return &explorer.Address{Hash: hash, CreatedBlock: height, CreatedTime: time}
}

func CreateAddressHistory(history *navcoind.AddressHistory, tx *explorer.BlockTransaction, block *explorer.Block) *explorer.AddressHistory {
	h := &explorer.AddressHistory{
		Height:  history.Block,
		TxIndex: history.TxIndex,
		Time:    time.Unix(history.Time, 0),
		TxId:    history.TxId,
		Hash:    history.Address,
		Changes: explorer.AddressChanges{
			Spending: history.Changes.Balance,
			Staking:  history.Changes.Stakable,
			Voting:   history.Changes.VotingWeight,
		},
		Balance: explorer.AddressBalance{
			Spending: history.Result.Balance,
			Staking:  history.Result.Stakable,
			Voting:   history.Result.VotingWeight,
		},
	}

	hasPubKeyHashOutput := func() bool {
		for _, v := range tx.Vout.WithAddress(h.Hash) {
			if v.ScriptPubKey.Type == explorer.VoutPubkeyhash {
				return true
			}
		}
		return false
	}
	if history.Changes.Flags == 1 {
		h.CfundPayout = tx.Type == explorer.TxCoinbase && tx.Version == 3 && hasPubKeyHashOutput()

		if tx.Vout.Count() > 1 && !tx.Vout[1].HasAddress(h.Hash) {
			h.StakePayout = true
		} else {
			h.Stake = true
		}
	}

	if h.IsSpend() {
		switch tx.Version {
		case 4:
			h.Changes.Proposal = true
		case 5:
			h.Changes.PaymentRequest = true
		case 6:
			h.Changes.Consultation = true
		}
	}

	return h
}
