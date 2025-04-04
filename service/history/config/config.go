// Copyright (c) 2020 Uber Technologies, Inc.
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

package config

import (
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
)

// Config represents configuration for cadence-history service
type Config struct {
	NumberOfShards                  int
	IsAdvancedVisConfigExist        bool
	RPS                             dynamicconfig.IntPropertyFn
	MaxIDLengthWarnLimit            dynamicconfig.IntPropertyFn
	DomainNameMaxLength             dynamicconfig.IntPropertyFnWithDomainFilter
	IdentityMaxLength               dynamicconfig.IntPropertyFnWithDomainFilter
	WorkflowIDMaxLength             dynamicconfig.IntPropertyFnWithDomainFilter
	SignalNameMaxLength             dynamicconfig.IntPropertyFnWithDomainFilter
	WorkflowTypeMaxLength           dynamicconfig.IntPropertyFnWithDomainFilter
	RequestIDMaxLength              dynamicconfig.IntPropertyFnWithDomainFilter
	TaskListNameMaxLength           dynamicconfig.IntPropertyFnWithDomainFilter
	ActivityIDMaxLength             dynamicconfig.IntPropertyFnWithDomainFilter
	ActivityTypeMaxLength           dynamicconfig.IntPropertyFnWithDomainFilter
	MarkerNameMaxLength             dynamicconfig.IntPropertyFnWithDomainFilter
	TimerIDMaxLength                dynamicconfig.IntPropertyFnWithDomainFilter
	PersistenceMaxQPS               dynamicconfig.IntPropertyFn
	PersistenceGlobalMaxQPS         dynamicconfig.IntPropertyFn
	EnableVisibilitySampling        dynamicconfig.BoolPropertyFn
	EnableReadFromClosedExecutionV2 dynamicconfig.BoolPropertyFn
	VisibilityOpenMaxQPS            dynamicconfig.IntPropertyFnWithDomainFilter
	VisibilityClosedMaxQPS          dynamicconfig.IntPropertyFnWithDomainFilter
	AdvancedVisibilityWritingMode   dynamicconfig.StringPropertyFn
	EmitShardDiffLog                dynamicconfig.BoolPropertyFn
	MaxAutoResetPoints              dynamicconfig.IntPropertyFnWithDomainFilter
	ThrottledLogRPS                 dynamicconfig.IntPropertyFn
	EnableStickyQuery               dynamicconfig.BoolPropertyFnWithDomainFilter
	ShutdownDrainDuration           dynamicconfig.DurationPropertyFn
	WorkflowDeletionJitterRange     dynamicconfig.IntPropertyFnWithDomainFilter
	MaxResponseSize                 dynamicconfig.IntPropertyFn

	// HistoryCache settings
	// Change of these configs require shard restart
	HistoryCacheInitialSize dynamicconfig.IntPropertyFn
	HistoryCacheMaxSize     dynamicconfig.IntPropertyFn
	HistoryCacheTTL         dynamicconfig.DurationPropertyFn

	// EventsCache settings
	// Change of these configs require shard restart
	EventsCacheInitialCount       dynamicconfig.IntPropertyFn
	EventsCacheMaxCount           dynamicconfig.IntPropertyFn
	EventsCacheMaxSize            dynamicconfig.IntPropertyFn
	EventsCacheTTL                dynamicconfig.DurationPropertyFn
	EventsCacheGlobalEnable       dynamicconfig.BoolPropertyFn
	EventsCacheGlobalInitialCount dynamicconfig.IntPropertyFn
	EventsCacheGlobalMaxCount     dynamicconfig.IntPropertyFn

	// ShardController settings
	RangeSizeBits           uint
	AcquireShardInterval    dynamicconfig.DurationPropertyFn
	AcquireShardConcurrency dynamicconfig.IntPropertyFn

	// the artificial delay added to standby cluster's view of active cluster's time
	StandbyClusterDelay                  dynamicconfig.DurationPropertyFn
	StandbyTaskMissingEventsResendDelay  dynamicconfig.DurationPropertyFn
	StandbyTaskMissingEventsDiscardDelay dynamicconfig.DurationPropertyFn

	// Task process settings
	TaskProcessRPS                          dynamicconfig.IntPropertyFnWithDomainFilter
	TaskSchedulerType                       dynamicconfig.IntPropertyFn
	TaskSchedulerWorkerCount                dynamicconfig.IntPropertyFn
	TaskSchedulerShardWorkerCount           dynamicconfig.IntPropertyFn
	TaskSchedulerQueueSize                  dynamicconfig.IntPropertyFn
	TaskSchedulerShardQueueSize             dynamicconfig.IntPropertyFn
	TaskSchedulerDispatcherCount            dynamicconfig.IntPropertyFn
	TaskSchedulerRoundRobinWeights          dynamicconfig.MapPropertyFn
	TaskCriticalRetryCount                  dynamicconfig.IntPropertyFn
	ActiveTaskRedispatchInterval            dynamicconfig.DurationPropertyFn
	StandbyTaskRedispatchInterval           dynamicconfig.DurationPropertyFn
	TaskRedispatchIntervalJitterCoefficient dynamicconfig.FloatPropertyFn
	StandbyTaskReReplicationContextTimeout  dynamicconfig.DurationPropertyFnWithDomainIDFilter
	EnableDropStuckTaskByDomainID           dynamicconfig.BoolPropertyFnWithDomainIDFilter
	ResurrectionCheckMinDelay               dynamicconfig.DurationPropertyFnWithDomainFilter

	// QueueProcessor settings
	QueueProcessorEnableSplit                          dynamicconfig.BoolPropertyFn
	QueueProcessorSplitMaxLevel                        dynamicconfig.IntPropertyFn
	QueueProcessorEnableRandomSplitByDomainID          dynamicconfig.BoolPropertyFnWithDomainIDFilter
	QueueProcessorRandomSplitProbability               dynamicconfig.FloatPropertyFn
	QueueProcessorEnablePendingTaskSplitByDomainID     dynamicconfig.BoolPropertyFnWithDomainIDFilter
	QueueProcessorPendingTaskSplitThreshold            dynamicconfig.MapPropertyFn
	QueueProcessorEnableStuckTaskSplitByDomainID       dynamicconfig.BoolPropertyFnWithDomainIDFilter
	QueueProcessorStuckTaskSplitThreshold              dynamicconfig.MapPropertyFn
	QueueProcessorSplitLookAheadDurationByDomainID     dynamicconfig.DurationPropertyFnWithDomainIDFilter
	QueueProcessorPollBackoffInterval                  dynamicconfig.DurationPropertyFn
	QueueProcessorPollBackoffIntervalJitterCoefficient dynamicconfig.FloatPropertyFn
	QueueProcessorEnablePersistQueueStates             dynamicconfig.BoolPropertyFn
	QueueProcessorEnableLoadQueueStates                dynamicconfig.BoolPropertyFn

	// TimerQueueProcessor settings
	TimerTaskBatchSize                                dynamicconfig.IntPropertyFn
	TimerTaskDeleteBatchSize                          dynamicconfig.IntPropertyFn
	TimerProcessorGetFailureRetryCount                dynamicconfig.IntPropertyFn
	TimerProcessorCompleteTimerFailureRetryCount      dynamicconfig.IntPropertyFn
	TimerProcessorUpdateAckInterval                   dynamicconfig.DurationPropertyFn
	TimerProcessorUpdateAckIntervalJitterCoefficient  dynamicconfig.FloatPropertyFn
	TimerProcessorCompleteTimerInterval               dynamicconfig.DurationPropertyFn
	TimerProcessorFailoverMaxStartJitterInterval      dynamicconfig.DurationPropertyFn
	TimerProcessorFailoverMaxPollRPS                  dynamicconfig.IntPropertyFn
	TimerProcessorMaxPollRPS                          dynamicconfig.IntPropertyFn
	TimerProcessorMaxPollInterval                     dynamicconfig.DurationPropertyFn
	TimerProcessorMaxPollIntervalJitterCoefficient    dynamicconfig.FloatPropertyFn
	TimerProcessorSplitQueueInterval                  dynamicconfig.DurationPropertyFn
	TimerProcessorSplitQueueIntervalJitterCoefficient dynamicconfig.FloatPropertyFn
	TimerProcessorMaxRedispatchQueueSize              dynamicconfig.IntPropertyFn
	TimerProcessorMaxTimeShift                        dynamicconfig.DurationPropertyFn
	TimerProcessorHistoryArchivalSizeLimit            dynamicconfig.IntPropertyFn
	TimerProcessorArchivalTimeLimit                   dynamicconfig.DurationPropertyFn

	// TransferQueueProcessor settings
	TransferTaskBatchSize                                dynamicconfig.IntPropertyFn
	TransferTaskDeleteBatchSize                          dynamicconfig.IntPropertyFn
	TransferProcessorCompleteTransferFailureRetryCount   dynamicconfig.IntPropertyFn
	TransferProcessorFailoverMaxStartJitterInterval      dynamicconfig.DurationPropertyFn
	TransferProcessorFailoverMaxPollRPS                  dynamicconfig.IntPropertyFn
	TransferProcessorMaxPollRPS                          dynamicconfig.IntPropertyFn
	TransferProcessorMaxPollInterval                     dynamicconfig.DurationPropertyFn
	TransferProcessorMaxPollIntervalJitterCoefficient    dynamicconfig.FloatPropertyFn
	TransferProcessorSplitQueueInterval                  dynamicconfig.DurationPropertyFn
	TransferProcessorSplitQueueIntervalJitterCoefficient dynamicconfig.FloatPropertyFn
	TransferProcessorUpdateAckInterval                   dynamicconfig.DurationPropertyFn
	TransferProcessorUpdateAckIntervalJitterCoefficient  dynamicconfig.FloatPropertyFn
	TransferProcessorCompleteTransferInterval            dynamicconfig.DurationPropertyFn
	TransferProcessorMaxRedispatchQueueSize              dynamicconfig.IntPropertyFn
	TransferProcessorEnableValidator                     dynamicconfig.BoolPropertyFn
	TransferProcessorValidationInterval                  dynamicconfig.DurationPropertyFn
	TransferProcessorVisibilityArchivalTimeLimit         dynamicconfig.DurationPropertyFn

	// CrossClusterQueueProcessor settings
	CrossClusterTaskBatchSize                                     dynamicconfig.IntPropertyFn
	CrossClusterTaskDeleteBatchSize                               dynamicconfig.IntPropertyFn
	CrossClusterTaskFetchBatchSize                                dynamicconfig.IntPropertyFnWithShardIDFilter
	CrossClusterSourceProcessorMaxPollRPS                         dynamicconfig.IntPropertyFn
	CrossClusterSourceProcessorMaxPollInterval                    dynamicconfig.DurationPropertyFn
	CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient   dynamicconfig.FloatPropertyFn
	CrossClusterSourceProcessorUpdateAckInterval                  dynamicconfig.DurationPropertyFn
	CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient dynamicconfig.FloatPropertyFn
	CrossClusterSourceProcessorMaxRedispatchQueueSize             dynamicconfig.IntPropertyFn
	CrossClusterSourceProcessorMaxPendingTaskSize                 dynamicconfig.IntPropertyFn

	// CrossClusterTargetTaskProcessor settings
	CrossClusterTargetProcessorMaxPendingTasks            dynamicconfig.IntPropertyFn
	CrossClusterTargetProcessorMaxRetryCount              dynamicconfig.IntPropertyFn
	CrossClusterTargetProcessorTaskWaitInterval           dynamicconfig.DurationPropertyFn
	CrossClusterTargetProcessorServiceBusyBackoffInterval dynamicconfig.DurationPropertyFn
	CrossClusterTargetProcessorJitterCoefficient          dynamicconfig.FloatPropertyFn

	// CrossClusterTaskFetcher settings
	CrossClusterFetcherParallelism                dynamicconfig.IntPropertyFn
	CrossClusterFetcherAggregationInterval        dynamicconfig.DurationPropertyFn
	CrossClusterFetcherServiceBusyBackoffInterval dynamicconfig.DurationPropertyFn
	CrossClusterFetcherErrorBackoffInterval       dynamicconfig.DurationPropertyFn
	CrossClusterFetcherJitterCoefficient          dynamicconfig.FloatPropertyFn

	// ReplicatorQueueProcessor settings
	ReplicatorTaskDeleteBatchSize          dynamicconfig.IntPropertyFn
	ReplicatorReadTaskMaxRetryCount        dynamicconfig.IntPropertyFn
	ReplicatorProcessorFetchTasksBatchSize dynamicconfig.IntPropertyFnWithShardIDFilter
	ReplicatorUpperLatency                 dynamicconfig.DurationPropertyFn
	ReplicatorCacheCapacity                dynamicconfig.IntPropertyFn

	// Persistence settings
	ExecutionMgrNumConns dynamicconfig.IntPropertyFn
	HistoryMgrNumConns   dynamicconfig.IntPropertyFn

	// System Limits
	MaximumBufferedEventsBatch dynamicconfig.IntPropertyFn
	MaximumSignalsPerExecution dynamicconfig.IntPropertyFnWithDomainFilter

	// ShardUpdateMinInterval the minimal time interval which the shard info can be updated
	ShardUpdateMinInterval dynamicconfig.DurationPropertyFn
	// ShardSyncMinInterval the minimal time interval which the shard info should be sync to remote
	ShardSyncMinInterval            dynamicconfig.DurationPropertyFn
	ShardSyncTimerJitterCoefficient dynamicconfig.FloatPropertyFn

	// Time to hold a poll request before returning an empty response
	// right now only used by GetMutableState
	LongPollExpirationInterval dynamicconfig.DurationPropertyFnWithDomainFilter

	// encoding the history events
	EventEncodingType dynamicconfig.StringPropertyFnWithDomainFilter
	// whether or not using ParentClosePolicy
	EnableParentClosePolicy dynamicconfig.BoolPropertyFnWithDomainFilter
	// whether or not enable system workers for processing parent close policy task
	EnableParentClosePolicyWorker dynamicconfig.BoolPropertyFn
	// whether to enable system workers for processing ElasticSearch Analyzer
	EnableESAnalyzer dynamicconfig.BoolPropertyFn
	// parent close policy will be processed by sys workers(if enabled) if
	// the number of children greater than or equal to this threshold
	ParentClosePolicyThreshold dynamicconfig.IntPropertyFnWithDomainFilter
	// total number of parentClosePolicy system workflows
	NumParentClosePolicySystemWorkflows dynamicconfig.IntPropertyFn

	// Archival settings
	NumArchiveSystemWorkflows        dynamicconfig.IntPropertyFn
	ArchiveRequestRPS                dynamicconfig.IntPropertyFn
	ArchiveInlineHistoryRPS          dynamicconfig.IntPropertyFn
	ArchiveInlineHistoryGlobalRPS    dynamicconfig.IntPropertyFn
	ArchiveInlineVisibilityRPS       dynamicconfig.IntPropertyFn
	ArchiveInlineVisibilityGlobalRPS dynamicconfig.IntPropertyFn
	AllowArchivingIncompleteHistory  dynamicconfig.BoolPropertyFn

	// Size limit related settings
	BlobSizeLimitError               dynamicconfig.IntPropertyFnWithDomainFilter
	BlobSizeLimitWarn                dynamicconfig.IntPropertyFnWithDomainFilter
	HistorySizeLimitError            dynamicconfig.IntPropertyFnWithDomainFilter
	HistorySizeLimitWarn             dynamicconfig.IntPropertyFnWithDomainFilter
	HistoryCountLimitError           dynamicconfig.IntPropertyFnWithDomainFilter
	HistoryCountLimitWarn            dynamicconfig.IntPropertyFnWithDomainFilter
	PendingActivitiesCountLimitError dynamicconfig.IntPropertyFn
	PendingActivitiesCountLimitWarn  dynamicconfig.IntPropertyFn
	PendingActivityValidationEnabled dynamicconfig.BoolPropertyFn

	// ValidSearchAttributes is legal indexed keys that can be used in list APIs
	ValidSearchAttributes             dynamicconfig.MapPropertyFn
	SearchAttributesNumberOfKeysLimit dynamicconfig.IntPropertyFnWithDomainFilter
	SearchAttributesSizeOfValueLimit  dynamicconfig.IntPropertyFnWithDomainFilter
	SearchAttributesTotalSizeLimit    dynamicconfig.IntPropertyFnWithDomainFilter

	// Decision settings
	// StickyTTL is to expire a sticky tasklist if no update more than this duration
	// TODO https://github.com/uber/cadence/issues/2357
	StickyTTL dynamicconfig.DurationPropertyFnWithDomainFilter
	// DecisionHeartbeatTimeout is to timeout behavior of: RespondDecisionTaskComplete with ForceCreateNewDecisionTask == true without any decisions
	// So that decision will be scheduled to another worker(by clear stickyness)
	DecisionHeartbeatTimeout dynamicconfig.DurationPropertyFnWithDomainFilter
	// MaxDecisionStartToCloseSeconds is the StartToCloseSeconds for decision
	MaxDecisionStartToCloseSeconds           dynamicconfig.IntPropertyFnWithDomainFilter
	DecisionRetryCriticalAttempts            dynamicconfig.IntPropertyFn
	DecisionRetryMaxAttempts                 dynamicconfig.IntPropertyFnWithDomainFilter
	NormalDecisionScheduleToStartMaxAttempts dynamicconfig.IntPropertyFnWithDomainFilter
	NormalDecisionScheduleToStartTimeout     dynamicconfig.DurationPropertyFnWithDomainFilter

	// The following is used by the new RPC replication stack
	ReplicationTaskFetcherParallelism                  dynamicconfig.IntPropertyFn
	ReplicationTaskFetcherAggregationInterval          dynamicconfig.DurationPropertyFn
	ReplicationTaskFetcherTimerJitterCoefficient       dynamicconfig.FloatPropertyFn
	ReplicationTaskFetcherErrorRetryWait               dynamicconfig.DurationPropertyFn
	ReplicationTaskFetcherServiceBusyWait              dynamicconfig.DurationPropertyFn
	ReplicationTaskProcessorErrorRetryWait             dynamicconfig.DurationPropertyFnWithShardIDFilter
	ReplicationTaskProcessorErrorRetryMaxAttempts      dynamicconfig.IntPropertyFnWithShardIDFilter
	ReplicationTaskProcessorErrorSecondRetryWait       dynamicconfig.DurationPropertyFnWithShardIDFilter
	ReplicationTaskProcessorErrorSecondRetryMaxWait    dynamicconfig.DurationPropertyFnWithShardIDFilter
	ReplicationTaskProcessorErrorSecondRetryExpiration dynamicconfig.DurationPropertyFnWithShardIDFilter
	ReplicationTaskProcessorNoTaskRetryWait            dynamicconfig.DurationPropertyFnWithShardIDFilter
	ReplicationTaskProcessorCleanupInterval            dynamicconfig.DurationPropertyFnWithShardIDFilter
	ReplicationTaskProcessorCleanupJitterCoefficient   dynamicconfig.FloatPropertyFnWithShardIDFilter
	ReplicationTaskProcessorStartWait                  dynamicconfig.DurationPropertyFnWithShardIDFilter
	ReplicationTaskProcessorStartWaitJitterCoefficient dynamicconfig.FloatPropertyFnWithShardIDFilter
	ReplicationTaskProcessorHostQPS                    dynamicconfig.FloatPropertyFn
	ReplicationTaskProcessorShardQPS                   dynamicconfig.FloatPropertyFn
	ReplicationTaskGenerationQPS                       dynamicconfig.FloatPropertyFn
	EnableReplicationTaskGeneration                    dynamicconfig.BoolPropertyFnWithDomainIDAndWorkflowIDFilter
	EnableRecordWorkflowExecutionUninitialized         dynamicconfig.BoolPropertyFnWithDomainFilter

	// The following are used by consistent query
	EnableConsistentQuery         dynamicconfig.BoolPropertyFn
	EnableConsistentQueryByDomain dynamicconfig.BoolPropertyFnWithDomainFilter
	MaxBufferedQueryCount         dynamicconfig.IntPropertyFn

	EnableCrossClusterOperations dynamicconfig.BoolPropertyFnWithDomainFilter

	// Data integrity check related config knobs
	MutableStateChecksumGenProbability    dynamicconfig.IntPropertyFnWithDomainFilter
	MutableStateChecksumVerifyProbability dynamicconfig.IntPropertyFnWithDomainFilter
	MutableStateChecksumInvalidateBefore  dynamicconfig.FloatPropertyFn

	// History check for corruptions
	EnableHistoryCorruptionCheck dynamicconfig.BoolPropertyFnWithDomainFilter

	// Failover marker heartbeat
	NotifyFailoverMarkerInterval               dynamicconfig.DurationPropertyFn
	NotifyFailoverMarkerTimerJitterCoefficient dynamicconfig.FloatPropertyFn
	EnableGracefulFailover                     dynamicconfig.BoolPropertyFn

	// Allows worker to dispatch activity tasks through local tunnel after decisions are made. This is an performance optimization to skip activity scheduling efforts.
	EnableActivityLocalDispatchByDomain dynamicconfig.BoolPropertyFnWithDomainFilter
	// Max # of activity tasks to dispatch to matching before creating transfer tasks. This is an performance optimization to skip activity scheduling efforts.
	MaxActivityCountDispatchByDomain dynamicconfig.IntPropertyFnWithDomainFilter

	ActivityMaxScheduleToStartTimeoutForRetry dynamicconfig.DurationPropertyFnWithDomainFilter

	// Debugging configurations
	EnableDebugMode             bool // note that this value is initialized once on service start
	EnableTaskInfoLogByDomainID dynamicconfig.BoolPropertyFnWithDomainIDFilter
}

