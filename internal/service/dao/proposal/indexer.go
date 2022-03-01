package proposal

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
)

type Indexer interface {
	Index(txs []explorer.BlockTransaction)
	Update(blockCycle explorer.BlockCycle, block *explorer.Block)
}

type indexer struct {
	navcoin   *navcoind.Navcoind
	elastic   elastic_cache.Index
	indexSize uint64
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, indexSize uint64) Indexer {
	return indexer{navcoin, elastic, indexSize}
}

func (i indexer) Index(txs []explorer.BlockTransaction) {
	for _, tx := range txs {
		if !tx.IsSpend() && tx.Version != 4 {
			continue
		}

		if navP, err := i.navcoin.GetProposal(tx.Hash); err == nil {
			proposal := CreateProposal(navP, tx.Height)
			i.elastic.Save(elastic_cache.ProposalIndex.Get(), proposal)
			Proposals = append(Proposals, proposal)
		}
	}
}

func (i indexer) Update(blockCycle explorer.BlockCycle, block *explorer.Block) {
	for _, p := range Proposals {
		if p == nil {
			continue
		}

		navP, err := i.navcoin.GetProposal(p.Hash)
		if err != nil {
			zap.L().With(zap.String("proposal", p.Hash), zap.Error(err)).Fatal("Failed to find active proposal")
		}

		UpdateProposal(navP, block.Height, p)
		if p.UpdatedOnBlock == block.Height {
			i.elastic.AddUpdateRequest(elastic_cache.ProposalIndex.Get(), p)
		}

		if p.Status == explorer.ProposalExpired.Status || p.Status == explorer.ProposalRejected.Status {
			if block.Height-p.UpdatedOnBlock >= uint64(blockCycle.Size) {
				Proposals.Delete(p.Hash)
			}
		}
	}
}
