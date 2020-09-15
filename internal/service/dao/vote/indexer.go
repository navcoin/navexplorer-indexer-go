package vote

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	elastic *elastic_cache.Index
}

func NewIndexer(elastic *elastic_cache.Index) *Indexer {
	return &Indexer{elastic}
}

func (i *Indexer) IndexVotes(txs []*explorer.BlockTransaction, block *explorer.Block, blockHeader *navcoind.BlockHeader) {
	var votingAddress string
	for _, tx := range txs {
		var err error
		votingAddress, err = tx.Vout.GetVotingAddress()
		if err == nil && votingAddress != "" {
			break
		}
	}
	if votingAddress == "" {
		log.WithField("height", block.Height).Error("Unable to identify the voting address")
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
