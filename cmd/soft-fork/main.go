package main

import (
	"github.com/NavExplorer/navexplorer-indexer-go/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
)

var container *dic.Container

func main() {
	config.Init()

	container, _ = dic.NewContainer(dingo.App)
	container.GetElastic().InstallMappings()
	container.GetSoftforkService().InitSoftForks()
	bestBlock, err := container.GetBlockRepo().GetBestBlock()
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= int(bestBlock.Height); i += 2016 {
		blocks, err := container.GetBlockRepo().GetBlocksBetweenHeight(uint64(i), uint64(i+2016))
		if err != nil {
			log.Fatal(err)
		}
		for _, block := range blocks {
			container.GetSoftforkIndexer().Index(block)
		}
		container.GetElastic().Persist()
	}
	container.GetElastic().Persist()
	//block, err := container.GetBlockRepo().GetBlockByHeight()
	//if err == nil {
	//	log.Info("Rewind SoftForks to height: ", block.Height)
	////	container.GetSoftforkRewinder().Rewind(block.Height - 10)
	//}
}
