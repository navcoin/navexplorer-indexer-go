package block

import (
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	STATIC_REWARD uint64 = 200000000
	MULTISIG_ASM         = "^OP_COINSTAKE OP_IF OP_DUP OP_HASH160 (?P<hash>[0-9a-f]{40}) OP_EQUALVERIFY OP_CHECKSIG OP_ELSE (?P<signaturesRequired>[1-9]) (?P<signatures>(?:0[0-9a-f]{65} )*)(?P<signaturesTotal>[1-9]) OP_CHECKMULTISIG OP_ENDIF$"
	WRAPPED_HEX          = "c66376a914a456b36048ce2e732ef729d044a1f744738df5fa88ac6753210277fa3f4f6d447c5914d8d69c259f94c76aa6eae829c5bd54e3cd6fc3f7e12f2f21033a0879f9ab601b4ee20ec9fed77ea1a48e9026b48e0d2a425d874b40ef13d02221034a51aa6aafbd6c6075ecaee0fbcf2c9ffbac05a49007a0f02c9d6680dccee6d42103ad915271a0b327f5379585c00c42a732530f246b60f9bb1c19af7db59363897e54ae68"
)

func CreateBlock(block *navcoind.Block, previousBlock *explorer.Block, cycleSize uint) *explorer.Block {
	log.Debugf("Create Block %s", block.Hash)
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
		BlockCycle: createBlockCycle(cycleSize, previousBlock),
		TxCount:    uint(len(block.Tx)),
		SupplyBalance: func(previousBlock *explorer.Block) explorer.SupplyBalance {
			if previousBlock == nil {
				return explorer.SupplyBalance{}
			}
			return previousBlock.SupplyBalance
		}(previousBlock),
		SupplyChange: func(previousBlock *explorer.Block) explorer.SupplyChange {
			if previousBlock == nil {
				return explorer.SupplyChange{}
			}
			return previousBlock.SupplyChange
		}(previousBlock),
	}
}

func createBlockCycle(size uint, previousBlock *explorer.Block) explorer.BlockCycle {
	if previousBlock == nil {
		return explorer.BlockCycle{
			Size:  size,
			Cycle: 1,
			Index: 1,
		}
	}

	if !previousBlock.BlockCycle.IsEnd() {
		return explorer.BlockCycle{
			Size:  size,
			Cycle: previousBlock.BlockCycle.Cycle,
			Index: previousBlock.BlockCycle.Index + 1,
		}
	}

	bc := explorer.BlockCycle{
		Size:  size,
		Cycle: previousBlock.BlockCycle.Cycle + 1,
		Index: uint(previousBlock.Height+1) % size,
	}
	if bc.Index != 0 {
		bc.Transitory = true
		bc.TransitorySize = size - bc.Index
	}

	return bc
}

func CreateBlockTransaction(rawTx navcoind.RawTransaction, index uint, block *explorer.Block) explorer.BlockTransaction {
	tx := explorer.BlockTransaction{
		RawBlockTransaction: explorer.RawBlockTransaction{
			Hex:             rawTx.Hex,
			Txid:            rawTx.Txid,
			Hash:            rawTx.Hash,
			Size:            rawTx.Size,
			VSize:           rawTx.VSize,
			Version:         rawTx.Version,
			LockTime:        rawTx.LockTime,
			Strdzeel:        rawTx.Strdzeel,
			AnonDestination: rawTx.AnonDestination,
			BlockHash:       rawTx.BlockHash,
			Height:          rawTx.Height,
			Confirmations:   rawTx.Confirmations,
			Time:            time.Unix(rawTx.Time, 0),
			BlockTime:       time.Unix(rawTx.BlockTime, 0),
		},
		TxHeight: float64(block.Height) + (float64(index+1) / 10000),
		Index:    index,
		Vin:      createVin(rawTx.Vin),
		Vout:     createVout(rawTx.Vout),
	}

	applyType(&tx)
	applyMultiSigColdStake(&tx)
	applyWrappedStatus(&tx)
	applyPrivateStatus(&tx, block)
	applyStaking(&tx, block)
	applySpend(&tx, block)
	applyCFundPayout(&tx, block)
	applyFees(&tx, block)

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
			PreviousOutput: nil,
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
				SpendingKey:  o.SpendingKey,
				OutputKey:    o.OutputKey,
				EphemeralKey: o.EphemeralKey,
				RangeProof:   o.RangeProof,
				SpentTxId:    o.SpentTxId,
				SpentIndex:   o.SpentIndex,
				SpentHeight:  uint64(o.SpentHeight),
			},
			Redeemed: false,
		})
	}

	return output
}

