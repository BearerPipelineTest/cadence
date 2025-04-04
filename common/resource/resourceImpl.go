// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package resource

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	persistenceClient "github.com/uber/cadence/common/persistence/client"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/service"
)

type (

	// VisibilityManagerInitializer is the function each service should implement
	// for visibility manager initialization
	VisibilityManagerInitializer func(
		persistenceBean persistenceClient.Bean,
		logger log.Logger,
	) (persistence.VisibilityManager, error)

	// Impl contains all common resources shared across frontend / matching / history / worker
	Impl struct {
		status int32

		// static infos
		numShards       int
		serviceName     string
		hostInfo        membership.HostInfo
		metricsScope    tally.Scope
		clusterMetadata cluster.Metadata

		// other common resources

		domainCache             cache.DomainCache
		domainMetricsScopeCache cache.DomainMetricsScopeCache
		timeSource              clock.TimeSource
		payloadSerializer       persistence.PayloadSerializer
		metricsClient           metrics.Client
		messagingClient         messaging.Client
		blobstoreClient         blobstore.Client
		archivalMetadata        archiver.ArchivalMetadata
		archiverProvider        provider.ArchiverProvider
		domainReplicationQueue  domain.ReplicationQueue

		// membership infos

		membershipResolver membership.Resolver

		// internal services clients

		sdkClient         workflowserviceclient.Interface
		frontendRawClient frontend.Client
		frontendClient    frontend.Client
		matchingRawClient matching.Client
		matchingClient    matching.Client
		historyRawClient  history.Client
		historyClient     history.Client
		clientBean        client.Bean

		// persistence clients
		persistenceBean persistenceClient.Bean

		// loggers
		logger          log.Logger
		throttledLogger log.Logger

		// for registering handlers
		dispatcher *yarpc.Dispatcher

		// internal vars

		pprofInitializer       common.PProfInitializer
		runtimeMetricsReporter *metrics.RuntimeMetricsReporter
		rpcFactory             common.RPCFactory
	}
)

var _ Resource = (*Impl)(nil)

// New create a new resource containing common dependencies
func New(
	params *Params,
	serviceName string,
	serviceConfig *service.Config,
) (impl *Impl, retError error) {

	logger := params.Logger
	throttledLogger := loggerimpl.NewThrottledLogger(logger, serviceConfig.ThrottledLoggerMaxRPS)

	numShards := params.PersistenceConfig.NumHistoryShards
	dispatcher := params.RPCFactory.GetDispatcher()
	membershipResolver := params.MembershipResolver

	dynamicCollection := dynamicconfig.NewCollection(
		params.DynamicConfig,
		logger,
		dynamicconfig.ClusterNameFilter(params.ClusterMetadata.GetCurrentClusterName()),
	)
	clientBean, err := client.NewClientBean(
		client.NewRPCClientFactory(
			params.RPCFactory,
			membershipResolver,
			params.MetricsClient,
			dynamicCollection,
			numShards,
			logger,
		),
		params.RPCFactory.GetDispatcher(),
		params.ClusterMetadata,
	)
	if err != nil {
		return nil, err
	}

	persistenceBean, err := persistenceClient.NewBeanFromFactory(persistenceClient.NewFactory(
		&params.PersistenceConfig,
		quotas.PerMemberDynamic(
			serviceName,
			serviceConfig.PersistenceGlobalMaxQPS.AsFloat64(),
			serviceConfig.PersistenceMaxQPS.AsFloat64(),
			membershipResolver,
		),
		params.ClusterMetadata.GetCurrentClusterName(),
		params.MetricsClient,
		logger,
		persistence.NewDynamicConfiguration(dynamicCollection),
	), &persistenceClient.Params{
		PersistenceConfig: params.PersistenceConfig,
		MetricsClient:     params.MetricsClient,
		MessagingClient:   params.MessagingClient,
		ESClient:          params.ESClient,
		ESConfig:          params.ESConfig,
	}, serviceConfig)
	if err != nil {
		return nil, err
	}

	domainCache := cache.NewDomainCache(
		persistenceBean.GetDomainManager(),
		params.MetricsClient,
		logger,
	)

	domainMetricsScopeCache := cache.NewDomainMetricsScopeCache()
	domainReplicationQueue := domain.NewReplicationQueue(
		persistenceBean.GetDomainReplicationQueueManager(),
		params.ClusterMetadata.GetCurrentClusterName(),
		params.MetricsClient,
		logger,
	)

	frontendRawClient := clientBean.GetFrontendClient()
	frontendClient := frontend.NewRetryableClient(
		frontendRawClient,
		common.CreateFrontendServiceRetryPolicy(),
		common.IsServiceTransientError,
	)

	matchingRawClient, err := clientBean.GetMatchingClient(domainCache.GetDomainName)
	if err != nil {
		return nil, err
	}
	matchingClient := matching.NewRetryableClient(
		matchingRawClient,
		common.CreateMatchingServiceRetryPolicy(),
		common.IsServiceTransientError,
	)

	historyRawClient := clientBean.GetHistoryClient()
	historyClient := history.NewRetryableClient(
		historyRawClient,
		common.CreateHistoryServiceRetryPolicy(),
		common.IsServiceTransientError,
	)

	historyArchiverBootstrapContainer := &archiver.HistoryBootstrapContainer{
		HistoryV2Manager: persistenceBean.GetHistoryManager(),
		Logger:           logger,
		MetricsClient:    params.MetricsClient,
		ClusterMetadata:  params.ClusterMetadata,
		DomainCache:      domainCache,
	}
	visibilityArchiverBootstrapContainer := &archiver.VisibilityBootstrapContainer{
		Logger:          logger,
		MetricsClient:   params.MetricsClient,
		ClusterMetadata: params.ClusterMetadata,
		DomainCache:     domainCache,
	}
	if err := params.ArchiverProvider.RegisterBootstrapContainer(
		serviceName,
		historyArchiverBootstrapContainer,
		visibilityArchiverBootstrapContainer,
	); err != nil {
		return nil, err
	}

	impl = &Impl{
		status: common.DaemonStatusInitialized,

		// static infos

		numShards:       numShards,
		serviceName:     params.Name,
		metricsScope:    params.MetricScope,
		clusterMetadata: params.ClusterMetadata,

		// other common resources

		domainCache:             domainCache,
		domainMetricsScopeCache: domainMetricsScopeCache,
		timeSource:              clock.NewRealTimeSource(),
		payloadSerializer:       persistence.NewPayloadSerializer(),
		metricsClient:           params.MetricsClient,
		messagingClient:         params.MessagingClient,
		blobstoreClient:         params.BlobstoreClient,
		archivalMetadata:        params.ArchivalMetadata,
		archiverProvider:        params.ArchiverProvider,
		domainReplicationQueue:  domainReplicationQueue,

		// membership infos
		membershipResolver: membershipResolver,

		// internal services clients

		sdkClient:         params.PublicClient,
		frontendRawClient: frontendRawClient,
		frontendClient:    frontendClient,
		matchingRawClient: matchingRawClient,
		matchingClient:    matchingClient,
		historyRawClient:  historyRawClient,
		historyClient:     historyClient,
		clientBean:        clientBean,

		// persistence clients
		persistenceBean: persistenceBean,

		// loggers

		logger:          logger,
		throttledLogger: throttledLogger,

		// for registering handlers
		dispatcher: dispatcher,

		// internal vars
		pprofInitializer: params.PProfInitializer,
		runtimeMetricsReporter: metrics.NewRuntimeMetricsReporter(
			params.MetricScope,
			time.Minute,
			logger,
			params.InstanceID,
		),
		rpcFactory: params.RPCFactory,
	}
	return impl, nil
}

