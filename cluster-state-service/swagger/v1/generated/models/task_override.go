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

// TaskOverride task override
// swagger:model TaskOverride
type TaskOverride struct {

	// container overrides
	// Required: true
	ContainerOverrides TaskOverrideContainerOverrides `json:"containerOverrides"`

	// task role arn
	TaskRoleArn string `json:"taskRoleArn,omitempty"`
}

// Validate validates this task override
func (m *TaskOverride) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateContainerOverrides(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *TaskOverride) validateContainerOverrides(formats strfmt.Registry) error {

	if err := validate.Required("containerOverrides", "body", m.ContainerOverrides); err != nil {
		return err
	}

	if err := m.ContainerOverrides.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("containerOverrides")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *TaskOverride) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TaskOverride) UnmarshalBinary(b []byte) error {
	var res TaskOverride
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
