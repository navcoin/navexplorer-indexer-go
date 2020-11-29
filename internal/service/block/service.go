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
	ErrBlockNotFound            = errors.New("Block not found")
	ErrBlockTransactionNotFound = errors.New("Transaction not found")
)

func (s *Service) GetLastBlockIndexed() *explorer.Block {
	if LastBlockIndexed != nil {
		return LastBlockIndexed
	}

	block, err := s.repo.GetBestBlock()
	if err != nil {
		return nil
	}

	return block
}
