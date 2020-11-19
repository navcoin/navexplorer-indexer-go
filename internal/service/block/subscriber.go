package block

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

func (i *Indexer) OnIndexed(block *explorer.Block, txs []*explorer.BlockTransaction, header *navcoind.BlockHeader) {
	BlockData[block.Height] = BlockElement{block, txs, header}
}
