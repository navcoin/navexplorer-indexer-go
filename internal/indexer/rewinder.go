package indexer

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/address"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"time"
)

type Rewinder interface {
	RewindToHeight(height uint64) error
}

type rewinder struct {
	elastic          elastic_cache.Index
	blockRewinder    block.Rewinder
	addressRewinder  address.Rewinder
	softforkRewinder softfork.Rewinder
	daoRewinder      dao.Rewinder
	blockService     block.Service
	blockRepo        block.Repository
}

func NewRewinder(
	elastic elastic_cache.Index,
	blockRewinder block.Rewinder,
	addressRewinder address.Rewinder,
	softforkRewinder softfork.Rewinder,
	daoRewinder dao.Rewinder,
	blockService block.Service,
	blockRepo block.Repository,
) Rewinder {
	return rewinder{
		elastic,
		blockRewinder,
		addressRewinder,
		softforkRewinder,
		daoRewinder,
		blockService,
		blockRepo,
	}
}

func (r rewinder) RewindToHeight(height uint64) error {
	zap.L().With(zap.Uint64("height", height)).Info("Rewinding to height")

	r.elastic.ClearRequests()

	if err := r.addressRewinder.Rewind(height); err != nil {
		return err
	}
	if err := r.blockRewinder.Rewind(height); err != nil {
		return err
	}
	if err := r.softforkRewinder.Rewind(height); err != nil {
		return err
	}
	if err := r.daoRewinder.Rewind(height); err != nil {
		return err
	}

	zap.L().With(zap.Uint64("height", height)).Info("Rewound to height")
	r.elastic.Persist()

	r.checkBlock(height)

	return nil
}

func (r rewinder) checkBlock(height uint64) {
	zap.L().With(zap.Uint64("height", height)).Info("Checking block height")
	exists, bestBlock := r.isBestBlockAtHeight(height)
	if !exists {
		zap.L().With(zap.Uint64("height", height)).Info("BestBlock is not at expected height")
		time.Sleep(5 * time.Second)
		r.checkBlock(height)
	}
	zap.L().With(zap.Uint64("height", height)).Info("Best block rewound to height")
	zap.L().With(zap.String("hash", bestBlock.Hash), zap.Uint64("height", bestBlock.Height)).Info("Set last block indexed")

	r.blockService.SetLastBlockIndexed(bestBlock)
}

func (r rewinder) isBestBlockAtHeight(height uint64) (bool, *explorer.Block) {
	lastBlock, err := r.blockRepo.GetBestBlock()
	if err != nil {
		zap.L().With(zap.Error(err), zap.Uint64("height", height)).Fatal("Failed to get last block indexed")
	}
	zap.L().With(zap.Uint64("height", lastBlock.Height)).Info("Best block height")

	return lastBlock.Height == height, lastBlock
}
