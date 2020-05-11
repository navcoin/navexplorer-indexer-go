package block

import (
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"time"
)

func CreateBlock(block *navcoind.Block, cycleSize uint) *explorer.Block {
	return &explorer.Block{
		RawBlock: explorer.RawBlock{
			Hash:              block.Hash,
			Confirmations:     block.Confirmations,
			StrippedSize:      block.StrippedSize,
			Size:              block.Size,
			Weight:            block.Weight,
			Height:            block.Height,
			Version:           block.Version,
			VersionHex:        block.VersionHex,
			Merkleroot:        block.MerkleRoot,
			Tx:                block.Tx,
			Time:              time.Unix(block.Time, 0),
			MedianTime:        time.Unix(block.MedianTime, 0),
			Nonce:             block.Nonce,
			Bits:              block.Bits,
			Difficulty:        fmt.Sprintf("%f", block.Difficulty),
			Chainwork:         block.ChainWork,
			Previousblockhash: block.PreviousBlockHash,
			Nextblockhash:     block.NextBlockHash,
		},
		BlockCycle: createBlockCycle(cycleSize, block),
		TxCount:    uint(len(block.Tx)),
	}
}

func createBlockCycle(size uint, block *navcoind.Block) *explorer.BlockCycle {
	cycle := getCycleForHeight(size, block.Height)

	return &explorer.BlockCycle{
		Size:  size,
		Cycle: cycle,
		Index: getCycleIndex(block.Height, cycle, size),
	}
}

func getCycleForHeight(size uint, height uint64) uint {
	return (uint(height) / size) + 1
}

func getCycleIndex(height uint64, cycle uint, size uint) uint {
	base := (cycle * size) - size
	return uint(height) - base
}

func CreateBlockTransaction(navTx navcoind.RawTransaction, index uint) *explorer.BlockTransaction {
	tx := &explorer.BlockTransaction{
		RawBlockTransaction: explorer.RawBlockTransaction{
			Hex:             navTx.Hex,
			Txid:            navTx.Txid,
			Hash:            navTx.Hash,
			Size:            navTx.Size,
			VSize:           navTx.VSize,
			Version:         navTx.Version,
			LockTime:        navTx.LockTime,
			Strdzeel:        navTx.Strdzeel,
			AnonDestination: navTx.AnonDestination,
			BlockHash:       navTx.BlockHash,
			Height:          navTx.Height,
			Confirmations:   navTx.Confirmations,
			Time:            time.Unix(navTx.Time, 0),
			BlockTime:       time.Unix(navTx.BlockTime, 0),
		},
		Index: index,
		Vin:   createVin(navTx.Vin),
		Vout:  createVout(navTx.Vout),
	}

	return tx
}

func createVin(vins []navcoind.Vin) []explorer.Vin {
	var inputs = make([]explorer.Vin, 0)
	for idx, _ := range vins {
		input := explorer.Vin{
			RawVin: explorer.RawVin{
				Coinbase: vins[idx].Coinbase,
				Sequence: vins[idx].Sequence,
			},
		}
		if vins[idx].Txid != "" {
			input.Txid = &vins[idx].Txid
			input.Vout = &vins[idx].Vout
		}

		if vins[idx].Value != 0 {
			input.Value = vins[idx].Value
			input.ValueSat = vins[idx].ValueSat
		}

		if vins[idx].Address != "" {
			input.Addresses = []string{vins[idx].Address}
		}

		inputs = append(inputs, input)
	}

	return inputs
}

func createVout(vouts []navcoind.Vout) []explorer.Vout {
	var output = make([]explorer.Vout, 0)
	for _, o := range vouts {
		output = append(output, explorer.Vout{
			RawVout: explorer.RawVout{
				Value:    o.Value,
				ValueSat: o.ValueSat,
				N:        o.N,
				ScriptPubKey: explorer.ScriptPubKey{
					Asm:       o.ScriptPubKey.Asm,
					Hex:       o.ScriptPubKey.Hex,
					ReqSigs:   o.ScriptPubKey.ReqSigs,
					Type:      explorer.VoutTypes[o.ScriptPubKey.Type],
					Addresses: o.ScriptPubKey.Addresses,
					Hash:      o.ScriptPubKey.Hash,
				},
			},
		})
	}

	return output
}

func applyType(tx *explorer.BlockTransaction) {
	if tx.IsCoinbase() {
		tx.Type = explorer.TxCoinbase
	} else if tx.Vout.GetAmount() <= tx.Vin.GetAmount() {
		tx.Type = explorer.TxSpend
	} else if len(tx.Vout) > 1 && tx.Vout[1].ScriptPubKey.Type == explorer.VoutColdStaking {
		tx.Type = explorer.TxColdStaking
	} else if len(tx.Vout) > 1 && tx.Vout[1].ScriptPubKey.Type == explorer.VoutColdStakingV2 {
		tx.Type = explorer.TxColdStakingV2
	} else {
		tx.Type = explorer.TxStaking
	}
}

func applyStaking(tx *explorer.BlockTransaction, block *explorer.Block) {
	if tx.IsSpend() {
		return
	}

	if tx.IsAnyStaking() {
		if tx.Height >= 2761920 {
			tx.Stake = 200000000 // hard coded to 2 as static rewards arrived after block_indexer 2761920
			block.Stake += tx.Stake
		} else {
			tx.Stake = tx.Vout.GetSpendableAmount() - tx.Vin.GetAmount()
			block.Stake += tx.Stake
		}
	} else if tx.IsCoinbase() {
		for _, o := range tx.Vout {
			if o.ScriptPubKey.Type == explorer.VoutPubkey {
				tx.Stake = o.ValueSat
				block.Stake = o.ValueSat
			}
		}
	}

	voutsWithAddresses := tx.Vout.FilterWithAddresses()
	vinsWithAddresses := tx.Vin.FilterWithAddresses()

	if tx.IsColdStaking() {
		for _, vout := range tx.Vout {
			if vout.ScriptPubKey.Type == explorer.VoutColdStaking {
				block.StakedBy = vout.ScriptPubKey.Addresses[0]
				break
			}
		}
		block.StakedBy = voutsWithAddresses[0].ScriptPubKey.Addresses[0]
	} else if len(vinsWithAddresses) != 0 {
		block.StakedBy = vinsWithAddresses[0].Addresses[0]
	} else if len(voutsWithAddresses) != 0 {
		block.StakedBy = voutsWithAddresses[0].ScriptPubKey.Addresses[0]
	}
}

func applySpend(tx *explorer.BlockTransaction, block *explorer.Block) {
	if tx.Type == explorer.TxSpend {
		tx.Spend = tx.Vout.GetAmount()
		tx.Fees = tx.Vin.GetAmount() - tx.Vout.GetAmount()
		block.Spend += tx.Spend
		block.Fees += tx.Fees
	}
}

func applyCFundPayout(tx *explorer.BlockTransaction, block *explorer.Block) {
	if tx.IsCoinbase() {
		for _, o := range tx.Vout {
			if o.ScriptPubKey.Type == explorer.VoutPubkeyhash && tx.Version == 3 {
				block.CFundPayout += o.ValueSat
			}
		}
	}
}
