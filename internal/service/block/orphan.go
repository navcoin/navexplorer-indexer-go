package block

import (
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
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

	previousBlock, err := o.repository.GetBlockByHeight(block.Height - 1)
	if err != nil {
		log.WithError(err).WithField("height", block.Height-1).Fatal("Failed to get previous block")
	}

	orphan := previousBlock.Hash != block.Previousblockhash
	if orphan == true {
		raven.CaptureError(err, nil)
		log.WithFields(log.Fields{"height": block.Height, "hash": block.Hash}).Info("Orphan block found")
	}

	return orphan, nil
}
