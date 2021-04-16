package block

import (
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/patrickmn/go-cache"
)

type Service struct {
	repo  *Repository
	cache *cache.Cache
}

func NewService(repo *Repository, cache *cache.Cache) *Service {
	return &Service{repo, cache}
}

var (
	ErrBlockNotFound            = errors.New("Block not found")
	ErrBlockTransactionNotFound = errors.New("Transaction not found")
)

func (s *Service) SetLastBlockIndexed(block *explorer.Block) {
	s.cache.Set("lastBlockIndexed", *block, cache.NoExpiration)
}

func (s *Service) GetLastBlockIndexed() *explorer.Block {
	if lastBlockIndexed, exists := s.cache.Get("lastBlockIndexed"); exists {
		block := lastBlockIndexed.(explorer.Block)
		return &block
	}

	block, err := s.repo.GetBestBlock()
	if err != nil {
		return nil
	}
	s.SetLastBlockIndexed(block)

	return block
}
