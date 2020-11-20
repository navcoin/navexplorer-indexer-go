package softfork

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"time"
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
	if err != nil && err != elastic_cache.ErrResultsNotFound {
		log.WithError(err).Fatal("Failed to get soft forks")
		return
	}

	for name, bip9fork := range info.Bip9SoftForks {
		if !SoftForks.HasSoftFork(name) {
			softFork := &explorer.SoftFork{
				Name:             name,
				SignalBit:        bip9fork.Bit,
				State:            explorer.SoftForkDefined,
				StartTime:        time.Unix(int64(bip9fork.StartTime), 0),
				Timeout:          time.Unix(int64(bip9fork.Timeout), 0),
				ActivationHeight: 0,
				LockedInHeight:   0,
			}

			i.elastic.Save(elastic_cache.SoftForkIndex, softFork)

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
