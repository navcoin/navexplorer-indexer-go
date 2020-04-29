package softfork

import (
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/softfork/signal"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
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
	signal := signal.CreateSignal(block, &SoftForks)
	if signal != nil {
		i.updateSoftForks(signal, block)
	}

	i.updateState(block)

	for _, softFork := range SoftForks {
		i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), softFork)
	}

	if signal != nil {
		for _, s := range signal.SoftForks {
			if SoftForks.GetSoftFork(s) != nil && SoftForks.GetSoftFork(s).State == explorer.SoftForkLockedIn {
				signal.DeleteSoftFork(s)
			}
		}
		if len(signal.SoftForks) > 0 {
			i.elastic.AddIndexRequest(elastic_cache.SignalIndex.Get(), signal)
		}
	}

}

func (i *Indexer) updateSoftForks(signal *explorer.Signal, block *explorer.Block) {
	if signal == nil || !signal.IsSignalling() {
		return
	}
	blockCycle := GetSoftForkBlockCycle(i.blocksInCycle, block.Height)

	for _, s := range signal.SoftForks {
		softFork := SoftForks.GetSoftFork(s)
		if softFork == nil || softFork.SignalHeight >= signal.Height {
			continue
		}

		softFork.SignalHeight = signal.Height
		if softFork.State == explorer.SoftForkDefined {
			softFork.State = explorer.SoftForkStarted
		}

		var cycle *explorer.SoftForkCycle
		if cycle = softFork.GetCycle(blockCycle.Cycle); cycle == nil {
			softFork.Cycles = append(softFork.Cycles, explorer.SoftForkCycle{Cycle: blockCycle.Cycle, BlocksSignalling: 0})
			cycle = softFork.GetCycle(blockCycle.Cycle)
		}

		cycle.BlocksSignalling++
	}
}

func (i *Indexer) updateState(block *explorer.Block) {
	for idx, s := range SoftForks {
		if s.Cycles == nil {
			continue
		}

		if s.State == explorer.SoftForkStarted && s.LatestCycle().BlocksSignalling >= explorer.GetQuorum(i.blocksInCycle, i.quorum) {
			SoftForks[idx].LockedInHeight = uint64(i.blocksInCycle * block.BlockCycle.Cycle)
			SoftForks[idx].ActivationHeight = SoftForks[idx].LockedInHeight + uint64(i.blocksInCycle)
		}

		if s.LockedInHeight != 0 && s.ActivationHeight != 0 {
			if s.State == explorer.SoftForkStarted && block.Height >= s.LockedInHeight {
				SoftForks[idx].State = explorer.SoftForkLockedIn
				if SoftForks[idx].Cycles[len(SoftForks[idx].Cycles)-1].BlocksSignalling == 1 {
					SoftForks[idx].Cycles = SoftForks[idx].Cycles[:len(SoftForks[idx].Cycles)-1]
				}
			}
			if s.State == explorer.SoftForkLockedIn && block.Height >= s.ActivationHeight {
				SoftForks[idx].State = explorer.SoftForkActive
			}
		}
	}
}
