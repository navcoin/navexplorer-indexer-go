package indexer

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/address"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork"
	log "github.com/sirupsen/logrus"
)

type Rewinder struct {
	elastic          *elastic_cache.Index
	blockRewinder    *block.Rewinder
	addressRewinder  *address.Rewinder
	softforkRewinder *softfork.Rewinder
	daoRewinder      *dao.Rewinder
	blockService     *block.Service
	blockRepo        *block.Repository
}

func NewRewinder(
	elastic *elastic_cache.Index,
	blockRewinder *block.Rewinder,
	addressRewinder *address.Rewinder,
	softforkRewinder *softfork.Rewinder,
	daoRewinder *dao.Rewinder,
	blockService *block.Service,
	blockRepo *block.Repository,
) *Rewinder {
	return &Rewinder{
		elastic,
		blockRewinder,
		addressRewinder,
		softforkRewinder,
		daoRewinder,
		blockService,
		blockRepo,
	}
}

func (r *Rewinder) RewindToHeight(height uint64) error {
	log.Infof("Rewinding to height: %d", height)

	lastBlock, err := r.blockRepo.GetBlockByHeight(height)
	if err != nil {
		r.blockService.SetLastBlockIndexed(lastBlock)
	}

	if err := r.addressRewinder.Rewind(height); err != nil {
		return err
	}
	if err := r.blockRewinder.Rewind(height); err != nil {
		return err
	}
	if err := r.softforkRewinder.Rewind(height); err != nil {
		return err
	}
	if err := r.daoRewinder.Rewind(height); err != nil {
		return err
	}

	log.Infof("Rewound to height: %d", height)
	r.elastic.Persist()

	return nil
}
