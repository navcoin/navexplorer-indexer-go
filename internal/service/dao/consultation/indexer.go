package consultation

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
	"strconv"
)

type Indexer interface {
	Index(txs []explorer.BlockTransaction)
	Update(blockCycle explorer.BlockCycle, block *explorer.Block)
}

type indexer struct {
	navcoin          *navcoind.Navcoind
	elastic          elastic_cache.Index
	blockRepository  block.Repository
	consensusService consensus.Service
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, blockRepository block.Repository, consensusService consensus.Service) Indexer {
	return indexer{navcoin, elastic, blockRepository, consensusService}
}

func (i indexer) Index(txs []explorer.BlockTransaction) {
	for _, tx := range txs {
		if tx.Version != 6 {
			continue
		}

		if navC, err := i.navcoin.GetConsultation(tx.Hash); err == nil {
			consultation := CreateConsultation(navC, &tx)
			i.elastic.Save(elastic_cache.DaoConsultationIndex.Get(), consultation)
			Consultations.Add(consultation)
		} else {
			zap.L().With(zap.String("hash", tx.Hash), zap.Error(err)).Error("Failed to find consultation")
		}
	}
}

func (i indexer) Update(blockCycle explorer.BlockCycle, block *explorer.Block) {
	consensusParameters := i.consensusService.GetConsensusParameters()
	for _, c := range Consultations {
		navC, err := i.navcoin.GetConsultation(c.Hash)
		if err != nil {
			zap.L().
				With(zap.Error(err), zap.String("consultation", c.Hash)).
				Fatal("ConsensusIndexer: Failed to find active consultation")
		}

		if UpdateConsultation(navC, &c, consensusParameters) {
			c.UpdatedOnBlock = block.Height
			zap.L().
				With(zap.String("consultation", c.Hash), zap.Uint64("height", c.Height)).
				Debug("ConsensusIndexer: Consultation updated")
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

func (i indexer) updateConsensusParameter(c explorer.Consultation, b *explorer.Block) {
	answer := c.GetPassedAnswer()
	if answer == nil {
		zap.L().With(zap.String("consultation", c.Hash)).Fatal("ConsensusIndexer: Passed Consultation with no passed answer")
		return
	}

	parameters := i.consensusService.GetConsensusParameters()
	parameter := parameters.GetConsensusParameterById(c.Min)
	if parameter == nil {
		zap.L().With(zap.String("consultation", c.Hash)).Fatal("ConsensusIndexer: Parameter not found")
		return
	}

	value, _ := strconv.Atoi(answer.Answer)
	parameter.Value = value
	parameter.UpdatedOnBlock = b.Height

	i.consensusService.Update(parameters, false)

	zap.L().With(
		zap.String("parameter", parameter.Description),
		zap.Int("value", parameter.Value),
		zap.Uint64("height", b.Height),
	).Info("ConsensusIndexer: Updated Consensus Parameter")
}
