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

import "github.com/goguardian/blox/daemon-scheduler/swagger/v1/generated/models"

type Exception struct {
	exceptionType string
	exceptionMsg  string
}

var (
	taskDefinition  string
	cluster         string
	clusterARN      string
	environment     string
	asg             string
	deploymentID    string
	deploymentToken string
	environmentList []*models.Environment
	err             error
	exception       Exception
)

var deploymentIDs = make(map[string]*models.Deployment)
