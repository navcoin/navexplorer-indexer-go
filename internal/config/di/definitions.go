package di

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/indexer"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/queue"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/address"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/consultation"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/payment_request"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/proposal"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/dao/vote"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/softfork/signal"
	"github.com/NavExplorer/subscriber"
	"github.com/patrickmn/go-cache"
	"github.com/sarulabs/dingo/v3"
	"go.uber.org/zap"
	"time"
)

var Definitions = []dingo.Def{
	{
		Name: "navcoin",
		Build: func() (*navcoind.Navcoind, error) {
			c := config.Get().Navcoind
			navcoin, err := navcoind.New(c.Host, c.Port, c.User, c.Password, c.Ssl, config.Get().Debug, c.Timeout)
			if err != nil {
				zap.L().With(zap.Error(err)).Fatal("Failed to initialize Navcoind")
			}
			return navcoin, nil
		},
	},
	{
		Name: "elastic",
		Build: func() (elastic_cache.Index, error) {
			elastic, err := elastic_cache.New()
			if err != nil {
				zap.L().With(zap.Error(err)).Fatal("Failed to start ES")
			}

			return elastic, nil
		},
	},
	{
		Name: "cache",
		Build: func() (*cache.Cache, error) {
			return cache.New(5*time.Minute, 10*time.Minute), nil
		},
	},
	{
		Name: "address.repo",
		Build: func(elastic elastic_cache.Index) (address.Repository, error) {
			return address.NewRepo(elastic), nil
		},
	},
	{
		Name: "block.repo",
		Build: func(elastic elastic_cache.Index) (block.Repository, error) {
			return block.NewRepo(elastic), nil
		},
	},
	{
		Name: "signal.repo",
		Build: func(elastic elastic_cache.Index) (signal.Repository, error) {
			return signal.NewRepo(elastic), nil
		},
	},
	{
		Name: "softfork.repo",
		Build: func(elastic elastic_cache.Index) (softfork.Repository, error) {
			return softfork.NewRepo(elastic), nil
		},
	},
	{
		Name: "indexer",
		Build: func(
			elastic elastic_cache.Index,
			publisher queue.Publisher,
			blockIndexer block.Indexer,
			blockService block.Service,
			addressIndexer address.Indexer,
			softForkIndexer softfork.Indexer,
			daoIndexer dao.Indexer,
			rewinder indexer.Rewinder,
		) (indexer.Indexer, error) {
			return indexer.NewIndexer(elastic, publisher, blockIndexer, blockService, addressIndexer, softForkIndexer, daoIndexer, rewinder), nil
		},
	},
	{
		Name: "rewinder",
		Build: func(
			elastic elastic_cache.Index,
			addressRewinder address.Rewinder,
			blockRewinder block.Rewinder,
			softforkRewinder softfork.Rewinder,
			daoRewinder dao.Rewinder,
			blockService block.Service,
			blockRepo block.Repository,
		) (indexer.Rewinder, error) {
			return indexer.NewRewinder(elastic, blockRewinder, addressRewinder, softforkRewinder, daoRewinder, blockService, blockRepo), nil
		},
	},
	{
		Name: "block.service",
		Build: func(repository block.Repository, cache *cache.Cache) (block.Service, error) {
			return block.NewService(repository, cache), nil
		},
	},
	{
		Name: "block.indexer",
		Build: func(
			navcoin *navcoind.Navcoind,
			elastic elastic_cache.Index,
			orphanedService block.OrphanService,
			repository block.Repository,
			service block.Service,
			consensusService consensus.Service,
		) (block.Indexer, error) {
			return block.NewIndexer(navcoin, elastic, orphanedService, repository, service, consensusService), nil
		},
	},
	{
		Name: "block.rewinder",
		Build: func(elastic elastic_cache.Index) (block.Rewinder, error) {
			return block.NewRewinder(elastic), nil
		},
	},
	{
		Name: "address.indexer",
		Build: func(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, cache *cache.Cache, addressRepo address.Repository, blockService block.Service, blockRepo block.Repository) (address.Indexer, error) {
			return address.NewIndexer(navcoin, elastic, cache, addressRepo, blockService, blockRepo), nil
		},
	},
	{
		Name: "address.rewinder",
		Build: func(elastic elastic_cache.Index, repository address.Repository, indexer address.Indexer) (address.Rewinder, error) {
			return address.NewRewinder(elastic, repository, indexer), nil
		},
	},
	{
		Name: "block.orphan.service",
		Build: func(blockRepository block.Repository) (block.OrphanService, error) {
			return block.NewOrphanService(blockRepository), nil
		},
	},
	{
		Name: "softfork.indexer",
		Build: func(elastic elastic_cache.Index) (softfork.Indexer, error) {
			return softfork.NewIndexer(elastic, uint(config.Get().SoftForkBlockCycle), config.Get().SoftForkQuorum), nil
		},
	},
	{
		Name: "softfork.rewinder",
		Build: func(elastic elastic_cache.Index, signalRepo signal.Repository) (softfork.Rewinder, error) {
			return softfork.NewRewinder(elastic, signalRepo, uint(config.Get().SoftForkBlockCycle), config.Get().SoftForkQuorum), nil
		},
	},
	{
		Name: "softfork.service",
		Build: func(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, repository softfork.Repository) (softfork.Service, error) {
			return softfork.New(navcoin, elastic, repository), nil
		},
	},
	{
		Name: "dao.Indexer",
		Build: func(navcoin *navcoind.Navcoind, proposalIndexer proposal.Indexer, paymentRequestIndexer payment_request.Indexer, consultationIndexer consultation.Indexer, voteIndexer vote.Indexer, consensusIndexer consensus.Indexer) (dao.Indexer, error) {
			return dao.NewIndexer(
				navcoin,
				proposalIndexer,
				paymentRequestIndexer,
				consultationIndexer,
				voteIndexer,
				consensusIndexer,
			), nil
		},
	},
	{
		Name: "dao.consensus.Indexer",
		Build: func(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, repository consensus.Repository, service consensus.Service) (consensus.Indexer, error) {
			return consensus.NewIndexer(navcoin, elastic, repository, service), nil
		},
	},
	{
		Name: "dao.consensus.Rewinder",
		Build: func(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, repository consensus.Repository, service consensus.Service, consultationRepo consultation.Repository) (consensus.Rewinder, error) {
			return consensus.NewRewinder(navcoin, elastic, repository, service), nil
		},
	},
	{
		Name: "dao.proposal.Indexer",
		Build: func(navcoin *navcoind.Navcoind, elastic elastic_cache.Index) (proposal.Indexer, error) {
			return proposal.NewIndexer(navcoin, elastic, config.Get().ReindexSize), nil
		},
	},
	{
		Name: "dao.payment_request.Indexer",
		Build: func(navcoin *navcoind.Navcoind, elastic elastic_cache.Index) (payment_request.Indexer, error) {
			return payment_request.NewIndexer(navcoin, elastic), nil
		},
	},
	{
		Name: "dao.consultation.Indexer",
		Build: func(navcoin *navcoind.Navcoind, elastic elastic_cache.Index, blockRepo block.Repository, consensusService consensus.Service) (consultation.Indexer, error) {
			return consultation.NewIndexer(navcoin, elastic, blockRepo, consensusService), nil
		},
	},
	{
		Name: "dao.vote.Indexer",
		Build: func(elastic elastic_cache.Index) (vote.Indexer, error) {
			return vote.NewIndexer(elastic), nil
		},
	},
	{
		Name: "dao.Rewinder",
		Build: func(elastic elastic_cache.Index, consensusRewinder consensus.Rewinder, repository consultation.Repository) (dao.Rewinder, error) {
			return dao.NewRewinder(elastic, consensusRewinder, repository), nil
		},
	},
	{
		Name: "dao.consultation.Service",
		Build: func(repository consultation.Repository) (consultation.Service, error) {
			return consultation.NewService(repository), nil
		},
	},
	{
		Name: "dao.consensus.Service",
		Build: func(elastic elastic_cache.Index, cache *cache.Cache, repository consensus.Repository) (consensus.Service, error) {
			return consensus.NewService(config.Get().Network, elastic, cache, repository), nil
		},
	},
	{
		Name: "dao.proposal.Service",
		Build: func(repository proposal.Repository) (proposal.Service, error) {
			return proposal.NewService(repository), nil
		},
	},
	{
		Name: "dao.payment_request.Service",
		Build: func(repository payment_request.Repository) (payment_request.Service, error) {
			return payment_request.NewService(repository), nil
		},
	},
	{
		Name: "dao.payment_request.repo",
		Build: func(elastic elastic_cache.Index) (payment_request.Repository, error) {
			return payment_request.NewRepo(elastic), nil
		},
	},
	{
		Name: "dao.consultation.repo",
		Build: func(elastic elastic_cache.Index) (consultation.Repository, error) {
			return consultation.NewRepo(elastic), nil
		},
	},
	{
		Name: "dao.consensus.repo",
		Build: func(elastic elastic_cache.Index) (consensus.Repository, error) {
			return consensus.NewRepo(elastic), nil
		},
	},
	{
		Name: "dao.proposal.repo",
		Build: func(elastic elastic_cache.Index) (proposal.Repository, error) {
			return proposal.NewRepo(elastic), nil
		},
	},
	{
		Name: "subscriber",
		Build: func() (*subscriber.Subscriber, error) {
			return subscriber.NewSubscriber(config.Get().ZeroMq.Address), nil
		},
	},
	{
		Name: "queue.publisher",
		Build: func() (queue.Publisher, error) {
			return queue.NewPublisher(
				config.Get().Network,
				config.Get().Index,
				config.Get().RabbitMq.User,
				config.Get().RabbitMq.Password,
				config.Get().RabbitMq.Host,
				config.Get().RabbitMq.Port,
			), nil
		},
	},
	{
		Name: "queue.consumer",
		Build: func() (*queue.Consumer, error) {
			return queue.NewConsumer(
				config.Get().RabbitMq.User,
				config.Get().RabbitMq.Password,
				config.Get().RabbitMq.Host,
				config.Get().RabbitMq.Port,
			), nil
		},
	},
}
