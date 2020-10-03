package main

import (
	"github.com/NavExplorer/navexplorer-indexer-go/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/sarulabs/dingo/v3"
	"os"
)

func main() {
	config.Init()

	container, _ := dic.NewContainer(dingo.App)
	container.GetElastic().InstallMappings()

	addressRepo := container.GetAddressRepo()
	rewinder := container.GetAddressRewinder()

	address, err := addressRepo.GetAddress(os.Args[1])
	if err == nil {
		rewinder.ResetAddress(address)
	}
}
