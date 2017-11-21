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

package steps

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/goguardian/blox/daemon-scheduler/internal/features/wrappers"
	"github.com/goguardian/blox/daemon-scheduler/swagger/v1/generated/client/operations"
	"github.com/goguardian/blox/daemon-scheduler/swagger/v1/generated/models"
	. "github.com/gucumber/gucumber"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

var (
	roleName            = "DSTestRole"
	instanceProfileName = "DSTestInstance"
	policyARN           = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
)

const (
	deploymentCompleted           = "completed"
	taskRunning                   = "RUNNING"
	deploymentCompleteWaitSeconds = 50
	invalidCluster                = "cluster/cluster"
	badRequestHTTPResponse        = "400 Bad Request"
	listEnvironmentsBadRequest    = "ListEnvironmentsBadRequest"

	errorCode            	      = 1
	launchConfigurationName	      = "DSASGLaunchConfiguration"
	defaultASGClusterName 	      = "DSClusterASG"
	defaultECSClusterName 	      = "DSTestCluster"
)

func init() {
	asgWrapper := wrappers.NewAutoScalingWrapper()
	ecsWrapper := wrappers.NewECSWrapper()
	ec2Wrapper := wrappers.NewEC2Wrapper()
	iamWrapper := wrappers.NewIAMWrapper()
	edsWrapper := wrappers.NewEDSWrapper()
	ctx := context.Background()

	css, err := wrappers.NewClusterState()
	if err != nil {
		T.Errorf("Error creating CSS client: %v", err)
		return
	}

	// TODO: Change these os.Exit calls to T.Errorf. Currently unable to do so because T is not initialized until the first test.
	// (https://github.com/gucumber/gucumber/issues/28)
	BeforeAll(func() {
		clusterName := wrappers.GetClusterName()

		_, err := ecsWrapper.CreateCluster(clusterName)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(errorCode)
		}

		err = terminateAllContainerInstances(ec2Wrapper, ecsWrapper, clusterName)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(errorCode)
		}

		azs, err := ec2Wrapper.DescribeAvailabilityZones()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(errorCode)
		}

		asg = wrappers.GetASGName()
		if asg == defaultASGClusterName {
			keyPair := wrappers.GetKeyPairName()

			err = createInstanceProfile(iamWrapper)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(errorCode)
			}

			amiID, err := wrappers.GetLatestECSOptimizedAMIID()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(errorCode)
			}

			err = asgWrapper.CreateLaunchConfiguration(launchConfigurationName, clusterName, instanceProfileName, keyPair, amiID)
			if err != nil {
				if awsErr, ok := errors.Cause(err).(awserr.Error); !ok || awsErr.Code() != "AlreadyExists" {
					fmt.Println(err.Error())
					os.Exit(errorCode)
				}
			}

			err = asgWrapper.CreateAutoScalingGroup(asg, launchConfigurationName, azs)
			if err != nil {
				if awsErr, ok := errors.Cause(err).(awserr.Error); !ok || awsErr.Code() != "AlreadyExists" {
					fmt.Println(err.Error())
					os.Exit(errorCode)
				} else {
					asgStatus, err := asgWrapper.GetAutoScalingGroupStatus(asg)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(errorCode)
					}
					if asgStatus == "Delete in progress" {
						fmt.Println("The current Autoscaling Group is in progress of deleting. Wait for the deletion to complete and restart the test.")
						os.Exit(errorCode)
					}
				}
			}
		}
	})

	AfterAll(func() {
		clusterName := wrappers.GetClusterName()

		err := stopAllTasks(ecsWrapper, clusterName)
		if err != nil {
			T.Errorf(err.Error())
			return
		}

		if wrappers.GetASGName() == defaultASGClusterName {
			forceDelete := true
			// With ForceDelete set to true, this will delete all instances attached to the Autoscaling group
			err = asgWrapper.DeleteAutoScalingGroup(asg, forceDelete)
			if err != nil {
				T.Errorf(err.Error())
				return
			}

			err = asgWrapper.DeleteLaunchConfiguration(launchConfigurationName)
			if err != nil {
				T.Errorf(err.Error())
				return
			}
		} else {
			err = terminateAllContainerInstances(ec2Wrapper, ecsWrapper, clusterName)
			if err != nil {
				T.Errorf(err.Error())
				return
			}
		}

		if wrappers.GetClusterName() == defaultECSClusterName {
			_, err = ecsWrapper.DeleteCluster(clusterName)
			if err != nil {
				T.Errorf(err.Error())
				return
			}
		}

		err = deleteInstanceProfile(iamWrapper)
		if err != nil {
			T.Errorf(err.Error())
			return
		}

	})

	When(`^I make a Ping call$`, func() {
		err = edsWrapper.Ping()
	})

	Then(`^the Ping response indicates that the service is healthy$`, func() {
		if err != nil {
			T.Errorf(err.Error())
		}
	})

	Given(`^A cluster "env.(.+?)" and asg "env.(.+?)"$`, func(cEnv string, aEnv string) {
		c := os.Getenv(cEnv)
		if len(c) == 0 {
			T.Errorf("ECS_CLUSTER env-var is not defined")
		}
		a := os.Getenv(aEnv)
		if len(a) == 0 {
			T.Errorf("ECS_CLUSTER_ASG env-var is not defined")
		}
		cluster = c
		asg = a
	})

	Given(`^A cluster named "env.(.+?)"$`, func(cEnv string) {
		c := os.Getenv(cEnv)
		if len(c) == 0 {
			T.Errorf("ECS_CLUSTER env-var is not defined")
		}
		cluster = c
	})

	Given(`^(?:a|another) cluster "(.+?)"$`, func(c string) {
		cARN, err := ecsWrapper.CreateCluster(c)
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		cluster = c
		clusterARN = *cARN
	})

	When(`^I update the desired-capacity of cluster to (\d+) instances and wait for a max of (\d+) seconds$`, func(count int, seconds int) {
		err := asgWrapper.SetDesiredCapacity(asg, int64(count))
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		ok, err := doSomething(time.Duration(seconds)*time.Second, 1*time.Second, func() (bool, error) {
			instances, err := css.ListInstances(cluster)
			if err != nil {
				return false, errors.Wrapf(err, "Error calling ListInstances for cluster %s", cluster)
			}
			activeCount := 0
			for _, instance := range instances {
				if "ACTIVE" == aws.StringValue(instance.Entity.Status) {
					activeCount++
				}
			}
			return count == activeCount, nil
		})
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		if !ok {
			T.Errorf("Expected %d instances in cluster %s", count, cluster)
			return
		}

	})

	And(`^a registered "(.+?)" task-definition$`, func(td string) {
		resp, err := ecsWrapper.RegisterTaskDefinition(td)
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		taskDefinition = resp
	})

	And(`^I deregister task-definition$`, func() {
		err := ecsWrapper.DeregisterTaskDefinition(taskDefinition)
		if err != nil {
			T.Errorf(err.Error())
			return
		}
	})

	When(`^I create an environment with name "(.+?)" in the cluster using the task-definition$`,
		func(e string) {
			environment = e
			err := edsWrapper.DeleteEnvironment(&environment)
			if err != nil {
				T.Errorf("Was not able to delete environment %v: %v", environment, err.Error())
			}

			createEnvReq := &models.CreateEnvironmentRequest{
				InstanceGroup: &models.InstanceGroup{
					Cluster: cluster,
				},
				Name:           &environment,
				TaskDefinition: &taskDefinition,
			}
			env, err := edsWrapper.CreateEnvironment(createEnvReq)
			if err != nil {
				_, ok := err.(*operations.CreateEnvironmentBadRequest)
				if !ok {
					T.Errorf(err.Error())
					return
				}
			} else {
				deploymentToken = env.DeploymentToken
			}
		})

	Then(`^creating the same environment should fail with BadRequest$`,
		func() {
			createEnvReq := &models.CreateEnvironmentRequest{
				InstanceGroup: &models.InstanceGroup{
					Cluster: cluster,
				},
				Name:           &environment,
				TaskDefinition: &taskDefinition,
			}
			_, err := edsWrapper.CreateEnvironment(createEnvReq)
			if err != nil {
				_, ok := err.(*operations.CreateEnvironmentBadRequest)
				if !ok {
					T.Errorf(err.Error())
					return
				}
			} else {
				T.Errorf("Expecting CreateEnvironmentBadRequest error")
				return
			}
		})

	Then(`^I create an environment with name "(.+?)" it should fail with NotFound$`, func(e string) {
		err := edsWrapper.DeleteEnvironment(&e)
		if err != nil {
			T.Errorf("Was not able to delete environment %v: %v", e, err.Error())
		}

		createEnvReq := &models.CreateEnvironmentRequest{
			InstanceGroup: &models.InstanceGroup{
				Cluster: cluster,
			},
			Name:           &e,
			TaskDefinition: &taskDefinition,
		}
		_, err = edsWrapper.CreateEnvironment(createEnvReq)
		if err != nil {
			_, ok := err.(*operations.CreateEnvironmentNotFound)
			if !ok {
				T.Errorf(err.Error())
				return
			}
		} else {
			T.Errorf("Expecting CreateEnvironmentNotFound error")
			return
		}
	})

	Then(`^I create an environment with name "(.+?)" it should fail with BadRequest$`,
		func(e string) {
			err := edsWrapper.DeleteEnvironment(&e)
			if err != nil {
				T.Errorf("Was not able to delete environment %v: %v", e, err.Error())
			}

			createEnvReq := &models.CreateEnvironmentRequest{
				InstanceGroup: &models.InstanceGroup{
					Cluster: cluster,
				},
				Name:           &e,
				TaskDefinition: &taskDefinition,
			}
			_, err = edsWrapper.CreateEnvironment(createEnvReq)
			if err != nil {
				_, ok := err.(*operations.CreateEnvironmentBadRequest)
				if !ok {
					T.Errorf(err.Error())
					return
				}
			} else {
				T.Errorf("Expecting CreateEnvironmentBadRequest error")
				return
			}
		})

	And(`^I delete cluster$`,
		func() {
			_, err := ecsWrapper.DeleteCluster(cluster)
			if err != nil {
				T.Errorf(err.Error())
				return
			}
		})

	Then(`^GetEnvironment should succeed$`,
		func() {
			_, err := edsWrapper.GetEnvironment(&environment)
			if err != nil {
				T.Errorf(err.Error())
				return
			}
		})

	Then(`^GetEnvironment with name "(.+?)" should fail with NotFound$`,
		func(e string) {
			_, err := edsWrapper.GetEnvironment(&e)
			if err != nil {
				_, ok := err.(*operations.GetEnvironmentNotFound)
				if !ok {
					T.Errorf(err.Error())
					return
				}
			} else {
				T.Errorf("Expecting GetEnvironmentNotFound error")
				return
			}
		})

	Then(`^GetDeployment with created deployment should succeed$`,
		func() {
			deploymentGet, err := edsWrapper.GetDeployment(&environment, &deploymentID)
			if err != nil {
				T.Errorf(err.Error())
				return
			}
			assert.Equal(T, deploymentID, *deploymentGet.ID, "DeploymentID should match")
		})

	Then(`^the environment should be returned in ListEnvironments call$`,
		func() {
			environments, err := edsWrapper.ListEnvironments()
			if err != nil {
				T.Errorf(err.Error())
				return
			}
			found := false
			for _, env := range environments {
				if *env.Name == environment {
					found = true
					break
				}
			}
			assert.Equal(T, true, found, "Did not find environment with name "+environment)
		})

	Then(`^there should be at least (\d+) environment returned when I call ListEnvironments with cluster filter set to the second cluster ARN$`, func(numEnvs int) {
		environments, err := edsWrapper.ListEnvironmentsWithClusterFilter(clusterARN)
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		assert.True(T, len(environments) >= numEnvs,
			"Number of environments in the response should be at least "+string(numEnvs))
		environmentList = environments
	})

	Then(`^there should be at least (\d+) environment returned when I call ListEnvironments with cluster filter set to the second cluster name$`, func(numEnvs int) {
		environments, err := edsWrapper.ListEnvironmentsWithClusterFilter(cluster)
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		assert.True(T, len(environments) >= numEnvs,
			"Number of environments in the response should be at least "+string(numEnvs))
		environmentList = environments
	})

	And(`^all the environments in the response should correspond to the second cluster$`,
		func() {
			for _, env := range environmentList {
				if env.InstanceGroup.Cluster != clusterARN {
					T.Errorf("Environment in list environments response with cluster filter set to '" +
						clusterARN + "' belongs to cluster + '" + env.InstanceGroup.Cluster + "'")
				}
			}
		})

	And(`^second environment should be one of the environments in the response$`,
		func() {
			found := false
			for _, env := range environmentList {
				if *env.Name == environment {
					found = true
					break
				}
			}
			assert.True(T, found, "Did not find environment with name "+environment)
		})

	When(`^I try to call ListEnvironments with redundant filters$`, func() {
		url := "http://localhost:2000/v1/environments?cluster=cluster1&cluster=cluster2"
		resp, err := http.Get(url)
		if err != nil {
			T.Errorf(err.Error())
		}

		var exceptionType string
		if resp.Status == badRequestHTTPResponse {
			exceptionType = listEnvironmentsBadRequest
		} else {
			T.Errorf("Unknown exception type '%s' when trying to list environments with redundant filters", resp.Status)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			T.Errorf("Error reading expection message when trying to list environments with redundant filters")
		}
		exceptionMsg := string(body)
		exception = Exception{exceptionType: exceptionType, exceptionMsg: exceptionMsg}
	})

	When(`^I try to call ListEnvironments with an invalid cluster filter$`, func() {
		exception = Exception{}
		exceptionMsg, exceptionType, err := edsWrapper.TryListEnvironmentsWithInvalidCluster(invalidCluster)
		if err != nil {
			T.Errorf(err.Error())
		}
		exception = Exception{exceptionType: exceptionType, exceptionMsg: exceptionMsg}
	})

	Then(`^I get a (.+?) exception$`, func(exceptionType string) {
		if exception.exceptionType == "" {
			T.Errorf("Error memorizing exception type")
		}
		if exceptionType != exception.exceptionType {
			T.Errorf("Expected exception type '%s' but got '%s'. ", exceptionType, exception.exceptionType)
		}
	})

	And(`^the exception message contains "(.+?)"$`, func(exceptionMsg string) {
		if exception.exceptionMsg == "" {
			T.Errorf("Error memorizing exception message")
		}
		if !strings.Contains(exception.exceptionMsg, exceptionMsg) {
			T.Errorf("Expected exception message returned '%s' to contain '%s'. ", exception.exceptionMsg, exceptionMsg)
		}
	})

	Then(`^I call CreateDeployment API$`, func() {
		createDeployment(deploymentToken, ctx, edsWrapper)
	})

	Then(`^creating another deployment with the same token should fail$`, func() {
		_, err := edsWrapper.CreateDeployment(context.TODO(), &environment, &deploymentToken)
		if err != nil {
			_, ok := err.(*operations.CreateDeploymentBadRequest)
			if !ok {
				T.Errorf(err.Error())
				return
			}
		} else {
			T.Errorf("Expecting CreateDeploymentBadRequest error")
			return
		}
	})

	Then(`^Deployment should be returned in ListDeployments call$`,
		func() {
			deployments, err := edsWrapper.ListDeployments(aws.String(environment))
			if err != nil {
				T.Errorf(err.Error())
				return
			}
			found := false
			for _, d := range deployments {
				if *d.ID == deploymentID {
					found = true
					break
				}
			}
			assert.Equal(T, true, found, fmt.Sprintf("Did not find deployment with id:%s under environment:%s", deploymentID, environment))
		})

	Then(`^the deployment should have (\d+) task(?:|s) running within (\d+) seconds$`, func(count int, seconds int) {
		ok, err := doSomething(time.Duration(seconds)*time.Second, 1*time.Second, func() (bool, error) {
			tasks, err := ecsWrapper.ListTasks(cluster, aws.String(deploymentID))
			if err != nil {
				return false, errors.Wrapf(err, "Error calling ListTasks for cluster %s and deployment %s", cluster, deploymentID)
			}

			runningTasks := filterTasksByStatusRunning(aws.String(cluster), tasks, ecsWrapper)
			return count == len(runningTasks), nil
		})

		if err != nil {
			T.Errorf(err.Error())
			return
		}

		if !ok {
			T.Errorf("Expecting at least %d tasks to be launched in the cluster %v", count, cluster)
			return
		}
	})

	Then(`^the deployment should complete in (\d+) seconds$`, func(seconds int) {
		ok, err := doSomething(time.Duration(seconds)*time.Second, 1*time.Second, func() (bool, error) {
			deployment, err := edsWrapper.GetDeployment(aws.String(environment), aws.String(deploymentID))
			if err != nil {
				return false, errors.Wrapf(err, "Error calling GetDeployment for environment %s and deployment %s", environment, deploymentID)
			}

			return aws.StringValue(deployment.Status) == models.DeploymentStatusCompleted, nil
		})

		if err != nil {
			T.Errorf(err.Error())
			return
		}

		if !ok {
			T.Errorf("Expecting the deployment status to be %v", taskRunning)
			return
		}
	})

	And(`^Deployment should be marked as completed$`, func() {
		deployment, err := edsWrapper.GetDeployment(aws.String(environment), aws.String(deploymentID))
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		if aws.StringValue(deployment.Status) != deploymentCompleted {
			T.Errorf("Expected deployment %s to be completed but was %s", deploymentID, *deployment.Status)
			return
		}
	})

	And(`^I stop the tasks running in cluster$`, func() {
		tasks, err := ecsWrapper.ListTasks(cluster, nil)
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		for _, task := range tasks {
			err := ecsWrapper.StopTask(cluster, *task)
			if err != nil {
				T.Errorf(err.Error())
				return
			}
		}
	})

	When(`^I call GetDeployment with environment "(.+?)", it should fail with NotFound$`,
		func(e string) {
			_, err := edsWrapper.GetDeployment(aws.String(e), aws.String(deploymentID))
			if err != nil {
				_, ok := err.(*operations.GetDeploymentNotFound)
				if !ok {
					T.Errorf(err.Error())
					return
				}
			} else {
				T.Errorf("Expecting GetDeploymentNotFound error")
				return
			}
		})

	When(`^I call GetDeployment with id "(.+?)", it should fail with NotFound$`,
		func(d string) {
			_, err := edsWrapper.GetDeployment(aws.String(environment), aws.String(d))
			if err != nil {
				_, ok := err.(*operations.GetDeploymentNotFound)
				if !ok {
					T.Errorf(err.Error())
					return
				}
			} else {
				T.Errorf("Expecting GetDeploymentNotFound error")
				return
			}
		})

	When(`^I call ListDeployments with environment "(.+?)", it should fail with NotFound$`,
		func(e string) {
			_, err := edsWrapper.ListDeployments(aws.String(e))
			if err != nil {
				_, ok := err.(*operations.ListDeploymentsNotFound)
				if !ok {
					T.Errorf(err.Error())
					return
				}
			} else {
				T.Errorf("Expecting ListDeploymentsNotFound error")
				return
			}
		})

	And(`^I call CreateDeployment API (.+?) times$`, func(count int) {
		for i := 0; i < count; i++ {
			deployment := createDeployment("", ctx, edsWrapper)
			waitForDeploymentToComplete(deploymentCompleteWaitSeconds, edsWrapper)
			deploymentIDs[*deployment.ID] = deployment
		}
	})

	And(`^ListDeployments should return (.+?) deployment$`, func(count int) {
		deployments, err := edsWrapper.ListDeployments(aws.String(environment))
		if err != nil {
			T.Errorf(err.Error())
			return
		}
		assert.True(T, count <= len(deployments), "Wrong number of deployments returned")
		deploymentsFromResponse := make(map[string]bool)
		for _, d := range deployments {
			deploymentsFromResponse[*d.ID] = true
		}
		for key, _ := range deploymentIDs {
			if !deploymentsFromResponse[key] {
				T.Errorf("Did not find deployment with id:%s under environment:%s", key, environment)
			}
		}
	})

	When(`^I delete the environment$`, func() {
		deleteEnvironment(environment, edsWrapper)
	})

	And(`^deleting the environment again should succeed$`, func() {
		deleteEnvironment(environment, edsWrapper)
	})

	Then(`^get environment should return empty$`, func() {
		_, err := edsWrapper.GetEnvironment(&environment)
		if err != nil {
			_, ok := err.(*operations.GetEnvironmentNotFound)
			if !ok {
				T.Errorf(err.Error())
				return
			}
		} else {
			T.Errorf("Expecting GetEnvironmentNotFound error")
			return
		}
	})
}

