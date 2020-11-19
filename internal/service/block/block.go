package block

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"sort"
)

var LastBlockIndexed *explorer.Block

var BlockData = BlockDataObj{}

func (b *BlockDataObj) Reset() {
	BlockData = BlockDataObj{}
}

func (b *BlockDataObj) First() BlockElement {
	keys := make([]int, len(BlockData))
	i := 0
	for k := range BlockData {
		keys[i] = int(k)
		i++
	}
	sort.Ints(keys)

	return BlockData[uint64(keys[0])]
}

func (b *BlockDataObj) Last() BlockElement {
	keys := make([]int, len(BlockData))
	i := 0
	for k := range BlockData {
		keys[i] = int(k)
		i++
	}
	sort.Ints(keys)

	return BlockData[uint64(keys[len(keys)-1])]
}

func (b *BlockDataObj) Addresses() []string {
	addressMap := map[string]struct{}{}

	for idx := range *b {
		for _, tx := range (*b)[idx].Txs {
			for _, address := range tx.GetAllAddresses() {
				addressMap[address] = struct{}{}
			}
		}
	}

	addresses := make([]string, 0)
	for a := range addressMap {
		addresses = append(addresses, a)
	}

	return addresses
}

func (b *BlockDataObj) Txs() []*explorer.BlockTransaction {
	txs := make([]*explorer.BlockTransaction, 0)
	for _, data := range *b {
		for _, tx := range data.Txs {
			txs = append(txs, tx)
		}
	}

	return txs
}

type BlockDataObj map[uint64]BlockElement
type BlockElement struct {
	Block  *explorer.Block
	Txs    []*explorer.BlockTransaction
	header *navcoind.BlockHeader
}
