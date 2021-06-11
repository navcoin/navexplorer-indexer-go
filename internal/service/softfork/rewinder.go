package softfork

import (
	"context"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork/signal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
)

type Rewinder interface {
	Rewind(height uint64) error
}

type rewinder struct {
	elastic       elastic_cache.Index
	signalRepo    signal.Repository
	blocksInCycle uint
	quorum        int
}

func NewRewinder(elastic elastic_cache.Index, signalRepo signal.Repository, blocksInCycle uint, quorum int) Rewinder {
	return rewinder{elastic, signalRepo, blocksInCycle, quorum}
}

func (r rewinder) Rewind(height uint64) error {
	zap.L().With(zap.Uint64("height", height)).Info("Rewinding soft fork index")

	zap.L().With(zap.Uint64("height", height)).Info("Delete Signals greater than height")
	if err := r.elastic.DeleteHeightGT(height, elastic_cache.SignalIndex.Get()); err != nil {
		return err
	}

	for idx, s := range SoftForks {
		zap.L().With(zap.String("softfork", s.Name)).Info("Resetting soft fork")
		SoftForks[idx] = &explorer.SoftFork{
			Name:      s.Name,
			SignalBit: s.SignalBit,
			State:     explorer.SoftForkDefined,
			StartTime: s.StartTime,
			Timeout:   s.Timeout,
		}
	}

	start := uint64(0)
	end := uint64(r.blocksInCycle) - 1

	for {
		if height == 0 || start >= height {
			break
		}
		if end >= height {
			end = height
		}

		for _, sig := range r.signalRepo.GetSignals(start, end) {
			AddSoftForkSignal(&sig, sig.Height, r.blocksInCycle)
		}

		if end-start == uint64(r.blocksInCycle)-1 {
			zap.L().With(
				zap.Uint64("index", end-start),
				zap.Uint64("height", end),
				zap.Uint("blocksInCycle", r.blocksInCycle),
				zap.Int("quorum", r.quorum),
			).Info("SoftFork: Block cycle end")
			UpdateSoftForksState(end-1, r.blocksInCycle, r.quorum)
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

	bulk := r.elastic.GetClient().Bulk()
	for _, sf := range SoftForks {
		bulk.Add(elastic.NewBulkUpdateRequest().Index(elastic_cache.SoftForkIndex.Get()).Id(sf.Slug()).Doc(sf))
	}

	if bulk.NumberOfActions() > 0 {
		if _, err := bulk.Do(context.Background()); err != nil {
			zap.L().With(zap.Error(err)).Fatal("Failed to rewind soft forks")
		}
	}

	return nil
}