func applyType(tx *explorer.BlockTransaction) {
	if tx.IsCoinbase() {
		tx.Type = explorer.TxCoinbase
	} else if !isStakingTx(tx) {
		tx.Type = explorer.TxSpend
	} else if tx.Vout.OutputAtIndexIsOfType(1, explorer.VoutColdStaking) {
		tx.Type = explorer.TxColdStaking
	} else if tx.Vout.OutputAtIndexIsOfType(1, explorer.VoutColdStakingV2) {
		tx.Type = explorer.TxColdStakingV2
	} else {
		tx.Type = explorer.TxStaking
	}
}

func isStakingTx(tx *explorer.BlockTransaction) bool {
	return tx.Vout.OutputAtIndexIsOfType(0, explorer.VoutNonstandard) &&
		tx.Vout.GetOutput(0).ScriptPubKey.Hex == ""
}

func applyMultiSigColdStake(tx *explorer.BlockTransaction) {
	for idx := range tx.Vout {
		if matched, err := regexp.MatchString(MULTISIG_ASM, tx.Vout[idx].ScriptPubKey.Asm); err != nil || matched == false {
			continue
		}

		multiSigParams := getRegExParams(MULTISIG_ASM, tx.Vout[idx].ScriptPubKey.Asm)

		signaturesRequired, err := strconv.Atoi(multiSigParams["signaturesRequired"])
		if err != nil {
			continue
		}

		signaturesTotal, err := strconv.Atoi(multiSigParams["signaturesTotal"])
		if err != nil {
			continue
		}

		multiSig := &explorer.MultiSig{
			Hash:       multiSigParams["hash"],
			Signatures: strings.Split(strings.TrimSpace(multiSigParams["signatures"]), " "),
			Required:   signaturesRequired,
			Total:      signaturesTotal,
		}

		tx.Vout[idx].MultiSig = multiSig
	}
}

func getRegExParams(regEx, url string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func applyWrappedStatus(tx *explorer.BlockTransaction) {
	if tx.IsCoinbase() {
		return
	}

	for idx := range tx.Vout {
		if outputIsWrapped(tx.Vout[idx]) {
			tx.Vout[idx].Wrapped = true
			tx.Wrapped = true
		}
	}
}

func outputIsWrapped(o explorer.Vout) bool {
	return o.IsMultiSig() && o.ScriptPubKey.Hex == WRAPPED_HEX
}

func applyPrivateStatus(tx *explorer.BlockTransaction, block *explorer.Block) {
	if tx.IsCoinbase() || tx.Wrapped {
		return
	}

	for idx := range tx.Vout {
		if idx == len(tx.Vout)-1 && tx.Vout[idx].ScriptPubKey.Asm == "OP_RETURN" && tx.Vout[idx].ScriptPubKey.Type == "nulldata" {
			tx.Private = true
			tx.Vout[idx].ScriptPubKey.Addresses = []string{block.StakedBy}
			tx.Vout[idx].RedeemedIn = &explorer.RedeemedIn{
				Hash:   block.Tx[1],
				Height: block.Height,
				Index:  1,
			}
		}
		if tx.Vout[idx].RangeProof == true {
			tx.Vout[idx].Private = true
			tx.Private = true
		}
	}
}

func applyStaking(tx *explorer.BlockTransaction, block *explorer.Block) {
	if tx.IsSpend() {
		return
	}

	if tx.IsAnyStaking() {
		tx.Stake = tx.Vout.GetSpendableAmount() - tx.Vin.GetAmount()
		block.Stake = tx.Vout.GetSpendableAmount() - tx.Vin.GetAmount()
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
		block.StakedBy = voutsWithAddresses[0].ScriptPubKey.Addresses[0]
	} else if len(vinsWithAddresses) != 0 {
		block.StakedBy = vinsWithAddresses[0].Addresses[0]
	} else if len(voutsWithAddresses) != 0 {
		block.StakedBy = voutsWithAddresses[0].ScriptPubKey.Addresses[0]
	}
}

func applySpend(tx *explorer.BlockTransaction, block *explorer.Block) {
	if tx.Type != explorer.TxSpend {
		return
	}

	block.Spend += tx.Vin.GetAmount()
	tx.Spend = tx.Vin.GetAmount()
}

func applyFees(tx *explorer.BlockTransaction, block *explorer.Block) {
	if tx.Type != explorer.TxSpend {
		return
	}

	if tx.Private == true {
		tx.Fees = tx.Vout.PrivateFees()
		log.Debugf("Fees for PRIVATE|%s %d %d %d", tx.Hash, tx.Vin.GetAmount(), tx.Vout.GetAmount(), tx.Fees)
	} else {
		tx.Fees = tx.Vin.GetAmount() - tx.Vout.GetAmount()
		log.Debugf("Fees for %s %d %d %d", tx.Hash, tx.Vin.GetAmount(), tx.Vout.GetAmount(), tx.Fees)
	}
	block.Fees += tx.Fees
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