func terminateAllContainerInstances(ec2Wrapper wrappers.EC2Wrapper, ecsWrapper wrappers.ECSWrapper, clusterName string) error {
	instanceARNs, err := ecsWrapper.ListContainerInstances(clusterName)
	if err != nil {
		return errors.Wrapf(err, "Failed to list container instances from cluster '%v'.", clusterName)
	}

	if len(instanceARNs) == 0 {
		return nil
	}

	err = ecsWrapper.DeregisterContainerInstances(&clusterName, instanceARNs)
	if err != nil {
		return errors.Wrapf(err, "Failed to deregister container instances '%v'.", instanceARNs)
	}

	ec2InstanceIDs := make([]*string, 0, len(instanceARNs))
	for _, v := range instanceARNs {
		containerInstance, err := ecsWrapper.DescribeContainerInstance(clusterName, *v)
		if err != nil {
			return errors.Wrapf(err, "Failed to describe container instance '%v'.", v)
		}
		ec2InstanceIDs = append(ec2InstanceIDs, containerInstance.Ec2InstanceId)
	}

	err = ec2Wrapper.TerminateInstances(ec2InstanceIDs)
	if err != nil {
		return errors.Wrapf(err, "Failed to terminate container instances '%v'.", ec2InstanceIDs)
	}

	return nil
}

