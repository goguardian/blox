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
	"net/http"

	"github.com/goguardian/blox/daemon-scheduler/pkg/types"
	log "github.com/cihub/seelog"
	"github.com/pkg/errors"
)

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	log.Errorf("Unexpected error : %+v", err)
	http.Error(w, "Server Error", http.StatusInternalServerError)
}

func writeBadRequestError(w http.ResponseWriter, errMsg string) {
	http.Error(w, errMsg, http.StatusBadRequest)
}

func handleBackendError(w http.ResponseWriter, err error) {
	_, ok := errors.Cause(err).(types.BadRequestError)
	if ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, ok = errors.Cause(err).(types.NotFoundError)
	if ok {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeInternalServerError(w, err)
}
