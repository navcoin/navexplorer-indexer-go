package vote

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

type Indexer interface {
	Index(txs []explorer.BlockTransaction, block *explorer.Block, blockHeader *navcoind.BlockHeader)
}

type indexer struct {
	elastic elastic_cache.Index
}

func NewIndexer(elastic elastic_cache.Index) Indexer {
	return indexer{elastic}
}

func (i indexer) Index(txs []explorer.BlockTransaction, block *explorer.Block, blockHeader *navcoind.BlockHeader) {
	var votingAddress string
	for _, tx := range txs {
		var err error
		votingAddress, err = tx.Vout.GetVotingAddress()
		if err == nil && votingAddress != "" {
			break
		}
	}
	if votingAddress == "" {
		return
	}

	for _, tx := range txs {
		if !tx.IsCoinbase() {
			continue
		}

		if v := CreateVotes(block, tx, blockHeader, votingAddress); v != nil {
			i.elastic.AddIndexRequest(elastic_cache.DaoVoteIndex.Get(), v)
			return
		}
	}
}
