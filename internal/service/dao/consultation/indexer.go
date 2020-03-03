package consultation

import (
	"context"
	"github.com/NavExplorer/navcoind-go"
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

		if navP, err := i.navcoin.GetConsultation(tx.Hash); err == nil {
			//strdzeel := new(strdzeel)
			//if err := json.Unmarshal([]byte(tx.Strdzeel), &strdzeel); err != nil {
			//	log.WithError(err).Fatal("Failed to unmarshall strdzeel")
			//}

			consultation := CreateConsultation(navP)

			index := elastic_cache.DaoConsultationIndex.Get()
			resp, err := i.elastic.Client.Index().Index(index).BodyJson(consultation).Do(context.Background())
			if err != nil {
				raven.CaptureError(err, nil)
				log.WithError(err).Fatal("Failed to save new payment request")
			}

			consultation.MetaData = explorer.NewMetaData(resp.Id, resp.Index)
			Consultations = append(Consultations, consultation)
		}
	}
}

func (i *Indexer) Update(blockCycle *explorer.BlockCycle, block *explorer.Block) {
	for _, c := range Consultations {
		if c == nil {
			continue
		}
	}
}
