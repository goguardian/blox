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

// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// TaskContainer task container
// swagger:model TaskContainer
type TaskContainer struct {

	// container a r n
	// Required: true
	ContainerARN *string `json:"containerARN"`

	// exit code
	ExitCode int64 `json:"exitCode,omitempty"`

	// last status
	// Required: true
	LastStatus *string `json:"lastStatus"`

	// name
	// Required: true
	Name *string `json:"name"`

	// network bindings
	NetworkBindings TaskContainerNetworkBindings `json:"networkBindings"`

	// reason
	Reason string `json:"reason,omitempty"`
}

// Validate validates this task container
func (m *TaskContainer) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateContainerARN(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateLastStatus(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *TaskContainer) validateContainerARN(formats strfmt.Registry) error {

	if err := validate.Required("containerARN", "body", m.ContainerARN); err != nil {
		return err
	}

	return nil
}

func (m *TaskContainer) validateLastStatus(formats strfmt.Registry) error {

	if err := validate.Required("lastStatus", "body", m.LastStatus); err != nil {
		return err
	}

	return nil
}

func (m *TaskContainer) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *TaskContainer) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TaskContainer) UnmarshalBinary(b []byte) error {
	var res TaskContainer
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
