package softfork

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork/signal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
)

type Indexer interface {
	Index(block *explorer.Block)
}

type indexer struct {
	elastic       elastic_cache.Index
	blocksInCycle uint
	quorum        int
}

func NewIndexer(elastic elastic_cache.Index, blocksInCycle uint, quorum int) Indexer {
	return indexer{elastic, blocksInCycle, quorum}
}

func (i indexer) Index(block *explorer.Block) {
	sig := signal.CreateSignal(block, &SoftForks)
	if sig != nil {
		AddSoftForkSignal(sig, block.Height, i.blocksInCycle)
	}

	if block.BlockCycle.IsEnd() {
		zap.L().With(
			zap.Uint64("height", block.Height),
			zap.Uint("blocksInCycle", i.blocksInCycle),
			zap.Int("quorum", i.quorum),
		).Debug("SoftFork: Block cycle end")
		UpdateSoftForksState(block.Height, i.blocksInCycle, i.quorum)
	}

	for _, softFork := range SoftForks {
		i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), softFork)
	}

	if sig != nil {
		for _, s := range sig.SoftForks {
			if SoftForks.GetSoftFork(s) != nil && SoftForks.GetSoftFork(s).State == explorer.SoftForkActive {
				zap.L().Info("SoftFork: Delete active softForks")
				sig.DeleteSoftFork(s)
			}
		}
		if len(sig.SoftForks) > 0 {
			i.elastic.AddIndexRequest(elastic_cache.SignalIndex.Get(), sig)
		}
	}
}
