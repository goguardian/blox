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

package types

import "github.com/goguardian/blox/daemon-scheduler/pkg/environment/types"

// ValidateAndUpdateEnvironment - implementation should validate the environment
// being passed into the function and return an updated environment based on the
// action being performed (like creating environment, updating deployments, etc.)
type ValidateAndUpdateEnvironment func(*types.Environment) (*types.Environment, error)
