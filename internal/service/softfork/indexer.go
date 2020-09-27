package softfork

import (
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/softfork/signal"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
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
	signal := signal.CreateSignal(block, &SoftForks)
	if signal != nil {
		i.updateSoftForks(signal, block.Height)
	}

	if block.BlockCycle.IsEnd() {
		i.updateState(block)
	}

	for _, softFork := range SoftForks {
		if softFork.State == explorer.SoftForkStarted {
			softFork.LockedInHeight = new(explorer.SoftFork).LockedInHeight
			softFork.ActivationHeight = new(explorer.SoftFork).ActivationHeight
			softFork.SignalHeight = block.Height
		}
		i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), softFork)
	}

	if signal != nil {
		for _, s := range signal.SoftForks {
			if SoftForks.GetSoftFork(s) != nil && SoftForks.GetSoftFork(s).State == explorer.SoftForkActive {
				log.Info("Delete the active softForks")
				signal.DeleteSoftFork(s)
			}
		}
		if len(signal.SoftForks) > 0 {
			i.elastic.AddIndexRequest(elastic_cache.SignalIndex.Get(), signal)
		}
	}
}

func (i *Indexer) updateSoftForks(signal *explorer.Signal, height uint64) {
	if signal == nil || !signal.IsSignalling() {
		return
	}
	blockCycle := GetSoftForkBlockCycle(i.blocksInCycle, height)

	for _, s := range signal.SoftForks {
		softFork := SoftForks.GetSoftFork(s)
		if softFork == nil || softFork.SignalHeight >= signal.Height {
			continue
		}

		if softFork.State == explorer.SoftForkDefined {
			softFork.State = explorer.SoftForkStarted
		}

		var cycle *explorer.SoftForkCycle
		if cycle = softFork.GetCycle(blockCycle.Cycle); cycle == nil {
			softFork.Cycles = append(softFork.Cycles, explorer.SoftForkCycle{Cycle: blockCycle.Cycle, BlocksSignalling: 0})
			cycle = softFork.GetCycle(blockCycle.Cycle)
		}

		cycle.BlocksSignalling++
		i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), softFork)
	}
}

func (i *Indexer) updateState(block *explorer.Block) {
	log.Info("Update Softfork State at height ", block.Height)
	for idx, _ := range SoftForks {
		if SoftForks[idx].Cycles == nil {
			continue
		}

		if SoftForks[idx].State == explorer.SoftForkStarted && block.Height >= SoftForks[idx].LockedInHeight {
			if SoftForks[idx].LatestCycle().BlocksSignalling >= explorer.GetQuorum(i.blocksInCycle, i.quorum) {
				log.WithField("softfork", SoftForks[idx].Name).Infof("Softfork locked in with %d signals", SoftForks[idx].LatestCycle().BlocksSignalling)
				SoftForks[idx].State = explorer.SoftForkLockedIn
				log.Info("Set State to LockedIn")

				SoftForks[idx].LockedInHeight = uint64(i.blocksInCycle * GetSoftForkBlockCycle(i.blocksInCycle, block.Height).Cycle)
				log.Info("Set LockedInHeight to ", SoftForks[idx].LockedInHeight)

				SoftForks[idx].ActivationHeight = SoftForks[idx].LockedInHeight + uint64(i.blocksInCycle)
				log.Info("Set ActivationHeight to ", SoftForks[idx].ActivationHeight)

				i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), SoftForks[idx])
			}
		}

		if SoftForks[idx].State == explorer.SoftForkLockedIn && block.Height >= SoftForks[idx].ActivationHeight-1 {
			log.WithField("softfork", SoftForks[idx].Name).Info("SoftFork Activated at height ", block.Height)
			SoftForks[idx].State = explorer.SoftForkActive
			i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), SoftForks[idx])
		}
	}
}
