package block

import (
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo}
}

var (
	ErrBlockNotFound                            = errors.New("Block not found")
	ErrBlockTransactionNotFound                 = errors.New("Transaction not found")
	lastBlockIndexed            *explorer.Block = nil
)

func (s *Service) setLastBlockIndexed(block *explorer.Block) {
	lastBlockIndexed = block
}

func (s *Service) GetLastBlockIndexed() *explorer.Block {
	if lastBlockIndexed != nil {
		return lastBlockIndexed
	}

	block, err := s.repo.GetBestBlock()
	if err != nil {
		return nil
	}
	lastBlockIndexed = block

	return lastBlockIndexed
}
