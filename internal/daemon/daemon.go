package daemon

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
	"sync"
)

var container *dic.Container

func Execute() {
	initialize()

	if config.Get().Reindex == true {
		log.Info("Reindex complete")
		return
	}

	if config.Get().VerifySupply == true {
		verifySupply()
		return
	}

	if config.Get().BulkIndex == true {
		bulkIndex()
	}

	if config.Get().Subscribe == true {
		rewind()
		subscribe()
	}
}

func initialize() {
	config.Init()

	container, _ = dic.NewContainer(dingo.App)
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

	log.Infof("Rewind from %d to %d", bestBlock.Height, target)
	if err := container.GetRewinder().RewindToHeight(target); err != nil {
		log.WithError(err).Fatal("Failed to rewind index")
	}

	bestBlock = container.GetBlockService().GetLastBlockIndexed()

	container.GetDaoProposalService().LoadVotingProposals(bestBlock)
	container.GetDaoPaymentRequestService().LoadVotingPaymentRequests(bestBlock)
	container.GetDaoConsultationService().LoadOpenConsultations(bestBlock)
}

func bulkIndex() {
	targetHeight := config.Get().BulkTargetHeight

	addressHeight, err := container.GetAddressRepo().GetBestHeight()
	if err != nil {
		log.WithError(err).Fatalf("Failed to get address height", addressHeight)
	}

	hash, err := container.GetNavcoin().GetBestBlockhash()
	if err != nil {
		log.WithError(err).Fatal("Failed to get best block hash")
	}

	bestNavBlock, err := container.GetNavcoin().GetBlock(hash)
	if err != nil {
		log.WithError(err).Fatal("Failed to get best block from navcoind")
	}

	if targetHeight == 0 {
		targetHeight = bestNavBlock.Height
	}

	bestIndexedBlock, err := container.GetBlockRepo().GetBestBlock()
	if err != nil {
		log.WithError(err).Fatal("Failed to get best block from index")
	}
	log.Infof("Blocks indexed up to %d", bestIndexedBlock.Height)

	log.Infof("Bulk indexing blocks to %d", targetHeight)
	if bestIndexedBlock.Height < targetHeight && addressHeight == 0 {
		if err := container.GetIndexer().Index(IndexOption.BatchIndex, targetHeight); err != nil {
			log.WithError(err).Fatal("Failed to bulk index blocks")
		}
		container.GetElastic().Persist()
	}
	log.Infof("Blocks indexed up to %d", targetHeight)

	log.Infof("Bulk indexing addresses to %d", targetHeight)
	if err := container.GetAddressIndexer().BulkIndex(targetHeight); err != nil {
		log.WithError(err).Error("Failed to bulk index addresses")
	}
	log.Infof("Addresses indexed up to %d", targetHeight)

	container.GetElastic().Persist()

	log.Infof("Bulk index complete to height %d", targetHeight)
}

func verifySupply() {
	bestBlock, err := container.GetBlockRepo().GetBestBlock()
	if err != nil {
		log.WithError(err).Fatalf("Failed to get best block")
	}
	height := bestBlock.Height

	log.Infof("Verifying supply to height %d", height)

	for i := config.Get().VerifySupplyFrom; i <= height; i++ {
		var wg sync.WaitGroup
		wg.Add(2)

		var block *explorer.Block
		go func() {
			defer wg.Done()
			block, _ = container.GetBlockRepo().GetBlockByHeight(i)
		}()

		var addressBalance int64
		go func() {
			defer wg.Done()
			addressBalance, _ = container.GetAddressRepo().GetAddressBalanceAtHeight(i)
		}()

		wg.Wait()

		if block.SupplyBalance.Public != uint64(addressBalance) {
			log.WithFields(log.Fields{
				"address_balance":       addressBalance,
				"block_balance":         block.SupplyBalance.Public,
				"block_changes_public":  block.SupplyChange.Public,
				"block_changes_private": block.SupplyChange.Private,
				"block_changes_wrapped": block.SupplyChange.Wrapped,
				"difference":            int64(block.SupplyBalance.Public) - addressBalance,
			}).Errorf("Balance is different at height %d", i)
		} else {
			log.WithFields(log.Fields{
				"address_balance": addressBalance,
				"block_balance":   block.SupplyBalance.Public,
				"difference":      int64(block.SupplyBalance.Public) - addressBalance,
			}).Infof("Accurate balance at height %d", i)
		}
	}
	log.Info("Supply Verification complete")
}

func subscribe() {
	err := container.GetSubscriber().Subscribe(container.GetIndexer().SingleIndex)
	if err != nil {
		log.WithError(err).Fatal("Failed to subscribe to ZMQ")
	}
}

func targetHeight(bestBlock *explorer.Block) uint64 {
	if config.Get().RewindToHeight != 0 {
		log.Info("Rewinding to height from config: %d", config.Get().RewindToHeight)
		return config.Get().RewindToHeight
	}

	height := bestBlock.Height

	if height >= config.Get().ReindexSize {
		return height - config.Get().ReindexSize
	}

	return 0
}
