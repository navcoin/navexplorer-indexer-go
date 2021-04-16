package proposal

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	navcoin   *navcoind.Navcoind
	elastic   *elastic_cache.Index
	indexSize uint64
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, indexSize uint64) *Indexer {
	return &Indexer{navcoin, elastic, indexSize}
}

func (i *Indexer) Index(txs []explorer.BlockTransaction) {
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

func (i *Indexer) Update(blockCycle explorer.BlockCycle, block *explorer.Block) {
	for _, p := range Proposals {
		if p == nil {
			continue
		}

		navP, err := i.navcoin.GetProposal(p.Hash)
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithError(err).Fatalf("Failed to find active proposal: %s", p.Hash)
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