func stopAllTasks(ecsWrapper wrappers.ECSWrapper, clusterName string) error {
	taskARNList, err := ecsWrapper.ListTasks(clusterName, nil)
	if err != nil {
		return err
	}
	for _, t := range taskARNList {
		err = ecsWrapper.StopTask(clusterName, *t)
		if err != nil {
			return err
		}
	}
	return nil
}

func createInstanceProfile(iamWrapper wrappers.IAMWrapper) error {
	assumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"Service": "ec2.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}
		]
	}`

	err := iamWrapper.GetRole(&roleName)
	if err != nil {
		if awsErr, ok := errors.Cause(err).(awserr.Error); ok && awsErr.Code() == "NoSuchEntity" {
			err = iamWrapper.CreateRole(&roleName, &assumeRolePolicy)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = iamWrapper.GetInstanceProfile(&instanceProfileName)
	if err != nil {
		if awsErr, ok := errors.Cause(err).(awserr.Error); ok && awsErr.Code() == "NoSuchEntity" {
			err = iamWrapper.CreateInstanceProfile(&instanceProfileName)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = iamWrapper.AttachRolePolicy(&policyARN, &roleName)
	if err != nil {
		return err
	}

	err = iamWrapper.AddRoleToInstanceProfile(&roleName, &instanceProfileName)
	if err != nil {
		if awsErr, ok := errors.Cause(err).(awserr.Error); ok {
			if awsErr.Code() != "EntityAlreadyExists" && awsErr.Code() != "LimitExceeded" {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func deleteInstanceProfile(iamWrapper wrappers.IAMWrapper) error {
	err := iamWrapper.DetachRolePolicy(&policyARN, &roleName)
	if err != nil {
		return err
	}

	err = iamWrapper.RemoveRoleFromInstanceProfile(&roleName, &instanceProfileName)
	if err != nil {
		return err
	}

	err = iamWrapper.DeleteRole(&roleName)
	if err != nil {
		return err
	}

	err = iamWrapper.DeleteInstanceProfile(&instanceProfileName)
	if err != nil {
		return err
	}

	return nil
}

func deleteEnvironment(environment string, edsWrapper wrappers.EDSWrapper) {
	err := edsWrapper.DeleteEnvironment(&environment)
	if err != nil {
		T.Errorf("Was not able to delete environment %v: %v", environment, err.Error())
		return
	}
}

func createDeployment(deploymentToken string, ctx context.Context, edsWrapper wrappers.EDSWrapper) *models.Deployment {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	deployment, err := edsWrapper.CreateDeployment(opCtx, &environment, &deploymentToken)
	if err != nil {
		T.Errorf(err.Error())
		return nil
	}
	deploymentID = *deployment.ID
	return deployment
}

func waitForDeploymentToComplete(seconds int, edsWrapper wrappers.EDSWrapper) {
	ok, err := doSomething(time.Duration(seconds)*time.Second, 1*time.Second, func() (bool, error) {
		deployment, err := edsWrapper.GetDeployment(aws.String(environment), aws.String(deploymentID))
		if err != nil {
			return false, errors.Wrapf(err, "Error calling GetDeployment for environment %s and deployment %s", environment, deploymentID)
		}

		return strings.ToLower(taskRunning) == strings.ToLower(aws.StringValue(deployment.Status)), nil
	})

	if err != nil {
		T.Errorf(err.Error())
		return
	}

	if !ok {
		T.Errorf("Expecting the deployment status to be %v", taskRunning)
		return
	}
}

func doSomething(ttl time.Duration, tickTime time.Duration, fn func() (bool, error)) (bool, error) {
	timeout := time.After(ttl)
	tick := time.Tick(tickTime)
	for {
		select {
		case <-timeout:
			return false, errors.New("timed out")
		case <-tick:
			ok, err := fn()
			if err != nil {
				return false, err
			} else if ok {
				return true, nil
			}
		}
	}
}

func filterTasksByStatusRunning(cluster *string, taskARNs []*string, ecsWrapper wrappers.ECSWrapper) []*string {
	runningTasks := make([]*string, 0, len(taskARNs))
	if len(taskARNs) == 0 {
		return runningTasks
	}
	tasks, err := ecsWrapper.DescribeTasks(cluster, taskARNs)
	if err != nil {
		T.Errorf(err.Error())
		return runningTasks
	}
	if len(tasks) > len(taskARNs) {
		T.Errorf("Expecting at most %d tasks to be returned", len(taskARNs))
		return runningTasks
	}
	for _, t := range tasks {
		if aws.StringValue(t.LastStatus) == taskRunning {
			runningTasks = append(runningTasks, t.TaskArn)
		}
	}
	return runningTasks
}