// New returns new service config with default values
func New(dc *dynamicconfig.Collection, numberOfShards int, storeType string, isAdvancedVisConfigExist bool) *Config {
	cfg := &Config{
		NumberOfShards:                       numberOfShards,
		IsAdvancedVisConfigExist:             isAdvancedVisConfigExist,
		RPS:                                  dc.GetIntProperty(dynamicconfig.HistoryRPS),
		MaxIDLengthWarnLimit:                 dc.GetIntProperty(dynamicconfig.MaxIDLengthWarnLimit),
		DomainNameMaxLength:                  dc.GetIntPropertyFilteredByDomain(dynamicconfig.DomainNameMaxLength),
		IdentityMaxLength:                    dc.GetIntPropertyFilteredByDomain(dynamicconfig.IdentityMaxLength),
		WorkflowIDMaxLength:                  dc.GetIntPropertyFilteredByDomain(dynamicconfig.WorkflowIDMaxLength),
		SignalNameMaxLength:                  dc.GetIntPropertyFilteredByDomain(dynamicconfig.SignalNameMaxLength),
		WorkflowTypeMaxLength:                dc.GetIntPropertyFilteredByDomain(dynamicconfig.WorkflowTypeMaxLength),
		RequestIDMaxLength:                   dc.GetIntPropertyFilteredByDomain(dynamicconfig.RequestIDMaxLength),
		TaskListNameMaxLength:                dc.GetIntPropertyFilteredByDomain(dynamicconfig.TaskListNameMaxLength),
		ActivityIDMaxLength:                  dc.GetIntPropertyFilteredByDomain(dynamicconfig.ActivityIDMaxLength),
		ActivityTypeMaxLength:                dc.GetIntPropertyFilteredByDomain(dynamicconfig.ActivityTypeMaxLength),
		MarkerNameMaxLength:                  dc.GetIntPropertyFilteredByDomain(dynamicconfig.MarkerNameMaxLength),
		TimerIDMaxLength:                     dc.GetIntPropertyFilteredByDomain(dynamicconfig.TimerIDMaxLength),
		PersistenceMaxQPS:                    dc.GetIntProperty(dynamicconfig.HistoryPersistenceMaxQPS),
		PersistenceGlobalMaxQPS:              dc.GetIntProperty(dynamicconfig.HistoryPersistenceGlobalMaxQPS),
		ShutdownDrainDuration:                dc.GetDurationProperty(dynamicconfig.HistoryShutdownDrainDuration),
		EnableVisibilitySampling:             dc.GetBoolProperty(dynamicconfig.EnableVisibilitySampling),
		EnableReadFromClosedExecutionV2:      dc.GetBoolProperty(dynamicconfig.EnableReadFromClosedExecutionV2),
		VisibilityOpenMaxQPS:                 dc.GetIntPropertyFilteredByDomain(dynamicconfig.HistoryVisibilityOpenMaxQPS),
		VisibilityClosedMaxQPS:               dc.GetIntPropertyFilteredByDomain(dynamicconfig.HistoryVisibilityClosedMaxQPS),
		MaxAutoResetPoints:                   dc.GetIntPropertyFilteredByDomain(dynamicconfig.HistoryMaxAutoResetPoints),
		MaxDecisionStartToCloseSeconds:       dc.GetIntPropertyFilteredByDomain(dynamicconfig.MaxDecisionStartToCloseSeconds),
		AdvancedVisibilityWritingMode:        dc.GetStringProperty(dynamicconfig.AdvancedVisibilityWritingMode),
		EmitShardDiffLog:                     dc.GetBoolProperty(dynamicconfig.EmitShardDiffLog),
		HistoryCacheInitialSize:              dc.GetIntProperty(dynamicconfig.HistoryCacheInitialSize),
		HistoryCacheMaxSize:                  dc.GetIntProperty(dynamicconfig.HistoryCacheMaxSize),
		HistoryCacheTTL:                      dc.GetDurationProperty(dynamicconfig.HistoryCacheTTL),
		EventsCacheInitialCount:              dc.GetIntProperty(dynamicconfig.EventsCacheInitialCount),
		EventsCacheMaxCount:                  dc.GetIntProperty(dynamicconfig.EventsCacheMaxCount),
		EventsCacheMaxSize:                   dc.GetIntProperty(dynamicconfig.EventsCacheMaxSize),
		EventsCacheTTL:                       dc.GetDurationProperty(dynamicconfig.EventsCacheTTL),
		EventsCacheGlobalEnable:              dc.GetBoolProperty(dynamicconfig.EventsCacheGlobalEnable),
		EventsCacheGlobalInitialCount:        dc.GetIntProperty(dynamicconfig.EventsCacheGlobalInitialCount),
		EventsCacheGlobalMaxCount:            dc.GetIntProperty(dynamicconfig.EventsCacheGlobalMaxCount),
		RangeSizeBits:                        20, // 20 bits for sequencer, 2^20 sequence number for any range
		AcquireShardInterval:                 dc.GetDurationProperty(dynamicconfig.AcquireShardInterval),
		AcquireShardConcurrency:              dc.GetIntProperty(dynamicconfig.AcquireShardConcurrency),
		StandbyClusterDelay:                  dc.GetDurationProperty(dynamicconfig.StandbyClusterDelay),
		StandbyTaskMissingEventsResendDelay:  dc.GetDurationProperty(dynamicconfig.StandbyTaskMissingEventsResendDelay),
		StandbyTaskMissingEventsDiscardDelay: dc.GetDurationProperty(dynamicconfig.StandbyTaskMissingEventsDiscardDelay),
		WorkflowDeletionJitterRange:          dc.GetIntPropertyFilteredByDomain(dynamicconfig.WorkflowDeletionJitterRange),
		MaxResponseSize:                      dc.GetIntProperty(dynamicconfig.GRPCMaxSizeInByte),

		TaskProcessRPS:                          dc.GetIntPropertyFilteredByDomain(dynamicconfig.TaskProcessRPS),
		TaskSchedulerType:                       dc.GetIntProperty(dynamicconfig.TaskSchedulerType),
		TaskSchedulerWorkerCount:                dc.GetIntProperty(dynamicconfig.TaskSchedulerWorkerCount),
		TaskSchedulerShardWorkerCount:           dc.GetIntProperty(dynamicconfig.TaskSchedulerShardWorkerCount),
		TaskSchedulerQueueSize:                  dc.GetIntProperty(dynamicconfig.TaskSchedulerQueueSize),
		TaskSchedulerShardQueueSize:             dc.GetIntProperty(dynamicconfig.TaskSchedulerShardQueueSize),
		TaskSchedulerDispatcherCount:            dc.GetIntProperty(dynamicconfig.TaskSchedulerDispatcherCount),
		TaskSchedulerRoundRobinWeights:          dc.GetMapProperty(dynamicconfig.TaskSchedulerRoundRobinWeights),
		TaskCriticalRetryCount:                  dc.GetIntProperty(dynamicconfig.TaskCriticalRetryCount),
		ActiveTaskRedispatchInterval:            dc.GetDurationProperty(dynamicconfig.ActiveTaskRedispatchInterval),
		StandbyTaskRedispatchInterval:           dc.GetDurationProperty(dynamicconfig.StandbyTaskRedispatchInterval),
		TaskRedispatchIntervalJitterCoefficient: dc.GetFloat64Property(dynamicconfig.TaskRedispatchIntervalJitterCoefficient),
		StandbyTaskReReplicationContextTimeout:  dc.GetDurationPropertyFilteredByDomainID(dynamicconfig.StandbyTaskReReplicationContextTimeout),
		EnableDropStuckTaskByDomainID:           dc.GetBoolPropertyFilteredByDomainID(dynamicconfig.EnableDropStuckTaskByDomainID),
		ResurrectionCheckMinDelay:               dc.GetDurationPropertyFilteredByDomain(dynamicconfig.ResurrectionCheckMinDelay),

		QueueProcessorEnableSplit:                          dc.GetBoolProperty(dynamicconfig.QueueProcessorEnableSplit),
		QueueProcessorSplitMaxLevel:                        dc.GetIntProperty(dynamicconfig.QueueProcessorSplitMaxLevel),
		QueueProcessorEnableRandomSplitByDomainID:          dc.GetBoolPropertyFilteredByDomainID(dynamicconfig.QueueProcessorEnableRandomSplitByDomainID),
		QueueProcessorRandomSplitProbability:               dc.GetFloat64Property(dynamicconfig.QueueProcessorRandomSplitProbability),
		QueueProcessorEnablePendingTaskSplitByDomainID:     dc.GetBoolPropertyFilteredByDomainID(dynamicconfig.QueueProcessorEnablePendingTaskSplitByDomainID),
		QueueProcessorPendingTaskSplitThreshold:            dc.GetMapProperty(dynamicconfig.QueueProcessorPendingTaskSplitThreshold),
		QueueProcessorEnableStuckTaskSplitByDomainID:       dc.GetBoolPropertyFilteredByDomainID(dynamicconfig.QueueProcessorEnableStuckTaskSplitByDomainID),
		QueueProcessorStuckTaskSplitThreshold:              dc.GetMapProperty(dynamicconfig.QueueProcessorStuckTaskSplitThreshold),
		QueueProcessorSplitLookAheadDurationByDomainID:     dc.GetDurationPropertyFilteredByDomainID(dynamicconfig.QueueProcessorSplitLookAheadDurationByDomainID),
		QueueProcessorPollBackoffInterval:                  dc.GetDurationProperty(dynamicconfig.QueueProcessorPollBackoffInterval),
		QueueProcessorPollBackoffIntervalJitterCoefficient: dc.GetFloat64Property(dynamicconfig.QueueProcessorPollBackoffIntervalJitterCoefficient),
		QueueProcessorEnablePersistQueueStates:             dc.GetBoolProperty(dynamicconfig.QueueProcessorEnablePersistQueueStates),
		QueueProcessorEnableLoadQueueStates:                dc.GetBoolProperty(dynamicconfig.QueueProcessorEnableLoadQueueStates),

		TimerTaskBatchSize:                                dc.GetIntProperty(dynamicconfig.TimerTaskBatchSize),
		TimerTaskDeleteBatchSize:                          dc.GetIntProperty(dynamicconfig.TimerTaskDeleteBatchSize),
		TimerProcessorGetFailureRetryCount:                dc.GetIntProperty(dynamicconfig.TimerProcessorGetFailureRetryCount),
		TimerProcessorCompleteTimerFailureRetryCount:      dc.GetIntProperty(dynamicconfig.TimerProcessorCompleteTimerFailureRetryCount),
		TimerProcessorUpdateAckInterval:                   dc.GetDurationProperty(dynamicconfig.TimerProcessorUpdateAckInterval),
		TimerProcessorUpdateAckIntervalJitterCoefficient:  dc.GetFloat64Property(dynamicconfig.TimerProcessorUpdateAckIntervalJitterCoefficient),
		TimerProcessorCompleteTimerInterval:               dc.GetDurationProperty(dynamicconfig.TimerProcessorCompleteTimerInterval),
		TimerProcessorFailoverMaxStartJitterInterval:      dc.GetDurationProperty(dynamicconfig.TimerProcessorFailoverMaxStartJitterInterval),
		TimerProcessorFailoverMaxPollRPS:                  dc.GetIntProperty(dynamicconfig.TimerProcessorFailoverMaxPollRPS),
		TimerProcessorMaxPollRPS:                          dc.GetIntProperty(dynamicconfig.TimerProcessorMaxPollRPS),
		TimerProcessorMaxPollInterval:                     dc.GetDurationProperty(dynamicconfig.TimerProcessorMaxPollInterval),
		TimerProcessorMaxPollIntervalJitterCoefficient:    dc.GetFloat64Property(dynamicconfig.TimerProcessorMaxPollIntervalJitterCoefficient),
		TimerProcessorSplitQueueInterval:                  dc.GetDurationProperty(dynamicconfig.TimerProcessorSplitQueueInterval),
		TimerProcessorSplitQueueIntervalJitterCoefficient: dc.GetFloat64Property(dynamicconfig.TimerProcessorSplitQueueIntervalJitterCoefficient),
		TimerProcessorMaxRedispatchQueueSize:              dc.GetIntProperty(dynamicconfig.TimerProcessorMaxRedispatchQueueSize),
		TimerProcessorMaxTimeShift:                        dc.GetDurationProperty(dynamicconfig.TimerProcessorMaxTimeShift),
		TimerProcessorHistoryArchivalSizeLimit:            dc.GetIntProperty(dynamicconfig.TimerProcessorHistoryArchivalSizeLimit),
		TimerProcessorArchivalTimeLimit:                   dc.GetDurationProperty(dynamicconfig.TimerProcessorArchivalTimeLimit),

		TransferTaskBatchSize:                                dc.GetIntProperty(dynamicconfig.TransferTaskBatchSize),
		TransferTaskDeleteBatchSize:                          dc.GetIntProperty(dynamicconfig.TransferTaskDeleteBatchSize),
		TransferProcessorFailoverMaxStartJitterInterval:      dc.GetDurationProperty(dynamicconfig.TransferProcessorFailoverMaxStartJitterInterval),
		TransferProcessorFailoverMaxPollRPS:                  dc.GetIntProperty(dynamicconfig.TransferProcessorFailoverMaxPollRPS),
		TransferProcessorMaxPollRPS:                          dc.GetIntProperty(dynamicconfig.TransferProcessorMaxPollRPS),
		TransferProcessorCompleteTransferFailureRetryCount:   dc.GetIntProperty(dynamicconfig.TransferProcessorCompleteTransferFailureRetryCount),
		TransferProcessorMaxPollInterval:                     dc.GetDurationProperty(dynamicconfig.TransferProcessorMaxPollInterval),
		TransferProcessorMaxPollIntervalJitterCoefficient:    dc.GetFloat64Property(dynamicconfig.TransferProcessorMaxPollIntervalJitterCoefficient),
		TransferProcessorSplitQueueInterval:                  dc.GetDurationProperty(dynamicconfig.TransferProcessorSplitQueueInterval),
		TransferProcessorSplitQueueIntervalJitterCoefficient: dc.GetFloat64Property(dynamicconfig.TransferProcessorSplitQueueIntervalJitterCoefficient),
		TransferProcessorUpdateAckInterval:                   dc.GetDurationProperty(dynamicconfig.TransferProcessorUpdateAckInterval),
		TransferProcessorUpdateAckIntervalJitterCoefficient:  dc.GetFloat64Property(dynamicconfig.TransferProcessorUpdateAckIntervalJitterCoefficient),
		TransferProcessorCompleteTransferInterval:            dc.GetDurationProperty(dynamicconfig.TransferProcessorCompleteTransferInterval),
		TransferProcessorMaxRedispatchQueueSize:              dc.GetIntProperty(dynamicconfig.TransferProcessorMaxRedispatchQueueSize),
		TransferProcessorEnableValidator:                     dc.GetBoolProperty(dynamicconfig.TransferProcessorEnableValidator),
		TransferProcessorValidationInterval:                  dc.GetDurationProperty(dynamicconfig.TransferProcessorValidationInterval),
		TransferProcessorVisibilityArchivalTimeLimit:         dc.GetDurationProperty(dynamicconfig.TransferProcessorVisibilityArchivalTimeLimit),

		CrossClusterTaskBatchSize:                                     dc.GetIntProperty(dynamicconfig.CrossClusterTaskBatchSize),
		CrossClusterTaskDeleteBatchSize:                               dc.GetIntProperty(dynamicconfig.CrossClusterTaskDeleteBatchSize),
		CrossClusterTaskFetchBatchSize:                                dc.GetIntPropertyFilteredByShardID(dynamicconfig.CrossClusterTaskFetchBatchSize),
		CrossClusterSourceProcessorMaxPollRPS:                         dc.GetIntProperty(dynamicconfig.CrossClusterSourceProcessorMaxPollRPS),
		CrossClusterSourceProcessorMaxPollInterval:                    dc.GetDurationProperty(dynamicconfig.CrossClusterSourceProcessorMaxPollInterval),
		CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient:   dc.GetFloat64Property(dynamicconfig.CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient),
		CrossClusterSourceProcessorUpdateAckInterval:                  dc.GetDurationProperty(dynamicconfig.CrossClusterSourceProcessorUpdateAckInterval),
		CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient: dc.GetFloat64Property(dynamicconfig.CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient),
		CrossClusterSourceProcessorMaxRedispatchQueueSize:             dc.GetIntProperty(dynamicconfig.CrossClusterSourceProcessorMaxRedispatchQueueSize),
		CrossClusterSourceProcessorMaxPendingTaskSize:                 dc.GetIntProperty(dynamicconfig.CrossClusterSourceProcessorMaxPendingTaskSize),

		CrossClusterTargetProcessorMaxPendingTasks:            dc.GetIntProperty(dynamicconfig.CrossClusterTargetProcessorMaxPendingTasks),
		CrossClusterTargetProcessorMaxRetryCount:              dc.GetIntProperty(dynamicconfig.CrossClusterTargetProcessorMaxRetryCount),
		CrossClusterTargetProcessorTaskWaitInterval:           dc.GetDurationProperty(dynamicconfig.CrossClusterTargetProcessorTaskWaitInterval),
		CrossClusterTargetProcessorServiceBusyBackoffInterval: dc.GetDurationProperty(dynamicconfig.CrossClusterTargetProcessorServiceBusyBackoffInterval),
		CrossClusterTargetProcessorJitterCoefficient:          dc.GetFloat64Property(dynamicconfig.CrossClusterTargetProcessorJitterCoefficient),

		CrossClusterFetcherParallelism:                dc.GetIntProperty(dynamicconfig.CrossClusterFetcherParallelism),
		CrossClusterFetcherAggregationInterval:        dc.GetDurationProperty(dynamicconfig.CrossClusterFetcherAggregationInterval),
		CrossClusterFetcherServiceBusyBackoffInterval: dc.GetDurationProperty(dynamicconfig.CrossClusterFetcherServiceBusyBackoffInterval),
		CrossClusterFetcherErrorBackoffInterval:       dc.GetDurationProperty(dynamicconfig.CrossClusterFetcherErrorBackoffInterval),
		CrossClusterFetcherJitterCoefficient:          dc.GetFloat64Property(dynamicconfig.CrossClusterFetcherJitterCoefficient),

		ReplicatorTaskDeleteBatchSize:          dc.GetIntProperty(dynamicconfig.ReplicatorTaskDeleteBatchSize),
		ReplicatorReadTaskMaxRetryCount:        dc.GetIntProperty(dynamicconfig.ReplicatorReadTaskMaxRetryCount),
		ReplicatorProcessorFetchTasksBatchSize: dc.GetIntPropertyFilteredByShardID(dynamicconfig.ReplicatorTaskBatchSize),
		ReplicatorUpperLatency:                 dc.GetDurationProperty(dynamicconfig.ReplicatorUpperLatency),
		ReplicatorCacheCapacity:                dc.GetIntProperty(dynamicconfig.ReplicatorCacheCapacity),

		ExecutionMgrNumConns:            dc.GetIntProperty(dynamicconfig.ExecutionMgrNumConns),
		HistoryMgrNumConns:              dc.GetIntProperty(dynamicconfig.HistoryMgrNumConns),
		MaximumBufferedEventsBatch:      dc.GetIntProperty(dynamicconfig.MaximumBufferedEventsBatch),
		MaximumSignalsPerExecution:      dc.GetIntPropertyFilteredByDomain(dynamicconfig.MaximumSignalsPerExecution),
		ShardUpdateMinInterval:          dc.GetDurationProperty(dynamicconfig.ShardUpdateMinInterval),
		ShardSyncMinInterval:            dc.GetDurationProperty(dynamicconfig.ShardSyncMinInterval),
		ShardSyncTimerJitterCoefficient: dc.GetFloat64Property(dynamicconfig.TransferProcessorMaxPollIntervalJitterCoefficient),

		// history client: client/history/client.go set the client timeout 30s
		LongPollExpirationInterval:          dc.GetDurationPropertyFilteredByDomain(dynamicconfig.HistoryLongPollExpirationInterval),
		EventEncodingType:                   dc.GetStringPropertyFilteredByDomain(dynamicconfig.DefaultEventEncoding),
		EnableParentClosePolicy:             dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableParentClosePolicy),
		NumParentClosePolicySystemWorkflows: dc.GetIntProperty(dynamicconfig.NumParentClosePolicySystemWorkflows),
		EnableParentClosePolicyWorker:       dc.GetBoolProperty(dynamicconfig.EnableParentClosePolicyWorker),
		ParentClosePolicyThreshold:          dc.GetIntPropertyFilteredByDomain(dynamicconfig.ParentClosePolicyThreshold),

		NumArchiveSystemWorkflows:        dc.GetIntProperty(dynamicconfig.NumArchiveSystemWorkflows),
		ArchiveRequestRPS:                dc.GetIntProperty(dynamicconfig.ArchiveRequestRPS),
		ArchiveInlineHistoryRPS:          dc.GetIntProperty(dynamicconfig.ArchiveInlineHistoryRPS),
		ArchiveInlineHistoryGlobalRPS:    dc.GetIntProperty(dynamicconfig.ArchiveInlineHistoryGlobalRPS),
		ArchiveInlineVisibilityRPS:       dc.GetIntProperty(dynamicconfig.ArchiveInlineVisibilityRPS),
		ArchiveInlineVisibilityGlobalRPS: dc.GetIntProperty(dynamicconfig.ArchiveInlineVisibilityGlobalRPS),
		AllowArchivingIncompleteHistory:  dc.GetBoolProperty(dynamicconfig.AllowArchivingIncompleteHistory),

		BlobSizeLimitError:               dc.GetIntPropertyFilteredByDomain(dynamicconfig.BlobSizeLimitError),
		BlobSizeLimitWarn:                dc.GetIntPropertyFilteredByDomain(dynamicconfig.BlobSizeLimitWarn),
		HistorySizeLimitError:            dc.GetIntPropertyFilteredByDomain(dynamicconfig.HistorySizeLimitError),
		HistorySizeLimitWarn:             dc.GetIntPropertyFilteredByDomain(dynamicconfig.HistorySizeLimitWarn),
		HistoryCountLimitError:           dc.GetIntPropertyFilteredByDomain(dynamicconfig.HistoryCountLimitError),
		HistoryCountLimitWarn:            dc.GetIntPropertyFilteredByDomain(dynamicconfig.HistoryCountLimitWarn),
		PendingActivitiesCountLimitError: dc.GetIntProperty(dynamicconfig.PendingActivitiesCountLimitError),
		PendingActivitiesCountLimitWarn:  dc.GetIntProperty(dynamicconfig.PendingActivitiesCountLimitWarn),
		PendingActivityValidationEnabled: dc.GetBoolProperty(dynamicconfig.EnablePendingActivityValidation),

		ThrottledLogRPS:   dc.GetIntProperty(dynamicconfig.HistoryThrottledLogRPS),
		EnableStickyQuery: dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableStickyQuery),

		ValidSearchAttributes:                    dc.GetMapProperty(dynamicconfig.ValidSearchAttributes),
		SearchAttributesNumberOfKeysLimit:        dc.GetIntPropertyFilteredByDomain(dynamicconfig.SearchAttributesNumberOfKeysLimit),
		SearchAttributesSizeOfValueLimit:         dc.GetIntPropertyFilteredByDomain(dynamicconfig.SearchAttributesSizeOfValueLimit),
		SearchAttributesTotalSizeLimit:           dc.GetIntPropertyFilteredByDomain(dynamicconfig.SearchAttributesTotalSizeLimit),
		StickyTTL:                                dc.GetDurationPropertyFilteredByDomain(dynamicconfig.StickyTTL),
		DecisionHeartbeatTimeout:                 dc.GetDurationPropertyFilteredByDomain(dynamicconfig.DecisionHeartbeatTimeout),
		DecisionRetryCriticalAttempts:            dc.GetIntProperty(dynamicconfig.DecisionRetryCriticalAttempts),
		DecisionRetryMaxAttempts:                 dc.GetIntPropertyFilteredByDomain(dynamicconfig.DecisionRetryMaxAttempts),
		NormalDecisionScheduleToStartMaxAttempts: dc.GetIntPropertyFilteredByDomain(dynamicconfig.NormalDecisionScheduleToStartMaxAttempts),
		NormalDecisionScheduleToStartTimeout:     dc.GetDurationPropertyFilteredByDomain(dynamicconfig.NormalDecisionScheduleToStartTimeout),

		ReplicationTaskFetcherParallelism:                  dc.GetIntProperty(dynamicconfig.ReplicationTaskFetcherParallelism),
		ReplicationTaskFetcherAggregationInterval:          dc.GetDurationProperty(dynamicconfig.ReplicationTaskFetcherAggregationInterval),
		ReplicationTaskFetcherTimerJitterCoefficient:       dc.GetFloat64Property(dynamicconfig.ReplicationTaskFetcherTimerJitterCoefficient),
		ReplicationTaskFetcherErrorRetryWait:               dc.GetDurationProperty(dynamicconfig.ReplicationTaskFetcherErrorRetryWait),
		ReplicationTaskFetcherServiceBusyWait:              dc.GetDurationProperty(dynamicconfig.ReplicationTaskFetcherServiceBusyWait),
		ReplicationTaskProcessorErrorRetryWait:             dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorErrorRetryWait),
		ReplicationTaskProcessorErrorRetryMaxAttempts:      dc.GetIntPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorErrorRetryMaxAttempts),
		ReplicationTaskProcessorErrorSecondRetryWait:       dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorErrorSecondRetryWait),
		ReplicationTaskProcessorErrorSecondRetryMaxWait:    dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorErrorSecondRetryMaxWait),
		ReplicationTaskProcessorErrorSecondRetryExpiration: dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorErrorSecondRetryExpiration),
		ReplicationTaskProcessorNoTaskRetryWait:            dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorNoTaskInitialWait),
		ReplicationTaskProcessorCleanupInterval:            dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorCleanupInterval),
		ReplicationTaskProcessorCleanupJitterCoefficient:   dc.GetFloat64PropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorCleanupJitterCoefficient),
		ReplicationTaskProcessorStartWait:                  dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorStartWait),
		ReplicationTaskProcessorStartWaitJitterCoefficient: dc.GetFloat64PropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorStartWaitJitterCoefficient),
		ReplicationTaskProcessorHostQPS:                    dc.GetFloat64Property(dynamicconfig.ReplicationTaskProcessorHostQPS),
		ReplicationTaskProcessorShardQPS:                   dc.GetFloat64Property(dynamicconfig.ReplicationTaskProcessorShardQPS),
		ReplicationTaskGenerationQPS:                       dc.GetFloat64Property(dynamicconfig.ReplicationTaskGenerationQPS),
		EnableReplicationTaskGeneration:                    dc.GetBoolPropertyFilteredByDomainIDAndWorkflowID(dynamicconfig.EnableReplicationTaskGeneration),
		EnableRecordWorkflowExecutionUninitialized:         dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableRecordWorkflowExecutionUninitialized),

		EnableConsistentQuery:                 dc.GetBoolProperty(dynamicconfig.EnableConsistentQuery),
		EnableConsistentQueryByDomain:         dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableConsistentQueryByDomain),
		EnableCrossClusterOperations:          dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableCrossClusterOperations),
		MaxBufferedQueryCount:                 dc.GetIntProperty(dynamicconfig.MaxBufferedQueryCount),
		MutableStateChecksumGenProbability:    dc.GetIntPropertyFilteredByDomain(dynamicconfig.MutableStateChecksumGenProbability),
		MutableStateChecksumVerifyProbability: dc.GetIntPropertyFilteredByDomain(dynamicconfig.MutableStateChecksumVerifyProbability),
		MutableStateChecksumInvalidateBefore:  dc.GetFloat64Property(dynamicconfig.MutableStateChecksumInvalidateBefore),

		EnableHistoryCorruptionCheck: dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableHistoryCorruptionCheck),

		NotifyFailoverMarkerInterval:               dc.GetDurationProperty(dynamicconfig.NotifyFailoverMarkerInterval),
		NotifyFailoverMarkerTimerJitterCoefficient: dc.GetFloat64Property(dynamicconfig.NotifyFailoverMarkerTimerJitterCoefficient),
		EnableGracefulFailover:                     dc.GetBoolProperty(dynamicconfig.EnableGracefulFailover),

		EnableActivityLocalDispatchByDomain: dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableActivityLocalDispatchByDomain),
		MaxActivityCountDispatchByDomain:    dc.GetIntPropertyFilteredByDomain(dynamicconfig.MaxActivityCountDispatchByDomain),

		ActivityMaxScheduleToStartTimeoutForRetry: dc.GetDurationPropertyFilteredByDomain(dynamicconfig.ActivityMaxScheduleToStartTimeoutForRetry),

		EnableDebugMode:             dc.GetBoolProperty(dynamicconfig.EnableDebugMode)(),
		EnableTaskInfoLogByDomainID: dc.GetBoolPropertyFilteredByDomainID(dynamicconfig.HistoryEnableTaskInfoLogByDomainID),
	}

	return cfg
}

