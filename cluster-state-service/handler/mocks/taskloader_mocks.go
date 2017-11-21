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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/goguardian/blox/cluster-state-service/handler/reconcile/loader (interfaces: TaskLoader)

package mocks

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of TaskLoader interface
type MockTaskLoader struct {
	ctrl     *gomock.Controller
	recorder *_MockTaskLoaderRecorder
}

// Recorder for MockTaskLoader (not exported)
type _MockTaskLoaderRecorder struct {
	mock *MockTaskLoader
}

func NewMockTaskLoader(ctrl *gomock.Controller) *MockTaskLoader {
	mock := &MockTaskLoader{ctrl: ctrl}
	mock.recorder = &_MockTaskLoaderRecorder{mock}
	return mock
}

func (_m *MockTaskLoader) EXPECT() *_MockTaskLoaderRecorder {
	return _m.recorder
}

func (_m *MockTaskLoader) LoadTasks() error {
	ret := _m.ctrl.Call(_m, "LoadTasks")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockTaskLoaderRecorder) LoadTasks() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "LoadTasks")
}
