package daemon

import (
	"github.com/NavExplorer/navexplorer-indexer-go/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/indexer"
	"github.com/getsentry/raven-go"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
)

var container *dic.Container

func Execute() {
	config.Init()

	container, _ = dic.NewContainer(dingo.App)
	container.GetElastic().InstallMappings()
	container.GetSoftforkService().InitSoftForks()
	container.GetDaoConsensusService().InitConsensusParameters()

	if config.Get().Sentry.Active {
		_ = raven.SetDSN(config.Get().Sentry.DSN)
	}

	indexer.LastBlockIndexed = getHeight()
	if indexer.LastBlockIndexed != 0 {
		log.Infof("Rewind from %d to %d", indexer.LastBlockIndexed+uint64(config.Get().BulkIndexSize), indexer.LastBlockIndexed)
		if err := container.GetRewinder().RewindToHeight(indexer.LastBlockIndexed); err != nil {
			log.WithError(err).Fatal("Failed to rewind index")
		}

		block, err := container.GetBlockRepo().GetBlockByHeight(indexer.LastBlockIndexed)
		if err != nil {
			log.WithError(err).Fatal("Failed to get block at height: ", indexer.LastBlockIndexed)
		}

		log.Debug("Get block cycle")
		container.GetDaoProposalService().LoadVotingProposals(block)
		container.GetDaoPaymentRequestService().LoadVotingPaymentRequests(block)
		container.GetDaoConsultationService().LoadOpenConsultations(block)
	}

	log.Debug("Bulk index the backlog")
	container.GetIndexer().BulkIndex()

	log.Debug("Subscribe to 0MQ")
	container.GetSubscriber().Subscribe()
}

func getHeight() uint64 {
	if height, err := container.GetBlockRepo().GetHeight(); err != nil {
		log.WithError(err).Fatal("Failed to get block height")
	} else {
		if height >= uint64(config.Get().BulkIndexSize) {
			return height - uint64(config.Get().BulkIndexSize)
		}
	}

	return 0
}