// Start start all resources
func (h *Impl) Start() {

	if !atomic.CompareAndSwapInt32(
		&h.status,
		common.DaemonStatusInitialized,
		common.DaemonStatusStarted,
	) {
		return
	}

	h.metricsScope.Counter(metrics.RestartCount).Inc(1)
	h.runtimeMetricsReporter.Start()

	if err := h.pprofInitializer.Start(); err != nil {
		h.logger.WithTags(tag.Error(err)).Fatal("fail to start PProf")
	}
	if err := h.dispatcher.Start(); err != nil {
		h.logger.WithTags(tag.Error(err)).Fatal("fail to start dispatcher")
	}
	h.membershipResolver.Start()
	h.domainCache.Start()
	h.domainMetricsScopeCache.Start()

	hostInfo, err := h.membershipResolver.WhoAmI()
	if err != nil {
		h.logger.WithTags(tag.Error(err)).Fatal("fail to get host info from membership monitor")
	}
	h.hostInfo = hostInfo

	// The service is now started up
	h.logger.Info("service started")
	// seed the random generator once for this service
	rand.Seed(time.Now().UTC().UnixNano())
}

// Stop stops all resources
func (h *Impl) Stop() {

	if !atomic.CompareAndSwapInt32(
		&h.status,
		common.DaemonStatusStarted,
		common.DaemonStatusStopped,
	) {
		return
	}

	h.domainCache.Stop()
	h.domainMetricsScopeCache.Stop()
	h.membershipResolver.Stop()
	if err := h.dispatcher.Stop(); err != nil {
		h.logger.WithTags(tag.Error(err)).Error("failed to stop dispatcher")
	}
	h.runtimeMetricsReporter.Stop()
	h.persistenceBean.Close()
}

// GetServiceName return service name
func (h *Impl) GetServiceName() string {
	return h.serviceName
}

// GetHostInfo return host info
func (h *Impl) GetHostInfo() membership.HostInfo {
	return h.hostInfo
}

// GetClusterMetadata return cluster metadata
func (h *Impl) GetClusterMetadata() cluster.Metadata {
	return h.clusterMetadata
}

// other common resources

