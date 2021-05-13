package block

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	log "github.com/sirupsen/logrus"
)

type Rewinder interface {
	Rewind(height uint64) error
}

type rewinder struct {
	elastic elastic_cache.Index
}

func NewRewinder(elastic elastic_cache.Index) Rewinder {
	return rewinder{elastic}
}

func (r rewinder) Rewind(height uint64) error {
	log.Infof("Rewinding block index to height: %d", height)
	return r.elastic.DeleteHeightGT(height, elastic_cache.BlockTransactionIndex.Get(), elastic_cache.BlockIndex.Get())
}
