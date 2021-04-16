package softfork

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork/signal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	elastic       *elastic_cache.Index
	Service       *Service
	signalRepo    *signal.Repository
	blocksInCycle uint
	quorum        int
}

func NewRewinder(elastic *elastic_cache.Index, service *Service, signalRepo *signal.Repository, blocksInCycle uint, quorum int) *Rewinder {
	return &Rewinder{elastic, service, signalRepo, blocksInCycle, quorum}
}

func (r *Rewinder) Rewind(height uint64) error {
	softForks := r.Service.GetSoftForks()
	defer r.Service.Update(softForks, true)

	log.Infof("Rewinding soft fork index to height: %d", height)

	if err := r.elastic.DeleteHeightGT(height, elastic_cache.SignalIndex.Get()); err != nil {
		return err
	}

	for idx, s := range softForks {
		softForks[idx] = &explorer.SoftFork{
			Name:      s.Name,
			SignalBit: s.SignalBit,
			State:     explorer.SoftForkDefined,
			StartTime: s.StartTime,
			Timeout:   s.Timeout,
		}
	}

	start := uint64(1)
	end := uint64(r.blocksInCycle)

	for {
		if height == 0 || start >= height {
			break
		}
		if end >= height {
			end = height
		}

		signals := r.signalRepo.GetSignals(start, end)

		for _, s := range signals {
			for _, sf := range s.SoftForks {
				softFork := softForks.GetSoftFork(sf)
				if (softFork.State == explorer.SoftForkLockedIn && height <= softFork.LockedInHeight) || softFork.IsOpen() {
					softFork.SignalHeight = end
					softFork.State = explorer.SoftForkStarted
					blockCycle := GetSoftForkBlockCycle(r.blocksInCycle, s.Height)

					var cycle *explorer.SoftForkCycle
					if cycle = softFork.GetCycle(blockCycle.Cycle); cycle == nil {
						softFork.Cycles = append(softFork.Cycles, explorer.SoftForkCycle{Cycle: blockCycle.Cycle, BlocksSignalling: 0})
						cycle = softFork.GetCycle(blockCycle.Cycle)
					}
					cycle.BlocksSignalling++
				}
			}
		}

		for _, s := range softForks {
			if s.State == explorer.SoftForkStarted && s.LatestCycle() != nil && s.LatestCycle().BlocksSignalling >= explorer.GetQuorum(r.blocksInCycle, r.quorum) {
				s.State = explorer.SoftForkLockedIn
				s.LockedInHeight = end
				s.ActivationHeight = end + uint64(r.blocksInCycle)
			}
			if s.State == explorer.SoftForkLockedIn && height >= s.ActivationHeight {
				s.State = explorer.SoftForkActive
			}
		}

		start, end = func(start uint64, end uint64, height uint64) (uint64, uint64) {
			start += uint64(config.Get().SoftForkBlockCycle)
			end += uint64(config.Get().SoftForkBlockCycle)
			if end > height {
				end = height
			}
			return start, end
		}(start, end, height)
	}

	return nil
}
