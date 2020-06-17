package validator

import (
	"context"
	_ "github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/service/address"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
	"time"
)

var container *dic.Container

type AddressValidator struct{}

func (v *AddressValidator) Execute() {
	log.Info("NavExplorer Address Validator")
	config.Init()
	container, _ = dic.NewContainer(dingo.App)

	log.Info("Loading all addresses")
	addresses, err := container.GetAddressRepo().GetAllAddresses()
	if err != nil {
		log.WithError(err).Fatal("Failed to get all addresses")
	}
	log.Infof("Validating %d addresses", len(addresses))

	for idx, a := range addresses {
		if a.Height < 4233114 || a.ValidatedAt != 0 {
			continue
		}
		log.Infof("%d - %s", idx, a.Hash)

		address.ResetAddress(a)
		if txs, err := container.GetBlockRepo().GetAllTransactionsThatIncludeAddress(a.Hash); err == nil {
			log.Infof("Loaded block transactions")
			for _, tx := range txs {
				block, err := container.GetBlockRepo().GetBlockByHeight(tx.Height)
				if err != nil {
					log.WithError(err).Fatalf("Failed to get block at height %d", tx.Height)
				}
				addressTxs := container.GetAddressIndexer().GenerateAddressTransactions(a, tx, block)
				for _, addressTx := range addressTxs {
					address.ApplyTxToAddress(a, addressTx)
					_, err = container.GetElastic().Client.Index().
						Index(elastic_cache.AddressTransactionIndex.Get()).
						BodyJson(addressTx).
						Id(addressTx.Slug()).
						Do(context.Background())
				}
			}

			a.ValidatedAt = uint64(time.Now().Unix())
			_, err = container.GetElastic().Client.Index().
				Index(elastic_cache.AddressIndex.Get()).
				BodyJson(a).
				Id(a.Slug()).
				Do(context.Background())
		}
	}
}
