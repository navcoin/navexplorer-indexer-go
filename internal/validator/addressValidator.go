package validator

import (
	"context"
	"fmt"
	_ "github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
	"time"
)

var container *dic.Container

var bestBlock uint64

type AddressValidator struct{}

func (v *AddressValidator) Execute() {
	config.Init()
	container, _ = dic.NewContainer(dingo.App)

	go loadBestBlock()

	execTime, execTimeCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer execTimeCancel()

	count := 0
	for {
		if execTime.Err() != nil {
			log.Infof("%d addresses validated", count)
			break
		}
		count++

		addresses, err := container.GetAddressRepo().GetAddressesByValidateAtDesc(10000)
		if err != nil {
			log.WithError(err).Fatal("Failed to get addresses")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		for {
			if ctx.Err() != nil {
				log.Fatal("Failed to find the best block")
			}
			if bestBlock != 0 {
				break
			}
		}

		for _, a := range addresses {
			currentBlock := bestBlock
			previousBalance := a.Balance
			err := container.GetAddressRewinder().ResetAddressToHeight(a, currentBlock)
			if err != nil {
				log.WithError(err).Fatal("Failed to reset address")
			}

			a.ValidatedAt = currentBlock

			if previousBalance != a.Balance {
				a.ValidatedAt = 0
				log.WithFields(log.Fields{
					"previous": previousBalance,
					"current":  a.Balance,
				}).Errorf("Validation error %s", a.Hash)
			}

			if currentBlock == bestBlock {
				a.ValidatedAt = currentBlock

				_, err = container.GetElastic().Client.Index().
					Index(elastic_cache.AddressIndex.Get()).
					BodyJson(a).
					Id(fmt.Sprintf("%s-%s", config.Get().Network, a.Slug())).
					Do(context.Background())
				log.Infof("Validation success %s", a.Hash)
			}
		}
	}
}

func loadBestBlock() {
	container.GetSubscriber().Subscribe(func() {
		block, err := container.GetBlockRepo().GetBestBlock()
		if err != nil {
			log.WithError(err).Fatal("Failed to get best block hash")
		}

		bestBlock = block.Height
	})
}
