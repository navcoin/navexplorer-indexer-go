package block

import (
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"time"
)

type OrphanService struct {
	repository *Repository
}

func NewOrphanService(repository *Repository) *OrphanService {
	return &OrphanService{repository}
}

var (
	ErrOrphanBlockFound = errors.New("Orphan block_indexer found")
)

func (o *OrphanService) IsOrphanBlock(block *explorer.Block) (bool, error) {
	if block.Height == 1 {
		return false, nil
	}

	getPreviousBlock := func(height uint64) (*explorer.Block, error) {
		return o.repository.GetBlockByHeight(height - 1)
	}

	previousBlock, err := getPreviousBlock(block.Height - 1)
	if err != nil {
		log.Info("Retry get previous block in 5 seconds")
		time.Sleep(5 * time.Second)
		previousBlock, err = getPreviousBlock(block.Height - 1)
		if err != nil {
			log.WithError(err).WithField("height", block.Height-1).Fatal("Failed to get previous block")
		}
	}

	orphan := previousBlock.Hash != block.Previousblockhash
	if orphan == true {
		raven.CaptureError(err, nil)
		log.WithFields(log.Fields{"height": block.Height, "hash": block.Hash}).Info("Orphan block found")
	}

	return orphan, nil
}
