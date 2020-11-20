package consultation

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Indexer struct {
	navcoin         *navcoind.Navcoind
	elastic         *elastic_cache.Index
	blockRepository *block.Repository
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, blockRepository *block.Repository) *Indexer {
	return &Indexer{navcoin, elastic, blockRepository}
}

func (i *Indexer) Index(txs []*explorer.BlockTransaction) {
	for _, tx := range txs {
		if tx.Version != 6 {
			continue
		}

		if navC, err := i.navcoin.GetConsultation(tx.Hash); err == nil {
			consultation := CreateConsultation(navC, tx)
			i.elastic.Save(elastic_cache.DaoConsultationIndex, consultation)
			Consultations.Add(consultation)
		} else {
			log.WithField("hash", tx.Hash).WithError(err).Error("Failed to find consultation")
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
			log.Infof("Consultation %s updated on block %d", c.Hash, block.Height)
			i.elastic.AddUpdateRequest(elastic_cache.DaoConsultationIndex.Get(), c)
		}

		if c.ConsensusParameter && uint(c.State) == explorer.ConsultationPassed.State && c.HasPassedAnswer() && c.StateChangedOnBlock == block.Hash {
			c.UpdatedOnBlock = block.Height
			i.updateConsensusParameter(c, block)
			i.elastic.AddUpdateRequest(elastic_cache.DaoConsultationIndex.Get(), c)
			Consultations.Delete(c.Hash)
		}

		if uint(c.State) == explorer.ConsultationExpired.State {
			if block.Height-c.UpdatedOnBlock >= uint64(blockCycle.Size) {
				Consultations.Delete(c.Hash)
			}
		}
	}
}

func (i *Indexer) updateConsensusParameter(c *explorer.Consultation, b *explorer.Block) {
	answer := c.GetPassedAnswer()
	if answer == nil {
		log.WithField("consultation", c).Fatal("Passed Consultation with no passed answer")
		return
	}

	parameter := consensus.Parameters.GetById(c.Min)
	if parameter == nil {
		log.WithField("consultation", c).Fatal("updateConsensusParameter: Consensus parameter not found")
		return
	}

	value, _ := strconv.Atoi(answer.Answer)
	parameter.Value = value
	parameter.UpdatedOnBlock = b.Height
	log.WithFields(log.Fields{
		"parameter": parameter.Description,
		"value":     parameter.Value,
		"height":    b.Height,
	}).Info("Updated Consensus Parameter")
}
