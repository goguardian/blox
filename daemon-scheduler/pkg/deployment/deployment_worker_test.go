// Copyright 2016-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package deployment

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	deploymenttypes "github.com/goguardian/blox/daemon-scheduler/pkg/deployment/types"
	environmenttypes "github.com/goguardian/blox/daemon-scheduler/pkg/environment/types"
	"github.com/goguardian/blox/daemon-scheduler/pkg/facade"
	"github.com/goguardian/blox/daemon-scheduler/pkg/mocks"
	"github.com/goguardian/blox/daemon-scheduler/pkg/types"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	TaskRunning = "RUNNING"
)

type DeploymentWorkerTestSuite struct {
	suite.Suite
	environmentService         *mocks.MockEnvironmentService
	environmentFacade          *mocks.MockEnvironmentFacade
	deploymentService          *mocks.MockDeploymentService
	clusterState               *facade.MockClusterState
	ecs                        *mocks.MockECS
	environmentObject          *environmenttypes.Environment
	pendingDeploymentObject    *deploymenttypes.Deployment
	inProgressDeploymentObject *deploymenttypes.Deployment
	clusterTaskARNs            []*string
	emptyDescribeTasksOutput   *ecs.DescribeTasksOutput
	deploymentWorker           DeploymentWorker
	ctx                        context.Context
}

func (suite *DeploymentWorkerTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())

	suite.environmentService = mocks.NewMockEnvironmentService(mockCtrl)
	suite.environmentFacade = mocks.NewMockEnvironmentFacade(mockCtrl)
	suite.deploymentService = mocks.NewMockDeploymentService(mockCtrl)
	suite.clusterState = facade.NewMockClusterState(mockCtrl)
	suite.ecs = mocks.NewMockECS(mockCtrl)
	suite.deploymentWorker = NewDeploymentWorker(suite.environmentService, suite.environmentFacade,
		suite.deploymentService, suite.ecs, suite.clusterState)

	var err error
	suite.environmentObject, err = environmenttypes.NewEnvironment(environmentName, taskDefinition, cluster1)
	assert.Nil(suite.T(), err, "Cannot initialize DeploymentWorkerTestSuite")

	suite.pendingDeploymentObject, err = deploymenttypes.NewDeployment(taskDefinition, suite.environmentObject.Token)
	assert.Nil(suite.T(), err, "Cannot initialize DeploymentWorkerTestSuite")

	inProgressDeployment := *suite.pendingDeploymentObject
	suite.inProgressDeploymentObject = &inProgressDeployment

	err = suite.inProgressDeploymentObject.UpdateDeploymentToInProgress(0, nil)
	assert.Nil(suite.T(), err, "Cannot initialize DeploymentWorkerTestSuite")

	suite.clusterTaskARNs = []*string{aws.String(taskARN1), aws.String(taskARN2)}
	suite.emptyDescribeTasksOutput = &ecs.DescribeTasksOutput{
		Tasks: []*ecs.Task{},
	}

	suite.ctx = context.TODO()
}

func TestDeploymentWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(DeploymentWorkerTestSuite))
}

func (suite *DeploymentWorkerTestSuite) TestNewDeploymentWorker() {
	w := NewDeploymentWorker(suite.environmentService, suite.environmentFacade, suite.deploymentService, suite.ecs, suite.clusterState)
	assert.NotNil(suite.T(), w, "Worker should not be nil")
}

func (suite *DeploymentWorkerTestSuite) TestStartPendingDeploymentEmptyEnvironmentName() {
	_, err := suite.deploymentWorker.StartPendingDeployment(suite.ctx, "")
	assert.Error(suite.T(), err, "Expected an error when env name is missing")
}

func (suite *DeploymentWorkerTestSuite) TestStartPendingDeploymentGetEnvironmentFails() {
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(nil, errors.New("Get environment failed"))
	suite.environmentFacade.EXPECT().InstanceARNs(suite.environmentObject).Times(0)
	suite.deploymentService.EXPECT().StartDeployment(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	_, err := suite.deploymentWorker.StartPendingDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when get environment fails")
}

func (suite *DeploymentWorkerTestSuite) TestStartPendingDeploymentGetEnvironmentIsEmpty() {
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(nil, nil).Times(1)
	suite.environmentFacade.EXPECT().InstanceARNs(suite.environmentObject).Times(0)
	suite.deploymentService.EXPECT().StartDeployment(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	d, err := suite.deploymentWorker.StartPendingDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err, "Unexpected error when get environment is empty")
	assert.Nil(suite.T(), d)
}

func (suite *DeploymentWorkerTestSuite) TestStartPendingDeploymentInstanceARNsFails() {
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil).Times(1)
	suite.environmentFacade.EXPECT().InstanceARNs(suite.environmentObject).Return(nil, errors.New("Instance ARNs fails")).Times(1)
	suite.deploymentService.EXPECT().StartDeployment(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	_, err := suite.deploymentWorker.StartPendingDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when get instance arns fails")
}

func (suite *DeploymentWorkerTestSuite) TestStartPendingDeploymentStartDeploymentFails() {
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil).Times(1)
	suite.environmentFacade.EXPECT().InstanceARNs(suite.environmentObject).Return(suite.clusterTaskARNs, nil).Times(1)
	suite.deploymentService.EXPECT().StartDeployment(suite.ctx, environmentName, suite.clusterTaskARNs).
		Return(nil, errors.New("Start deployment fails"))

	_, err := suite.deploymentWorker.StartPendingDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when start deployment fails")
}

