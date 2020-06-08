package validator

import (
	"context"
	_ "github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
	"time"
)

var container *dic.Container

var firstBlock uint64
var bestBlock uint64

type AddressValidator struct{}

func (v *AddressValidator) Execute() {
	log.Info("NavExplorer Address Validator")
	config.Init()
	container, _ = dic.NewContainer(dingo.App)

	go loadBestBlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for {
		if ctx.Err() != nil {
			log.Fatal("Failed to find the best block")
		}
		if bestBlock != 0 {
			log.Info("Validating from block ", bestBlock)
			firstBlock = bestBlock
			break
		}
	}

	log.Info("Validating")
	validateAddresses()
}

func loadBestBlock() {
	_ = container.GetSubscriber().Subscribe(func() {
		block, err := container.GetBlockRepo().GetBestBlock()
		if err != nil {
			log.WithError(err).Fatal("Failed to get best block hash")
		}
		bestBlock = block.Height
	})
}

func validateAddresses() {
	count := 0
	//for {
	addresses, err := container.GetAddressRepo().GetAddressesByValidateAtDesc(1000)
	if err != nil {
		log.WithError(err).Fatal("Failed to get addresses")
	}

	log.Infof("Validating %d addresses", len(addresses))
	count = count + len(addresses)

	for _, address := range addresses {
		clone := address
		if address.ValidatedAt > firstBlock {
			return
		}

		currentBlock := bestBlock
		previousBalance := address.Balance
		err := container.GetAddressRewinder().ResetAddressToHeight(clone, currentBlock)
		if err != nil {
			log.WithError(err).Fatal("Failed to reset address")
		}

		//a.ValidatedAt = currentBlock

		if previousBalance != clone.Balance {
			//a.ValidatedAt = 0
			log.WithFields(log.Fields{
				"previous": previousBalance,
				"current":  clone.Balance,
			}).Errorf("Validation error: The balance was incorrect %s", clone.Hash)
		}

		if clone.Received+clone.Staked+clone.Sent != clone.Balance {
			//a.ValidatedAt = 0
			log.WithFields(log.Fields{
				"staked":         clone.Staked,
				"sent":           clone.Sent,
				"received":       clone.Received,
				"balance-calc":   clone.Received + clone.Staked - clone.Sent,
				"balance-actual": clone.Balance,
				"height":         clone.Height,
			}).Errorf("Validation error: Transactions don't equate to balance %s", clone.Hash)
		}

		if clone.ColdReceived+clone.ColdStaked+clone.ColdSent != clone.ColdBalance {
			//a.ValidatedAt = 0
			log.Errorf("Validation error: Transactions don't equate to cold balance %s", clone.Hash)
		}

		if currentBlock == bestBlock {
			address.ValidatedAt = currentBlock

			//_, err = container.GetElastic().Client.Index().
			//	Index(elastic_cache.AddressIndex.Get()).
			//	BodyJson(address).
			//	Id(fmt.Sprintf("%s-%s", config.Get().Network, address.Slug())).
			//	Do(context.Background())
		}
	}
	//}
}
