package softfork

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork/signal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	elastic       *elastic_cache.Index
	blocksInCycle uint
	quorum        int
}

func NewIndexer(elastic *elastic_cache.Index, blocksInCycle uint, quorum int) *Indexer {
	return &Indexer{elastic, blocksInCycle, quorum}
}

func (i *Indexer) Index(block *explorer.Block) {
	sig := signal.CreateSignal(block, &SoftForks)
	if sig != nil {
		AddSoftForkSignal(sig, block.Height, i.blocksInCycle)
	}

	if block.BlockCycle.IsEnd() {
		log.WithFields(log.Fields{"height": block.Height, "blocksInCycle": i.blocksInCycle, "quorum": i.quorum}).Info("SoftFork: Block cycle end")
		UpdateSoftForksState(block.Height, i.blocksInCycle, i.quorum)
	}

	for _, softFork := range SoftForks {
		i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), softFork)
	}

	if sig != nil {
		for _, s := range sig.SoftForks {
			if SoftForks.GetSoftFork(s) != nil && SoftForks.GetSoftFork(s).State == explorer.SoftForkActive {
				log.Info("SoftFork: Delete active softForks")
				sig.DeleteSoftFork(s)
			}
		}
		if len(sig.SoftForks) > 0 {
			i.elastic.AddIndexRequest(elastic_cache.SignalIndex.Get(), sig)
		}
	}
}
