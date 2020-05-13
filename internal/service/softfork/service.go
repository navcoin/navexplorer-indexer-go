package softfork

import (
	"context"
	"fmt"
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

var SoftForks explorer.SoftForks

type Service struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
}

func New(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository) *Service {
	return &Service{navcoin, elastic, repo}
}

func (i *Service) InitSoftForks() {
	log.Info("Init SoftForks")

	info, err := i.navcoin.GetBlockchainInfo()
	if err != nil {
		log.WithError(err).Fatal("Failed to get blockchaininfo")
	}

	SoftForks, err = i.repo.GetSoftForks()
	if err != nil {
		log.WithError(err).Fatal("Failed to get softforks from repo")
	}

	for name, bip9fork := range info.Bip9SoftForks {
		if SoftForks.GetSoftFork(name) == nil {
			softFork := &explorer.SoftFork{Name: name, SignalBit: bip9fork.Bit, State: explorer.SoftForkDefined}
			_, err := i.elastic.Client.
				Index().
				Index(elastic_cache.SoftForkIndex.Get()).
				Id(fmt.Sprintf("%s-%s", config.Get().Network, softFork.Slug())).
				BodyJson(softFork).
				Do(context.Background())
			if err != nil {
				log.WithError(err).Fatal("Failed to save new softfork")
			}

			log.Info("Saving new softfork ", name)
			SoftForks = append(SoftForks, softFork)
		}
	}
}

func GetSoftForkBlockCycle(size uint, height uint64) *explorer.BlockCycle {
	cycle := (uint(height) / size) + 1

	return &explorer.BlockCycle{
		Size:  size,
		Cycle: cycle,
		Index: uint(height) - ((cycle * size) - size),
	}
}
