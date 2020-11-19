package daemon

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
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

	indexer.LastBlockIndexed = getHeight()
	if indexer.LastBlockIndexed != 0 {
		log.Infof("Rewind from %d to %d", indexer.LastBlockIndexed+config.Get().ReindexSize, indexer.LastBlockIndexed)
		if err := container.GetRewinder().RewindToHeight(indexer.LastBlockIndexed); err != nil {
			log.WithError(err).Fatal("Failed to rewind index")
		}

		b, err := container.GetBlockRepo().GetBlockByHeight(indexer.LastBlockIndexed)
		if err != nil {
			log.WithError(err).Fatal("Failed to get block at height: ", indexer.LastBlockIndexed)
		}

		log.Debug("Get block cycle")
		container.GetDaoProposalService().LoadVotingProposals(b)
		container.GetDaoPaymentRequestService().LoadVotingPaymentRequests(b)
		container.GetDaoConsultationService().LoadOpenConsultations(b)
	}

	if err := container.GetEvent().Subscribe("batch.persist.start", container.GetAddressIndexer().BulkIndex); err != nil {
		log.WithError(err).Fatal("Failed to subscribe to block.indexed event")
	}

	container.GetIndexer().BulkIndex()

	err := container.GetSubscriber().Subscribe(container.GetIndexer().SingleIndex)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.WithError(err).Fatal("Failed to subscribe to ZMQ")
	}
}

func getHeight() uint64 {
	height, err := container.GetBlockRepo().GetHeight()
	if err != nil {
		if err == block.ErrBlockNotFound {
			return 0
		}
		log.WithError(err).Fatal("Failed to get block height")
	}

	if height >= config.Get().ReindexSize {
		return height - config.Get().ReindexSize
	}

	return 0
}
