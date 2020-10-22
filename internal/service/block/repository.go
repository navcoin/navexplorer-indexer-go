package block

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/getsentry/raven-go"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type Repository struct {
	elastic *elastic_cache.Index
}

func NewRepo(elastic *elastic_cache.Index) *Repository {
	return &Repository{elastic}
}

func (r *Repository) GetBestBlock() (*explorer.Block, error) {
	result, err := r.elastic.Client.
		Search(elastic_cache.BlockIndex.Get()).
		Sort("height", false).
		Size(1).
		Do(context.Background())
	if err != nil || result == nil {
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, ErrBlockNotFound
	}

	var block *explorer.Block
	hit := result.Hits.Hits[0]
	if err = json.Unmarshal(hit.Source, &block); err != nil {
		return nil, err
	}
	block.SetId(hit.Id)
	return block, nil
}

func (r *Repository) GetHeight() (uint64, error) {
	b, err := r.GetBestBlock()
	if err != nil {
		return 0, err
	}

	return b.Height, nil
}

func (r *Repository) GetBlockByHash(hash string) (*explorer.Block, error) {
	results, err := r.elastic.Client.
		Search(elastic_cache.BlockIndex.Get()).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	if len(results.Hits.Hits) == 0 {
		return nil, elastic_cache.ErrRecordNotFound
	}

	var block *explorer.Block
	hit := results.Hits.Hits[0]
	if err = json.Unmarshal(hit.Source, &block); err != nil {
		return nil, err
	}
	block.SetId(hit.Id)

	return block, nil
}

func (r *Repository) GetTransactionsByBlock(block *explorer.Block) ([]*explorer.BlockTransaction, error) {
	result, err := r.elastic.Client.
		Search(elastic_cache.BlockTransactionIndex.Get()).
		Query(elastic.NewMatchQuery("height", block.Height)).
		Do(context.Background())
	if err != nil || result == nil {
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, elastic_cache.ErrRecordNotFound
	}

	var txs []*explorer.BlockTransaction
	for _, hit := range result.Hits.Hits {
		var tx *explorer.BlockTransaction
		if err = json.Unmarshal(hit.Source, &tx); err != nil {
			raven.CaptureError(err, nil)
			return nil, err
		}
		tx.SetId(hit.Id)
		txs = append(txs, tx)
	}

	return txs, nil
}

func (r *Repository) GetTransactionByHash(hash string) (*explorer.BlockTransaction, error) {
	request := r.elastic.GetRequest(explorer.CreateBlockTxSlug(hash))
	if request != nil {
		return request.Entity.(*explorer.BlockTransaction), nil
	}

	getTransactionByHash := func(hash string) (*elastic.SearchResult, error) {
		return r.elastic.Client.
			Search(elastic_cache.BlockTransactionIndex.Get()).
			Query(elastic.NewTermQuery("hash.keyword", hash)).
			Size(1).
			Do(context.Background())
	}

	results, err := getTransactionByHash(hash)
	if err != nil || results == nil {
		return nil, err
	}

	if len(results.Hits.Hits) == 0 {
		log.Infof("Failed to find record, retrying in 5 seconds")
		time.Sleep(5 * time.Second)
		results, err = getTransactionByHash(hash)
		if err != nil || results == nil {
			return nil, err
		}

		if len(results.Hits.Hits) == 0 {
			return nil, elastic_cache.ErrRecordNotFound
		}
	}

	var tx *explorer.BlockTransaction
	hit := results.Hits.Hits[0]
	if err = json.Unmarshal(hit.Source, &tx); err != nil {
		return nil, err
	}
	tx.SetId(hit.Id)

	return tx, nil
}

func (r *Repository) GetBlockByHeight(height uint64) (*explorer.Block, error) {
	results, err := r.elastic.Client.
		Search(elastic_cache.BlockIndex.Get()).
		Query(elastic.NewMatchQuery("height", height)).
		Size(1).
		Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	if len(results.Hits.Hits) == 0 {
		return nil, elastic_cache.ErrRecordNotFound
	}

	var block *explorer.Block
	hit := results.Hits.Hits[0]
	if err = json.Unmarshal(hit.Source, &block); err != nil {
		return nil, err
	}
	block.SetId(hit.Id)

	return block, nil
}

func (r *Repository) GetBlocksBetweenHeight(start uint64, end uint64) ([]*explorer.Block, error) {
	results, err := r.elastic.Client.
		Search(elastic_cache.BlockIndex.Get()).
		Query(elastic.NewRangeQuery("height").Gte(start).Lt(end)).
		Sort("height", true).
		Size(int(end - start)).
		Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	blocks := make([]*explorer.Block, 0)
	for _, hit := range results.Hits.Hits {
		var block *explorer.Block
		if err = json.Unmarshal(hit.Source, &block); err != nil {
			return nil, err
		}
		block.SetId(hit.Id)
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (r *Repository) GetTransactionsWithCfundPayment() error {
	results, err := r.elastic.Client.Search(elastic_cache.BlockTransactionIndex.Get()).
		Query(elastic.NewMatchQuery("vout.scriptPubKey.type.keyword", explorer.VoutCfundContribution)).
		Do(context.Background())

	if err != nil || results == nil {
		return err
	}

	return nil
}

func (r *Repository) GetAllTransactionsThatIncludeAddress(hash string) ([]*explorer.BlockTransaction, error) {
	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewNestedQuery("vin",
		elastic.NewBoolQuery().Must(elastic.NewMatchQuery("vin.addresses.keyword", hash))))
	query = query.Should(elastic.NewNestedQuery("vout",
		elastic.NewBoolQuery().Must(elastic.NewMatchQuery("vout.scriptPubKey.addresses.keyword", hash))))

	service := r.elastic.Client.Scroll(elastic_cache.BlockTransactionIndex.Get()).Query(query).Size(10000).Sort("height", true)
	txs := make([]*explorer.BlockTransaction, 0)

	for {
		results, err := service.Do(context.Background())
		if err == io.EOF {
			break
		}
		if err != nil || results == nil {
			log.Fatal(err)
		}

		for _, hit := range results.Hits.Hits {
			var tx *explorer.BlockTransaction
			if err = json.Unmarshal(hit.Source, &tx); err != nil {
				raven.CaptureError(err, nil)
				return nil, err
			}
			tx.SetId(hit.Id)
			txs = append(txs, tx)
		}
	}

	return txs, nil
}