// GetDomainCache return domain cache
func (h *Impl) GetDomainCache() cache.DomainCache {
	return h.domainCache
}

// GetDomainMetricsScopeCache return domainMetricsScope cache
func (h *Impl) GetDomainMetricsScopeCache() cache.DomainMetricsScopeCache {
	return h.domainMetricsScopeCache
}

// GetTimeSource return time source
func (h *Impl) GetTimeSource() clock.TimeSource {
	return h.timeSource
}

// GetPayloadSerializer return binary payload serializer
func (h *Impl) GetPayloadSerializer() persistence.PayloadSerializer {
	return h.payloadSerializer
}

// GetMetricsClient return metrics client
func (h *Impl) GetMetricsClient() metrics.Client {
	return h.metricsClient
}

// GetMessagingClient return messaging client
func (h *Impl) GetMessagingClient() messaging.Client {
	return h.messagingClient
}

// GetBlobstoreClient returns blobstore client
func (h *Impl) GetBlobstoreClient() blobstore.Client {
	return h.blobstoreClient
}

// GetArchivalMetadata return archival metadata
func (h *Impl) GetArchivalMetadata() archiver.ArchivalMetadata {
	return h.archivalMetadata
}

// GetArchiverProvider return archival provider
func (h *Impl) GetArchiverProvider() provider.ArchiverProvider {
	return h.archiverProvider
}

// GetDomainReplicationQueue return domain replication queue
func (h *Impl) GetDomainReplicationQueue() domain.ReplicationQueue {
	return h.domainReplicationQueue
}

// GetMembershipResolver return the membership resolver
func (h *Impl) GetMembershipResolver() membership.Resolver {
	return h.membershipResolver
}

// internal services clients

// GetSDKClient return sdk client
func (h *Impl) GetSDKClient() workflowserviceclient.Interface {
	return h.sdkClient
}

// GetFrontendRawClient return frontend client without retry policy
func (h *Impl) GetFrontendRawClient() frontend.Client {
	return h.frontendRawClient
}

// GetFrontendClient return frontend client with retry policy
func (h *Impl) GetFrontendClient() frontend.Client {
	return h.frontendClient
}

// GetMatchingRawClient return matching client without retry policy
func (h *Impl) GetMatchingRawClient() matching.Client {
	return h.matchingRawClient
}

// GetMatchingClient return matching client with retry policy
func (h *Impl) GetMatchingClient() matching.Client {
	return h.matchingClient
}

// GetHistoryRawClient return history client without retry policy
func (h *Impl) GetHistoryRawClient() history.Client {
	return h.historyRawClient
}

// GetHistoryClient return history client with retry policy
func (h *Impl) GetHistoryClient() history.Client {
	return h.historyClient
}

// GetRemoteAdminClient return remote admin client for given cluster name
func (h *Impl) GetRemoteAdminClient(
	cluster string,
) admin.Client {

	return h.clientBean.GetRemoteAdminClient(cluster)
}

// GetRemoteFrontendClient return remote frontend client for given cluster name
func (h *Impl) GetRemoteFrontendClient(
	cluster string,
) frontend.Client {

	return h.clientBean.GetRemoteFrontendClient(cluster)
}

// GetClientBean return RPC client bean
func (h *Impl) GetClientBean() client.Bean {
	return h.clientBean
}

// persistence clients

// GetMetadataManager return metadata manager
func (h *Impl) GetDomainManager() persistence.DomainManager {
	return h.persistenceBean.GetDomainManager()
}

// GetTaskManager return task manager
func (h *Impl) GetTaskManager() persistence.TaskManager {
	return h.persistenceBean.GetTaskManager()
}

// GetVisibilityManager return visibility manager
func (h *Impl) GetVisibilityManager() persistence.VisibilityManager {
	return h.persistenceBean.GetVisibilityManager()
}

// GetShardManager return shard manager
func (h *Impl) GetShardManager() persistence.ShardManager {
	return h.persistenceBean.GetShardManager()
}

// GetHistoryManager return history manager
func (h *Impl) GetHistoryManager() persistence.HistoryManager {
	return h.persistenceBean.GetHistoryManager()
}

// GetExecutionManager return execution manager for given shard ID
func (h *Impl) GetExecutionManager(
	shardID int,
) (persistence.ExecutionManager, error) {

	return h.persistenceBean.GetExecutionManager(shardID)
}

// GetPersistenceBean return persistence bean
func (h *Impl) GetPersistenceBean() persistenceClient.Bean {
	return h.persistenceBean
}

// loggers

// GetLogger return logger
func (h *Impl) GetLogger() log.Logger {
	return h.logger
}

// GetThrottledLogger return throttled logger
func (h *Impl) GetThrottledLogger() log.Logger {
	return h.throttledLogger
}

// GetDispatcher return YARPC dispatcher, used for registering handlers
func (h *Impl) GetDispatcher() *yarpc.Dispatcher {
	return h.dispatcher
}
