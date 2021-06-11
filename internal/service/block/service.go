package block

import (
	"errors"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/patrickmn/go-cache"
)

type Service interface {
	SetLastBlockIndexed(block *explorer.Block)
	GetLastBlockIndexed() *explorer.Block
	ClearLastBlockIndexed()
}

type service struct {
	repository Repository
	cache      *cache.Cache
}

func NewService(repository Repository, cache *cache.Cache) Service {
	return service{repository, cache}
}

var (
	ErrBlockNotFound            = errors.New("Block not found")
	ErrBlockTransactionNotFound = errors.New("Transaction not found")
)

func (s service) ClearLastBlockIndexed() {
	s.cache.Delete("lastBlockIndexed")
}

func (s service) SetLastBlockIndexed(block *explorer.Block) {
	s.cache.Set("lastBlockIndexed", *block, cache.NoExpiration)
}

func (s service) GetLastBlockIndexed() *explorer.Block {
	if lastBlockIndexed, exists := s.cache.Get("lastBlockIndexed"); exists {
		block := lastBlockIndexed.(explorer.Block)
		return &block
	}

	block, err := s.repository.GetBestBlock()
	if err != nil {
		return nil
	}
	s.SetLastBlockIndexed(block)

	return block
}
