package softfork

import (
	"time"

	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
)

var SoftForks explorer.SoftForks

type Service interface {
	InitSoftForks()
}

type service struct {
	navcoin    *navcoind.Navcoind
	elastic    elastic_cache.Index
	repository Repository
}

func New(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, repository Repository) Service {
	return service{navcoin, elastic, repository}
}

func (i service) InitSoftForks() {
	zap.L().Info("SoftFork: Init")

	info, err := i.navcoin.GetBlockchainInfo()
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("SoftFork: Failed to get blockchaininfo")
	}

	SoftForks, err = i.repository.GetSoftForks()
	if err != nil && err != elastic_cache.ErrResultsNotFound {
		zap.L().With(zap.Error(err)).Fatal("SoftFork: Failed to get soft forks")
	}

	for name, bip9fork := range info.Bip9SoftForks {
		if !SoftForks.HasSoftFork(name) {
			softFork := &explorer.SoftFork{
				Name:             name,
				SignalBit:        bip9fork.Bit,
				State:            explorer.SoftForkDefined,
				StartTime:        time.Unix(int64(bip9fork.StartTime), 0),
				Timeout:          time.Unix(int64(bip9fork.Timeout), 0),
				ActivationHeight: 0,
				LockedInHeight:   0,
			}

			i.elastic.Save(elastic_cache.SoftForkIndex.Get(), softFork)

			SoftForks = append(SoftForks, softFork)
		} else {
			if bip9fork.Bit != SoftForks.GetSoftFork(name).SignalBit {
				SoftForks.GetSoftFork(name).SignalBit = bip9fork.Bit
				i.elastic.Save(elastic_cache.SoftForkIndex.Get(), SoftForks.GetSoftFork(name))
			}
		}
	}
}

func GetSoftForkBlockCycle(size uint, height uint64) *explorer.BlockCycle {
	cycle := (uint(height) / size) + 1

	return &explorer.BlockCycle{
		Size:  size,
		Cycle: cycle,
		Index: uint(height) - ((cycle * size) - size),
	}
}

func AddSoftForkSignal(signal *explorer.Signal, height uint64, blocksInCycle uint) {
	if !signal.IsSignalling() {
		return
	}

	blockCycle := GetSoftForkBlockCycle(blocksInCycle, height)

	for _, signalSoftFork := range signal.SoftForks {
		softFork := SoftForks.GetSoftFork(signalSoftFork)
		if softFork == nil || !softFork.IsOpen() {
			continue
		}

		softFork.SignalHeight = height
		if softFork.State == explorer.SoftForkDefined {
			softFork.State = explorer.SoftForkStarted
		}

		var cycle *explorer.SoftForkCycle
		if cycle = softFork.GetCycle(blockCycle.Cycle); cycle == nil {
			softFork.Cycles = append(softFork.Cycles, explorer.SoftForkCycle{Cycle: blockCycle.Cycle, BlocksSignalling: 0})
			cycle = softFork.GetCycle(blockCycle.Cycle)

			zap.L().With(
				zap.String("softfork", softFork.Name),
				zap.Uint("cycle", cycle.Cycle),
				zap.Uint64("height", height),
			).Info("SoftFork: Create Next Cycle")
		}

		cycle.BlocksSignalling++
	}
}

func UpdateSoftForksState(height uint64, blocksInCycle uint, quorum int) {
	for idx, _ := range SoftForks {
		if SoftForks[idx].Cycles == nil {
			continue
		}

		if SoftForks[idx].State == explorer.SoftForkStarted && height >= SoftForks[idx].LockedInHeight {
			if SoftForks[idx].LatestCycle().BlocksSignalling >= explorer.GetQuorum(blocksInCycle, quorum) {
				SoftForks[idx].State = explorer.SoftForkLockedIn
				SoftForks[idx].LockedInHeight = uint64(blocksInCycle * GetSoftForkBlockCycle(blocksInCycle, height).Cycle)
				SoftForks[idx].ActivationHeight = SoftForks[idx].LockedInHeight + uint64(blocksInCycle)

				zap.L().With(
					zap.String("softFork", SoftForks[idx].Name),
					zap.Uint64("height", height),
					zap.Uint64("lockedInHeight", SoftForks[idx].LockedInHeight),
					zap.Int("signals", SoftForks[idx].LatestCycle().BlocksSignalling),
				).Info("SoftFork: Locked in")
			}
		}

		if SoftForks[idx].State == explorer.SoftForkLockedIn && height >= SoftForks[idx].ActivationHeight-1 {
			SoftForks[idx].State = explorer.SoftForkActive

			zap.L().With(
				zap.String("softfork", SoftForks[idx].Name),
				zap.Uint64("height", height),
				zap.Uint64("activationHeight", SoftForks[idx].ActivationHeight),
			).Info("SoftFork: Activated")
		}
	}
}
