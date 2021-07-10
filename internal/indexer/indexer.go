package indexer

import (
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/queue"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/address"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Indexer interface {
	SingleIndex()
	Index(option IndexOption.IndexOption, target uint64) error
}

type indexer struct {
	elastic         elastic_cache.Index
	publisher       queue.Publisher
	blockIndexer    block.Indexer
	blockService    block.Service
	addressIndexer  address.Indexer
	softForkIndexer softfork.Indexer
	daoIndexer      dao.Indexer
	rewinder        Rewinder
}

func NewIndexer(
	elastic elastic_cache.Index,
	publisher queue.Publisher,
	blockIndexer block.Indexer,
	blockService block.Service,
	addressIndexer address.Indexer,
	softForkIndexer softfork.Indexer,
	daoIndexer dao.Indexer,
	rewinder Rewinder,
) Indexer {
	return indexer{
		elastic,
		publisher,
		blockIndexer,
		blockService,
		addressIndexer,
		softForkIndexer,
		daoIndexer,
		rewinder,
	}
}

func (i indexer) SingleIndex() {
	err := i.Index(IndexOption.SingleIndex, 0)
	if err != nil && err.Error() != "-8: Block height out of range" {
		zap.L().With(zap.Error(err)).Fatal("Failed to index subscribed block")
	}

	i.publisher.PublishToQueue("indexed.block", fmt.Sprintf("%d", i.blockService.GetLastBlockIndexed().Height))
}

func (i indexer) Index(option IndexOption.IndexOption, target uint64) error {
	var height uint64 = 1
	if lastBlockIndexed := i.blockService.GetLastBlockIndexed(); lastBlockIndexed != nil {
		height = lastBlockIndexed.Height + 1
		if target != 0 && height == target {
			return nil
		}
	}

	err := i.index(height, target, option)

	if errors.Is(block.ErrOrphanBlockFound, err) && option == IndexOption.SingleIndex {
		targetHeight := i.blockService.GetLastBlockIndexed().Height - config.Get().ReindexSize
		if err = i.rewinder.RewindToHeight(targetHeight); err == nil {
			return i.Index(option, target)
		}
	}

	return err
}

func (i indexer) index(height, target uint64, option IndexOption.IndexOption) error {
	start := time.Now()
	b, txs, header, err := i.blockIndexer.Index(height, option)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		start := time.Now()

		if option == IndexOption.SingleIndex {
			i.addressIndexer.Index(height, height, txs, option)
			return
		}

		if option == IndexOption.BatchIndex {
			if height%i.elastic.GetBulkIndexSize() != 0 && height != target {
				return
			}

			from := height
			if option == IndexOption.BatchIndex {
				from = height - i.elastic.GetBulkIndexSize() + 1
			}
			i.addressIndexer.Index(from, height, txs, option)
		}
		zap.L().With(
			zap.Duration("elapsed", time.Since(start)),
			zap.Uint64("height", height),
		).Info("Index addresses")
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		i.softForkIndexer.Index(b)

		zap.L().With(
			zap.Duration("elapsed", time.Since(start)),
			zap.Uint64("height", height),
		).Info("Index softforks")
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		i.daoIndexer.Index(b, txs, header)

		zap.L().With(
			zap.Duration("elapsed", time.Since(start)),
			zap.Uint64("height", height),
		).Info("Index dao")
	}()

	wg.Wait()

	zap.L().With(
		zap.Duration("elapsed", time.Since(start)),
		zap.Uint64("height", height),
	).Info("Index block")

	i.blockService.SetLastBlockIndexed(b)

	if option == IndexOption.BatchIndex {
		i.elastic.BatchPersist(height)
	} else {
		i.elastic.Persist()
	}

	if target != 0 && height == target {
		return nil
	}

	return i.index(height+1, target, option)
}
