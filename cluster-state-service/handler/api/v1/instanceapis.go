// Copyright 2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"context"
	"encoding/json"
	"net/http"

	"github.com/blox/blox/cluster-state-service/handler/api/v1/models"
	"github.com/blox/blox/cluster-state-service/handler/regex"
	"github.com/blox/blox/cluster-state-service/handler/store"
	"github.com/blox/blox/cluster-state-service/handler/types"
	"github.com/gorilla/mux"
)

const (
	instanceARNKey     = "arn"
	instanceClusterKey = "cluster"

	instanceStatusFilter  = "status"
	instanceClusterFilter = "cluster"
)

var (
	// Using maps because arrays don't support easy lookup
	supportedInstanceFilters  = map[string]string{instanceStatusFilter: "", instanceClusterFilter: ""}
	supportedInstanceStatuses = map[string]string{"active": "", "inactive": ""}
)

// ContainerInstanceAPIs encapsulates the backend datastore with which the container instance APIs interact
type ContainerInstanceAPIs struct {
	instanceStore store.ContainerInstanceStore
}

// NewContainerInstanceAPIs initializes the ContainerInstanceAPIs struct
func NewContainerInstanceAPIs(instanceStore store.ContainerInstanceStore) ContainerInstanceAPIs {
	return ContainerInstanceAPIs{
		instanceStore: instanceStore,
	}
}

// GetInstance gets a container instance using the cluster name to which the instance belongs to and the instance ARN
func (instanceAPIs ContainerInstanceAPIs) GetInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceARN := vars[instanceARNKey]
	cluster := vars[instanceClusterKey]

	if len(instanceARN) == 0 || len(cluster) == 0 || !regex.IsInstanceARN(instanceARN) || !regex.IsClusterName(cluster) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(routingServerErrMsg)
		return
	}

	instance, err := instanceAPIs.instanceStore.GetContainerInstance(cluster, instanceARN)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalServerErrMsg)
		return
	}

	if instance == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(instanceNotFoundClientErrMsg)
		return
	}

	extInstance, err := ToContainerInstance(*instance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalServerErrMsg)
		return
	}

	w.Header().Set(contentTypeKey, contentTypeVal)
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(extInstance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(encodingServerErrMsg)
		return
	}
}

// ListInstances lists all container instances across all clusters after applying filters, if any
func (instanceAPIs ContainerInstanceAPIs) ListInstances(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if instanceAPIs.hasUnsupportedFilters(query) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(unsupportedFilterClientErrMsg)
		return
	}

	status := query.Get(instanceStatusFilter)
	cluster := query.Get(instanceClusterFilter)

	// TODO: Support filtering by both status and cluster
	if status != "" && cluster != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(unsupportedFilterCombinationClientErrMsg)
		return
	}

	if status != "" {
		if !instanceAPIs.isValidStatus(status) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(invalidStatusClientErrMsg)
			return
		}
	}

	if cluster != "" {
		if !regex.IsClusterARN(cluster) && !regex.IsClusterName(cluster) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(invalidClusterClientErrMsg)
			return
		}
	}

	var instances []types.ContainerInstance
	var err error

	switch {
	case len(status) != 0:
		instances, err = instanceAPIs.instanceStore.FilterContainerInstances(instanceStatusFilter, status)
	case len(cluster) != 0:
		instances, err = instanceAPIs.instanceStore.FilterContainerInstances(instanceClusterFilter, cluster)
	default:
		instances, err = instanceAPIs.instanceStore.ListContainerInstances()
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalServerErrMsg)
		return
	}

	w.Header().Set(contentTypeKey, contentTypeVal)
	w.WriteHeader(http.StatusOK)

	extInstanceItems := make([]*models.ContainerInstance, len(instances))
	for i := range instances {
		ins, err := ToContainerInstance(instances[i])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalServerErrMsg)
			return
		}
		extInstanceItems[i] = &ins
	}

	extInstances := models.ContainerInstances{
		Items: extInstanceItems,
	}

	err = json.NewEncoder(w).Encode(extInstances)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(encodingServerErrMsg)
		return
	}
}

// StreamInstances streams container instances that change (status, resources, etc.) across all clusters
func (instanceAPIs ContainerInstanceAPIs) StreamInstances(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	instanceRespChan, err := instanceAPIs.instanceStore.StreamContainerInstances(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalServerErrMsg)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalServerErrMsg)
		return
	}

	w.Header().Set(connectionKey, contentTypeVal)
	w.Header().Set(transferEncodingKey, transferEncodingVal)

	for instanceResp := range instanceRespChan {
		if instanceResp.Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalServerErrMsg)
			return
		}
		extInstance, err := ToContainerInstance(instanceResp.ContainerInstance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalServerErrMsg)
			return
		}
		err = json.NewEncoder(w).Encode(extInstance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(encodingServerErrMsg)
			return
		}
		flusher.Flush()
	}

	// TODO: Handle client-side termination (Ctrl+C) using w.(http.CloseNotifier).closeNotify()
}

func (instanceAPIs ContainerInstanceAPIs) isValidStatus(status string) bool {
	_, ok := supportedInstanceStatuses[status]
	if ok {
		return true
	}
	return false
}

func (instanceAPIs ContainerInstanceAPIs) hasUnsupportedFilters(filters map[string][]string) bool {
	if len(filters) > len(supportedInstanceFilters) {
		return true
	}

	for f := range filters {
		_, ok := supportedInstanceFilters[f]
		if !ok {
			return true
		}
	}
	return false
}