func (suite *DeploymentWorkerTestSuite) TestStartPendingDeployment() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(nil, nil).Times(1)
	suite.deploymentService.EXPECT().GetPendingDeployment(suite.ctx, environmentName).Return(suite.pendingDeploymentObject, nil).Times(1)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil).Times(1)
	suite.environmentFacade.EXPECT().InstanceARNs(suite.environmentObject).Return(suite.clusterTaskARNs, nil).Times(1)
	suite.deploymentService.EXPECT().StartDeployment(suite.ctx, environmentName, suite.clusterTaskARNs).
		Return(suite.inProgressDeploymentObject, nil).Times(1)

	d, err := suite.deploymentWorker.StartPendingDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err)
	verifyDeployment(suite.T(), suite.inProgressDeploymentObject, d)
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentEmptyEnvironmentName() {
	_, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, "")
	assert.Error(suite.T(), err, "Expected an error when env name is missing")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentGetInProgressDeploymentFails() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).
		Return(nil, errors.New("Get in progress deployment failed"))

	_, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when get in progress deployment fails")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentNoInProgressDeployment() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(nil, nil)

	d, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err, "Unexpected error when get in progress deployment returns empty")
	assert.Nil(suite.T(), d, "Deployment should be nil when get in progress Deployment returns empty")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentGetEnvironmentFails() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(nil, errors.New("Get environment failed"))

	_, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when get environment fails")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentGetEnvironmentIsNil() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(nil, nil)

	d, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err, "Unxpected error when get environment is empty")
	assert.Nil(suite.T(), d, "Deployment should be nil when get environment returns empty")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentListTasksFails() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil)
	suite.ecs.EXPECT().ListTasks(suite.environmentObject.Cluster, suite.inProgressDeploymentObject.ID).
		Return(nil, errors.New("ListTasks failed"))

	_, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when list tasks fails")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentDescribeTasksFails() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil)
	suite.ecs.EXPECT().ListTasks(suite.environmentObject.Cluster, suite.inProgressDeploymentObject.ID).
		Return(suite.clusterTaskARNs, nil)
	suite.ecs.EXPECT().DescribeTasks(suite.environmentObject.Cluster, suite.clusterTaskARNs).
		Return(nil, errors.New("DescribeTasks failed"))

	_, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when describe tasks fails")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentNoTasksStartedByTheDeployment() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil).Times(2)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil)
	suite.ecs.EXPECT().ListTasks(suite.environmentObject.Cluster, suite.inProgressDeploymentObject.ID).
		Return(suite.clusterTaskARNs, nil)

	noTasks := &ecs.DescribeTasksOutput{
		Tasks: []*ecs.Task{},
	}

	suite.ecs.EXPECT().DescribeTasks(suite.environmentObject.Cluster, suite.clusterTaskARNs).Return(noTasks, nil)
	suite.deploymentService.EXPECT().UpdateInProgressDeployment(suite.ctx, suite.environmentObject.Name, suite.inProgressDeploymentObject).
		Return(nil)

	d, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err, "Unexpected error when there are no tasks started by the deployment")
	assert.Equal(suite.T(), suite.inProgressDeploymentObject, d, "Expected deployments to match")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentTasksArePending() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil).Times(2)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil)
	suite.ecs.EXPECT().ListTasks(suite.environmentObject.Cluster, suite.inProgressDeploymentObject.ID).
		Return(suite.clusterTaskARNs, nil)

	pendingTask := &ecs.Task{
		TaskArn:    aws.String(taskARN1),
		LastStatus: aws.String(TaskPending),
	}

	runningTask := &ecs.Task{
		TaskArn:    aws.String(taskARN2),
		LastStatus: aws.String(TaskRunning),
	}

	tasks := &ecs.DescribeTasksOutput{
		Tasks: []*ecs.Task{runningTask, pendingTask},
	}

	suite.ecs.EXPECT().DescribeTasks(suite.environmentObject.Cluster, suite.clusterTaskARNs).Return(tasks, nil)
	suite.deploymentService.EXPECT().UpdateInProgressDeployment(suite.ctx, suite.environmentObject.Name, suite.inProgressDeploymentObject).
		Return(nil)

	d, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err, "Unexpected error when there is a pending task started by the deployment")
	assert.Equal(suite.T(), suite.inProgressDeploymentObject, d, "Expected deployments to match")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentDeploymentCompleted() {
	suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil).Times(2)
	suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil)
	suite.ecs.EXPECT().ListTasks(suite.environmentObject.Cluster, suite.inProgressDeploymentObject.ID).
		Return(suite.clusterTaskARNs, nil)

	runningTask1 := &ecs.Task{
		TaskArn:    aws.String(taskARN1),
		LastStatus: aws.String(TaskRunning),
	}

	runningTask2 := &ecs.Task{
		TaskArn:    aws.String(taskARN2),
		LastStatus: aws.String(TaskRunning),
	}

	tasks := &ecs.DescribeTasksOutput{
		Tasks: []*ecs.Task{runningTask1, runningTask2},
	}

	suite.ecs.EXPECT().DescribeTasks(suite.environmentObject.Cluster, suite.clusterTaskARNs).Return(tasks, nil)

	err := suite.inProgressDeploymentObject.UpdateDeploymentToCompleted(nil)
	completedDeployment := suite.inProgressDeploymentObject
	assert.Nil(suite.T(), err, "Unexpected error when moving deployment to completed")

	suite.deploymentService.EXPECT().UpdateInProgressDeployment(suite.ctx, suite.environmentObject.Name, gomock.Any()).Do(
		func(_ interface{}, _ interface{}, d *deploymenttypes.Deployment) {
			verifyDeploymentCompleted(suite.T(), completedDeployment, d)
		}).Return(nil)

	d, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err, "Unexpected error when the deployment is completed")
	verifyDeploymentCompleted(suite.T(), completedDeployment, d)
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentUpdateInProgressDeploymentFailsWithUnexpectedDeploymentStatusError() {
	gomock.InOrder(
		suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil),
		suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil),
		suite.ecs.EXPECT().ListTasks(suite.environmentObject.Cluster, suite.inProgressDeploymentObject.ID).
			Return(suite.clusterTaskARNs, nil),
		suite.ecs.EXPECT().DescribeTasks(suite.environmentObject.Cluster, suite.clusterTaskARNs).
			Return(suite.emptyDescribeTasksOutput, nil),
		suite.deploymentService.EXPECT().UpdateInProgressDeployment(suite.ctx, suite.environmentObject.Name, suite.inProgressDeploymentObject).
			Return(types.NewUnexpectedDeploymentStatusError(errors.New("Update deployment failed since deployment status was unexpected"))),
	)
	dep, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Nil(suite.T(), err, "Unexpected error when update in-progress deployment fails with unexpected deployment status error")
	assert.Nil(suite.T(), dep, "Expected deployment to be empty when update deployment fails due to unexpected deployment status error")
}

