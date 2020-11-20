package payment_request

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

type Indexer struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index) *Indexer {
	return &Indexer{navcoin, elastic}
}

func (i *Indexer) Index(txs []*explorer.BlockTransaction) {
	for _, tx := range txs {
		if tx.Version != 5 {
			continue
		}

		if navP, err := i.navcoin.GetPaymentRequest(tx.Hash); err == nil {
			paymentRequest := CreatePaymentRequest(navP, tx.Height)
			i.elastic.Save(elastic_cache.PaymentRequestIndex, paymentRequest)
			PaymentRequests = append(PaymentRequests, paymentRequest)
		}
	}
}

func (i *Indexer) Update(blockCycle *explorer.BlockCycle, block *explorer.Block) {
	for _, p := range PaymentRequests {
		if p == nil {
			continue
		}

		navP, err := i.navcoin.GetPaymentRequest(p.Hash)
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithError(err).Fatalf("Failed to find active payment request: %s", p.Hash)
		}

		UpdatePaymentRequest(navP, block.Height, p)
		if p.UpdatedOnBlock == block.Height {
			log.Debugf("Payment Request %s updated on block %d", p.Hash, block.Height)
			i.elastic.AddUpdateRequest(elastic_cache.PaymentRequestIndex.Get(), p)
		}

		if p.Status == explorer.PaymentRequestPaid.Status || p.Status == explorer.PaymentRequestExpired.Status || p.Status == explorer.PaymentRequestRejected.Status {
			if block.Height-p.UpdatedOnBlock >= uint64(blockCycle.Size) {
				PaymentRequests.Delete(p.Hash)
			}
		}
	}
}
