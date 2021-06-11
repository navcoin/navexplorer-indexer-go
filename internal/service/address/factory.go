package address

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"time"
)

func CreateAddress(hash string) explorer.Address {
	return explorer.Address{Hash: hash}
}

func CreateAddressHistory(order uint, history *navcoind.AddressHistory, tx *explorer.BlockTransaction) explorer.AddressHistory {
	h := explorer.AddressHistory{
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
		MultiSig: false,
		Order:    order,
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
			if tx.IsColdStaking() {
				h.ColdStake = true
			}
			h.Reward = explorer.AddressReward{}
			if history.Changes.Balance != 0 {
				h.Reward.Spendable = (float64(history.Result.Balance) / 100000000) / (float64(history.Changes.Balance) / 100000000) / 100
			}
			if history.Changes.Stakable != 0 {
				h.Reward.Stakable = (float64(history.Result.Stakable) / 100000000) / (float64(history.Changes.Stakable) / 100000000) / 100
			}
			if history.Changes.VotingWeight != 0 {
				h.Reward.VotingWeight = (float64(history.Result.VotingWeight) / 100000000) / (float64(history.Changes.VotingWeight) / 100000000) / 100
			}
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

func CreateMultiSigAddressHistory(tx explorer.BlockTransaction, multiSig *explorer.MultiSig, address explorer.Address) explorer.AddressHistory {
	addressHistory := explorer.AddressHistory{
		Height:      tx.Height,
		TxIndex:     tx.Index,
		TxId:        tx.Txid,
		Time:        tx.Time,
		Hash:        multiSig.Key(),
		CfundPayout: false,
		StakePayout: false,
		Changes: explorer.AddressChanges{
			Spendable:    0,
			Stakable:     0,
			VotingWeight: 0,
		},
		Balance: explorer.AddressBalance{
			Spendable:    address.Spendable,
			Stakable:     address.Stakable,
			VotingWeight: address.VotingWeight,
		},
		Stake:    tx.IsAnyStaking(),
		MultiSig: true,
	}
	if addressHistory.Stake {
		addressHistory.Reward = explorer.AddressReward{}
		if addressHistory.Changes.Spendable != 0 {
			addressHistory.Reward.Spendable = float64(addressHistory.Balance.Spendable/100000000) / float64(addressHistory.Changes.Spendable/100000000) / 100
		}
		if addressHistory.Changes.Stakable != 0 {
			addressHistory.Reward.Stakable = float64(addressHistory.Balance.Stakable/100000000) / float64(addressHistory.Changes.Stakable/100000000) / 100
		}
		if addressHistory.Changes.VotingWeight != 0 {
			addressHistory.Reward.VotingWeight = float64(addressHistory.Balance.VotingWeight/100000000) / float64(addressHistory.Changes.VotingWeight/100000000) / 100
		}
	}
	for _, vin := range tx.Vin {
		if vin.PreviousOutput != nil && vin.PreviousOutput.MultiSig != nil && vin.PreviousOutput.MultiSig.Key() == multiSig.Key() {
			addressHistory.Changes = explorer.AddressChanges{
				Spendable:    addressHistory.Changes.Spendable - int64(vin.ValueSat),
				Stakable:     addressHistory.Changes.Stakable - int64(vin.ValueSat),
				VotingWeight: addressHistory.Changes.VotingWeight - int64(vin.ValueSat),
			}
			addressHistory.Balance = explorer.AddressBalance{
				Spendable:    addressHistory.Balance.Spendable - int64(vin.ValueSat),
				Stakable:     addressHistory.Balance.Stakable - int64(vin.ValueSat),
				VotingWeight: addressHistory.Balance.VotingWeight - int64(vin.ValueSat),
			}
		}
	}
	for _, vout := range tx.Vout {
		if vout.MultiSig != nil && vout.MultiSig.Key() == multiSig.Key() {
			addressHistory.Changes = explorer.AddressChanges{
				Spendable:    addressHistory.Changes.Spendable + int64(vout.ValueSat),
				Stakable:     addressHistory.Changes.Stakable + int64(vout.ValueSat),
				VotingWeight: addressHistory.Changes.VotingWeight + int64(vout.ValueSat),
			}
			addressHistory.Balance = explorer.AddressBalance{
				Spendable:    addressHistory.Balance.Spendable + int64(vout.ValueSat),
				Stakable:     addressHistory.Balance.Stakable + int64(vout.ValueSat),
				VotingWeight: addressHistory.Balance.VotingWeight + int64(vout.ValueSat),
			}
		}
	}

	return addressHistory
}
