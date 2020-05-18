package dao

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/consultation"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/payment_request"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/proposal"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao/vote"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	proposalIndexer       *proposal.Indexer
	paymentRequestIndexer *payment_request.Indexer
	consultationIndexer   *consultation.Indexer
	voteIndexer           *vote.Indexer
	consensusIndexer      *consensus.Indexer
	navcoin               *navcoind.Navcoind
}

func NewIndexer(
	proposalIndexer *proposal.Indexer,
	paymentRequestIndexer *payment_request.Indexer,
	consultationIndexer *consultation.Indexer,
	voteIndexer *vote.Indexer,
	consensusIndexer *consensus.Indexer,
	navcoin *navcoind.Navcoind,
) *Indexer {
	return &Indexer{
		proposalIndexer,
		paymentRequestIndexer,
		consultationIndexer,
		voteIndexer,
		consensusIndexer,
		navcoin,
	}
}

func (i *Indexer) Index(block *explorer.Block, txs []*explorer.BlockTransaction) {
	if consensus.Parameters == nil {
		err := i.consensusIndexer.Index()
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithError(err).Fatal("Failed to get Consensus")
		}
	}

	i.proposalIndexer.Index(txs)
	i.paymentRequestIndexer.Index(txs)
	i.consultationIndexer.Index(txs)

	header, err := i.navcoin.GetBlockheader(block.Hash)
	if err != nil {
		raven.CaptureError(err, nil)
		log.WithError(err).Fatal("Failed to get blockHeader")
	}
	i.voteIndexer.IndexVotes(txs, block, header)

	if block.BlockCycle.IsEnd() {
		i.proposalIndexer.Update(block.BlockCycle, block)
		i.paymentRequestIndexer.Update(block.BlockCycle, block)
	}

	i.consultationIndexer.Update(block.BlockCycle, block)
	if block.BlockCycle.IsEnd() {
		log.WithFields(log.Fields{
			"size":   block.BlockCycle.Size,
			"height": block.Height,
		}).Infof("Blockcycle %d complete", block.BlockCycle.Cycle)
		i.consensusIndexer.Update(block)
	}
}
