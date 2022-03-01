package daemon

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"time"
)

var container *dic.Container

func Execute() {
	initialize()

	if config.Get().Reindex == true {
		zap.L().Info("Reindex complete")
		return
	}

	rewind()

	if config.Get().BulkIndex == true {
		bulkIndex()
	}

	if config.Get().Subscribe == true {
		subscribe()
	}
}

func initialize() {
	config.Init()

	container, _ = dic.NewContainer()
	container.GetElastic().InstallMappings()
	container.GetSoftforkService().InitSoftForks()
	container.GetDaoConsensusService().InitConsensusParameters()
}

func rewind() {
	bestBlock, err := container.GetBlockRepo().GetBestBlock()
	if err != nil {
		return
	}

	target := targetHeight(bestBlock)

	zap.L().Info("Rewind Index", zap.Uint64("from", bestBlock.Height), zap.Uint64("to", target))
	if err := container.GetRewinder().RewindToHeight(target); err != nil {
		zap.L().With(zap.Error(err)).Fatal("Failed to rewind index")
	}

	container.GetElastic().Persist()
	zap.L().Info("Sleep for 5 seconds")
	time.Sleep(5 * time.Second)

	bestBlock, err = container.GetBlockRepo().GetBestBlock()
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("Failed to get best block")
	}

	container.GetBlockService().SetLastBlockIndexed(bestBlock)
	zap.L().With(
		zap.Uint64("height", container.GetBlockService().GetLastBlockIndexed().Height),
	).Info("Best block")

	container.GetDaoProposalService().LoadVotingProposals(bestBlock)
	container.GetDaoPaymentRequestService().LoadVotingPaymentRequests(bestBlock)
	container.GetDaoConsultationService().LoadOpenConsultations(bestBlock)
}

func bulkIndex() {
	targetHeight := config.Get().BulkTargetHeight

	hash, err := container.GetNavcoin().GetBestBlockhash()
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("Failed to get best block hash")
	}

	bestNavBlock, err := container.GetNavcoin().GetBlock(hash)
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("Failed to get best block from navcoind")
	}

	if targetHeight == 0 {
		targetHeight = bestNavBlock.Height
	}

	currentHeight := uint64(0)
	lastBlockIndexed := container.GetBlockService().GetLastBlockIndexed()
	if lastBlockIndexed != nil {
		currentHeight = lastBlockIndexed.Height
	}

	zap.L().With(
		zap.Uint64("current-height", currentHeight),
		zap.Uint64("target", targetHeight),
	).Info("Bulk indexing blocks")

	if err := container.GetIndexer().Index(IndexOption.BatchIndex, targetHeight); err != nil {
		zap.L().With(zap.Error(err)).Fatal("Failed to bulk index blocks")
	}
	container.GetElastic().Persist()

	bestIndexedBlock, err := container.GetBlockRepo().GetBestBlock()
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("Failed to get best block from index")
	}
	zap.L().With(zap.Uint64("height", bestIndexedBlock.Height)).
		Info("Bulk index complete", zap.Uint64("height", bestIndexedBlock.Height))
}

func subscribe() {
	err := container.GetSubscriber().Subscribe(container.GetIndexer().SingleIndex)
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("Failed to subscribe to ZMQ")
	}
}

func targetHeight(bestBlock *explorer.Block) uint64 {
	if config.Get().RewindToHeight != 0 {
		zap.L().With(zap.Uint64("height", config.Get().RewindToHeight)).Info("Rewinding to height from config")
		return config.Get().RewindToHeight
	}

	height := bestBlock.Height

	if height >= config.Get().ReindexSize {
		return height - config.Get().ReindexSize
	}

	return 0
}
