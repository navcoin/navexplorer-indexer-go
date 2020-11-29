package elastic_cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/asaskevich/EventBus"
	"github.com/getsentry/raven-go"
	"github.com/olivere/elastic/v7"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"
)

type Index struct {
	Client        *elastic.Client
	cache         *cache.Cache
	event         EventBus.Bus
	bulkIndexSize uint64
}

type Request struct {
	Index  string
	Entity explorer.Entity
	Type   RequestType
}

type RequestType string

var (
	IndexRequest  RequestType = "index"
	UpdateRequest RequestType = "update"
)

var (
	ErrResultsNotFound = errors.New("Results not found")
	ErrRecordNotFound  = errors.New("Record not found")
)

func New(event EventBus.Bus) (*Index, error) {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(strings.Join(config.Get().ElasticSearch.Hosts, ",")),
		elastic.SetSniff(config.Get().ElasticSearch.Sniff),
		elastic.SetHealthcheck(config.Get().ElasticSearch.HealthCheck),
	}

	if config.Get().ElasticSearch.Username != "" {
		opts = append(opts, elastic.SetBasicAuth(
			config.Get().ElasticSearch.Username,
			config.Get().ElasticSearch.Password,
		))
	}

	if config.Get().ElasticSearch.Debug {
		opts = append(opts, elastic.SetTraceLog(logrus.StandardLogger()))
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		log.Println("Error: ", err)
	}

	return &Index{
		Client:        client,
		event:         event,
		cache:         cache.New(5*time.Minute, 10*time.Minute),
		bulkIndexSize: config.Get().BulkIndexSize,
	}, err
}

func (i *Index) InstallMappings() {
	logrus.Info("Install Mappings")
	files, err := ioutil.ReadDir(config.Get().ElasticSearch.MappingDir)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize ES")
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.Get().ElasticSearch.MappingDir, f.Name()))
		if err != nil {
			logrus.WithError(err).Fatal("Failed to initialize ES")
		}

		index := fmt.Sprintf("%s.%s.%s", config.Get().Network, config.Get().Index, f.Name()[0:len(f.Name())-len(filepath.Ext(f.Name()))])
		if err = i.createIndex(index, b); err != nil {
			logrus.WithError(err).Fatal("Failed to initialize ES")
		}
	}
}

func (i *Index) AddIndexRequest(index string, entity explorer.Entity) {
	logrus.WithFields(logrus.Fields{"slug": entity.Slug()}).Debug("AddIndexRequest")
	i.AddRequest(index, entity, IndexRequest)
}

func (i *Index) AddUpdateRequest(index string, entity explorer.Entity) {
	logrus.WithFields(logrus.Fields{"slug": entity.Slug()}).Debug("AddUpdateRequest")
	i.AddRequest(index, entity, UpdateRequest)
}

func (i *Index) AddRequest(index string, entity explorer.Entity, reqType RequestType) {
	logrus.WithFields(logrus.Fields{
		"index": index,
		"type":  reqType,
		"slug":  entity.Slug(),
	}).Debugf("AddRequest")

	if cached, found := i.cache.Get(entity.Slug()); found == true && cached.(Request).Type == IndexRequest {
		logrus.Debugf("Switch update to index as not previously persisted %s", entity.Slug())
		reqType = IndexRequest
	}
	//i.cache.Delete(entity.Slug())
	//if reqType == UpdateRequest && entity.Slug() == "" {
	//	logrus.Debugf("Switch update to index as not previously persisted %s", entity.Slug())
	//	reqType = IndexRequest
	//}

	i.cache.Set(entity.Slug(), Request{index, entity, reqType}, cache.DefaultExpiration)
}

func (i *Index) GetRequests() []Request {
	requests := make([]Request, 0)

	for _, item := range i.cache.Items() {
		requests = append(requests, item.Object.(Request))
	}

	return requests
}

func (i *Index) GetRequest(id string) *Request {
	if item, found := i.cache.Get(id); found == true {
		req := item.(Request)
		return &req
	} else {
		return nil
	}
}

func (i *Index) Save(index Indices, entity explorer.Entity) {
	logrus.WithFields(logrus.Fields{"slug": entity.Slug()}).Debug("Save")

	_, err := i.Client.Index().
		Index(index.Get()).
		Id(entity.Slug()).
		BodyJson(entity).
		Do(context.Background())

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"index": index.Get(),
			"slug":  entity.Slug(),
		}).Fatal("Failed to save entity")
	}
}

func (i *Index) BatchPersist(height uint64) bool {
	if height%i.bulkIndexSize != 0 {
		return false
	}

	i.event.Publish("batch.persist.start", height)

	logrus.Infof("Persisting data at height   %d", height)
	i.Persist()
	waitOnReindex(height)

	return true
}

func (i *Index) Persist() int {
	bulk := i.Client.Bulk()
	for _, r := range i.GetRequests() {
		if r.Type == IndexRequest {
			bulk.Add(elastic.NewBulkIndexRequest().Index(r.Index).Id(r.Entity.Slug()).Doc(r.Entity))
		} else if r.Type == UpdateRequest {
			bulk.Add(elastic.NewBulkUpdateRequest().Index(r.Index).Id(r.Entity.Slug()).Doc(r.Entity))
		}
	}

	actions := bulk.NumberOfActions()
	if actions != 0 {
		i.persist(bulk)
	}

	return actions
}

func (i *Index) persist(bulk *elastic.BulkService) {
	actions := bulk.NumberOfActions()
	response, err := bulk.Do(context.Background())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to persist requests")
	}

	if response.Errors == true {
		for _, failed := range response.Failed() {
			logrus.WithFields(logrus.Fields{
				"error": failed.Error,
				"index": failed.Index,
				"id":    failed.Id,
			}).Fatal("Failed to persist to ES")
		}
	}
	logrus.Debugf("Persisted %d actions", actions)

	logrus.Debug("Flushing ES cache")
	i.cache.Flush()
}

func (i *Index) DeleteHeightGT(height uint64, indices ...string) error {
	_, err := i.Client.DeleteByQuery(indices...).
		Query(elastic.NewRangeQuery("height").Gt(height)).
		Do(context.Background())
	if err != nil {
		logrus.WithError(err).Fatalf("Could not rewind to %d", height)
		return err
	}

	i.Client.Flush(indices...)

	logrus.Debugf("Deleted height greater than %d", height)

	return nil
}

func (i *Index) createIndex(index string, mapping []byte) error {
	ctx := context.Background()
	client := i.Client

	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		raven.CaptureError(err, nil)
		return err
	}

	if exists && config.Get().Reindex {
		logrus.Infof("Deleting Index: %s", index)
		_, err = client.DeleteIndex(index).Do(ctx)
		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}
		exists = false
	}

	if !exists {
		createIndex, err := client.CreateIndex(index).BodyString(string(mapping)).Do(ctx)
		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}

		if createIndex.Acknowledged {
			logrus.Info("Created index: ", index)
		}
	}

	return nil
}

func waitOnReindex(height uint64) {
	if config.Get().Reindex == true && height == config.Get().BulkIndexSize {
		logrus.Error("Stopping reindex at height:", height)
		for {
			switch {
			}
		}
	}
}
