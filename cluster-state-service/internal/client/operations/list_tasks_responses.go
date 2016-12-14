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

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/blox/blox/cluster-state-service/internal/models"
)

// ListTasksReader is a Reader for the ListTasks structure.
type ListTasksReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListTasksReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewListTasksOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewListTasksBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 500:
		result := NewListTasksInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewListTasksOK creates a ListTasksOK with default headers values
func NewListTasksOK() *ListTasksOK {
	return &ListTasksOK{}
}

/*ListTasksOK handles this case with default header values.

List tasks - success
*/
type ListTasksOK struct {
	Payload *models.Tasks
}

func (o *ListTasksOK) Error() string {
	return fmt.Sprintf("[GET /tasks][%d] listTasksOK  %+v", 200, o.Payload)
}

func (o *ListTasksOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Tasks)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListTasksBadRequest creates a ListTasksBadRequest with default headers values
func NewListTasksBadRequest() *ListTasksBadRequest {
	return &ListTasksBadRequest{}
}

/*ListTasksBadRequest handles this case with default header values.

List tasks - bad input
*/
type ListTasksBadRequest struct {
	Payload string
}

func (o *ListTasksBadRequest) Error() string {
	return fmt.Sprintf("[GET /tasks][%d] listTasksBadRequest  %+v", 400, o.Payload)
}

func (o *ListTasksBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListTasksInternalServerError creates a ListTasksInternalServerError with default headers values
func NewListTasksInternalServerError() *ListTasksInternalServerError {
	return &ListTasksInternalServerError{}
}

/*ListTasksInternalServerError handles this case with default header values.

List tasks - unexpected error
*/
type ListTasksInternalServerError struct {
	Payload string
}

func (o *ListTasksInternalServerError) Error() string {
	return fmt.Sprintf("[GET /tasks][%d] listTasksInternalServerError  %+v", 500, o.Payload)
}

func (o *ListTasksInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
