package block

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"go.uber.org/zap"
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
	zap.L().With(zap.Uint64("height", height)).Info("Rewinding block index")
	return r.elastic.DeleteHeightGT(height, elastic_cache.BlockTransactionIndex.Get(), elastic_cache.BlockIndex.Get())
}
