package address

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"time"
)

func CreateAddress(hash string) *explorer.Address {
	return &explorer.Address{Hash: hash}
}

func CreateAddressHistory(history *navcoind.AddressHistory, tx *explorer.BlockTransaction) *explorer.AddressHistory {
	h := &explorer.AddressHistory{
		Height:  history.Block,
		TxIndex: history.TxIndex,
		Time:    time.Unix(history.Time, 0),
		TxId:    history.TxId,
		Hash:    history.Address,
		Changes: explorer.AddressChanges{
			Spendable:    history.Changes.Balance,
			Stakable:     history.Changes.Stakable,
			VotingWeight: history.Changes.VotingWeight,
		},
		Balance: explorer.AddressBalance{
			Spendable:    history.Result.Balance,
			Stakable:     history.Result.Stakable,
			VotingWeight: history.Result.VotingWeight,
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
