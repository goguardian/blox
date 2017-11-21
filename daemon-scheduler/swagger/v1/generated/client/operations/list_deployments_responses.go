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

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/goguardian/blox/daemon-scheduler/swagger/v1/generated/models"
)

// ListDeploymentsReader is a Reader for the ListDeployments structure.
type ListDeploymentsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListDeploymentsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewListDeploymentsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewListDeploymentsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewListDeploymentsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewListDeploymentsOK creates a ListDeploymentsOK with default headers values
func NewListDeploymentsOK() *ListDeploymentsOK {
	return &ListDeploymentsOK{}
}

/*ListDeploymentsOK handles this case with default header values.

OK
*/
type ListDeploymentsOK struct {
	Payload *models.Deployments
}

func (o *ListDeploymentsOK) Error() string {
	return fmt.Sprintf("[GET /environments/{name}/deployments][%d] listDeploymentsOK  %+v", 200, o.Payload)
}

func (o *ListDeploymentsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Deployments)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListDeploymentsBadRequest creates a ListDeploymentsBadRequest with default headers values
func NewListDeploymentsBadRequest() *ListDeploymentsBadRequest {
	return &ListDeploymentsBadRequest{}
}

/*ListDeploymentsBadRequest handles this case with default header values.

Bad Request
*/
type ListDeploymentsBadRequest struct {
	Payload string
}

func (o *ListDeploymentsBadRequest) Error() string {
	return fmt.Sprintf("[GET /environments/{name}/deployments][%d] listDeploymentsBadRequest  %+v", 400, o.Payload)
}

func (o *ListDeploymentsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListDeploymentsNotFound creates a ListDeploymentsNotFound with default headers values
func NewListDeploymentsNotFound() *ListDeploymentsNotFound {
	return &ListDeploymentsNotFound{}
}

/*ListDeploymentsNotFound handles this case with default header values.

Not Found
*/
type ListDeploymentsNotFound struct {
	Payload string
}

func (o *ListDeploymentsNotFound) Error() string {
	return fmt.Sprintf("[GET /environments/{name}/deployments][%d] listDeploymentsNotFound  %+v", 404, o.Payload)
}

func (o *ListDeploymentsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
