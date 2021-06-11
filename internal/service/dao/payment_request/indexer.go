package payment_request

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
	navcoin *navcoind.Navcoind
	elastic elastic_cache.Index
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic elastic_cache.Index) Indexer {
	return indexer{navcoin, elastic}
}

func (i indexer) Index(txs []explorer.BlockTransaction) {
	for _, tx := range txs {
		if tx.Version != 5 {
			continue
		}

		if navP, err := i.navcoin.GetPaymentRequest(tx.Hash); err == nil {
			paymentRequest := CreatePaymentRequest(navP, tx.Height)
			i.elastic.Save(elastic_cache.PaymentRequestIndex.Get(), paymentRequest)
			PaymentRequests = append(PaymentRequests, paymentRequest)
		}
	}
}

func (i indexer) Update(blockCycle explorer.BlockCycle, block *explorer.Block) {
	for _, p := range PaymentRequests {
		if p == nil {
			continue
		}

		navP, err := i.navcoin.GetPaymentRequest(p.Hash)
		if err != nil {
			zap.L().With(zap.Error(err), zap.String("paymentRequest", p.Hash)).Fatal("Failed to find active payment request")
		}

		UpdatePaymentRequest(navP, block.Height, p)
		if p.UpdatedOnBlock == block.Height {
			zap.L().With(zap.String("paymentRequest", p.Hash), zap.Uint64("height", block.Height)).
				Debug("Payment Request updated")
			i.elastic.AddUpdateRequest(elastic_cache.PaymentRequestIndex.Get(), p)
		}

		if p.Status == explorer.PaymentRequestPaid.Status || p.Status == explorer.PaymentRequestExpired.Status || p.Status == explorer.PaymentRequestRejected.Status {
			if block.Height-p.UpdatedOnBlock >= uint64(blockCycle.Size) {
				PaymentRequests.Delete(p.Hash)
			}
		}
	}
}
