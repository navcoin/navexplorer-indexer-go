package dao

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consultation"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/payment_request"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/proposal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/vote"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"sync"
)

type Indexer interface {
	Index(block *explorer.Block, txs []explorer.BlockTransaction, header *navcoind.BlockHeader)
}

type indexer struct {
	navcoin               *navcoind.Navcoind
	proposalIndexer       proposal.Indexer
	paymentRequestIndexer payment_request.Indexer
	consultationIndexer   consultation.Indexer
	voteIndexer           vote.Indexer
	consensusIndexer      consensus.Indexer
}

func NewIndexer(
	navcoin *navcoind.Navcoind,
	proposalIndexer proposal.Indexer,
	paymentRequestIndexer payment_request.Indexer,
	consultationIndexer consultation.Indexer,
	voteIndexer vote.Indexer,
	consensusIndexer consensus.Indexer,
) Indexer {
	return indexer{
		navcoin,
		proposalIndexer,
		paymentRequestIndexer,
		consultationIndexer,
		voteIndexer,
		consensusIndexer,
	}
}

func (i indexer) Index(block *explorer.Block, txs []explorer.BlockTransaction, header *navcoind.BlockHeader) {
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

	i.voteIndexer.Index(txs, block, header)
	i.proposalIndexer.Update(block.BlockCycle, block)
	i.paymentRequestIndexer.Update(block.BlockCycle, block)
	i.consultationIndexer.Update(block.BlockCycle, block)

	if block.BlockCycle.IsEnd() {
		zap.L().With(zap.Uint("cycle", block.BlockCycle.Cycle), zap.Uint64("height", block.Height)).
			Info("DaoIndexer: BlockCycle complete")
		i.consensusIndexer.Update(block)
	}
}
