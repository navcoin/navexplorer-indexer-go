package softfork

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"time"
)

type Service struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	cache   *cache.Cache
	repo    *Repository
}

var (
	cacheKey = "explorer.SoftForks"
)

func New(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, cache *cache.Cache, repo *Repository) *Service {
	return &Service{navcoin, elastic, cache, repo}
}

func (i *Service) GetSoftForks() explorer.SoftForks {
	softForks, exists := i.cache.Get(cacheKey)
	if exists == false {
		return nil
	}

	return softForks.(explorer.SoftForks)
}

func (i *Service) Update(softForks explorer.SoftForks, persist bool) {
	i.cache.Set(cacheKey, softForks, cache.NoExpiration)

	for _, softFork := range softForks {
		if persist {
			i.elastic.Save(elastic_cache.SoftForkIndex.Get(), softFork)
		} else {
			i.elastic.AddUpdateRequest(elastic_cache.SoftForkIndex.Get(), softFork)
		}
	}
}

func (i *Service) InitSoftForks() {
	log.Info("Init SoftForks")

	info, err := i.navcoin.GetBlockchainInfo()
	if err != nil {
		log.WithError(err).Fatal("Failed to get blockchaininfo")
	}

	softForks, err := i.repo.getSoftForks()
	if err != nil && err != elastic_cache.ErrResultsNotFound {
		log.WithError(err).Fatal("Failed to get soft forks")
		return
	}

	for name, bip9fork := range info.Bip9SoftForks {
		if !softForks.HasSoftFork(name) {
			softFork := &explorer.SoftFork{
				Name:             name,
				SignalBit:        bip9fork.Bit,
				State:            explorer.SoftForkDefined,
				StartTime:        time.Unix(int64(bip9fork.StartTime), 0),
				Timeout:          time.Unix(int64(bip9fork.Timeout), 0),
				ActivationHeight: 0,
				LockedInHeight:   0,
			}

			i.elastic.Save(elastic_cache.SoftForkIndex.Get(), softFork)

			softForks = append(softForks, softFork)
		} else {
			if bip9fork.Bit != softForks.GetSoftFork(name).SignalBit {
				softForks.GetSoftFork(name).SignalBit = bip9fork.Bit
			}
		}
	}

	for _, softFork := range softForks {
		log.WithFields(log.Fields{
			"signalBit": softFork.SignalBit,
			"state":     softFork.State,
		}).Infof("SoftFork %s", softFork.Name)
	}

	i.Update(softForks, true)
}

func GetSoftForkBlockCycle(size uint, height uint64) *explorer.BlockCycle {
	cycle := (uint(height-1) / size) + 1

	return &explorer.BlockCycle{
		Size:  size,
		Cycle: cycle,
		Index: uint(height) - ((cycle * size) - size),
	}
}
