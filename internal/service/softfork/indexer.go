package softfork

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork/signal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	elastic       *elastic_cache.Index
	Service       *Service
	blocksInCycle uint
	quorum        int
}

func NewIndexer(elastic *elastic_cache.Index, service *Service, blocksInCycle uint, quorum int) *Indexer {
	return &Indexer{elastic, service, blocksInCycle, quorum}
}

func (i *Indexer) Index(block *explorer.Block) {
	softForks := i.Service.GetSoftForks()
	defer i.Service.Update(softForks, false)

	signal := signal.CreateSignal(block, softForks)
	if signal != nil {
		i.updateSoftForks(signal, block.Height)
	}

	if block.BlockCycle.IsEnd() {
		i.updateState(block)
	}

	for _, softFork := range softForks {
		if softFork.State == explorer.SoftForkStarted {
			softFork.LockedInHeight = new(explorer.SoftFork).LockedInHeight
			softFork.ActivationHeight = new(explorer.SoftFork).ActivationHeight
			softFork.SignalHeight = block.Height
		}
		if softFork.State == explorer.SoftForkLockedIn && softFork.SignalHeight <= softFork.LockedInHeight {
			softFork.SignalHeight = block.Height - 1
		}
	}

	if signal != nil {
		for _, s := range signal.SoftForks {
			if softForks.GetSoftFork(s) != nil && softForks.GetSoftFork(s).State == explorer.SoftForkActive {
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
	softForks := i.Service.GetSoftForks()
	defer i.Service.Update(softForks, false)

	if signal == nil || !signal.IsSignalling() {
		return
	}
	blockCycle := GetSoftForkBlockCycle(i.blocksInCycle, height)

	for _, s := range signal.SoftForks {
		softFork := softForks.GetSoftFork(s)
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
	}
}

func (i *Indexer) updateState(block *explorer.Block) {
	softForks := i.Service.GetSoftForks()
	defer i.Service.Update(softForks, false)

	log.Info("Update Softfork State at height ", block.Height)
	for _, softFork := range softForks {
		if softFork.Cycles == nil {
			continue
		}

		if softFork.State == explorer.SoftForkStarted && block.Height >= softFork.LockedInHeight {
			if softFork.LatestCycle().BlocksSignalling >= explorer.GetQuorum(i.blocksInCycle, i.quorum) {
				log.WithField("softfork", softFork.Name).Infof("Softfork locked in with %d signals", softFork.LatestCycle().BlocksSignalling)
				softFork.State = explorer.SoftForkLockedIn

				softFork.LockedInHeight = uint64(i.blocksInCycle * GetSoftForkBlockCycle(i.blocksInCycle, block.Height).Cycle)
				log.Info("Set LockedInHeight to ", softFork.LockedInHeight)

				softFork.ActivationHeight = softFork.LockedInHeight + uint64(i.blocksInCycle)
				log.Info("Set ActivationHeight to ", softFork.ActivationHeight)
			}
		}

		if softFork.State == explorer.SoftForkLockedIn && block.Height >= softFork.ActivationHeight-1 {
			log.WithField("softfork", softFork.Name).Info("SoftFork Activated at height ", block.Height)
			softFork.State = explorer.SoftForkActive
		}
	}
}
