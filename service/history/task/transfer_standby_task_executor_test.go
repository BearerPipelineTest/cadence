// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

package task

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/ndc"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	test "github.com/uber/cadence/service/history/testing"
	warchiver "github.com/uber/cadence/service/worker/archiver"
)

type (
	transferStandbyTaskExecutorSuite struct {
		suite.Suite
		*require.Assertions

		controller             *gomock.Controller
		mockShard              *shard.TestContext
		mockDomainCache        *cache.MockDomainCache
		mockNDCHistoryResender *ndc.MockHistoryResender
		mockMatchingClient     *matching.MockClient

		mockVisibilityMgr    *mocks.VisibilityManager
		mockExecutionMgr     *mocks.ExecutionManager
		mockArchivalClient   *warchiver.ClientMock
		mockArchivalMetadata *archiver.MockArchivalMetadata
		mockArchiverProvider *provider.MockArchiverProvider

		logger               log.Logger
		domainID             string
		domainName           string
		domainEntry          *cache.DomainCacheEntry
		version              int64
		clusterName          string
		now                  time.Time
		timeSource           *clock.EventTimeSource
		fetchHistoryDuration time.Duration
		discardDuration      time.Duration

		transferStandbyTaskExecutor *transferStandbyTaskExecutor
	}
)

func TestTransferStandbyTaskExecutorSuite(t *testing.T) {
	s := new(transferStandbyTaskExecutorSuite)
	suite.Run(t, s)
}

func (s *transferStandbyTaskExecutorSuite) SetupSuite() {

}

func (s *transferStandbyTaskExecutorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	config := config.NewForTest()
	s.domainID = constants.TestDomainID
	s.domainName = constants.TestDomainName
	s.domainEntry = constants.TestGlobalDomainEntry
	s.version = s.domainEntry.GetFailoverVersion()
	s.now = time.Now()
	s.timeSource = clock.NewEventTimeSource().Update(s.now)
	s.fetchHistoryDuration = config.StandbyTaskMissingEventsResendDelay() +
		(config.StandbyTaskMissingEventsDiscardDelay()-config.StandbyTaskMissingEventsResendDelay())/2
	s.discardDuration = config.StandbyTaskMissingEventsDiscardDelay() * 2

	s.controller = gomock.NewController(s.T())
	s.mockNDCHistoryResender = ndc.NewMockHistoryResender(s.controller)

	s.mockShard = shard.NewTestContext(
		s.controller,
		&persistence.ShardInfo{
			RangeID:          1,
			TransferAckLevel: 0,
		},
		config,
	)
	s.mockShard.SetEventsCache(events.NewCache(
		s.mockShard.GetShardID(),
		s.mockShard.GetHistoryManager(),
		s.mockShard.GetConfig(),
		s.mockShard.GetLogger(),
		s.mockShard.GetMetricsClient(),
		s.mockShard.GetDomainCache(),
	))
	s.mockShard.Resource.TimeSource = s.timeSource

	s.mockMatchingClient = s.mockShard.Resource.MatchingClient
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr
	s.mockVisibilityMgr = s.mockShard.Resource.VisibilityMgr
	s.mockArchivalClient = &warchiver.ClientMock{}
	s.mockArchivalMetadata = s.mockShard.Resource.ArchivalMetadata
	s.mockArchiverProvider = s.mockShard.Resource.ArchiverProvider
	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestTargetDomainID).Return(constants.TestTargetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestTargetDomainName).Return(constants.TestTargetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestParentDomainID).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestParentDomainID).Return(constants.TestParentDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestParentDomainName).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestChildDomainID).Return(constants.TestGlobalChildDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestChildDomainID).Return(constants.TestChildDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestChildDomainName).Return(constants.TestChildDomainID, nil).AnyTimes()

	s.logger = s.mockShard.GetLogger()
	s.clusterName = cluster.TestAlternativeClusterName
	s.transferStandbyTaskExecutor = NewTransferStandbyTaskExecutor(
		s.mockShard,
		s.mockArchivalClient,
		execution.NewCache(s.mockShard),
		s.mockNDCHistoryResender,
		s.logger,
		s.clusterName,
		config,
	).(*transferStandbyTaskExecutor)
}

func (s *transferStandbyTaskExecutorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
	s.mockArchivalClient.AssertExpectations(s.T())
}

func (s *transferStandbyTaskExecutorSuite) TestProcessActivityTask_Pending() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	event, _ := test.AddActivityTaskScheduledEvent(
		mutableState,
		decisionCompletionID,
		"activity-1",
		"some random activity type",
		mutableState.GetExecutionInfo().TaskList,
		[]byte{}, 1, 1, 1, 1,
	)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeActivityTask,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))
}

func (s *transferStandbyTaskExecutorSuite) TestProcessActivityTask_Pending_PushToMatching() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	event, _ := test.AddActivityTaskScheduledEvent(
		mutableState,
		decisionCompletionID,
		"activity-1",
		"some random activity type",
		mutableState.GetExecutionInfo().TaskList,
		[]byte{}, 1, 1, 1, 1,
	)

	now := time.Now()
	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeActivityTask,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddActivityTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessActivityTask_Success() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	event, _ := test.AddActivityTaskScheduledEvent(
		mutableState,
		decisionCompletionID,
		"activity-1",
		"some random activity type",
		mutableState.GetExecutionInfo().TaskList,
		[]byte{}, 1, 1, 1, 1,
	)
	event = test.AddActivityTaskStartedEvent(mutableState, event.ID, "")
	mutableState.FlushBufferedEvents()

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeActivityTask,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Pending() {

	workflowExecution, mutableState, err := test.StartWorkflow(s.mockShard, s.domainID)
	s.NoError(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Pending_PushToMatching() {

	workflowExecution, mutableState, err := test.StartWorkflow(s.mockShard, s.domainID)
	s.NoError(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Success_FirstDecision() {

	workflowExecution, mutableState, err := test.StartWorkflow(s.mockShard, s.domainID)
	s.NoError(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	})

	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, mutableState.GetExecutionInfo().TaskList, uuid.New())
	di.StartedID = event.ID

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Success_NonFirstDecision() {

	workflowExecution, mutableState, _, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	})

	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, mutableState.GetExecutionInfo().TaskList, uuid.New())
	di.StartedID = event.ID

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessCloseExecution() {
	s.testProcessCloseExecution(persistence.TransferTaskTypeCloseExecution)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessRecordWorkflowClosedTask() {
	s.testProcessCloseExecution(persistence.TransferTaskTypeRecordWorkflowClosed)
}

func (s *transferStandbyTaskExecutorSuite) testProcessCloseExecution(taskType int) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	event := test.AddCompleteWorkflowEvent(mutableState, decisionCompletionID, nil)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            taskType,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessCancelExecution_Pending() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	event, _ := test.AddRequestCancelInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
	)
	nextEventID := event.ID

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeCancelExecution,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	s.mockNDCHistoryResender.EXPECT().SendSingleWorkflowHistory(
		transferTask.GetDomainID(),
		transferTask.GetWorkflowID(),
		transferTask.GetRunID(),
		common.Int64Ptr(nextEventID),
		common.Int64Ptr(s.version),
		nil,
		nil,
	).Return(nil).Times(1)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.discardDuration))
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskDiscarded, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessCancelExecution_Success() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	event, _ := test.AddRequestCancelInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
	)
	event = test.AddCancelRequestedEvent(mutableState, event.ID, constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
	mutableState.FlushBufferedEvents()

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeCancelExecution,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessSignalExecution_Pending() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	event, _ := test.AddRequestSignalInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
		"some random signal name", nil, nil,
	)
	nextEventID := event.ID

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeSignalExecution,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	s.mockNDCHistoryResender.EXPECT().SendSingleWorkflowHistory(
		transferTask.GetDomainID(),
		transferTask.GetWorkflowID(),
		transferTask.GetRunID(),
		common.Int64Ptr(nextEventID),
		common.Int64Ptr(s.version),
		nil,
		nil,
	).Return(nil).Times(1)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.discardDuration))
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskDiscarded, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessSignalExecution_Success() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	event, _ := test.AddRequestSignalInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
		"some random signal name", nil, nil,
	)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeSignalExecution,
		ScheduleID:          event.ID,
	})

	event = test.AddSignaledEvent(mutableState, event.ID, constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), nil)
	mutableState.FlushBufferedEvents()

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessStartChildExecution_Pending() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)
	childWorkflowID := "some random child workflow ID"

	event, _ := test.AddStartChildWorkflowExecutionInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestChildDomainName,
		childWorkflowID,
		"some random child workflow type",
		"some random child task list",
		nil, 1, 1, nil,
	)
	nextEventID := event.ID

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestChildDomainID,
		TargetWorkflowID:    childWorkflowID,
		TargetRunID:         "",
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	s.mockNDCHistoryResender.EXPECT().SendSingleWorkflowHistory(
		transferTask.GetDomainID(),
		transferTask.GetWorkflowID(),
		transferTask.GetRunID(),
		common.Int64Ptr(nextEventID),
		common.Int64Ptr(s.version),
		nil,
		nil,
	).Return(nil).Times(1)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.True(isRedispatchErr(err))

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.discardDuration))
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskDiscarded, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessStartChildExecution_Success() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	childWorkflowID := "some random child workflow ID"
	childWorkflowType := "some random child workflow type"
	event, childInfo := test.AddStartChildWorkflowExecutionInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestChildDomainName,
		childWorkflowID,
		childWorkflowType,
		"some random child task list",
		nil, 1, 1, nil,
	)

	event = test.AddChildWorkflowExecutionStartedEvent(mutableState, event.ID, constants.TestChildDomainName, childWorkflowID, uuid.New(), childWorkflowType)
	mutableState.FlushBufferedEvents()
	childInfo.StartedID = event.ID

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestChildDomainID,
		TargetWorkflowID:    childWorkflowID,
		TargetRunID:         "",
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:          event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessRecordWorkflowStartedTask() {

	workflowExecution, mutableState, err := test.StartWorkflow(s.mockShard, s.domainID)
	s.NoError(err)
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.CronSchedule = "@every 5s"
	startEvent, err := mutableState.GetStartEvent(context.Background())
	s.NoError(err)
	startEvent.WorkflowExecutionStartedEventAttributes.FirstDecisionTaskBackoffSeconds = common.Int32Ptr(5)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeRecordWorkflowStarted,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, startEvent.ID, startEvent.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	if s.mockShard.GetConfig().EnableRecordWorkflowExecutionUninitialized(s.domainName) {
		s.mockVisibilityMgr.On(
			"RecordWorkflowExecutionUninitialized",
			mock.Anything,
			createRecordWorkflowExecutionUninitializedRequest(transferTask, mutableState),
		).Once().Return(nil)
	}
	s.mockVisibilityMgr.On(
		"RecordWorkflowExecutionStarted",
		mock.Anything,
		createRecordWorkflowExecutionStartedRequest(
			s.domainName, startEvent, transferTask, mutableState, 2),
	).Return(nil).Once()

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessUpsertWorkflowSearchAttributesTask() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	now := time.Now()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              int64(59),
		TaskList:            mutableState.GetExecutionInfo().TaskList,
		TaskType:            persistence.TransferTaskTypeUpsertWorkflowSearchAttributes,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, decisionCompletionID, mutableState.GetCurrentVersion())
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	startEvent, err := mutableState.GetStartEvent(context.Background())
	s.NoError(err)
	s.mockVisibilityMgr.On(
		"UpsertWorkflowExecution",
		mock.Anything,
		createUpsertWorkflowSearchAttributesRequest(
			s.domainName, startEvent, transferTask, mutableState, 2),
	).Return(nil).Once()

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) newTransferTaskFromInfo(
	info *persistence.TransferTaskInfo,
) Task {
	return NewTransferTask(s.mockShard, info, QueueTypeStandbyTransfer, s.logger, nil, nil, nil, nil, nil)
}
