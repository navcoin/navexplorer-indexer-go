package indexer

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/event"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/indexer/IndexOption"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/address"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/dao"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/softfork"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Indexer struct {
	elastic         *elastic_cache.Index
	publisher       *event.Publisher
	blockIndexer    *block.Indexer
	addressIndexer  *address.Indexer
	softForkIndexer *softfork.Indexer
	daoIndexer      *dao.Indexer
	rewinder        *Rewinder
}

var (
	LastBlockIndexed uint64 = 0
)

func NewIndexer(
	elastic *elastic_cache.Index,
	publisher *event.Publisher,
	blockIndexer *block.Indexer,
	addressIndexer *address.Indexer,
	softForkIndexer *softfork.Indexer,
	daoIndexer *dao.Indexer,
	rewinder *Rewinder,
) *Indexer {
	return &Indexer{
		elastic,
		publisher,
		blockIndexer,
		addressIndexer,
		softForkIndexer,
		daoIndexer,
		rewinder,
	}
}

func (i *Indexer) BulkIndex() {
	log.Debug("Subscribe to 0MQ")

	if err := i.Index(IndexOption.BatchIndex); err != nil {
		if err.Error() == "-8: Block height out of range" {
			i.elastic.Persist()
		} else {
			log.WithError(err).Fatal("Failed to index blocks")
		}
	}
}

func (i *Indexer) SingleIndex() {
	err := i.Index(IndexOption.SingleIndex)
	if err != nil {
		if err.Error() != "-8: Block height out of range" {
			raven.CaptureErrorAndWait(err, nil)
			log.WithError(err).Fatal("Failed to index subscribed block")
		}
	}

	i.publisher.PublishToQueue("indexed.block", fmt.Sprintf("%d", LastBlockIndexed))
}

func (i *Indexer) Index(option IndexOption.IndexOption) error {
	err := i.index(LastBlockIndexed+1, option)
	if err == block.ErrOrphanBlockFound {
		err = i.rewinder.RewindToHeight(LastBlockIndexed - config.Get().ReindexSize)
	}

	if err == nil {
		return i.Index(option)
	}

	return err
}

func (i *Indexer) index(height uint64, option IndexOption.IndexOption) error {
	start := time.Now()
	b, txs, header, err := i.blockIndexer.Index(height, option)
	if err != nil {
		if err.Error() != "-8: Block height out of range" {
			raven.CaptureError(err, nil)
		}
		return err
	}
	elapsed := time.Since(start)
	log.WithField("took", elapsed).Infof("Indexed block     at height %d", height)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		start := time.Now()
		i.addressIndexer.Index(txs, b)
		elapsed := time.Since(start)
		log.WithField("took", elapsed).Infof("Indexed addresses at height %d", height)
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		i.softForkIndexer.Index(b)
		elapsed := time.Since(start)
		log.WithField("took", elapsed).Infof("Indexed softforks at height %d", height)
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		i.daoIndexer.Index(b, txs, header)
		elapsed := time.Since(start)
		log.WithField("took", elapsed).Infof("Indexed dao       at height %d", height)
	}()

	wg.Wait()

	LastBlockIndexed = height

	if option == IndexOption.BatchIndex {
		i.elastic.BatchPersist(height)
	} else {
		i.elastic.Persist()
	}

	return i.index(height+1, option)
}
