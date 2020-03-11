package vote

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type Indexer struct {
	elastic *elastic_cache.Index
}

func NewIndexer(elastic *elastic_cache.Index) *Indexer {
	return &Indexer{elastic}
}

func (i *Indexer) IndexVotes(txs []*explorer.BlockTransaction, block *explorer.Block, blockHeader *navcoind.BlockHeader) {
	for _, tx := range txs {
		if !tx.IsCoinbase() {
			continue
		}

		if v := CreateVotes(block, tx, blockHeader); v != nil {
			i.elastic.AddIndexRequest(elastic_cache.DaoVoteIndex.Get(), v.Slug(), v)
			return
		}
	}
}