func (suite *DeploymentWorkerTestSuite) TestUpdateInProgressDeploymentUpdateDeploymentFails() {
	gomock.InOrder(
		suite.deploymentService.EXPECT().GetInProgressDeployment(suite.ctx, environmentName).Return(suite.inProgressDeploymentObject, nil),
		suite.environmentService.EXPECT().GetEnvironment(suite.ctx, environmentName).Return(suite.environmentObject, nil),
		suite.ecs.EXPECT().ListTasks(suite.environmentObject.Cluster, suite.inProgressDeploymentObject.ID).
			Return(suite.clusterTaskARNs, nil),
		suite.ecs.EXPECT().DescribeTasks(suite.environmentObject.Cluster, suite.clusterTaskARNs).
			Return(suite.emptyDescribeTasksOutput, nil),
		suite.deploymentService.EXPECT().UpdateInProgressDeployment(suite.ctx, suite.environmentObject.Name, suite.inProgressDeploymentObject).
			Return(errors.New("Update deployment failed")),
	)

	_, err := suite.deploymentWorker.UpdateInProgressDeployment(suite.ctx, environmentName)
	assert.Error(suite.T(), err, "Expected an error when update deployment fails")
}

func verifyDeploymentCompleted(t *testing.T, expected *deploymenttypes.Deployment, actual *deploymenttypes.Deployment) {
	assert.Exactly(t, expected.ID, actual.ID, "Deployment ids should match")
	assert.Exactly(t, deploymenttypes.DeploymentCompleted, actual.Status, "Deployment status should be completed")
	assert.Exactly(t, expected.Health, actual.Health, "Deployment health should match")
	assert.Exactly(t, expected.TaskDefinition, actual.TaskDefinition, "Deployment task definition should match")
	assert.Exactly(t, expected.DesiredTaskCount, actual.DesiredTaskCount, "Deployment desired task count should match")
	assert.Exactly(t, expected.FailedInstances, actual.FailedInstances, "Deployment failed instances should match")
}
