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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/goguardian/blox/daemon-scheduler/pkg/deployment"
	"github.com/goguardian/blox/daemon-scheduler/pkg/environment"
	environmenttypes "github.com/goguardian/blox/daemon-scheduler/pkg/environment/types"
	"github.com/goguardian/blox/daemon-scheduler/pkg/facade"
	"github.com/goguardian/blox/daemon-scheduler/pkg/validate"
	"github.com/goguardian/blox/daemon-scheduler/swagger/v1/generated/models"
	log "github.com/cihub/seelog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	envNameKey      = "name"
	deploymentIDKey = "id"
	clusterFilter   = "cluster"

	// Client error messages
	unsupportedFilterError     = "At least one of the filters provided is not supported"
	invalidClusterError        = "Invalid cluster ARN or name"
	redundantFilterClientError = "At least one of the filters provided is specified multiple times"
)

var (
	// Using maps because arrays don't support easy lookup
	supportedEnvironmentFilters = map[string]string{clusterFilter: ""}
)

type API struct {
	environmentService environment.EnvironmentService
	deploymentService  deployment.DeploymentService
	ecs                facade.ECS
}

// NewAPI initializes the API struct
func NewAPI(e environment.EnvironmentService, d deployment.DeploymentService, ecs facade.ECS) API {
	return API{
		environmentService: e,
		deploymentService:  d,
		ecs:                ecs,
	}
}

// Ping is used to perform server health checks
func (api API) Ping(w http.ResponseWriter, r *http.Request) {
	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
}

// CreateEnvironment creates a new environment using details set in the request
func (api API) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var createEnvReq models.CreateEnvironmentRequest
	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &createEnvReq)

	err := createEnvReq.Validate(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ecsCluster, err := api.validateCluster(&createEnvReq.InstanceGroup.Cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ecsTaskDefinition, err := api.validateTaskDefinition(createEnvReq.TaskDefinition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	env, err := api.environmentService.CreateEnvironment(r.Context(), *createEnvReq.Name, *ecsTaskDefinition.TaskDefinitionArn, *ecsCluster.ClusterArn)
	if err != nil {
		handleBackendError(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(toEnvironmentModel(*env))
	if err != nil {
		log.Errorf("Error sending response for CreateEnvironment: %+v", err)
	}
}

// GetEnvironment gets an enironent by name
func (api API) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars[envNameKey]

	env, err := api.environmentService.GetEnvironment(r.Context(), name)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	if env == nil {
		http.Error(w, fmt.Sprintf("Environment %s does not exist", name), http.StatusNotFound)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(toEnvironmentModel(*env))
	if err != nil {
		log.Errorf("Error sending response for GetEnvironment: %+v", err)
	}
}

// ListEnvironments lists all environments across all clusters after applying filters, if any
func (api API) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if api.hasUnsupportedFilters(query) {
		writeBadRequestError(w, unsupportedFilterError)
		return
	}

	if api.hasRedundantFilters(query) {
		http.Error(w, redundantFilterClientError, http.StatusBadRequest)
		return
	}

	var envs []environmenttypes.Environment
	var err error

	cluster := query.Get(clusterFilter)
	if cluster != "" {
		if !validate.IsClusterARN(cluster) && !validate.IsClusterName(cluster) {
			writeBadRequestError(w, invalidClusterError)
			return
		}
		envs, err = api.environmentService.FilterEnvironments(r.Context(), clusterFilter, cluster)
	} else {
		envs, err = api.environmentService.ListEnvironments(r.Context())
	}

	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)
	envModels := []*models.Environment{}
	for _, envType := range envs {
		envModel := toEnvironmentModel(envType)
		envModels = append(envModels, &envModel)
	}
	environments := models.Environments{
		Items: envModels,
	}
	err = json.NewEncoder(w).Encode(environments)
	if err != nil {
		log.Errorf("Error sending response for ListEnvironments: %+v", err)
	}
}

// DeleteEnvironment deletes an environment by name
func (api API) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars[envNameKey]

	err := api.environmentService.DeleteEnvironment(r.Context(), name)
	if err != nil {
		handleBackendError(w, err)
		return
	}

	//TODO: return something when successful?
}

// CreateDeployment creates a deployment in an environment using details in the request
func (api API) CreateDeployment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars[envNameKey]
	token := vars[deploymentToken]

	d, err := api.deploymentService.CreateDeployment(r.Context(), name, token)
	if err != nil {
		handleBackendError(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	depModel := toDeploymentModel(&name, *d)
	err = json.NewEncoder(w).Encode(depModel)
	if err != nil {
		log.Errorf("Error sending response for CreateDeployment: %+v", err)
	}
}

// GetDeployment gets the deployment in an environment using the environment name and deployment ID
func (api API) GetDeployment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars[envNameKey]
	id := vars[deploymentIDKey]

	d, err := api.deploymentService.GetDeployment(r.Context(), name, id)
	if err != nil {
		handleBackendError(w, err)
		return
	}

	if d == nil {
		http.Error(w, fmt.Sprintf("Deployment %s does not exist for environment %s", id, name), http.StatusNotFound)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	depModel := toDeploymentModel(&name, *d)
	err = json.NewEncoder(w).Encode(depModel)
	if err != nil {
		log.Errorf("Error sending response for GetDeployment: %+v", err)
	}
}

// ListDeployments lists all deployments in an environment
func (api API) ListDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars[envNameKey]

	ds, err := api.deploymentService.ListDeploymentsSortedReverseChronologically(r.Context(), name)
	if err != nil {
		handleBackendError(w, err)
		return
	}

	setJSONContentType(w)
	w.WriteHeader(http.StatusOK)

	deploymentsModel := toDeploymentsModel(&name, ds)
	err = json.NewEncoder(w).Encode(deploymentsModel)
	if err != nil {
		log.Errorf("Error sending response for ListDeployments: %+v", err)
	}
}

func (api API) validateCluster(clusterName *string) (*ecs.Cluster, error) {
	cluster, err := api.ecs.DescribeCluster(clusterName)
	if err != nil {
		return nil, err
	}

	if *cluster.Status == "INACTIVE" {
		return nil, errors.New("Cluster is inactive")
	}

	return cluster, nil
}

func (api API) validateTaskDefinition(td *string) (*ecs.TaskDefinition, error) {
	taskDefinition, err := api.ecs.DescribeTaskDefinition(td)
	if err != nil {
		return nil, err
	}

	if *taskDefinition.Status == "INACTIVE" {
		return nil, errors.New("TaskDefinition is inactive")
	}

	return taskDefinition, nil
}

func (api API) hasUnsupportedFilters(filters map[string][]string) bool {
	if len(filters) > len(supportedEnvironmentFilters) {
		return true
	}

	for f := range filters {
		_, ok := supportedEnvironmentFilters[f]
		if !ok {
			return true
		}
	}
	return false
}

func (api API) hasRedundantFilters(filters map[string][]string) bool {
	for _, val := range filters {
		// Multiple values for a given filter implies that it has been specified multiple times
		if len(val) > 1 {
			return true
		}
	}
	return false
}
