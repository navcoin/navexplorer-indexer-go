package address

import (
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

func CreateAddress(hash string) *explorer.Address {
	return &explorer.Address{Hash: hash}
}

func ResetAddress(address *explorer.Address) {
	address.ColdBalance = 0
	address.ColdReceived = 0
	address.ColdReceivedCount = 0
	address.ColdSent = 0
	address.ColdSentCount = 0
	address.ColdStaked = 0
	address.ColdStakedCount = 0
	address.Balance = 0
	address.Received = 0
	address.ReceivedCount = 0
	address.Sent = 0
	address.SentCount = 0
	address.Staked = 0
	address.StakedCount = 0
}

func ApplyTxToAddress(address *explorer.Address, tx *explorer.AddressTransaction) {
	address.Height = tx.Height

	if tx.Cold == true {
		if explorer.IsColdStake(tx.Type) {
			address.ColdStaked += tx.Total
			address.ColdStakedCount++
		} else if tx.Type == explorer.TransferSend {
			address.ColdSent += int64(tx.Input)
			address.ColdSentCount++
		} else if tx.Type == explorer.TransferReceive {
			address.ColdReceived += int64(tx.Output)
			address.ColdReceivedCount++
		}
	} else {
		if explorer.IsStake(tx.Type) || explorer.IsColdStake(tx.Type) {
			address.Staked += tx.Total
			address.StakedCount++
		} else if tx.Type == explorer.TransferSend {
			address.Sent += int64(tx.Input)
			address.SentCount++
		} else if tx.Type == explorer.TransferReceive {
			address.Received += int64(tx.Output)
			address.ReceivedCount++
		} else if tx.Type == explorer.TransferPoolFee {
			address.Staked += tx.Total
			address.StakedCount++
		}
	}

	log.WithFields(log.Fields{"address": address.Hash, "tx": tx.Txid}).
		Debugf("Balance at height %d: %d", tx.Height, tx.Balance)
}

func CreateAddressTransaction(tx *explorer.BlockTransaction, block *explorer.Block) []*explorer.AddressTransaction {
	addressTxs := make([]*explorer.AddressTransaction, 0)
	for _, address := range tx.GetAllAddresses() {
		if tx.HasColdInput(address) || tx.HasColdStakeStake(address) || tx.HasColdStakeReceive(address) {
			if coldAddressTx := createColdTransaction(address, tx); coldAddressTx != nil {
				addressTxs = append(addressTxs, coldAddressTx)
			}
		}
		if addressTx := createTransaction(address, tx, block); addressTx != nil {
			addressTxs = append(addressTxs, addressTx)
		}
	}

	return addressTxs
}

func createTransaction(address string, tx *explorer.BlockTransaction, block *explorer.Block) *explorer.AddressTransaction {
	_, input := tx.Vin.GetAmountByAddress(address, false)
	_, output := tx.Vout.GetAmountByAddress(address, false)
	if input+output == 0 {
		return nil
	}

	addressTransaction := &explorer.AddressTransaction{
		Hash:   address,
		Txid:   tx.Hash,
		Height: tx.Height,
		Index:  tx.Index,
		Time:   tx.Time,
		Cold:   false,
		Input:  input,
		Output: output,
		Total:  int64(output - input),
	}

	if tx.IsStaking() {
		addressTransaction.Type = explorer.TransferStake
	} else if tx.IsCoinbase() {
		if block.StakedBy == address {
			// POW block_indexer
			addressTransaction.Type = explorer.TransferStake
		} else if tx.Version == 3 {
			addressTransaction.Type = explorer.TransferCommunityFundPayout
		} else {
			bt, _ := json.Marshal(tx)
			log.WithFields(log.Fields{"tx": string(bt)}).Fatal("Could not handle coinbase")
		}
	} else if tx.IsColdStaking() {
		if addressTransaction.Input == addressTransaction.Output {
			addressTransaction.Type = explorer.TransferStake
		} else if tx.HasColdStakeSpend(address) {
			if tx.Vin.HasAddress(address) {
				addressTransaction.Type = explorer.TransferColdStake
			} else {
				addressTransaction.Type = explorer.TransferColdDelegateStake
			}
		} else if !tx.Vin.HasAddress(address) && tx.Vout.HasAddress(address) {
			addressTransaction.Type = explorer.TransferPoolFee
		} else if addressTransaction.Input == 0 {
			addressTransaction.Type = explorer.TransferReceive
		}
	} else {
		if addressTransaction.Input > addressTransaction.Output {
			addressTransaction.Type = explorer.TransferSend
		} else {
			addressTransaction.Type = explorer.TransferReceive
		}
	}

	if addressTransaction.Type == "" {
		bt, _ := json.Marshal(tx)
		log.WithFields(log.Fields{"tx": string(bt)}).Fatal("addressTransaction.Type not identified: ", address)
		return nil
	}

	return addressTransaction
}

func createColdTransaction(address string, tx *explorer.BlockTransaction) *explorer.AddressTransaction {
	_, input := tx.Vin.GetAmountByAddress(address, true)
	_, output := tx.Vout.GetAmountByAddress(address, true)
	if input+output == 0 {
		return nil
	}

	addressTransaction := &explorer.AddressTransaction{
		Hash:   address,
		Txid:   tx.Hash,
		Height: tx.Height,
		Index:  tx.Index,
		Time:   tx.Time,
		Cold:   true,
		Input:  input,
		Output: output,
		Total:  int64(output - input),
	}

	if tx.IsSpend() {
		if addressTransaction.Input > addressTransaction.Output {
			addressTransaction.Type = explorer.TransferSend
		} else {
			addressTransaction.Type = explorer.TransferReceive
		}
	} else if tx.IsColdStaking() {
		addressTransaction.Type = explorer.TransferColdStake
	} else {
		bt, _ := json.Marshal(tx)
		log.WithFields(log.Fields{
			"address": address,
			"tx":      string(bt),
			"type":    tx.Type,
		}).Fatal("WE FOUND SOMETHING ELSE IN COLD TX")
	}

	return addressTransaction
}
