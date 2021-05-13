package block

import (
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"time"
)

type OrphanService interface {
	IsOrphanBlock(block *explorer.Block, previousBlock *explorer.Block) (bool, error)
}

type orphanService struct {
	repository Repository
}

func NewOrphanService(repository Repository) OrphanService {
	return orphanService{repository}
}

var (
	ErrOrphanBlockFound = errors.New("Orphan block_indexer found")
)

func (o orphanService) IsOrphanBlock(block *explorer.Block, previousBlock *explorer.Block) (bool, error) {
	if block.Height == 1 {
		return false, nil
	}

	getPreviousBlock := func(height uint64) (*explorer.Block, error) {
		return o.repository.GetBlockByHeight(height - 1)
	}

	if previousBlock == nil {
		var err error
		previousBlock, err = getPreviousBlock(block.Height)
		if err != nil {
			log.Info("Retry get previous block in 1 seconds")
			time.Sleep(1 * time.Second)
			previousBlock, err = getPreviousBlock(block.Height)
			if err != nil {
				log.WithError(err).WithField("height", block.Height-1).Fatal("Failed to get previous block")
			}
		}
	}

	orphan := previousBlock.Hash != block.Previousblockhash
	if orphan == true {
		log.WithFields(log.Fields{"height": block.Height, "block": block.Hash, "previous": previousBlock.Hash}).Info("Orphan block found")
	}

	return orphan, nil
}
