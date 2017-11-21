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

package v1

import (
	"github.com/goguardian/blox/cluster-state-service/handler/store"
)

type APIs struct {
	TaskApis              TaskAPIs
	ContainerInstanceApis ContainerInstanceAPIs
}

func NewAPIs(stores store.Stores) APIs {
	return APIs{
		TaskApis:              NewTaskAPIs(stores.TaskStore),
		ContainerInstanceApis: NewContainerInstanceAPIs(stores.ContainerInstanceStore),
	}
}
