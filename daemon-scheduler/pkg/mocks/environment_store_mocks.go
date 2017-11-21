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
// Source: pkg/store/environment.go

package mocks

import (
	context "context"
	types "github.com/goguardian/blox/daemon-scheduler/pkg/store/types"
	types0 "github.com/goguardian/blox/daemon-scheduler/pkg/environment/types"
	gomock "github.com/golang/mock/gomock"
)

// Mock of EnvironmentStore interface
type MockEnvironmentStore struct {
	ctrl     *gomock.Controller
	recorder *_MockEnvironmentStoreRecorder
}

// Recorder for MockEnvironmentStore (not exported)
type _MockEnvironmentStoreRecorder struct {
	mock *MockEnvironmentStore
}

func NewMockEnvironmentStore(ctrl *gomock.Controller) *MockEnvironmentStore {
	mock := &MockEnvironmentStore{ctrl: ctrl}
	mock.recorder = &_MockEnvironmentStoreRecorder{mock}
	return mock
}

func (_m *MockEnvironmentStore) EXPECT() *_MockEnvironmentStoreRecorder {
	return _m.recorder
}

func (_m *MockEnvironmentStore) PutEnvironment(ctx context.Context, name string, validateAndUpdateEnv types.ValidateAndUpdateEnvironment) error {
	ret := _m.ctrl.Call(_m, "PutEnvironment", ctx, name, validateAndUpdateEnv)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockEnvironmentStoreRecorder) PutEnvironment(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PutEnvironment", arg0, arg1, arg2)
}

func (_m *MockEnvironmentStore) GetEnvironment(ctx context.Context, name string) (*types0.Environment, error) {
	ret := _m.ctrl.Call(_m, "GetEnvironment", ctx, name)
	ret0, _ := ret[0].(*types0.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockEnvironmentStoreRecorder) GetEnvironment(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetEnvironment", arg0, arg1)
}

func (_m *MockEnvironmentStore) DeleteEnvironment(ctx context.Context, name string) error {
	ret := _m.ctrl.Call(_m, "DeleteEnvironment", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockEnvironmentStoreRecorder) DeleteEnvironment(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteEnvironment", arg0, arg1)
}

func (_m *MockEnvironmentStore) ListEnvironments(ctx context.Context) ([]types0.Environment, error) {
	ret := _m.ctrl.Call(_m, "ListEnvironments", ctx)
	ret0, _ := ret[0].([]types0.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockEnvironmentStoreRecorder) ListEnvironments(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ListEnvironments", arg0)
}
