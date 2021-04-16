package dao

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consultation"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/payment_request"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/proposal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/vote"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"sync"
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

func (i *Indexer) Index(block *explorer.Block, txs []explorer.BlockTransaction, header *navcoind.BlockHeader) {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		i.proposalIndexer.Index(txs)
	}()

	go func() {
		defer wg.Done()
		i.paymentRequestIndexer.Index(txs)
	}()

	go func() {
		defer wg.Done()
		i.consultationIndexer.Index(txs)
	}()

	wg.Wait()

	i.voteIndexer.IndexVotes(txs, block, header)
	i.proposalIndexer.Update(block.BlockCycle, block)
	i.paymentRequestIndexer.Update(block.BlockCycle, block)
	i.consultationIndexer.Update(block.BlockCycle, block)

	if block.BlockCycle.IsEnd() {
		log.
			WithFields(log.Fields{"size": block.BlockCycle.Size, "height": block.Height}).
			Infof("Blockcycle %d complete", block.BlockCycle.Cycle)
		i.consensusIndexer.Update(block)
	}
}
