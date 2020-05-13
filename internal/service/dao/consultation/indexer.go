package consultation

import (
	"context"
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
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
		if tx.Version != 6 {
			continue
		}

		if navC, err := i.navcoin.GetConsultation(tx.Hash); err == nil {
			consultation := CreateConsultation(navC, tx)

			index := elastic_cache.DaoConsultationIndex.Get()
			_, err := i.elastic.Client.Index().
				Index(index).
				Id(fmt.Sprintf("%s-%s", config.Get().Network, consultation.Slug())).
				BodyJson(consultation).
				Do(context.Background())
			if err != nil {
				raven.CaptureError(err, nil)
				log.WithError(err).Fatal("Failed to save new consultation")
			}

			Consultations = append(Consultations, consultation)
		}
	}
}

func (i *Indexer) Update(blockCycle *explorer.BlockCycle, block *explorer.Block) {
	for _, c := range Consultations {
		if c == nil {
			continue
		}

		navC, err := i.navcoin.GetConsultation(c.Hash)
		if err != nil {
			raven.CaptureError(err, nil)
			log.WithError(err).Fatalf("Failed to find active consultation: %s", c.Hash)
		}

		if UpdateConsultation(navC, c) {
			c.UpdatedOnBlock = block.Height
			log.Debugf("Consultation %s updated on block %d", c.Hash, block.Height)
			i.elastic.AddUpdateRequest(elastic_cache.DaoConsultationIndex.Get(), c)
		}

		if c.Status == "expired" {
			if block.Height-c.UpdatedOnBlock >= uint64(blockCycle.Size) {
				Consultations.Delete(c.Hash)
			}
		}
	}
}