// NewForTest create new history service config for test
func NewForTest() *Config {
	return NewForTestByShardNumber(1)
}

// NewForTestByShardNumber create new history service config for test
func NewForTestByShardNumber(shardNumber int) *Config {
	panicIfErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	inMem := dynamicconfig.NewInMemoryClient()
	panicIfErr(inMem.UpdateValue(dynamicconfig.HistoryLongPollExpirationInterval, 10*time.Second))
	panicIfErr(inMem.UpdateValue(dynamicconfig.EnableConsistentQueryByDomain, true))
	panicIfErr(inMem.UpdateValue(dynamicconfig.ReplicationTaskProcessorHostQPS, float64(10000)))
	panicIfErr(inMem.UpdateValue(dynamicconfig.ReplicationTaskProcessorShardQPS, float64(10000)))
	panicIfErr(inMem.UpdateValue(dynamicconfig.ReplicationTaskProcessorStartWait, time.Nanosecond))
	panicIfErr(inMem.UpdateValue(dynamicconfig.EnableActivityLocalDispatchByDomain, true))
	panicIfErr(inMem.UpdateValue(dynamicconfig.MaxActivityCountDispatchByDomain, 0))
	panicIfErr(inMem.UpdateValue(dynamicconfig.EnableCrossClusterOperations, true))
	panicIfErr(inMem.UpdateValue(dynamicconfig.NormalDecisionScheduleToStartMaxAttempts, 3))
	panicIfErr(inMem.UpdateValue(dynamicconfig.EnablePendingActivityValidation, true))
	dc := dynamicconfig.NewCollection(inMem, log.NewNoop())
	config := New(dc, shardNumber, config.StoreTypeCassandra, false)
	// reduce the duration of long poll to increase test speed
	config.LongPollExpirationInterval = dc.GetDurationPropertyFilteredByDomain(dynamicconfig.HistoryLongPollExpirationInterval)
	config.EnableConsistentQueryByDomain = dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableConsistentQueryByDomain)
	config.ReplicationTaskProcessorHostQPS = dc.GetFloat64Property(dynamicconfig.ReplicationTaskProcessorHostQPS)
	config.ReplicationTaskProcessorShardQPS = dc.GetFloat64Property(dynamicconfig.ReplicationTaskProcessorShardQPS)
	config.ReplicationTaskProcessorStartWait = dc.GetDurationPropertyFilteredByShardID(dynamicconfig.ReplicationTaskProcessorStartWait)
	config.EnableActivityLocalDispatchByDomain = dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableActivityLocalDispatchByDomain)
	config.MaxActivityCountDispatchByDomain = dc.GetIntPropertyFilteredByDomain(dynamicconfig.MaxActivityCountDispatchByDomain)
	config.EnableCrossClusterOperations = dc.GetBoolPropertyFilteredByDomain(dynamicconfig.EnableCrossClusterOperations)
	config.NormalDecisionScheduleToStartMaxAttempts = dc.GetIntPropertyFilteredByDomain(dynamicconfig.NormalDecisionScheduleToStartMaxAttempts)
	config.PendingActivityValidationEnabled = dc.GetBoolProperty(dynamicconfig.EnablePendingActivityValidation)
	return config
}

// GetShardID return the corresponding shard ID for a given workflow ID
func (config *Config) GetShardID(workflowID string) int {
	return common.WorkflowIDToHistoryShard(workflowID, config.NumberOfShards)
}
