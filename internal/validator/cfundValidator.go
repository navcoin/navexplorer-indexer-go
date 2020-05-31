package validator

import (
	"github.com/NavExplorer/navexplorer-indexer-go/generated/dic"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
	"github.com/sarulabs/dingo/v3"
)

type CfundValidator struct{}

func (v *CfundValidator) Execute() {
	config.Init()
	container, _ = dic.NewContainer(dingo.App)

	container.GetBlockRepo()
}
