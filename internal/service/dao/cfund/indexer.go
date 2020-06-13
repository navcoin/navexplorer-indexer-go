package cfund

import (
	"context"
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Indexer struct {
	elastic *elastic_cache.Index
}

func NewIndexer(elastic *elastic_cache.Index) *Indexer {
	return &Indexer{elastic}
}

func (i *Indexer) Index(header *navcoind.BlockHeader) {
	available, err := strconv.ParseFloat(header.NcfSupply, 64)
	if err != nil {
		log.WithError(err).Errorf("Failed to parse header.NcfSupply: %s", header.NcfSupply)
	}
	locked, err := strconv.ParseFloat(header.NcfLocked, 64)
	if err != nil {
		log.WithError(err).Errorf("Failed to parse header.NcfLocked: %s", header.NcfLocked)
	}

	cfund := explorer.Cfund{Available: available, Locked: locked}

	_, err = i.elastic.Client.Get().
		Index(elastic_cache.CfundIndex.Get()).
		Id(fmt.Sprintf("%s-%s", config.Get().Network, cfund.Slug())).
		Do(context.Background())
	if err != nil {
		i.elastic.AddIndexRequest(elastic_cache.CfundIndex.Get(), &cfund)
	} else {
		i.elastic.AddUpdateRequest(elastic_cache.CfundIndex.Get(), &cfund)
	}
}
