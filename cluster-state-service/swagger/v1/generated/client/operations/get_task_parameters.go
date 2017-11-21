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
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewGetTaskParams creates a new GetTaskParams object
// with the default values initialized.
func NewGetTaskParams() *GetTaskParams {
	var ()
	return &GetTaskParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetTaskParamsWithTimeout creates a new GetTaskParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetTaskParamsWithTimeout(timeout time.Duration) *GetTaskParams {
	var ()
	return &GetTaskParams{

		timeout: timeout,
	}
}

// NewGetTaskParamsWithContext creates a new GetTaskParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetTaskParamsWithContext(ctx context.Context) *GetTaskParams {
	var ()
	return &GetTaskParams{

		Context: ctx,
	}
}

// NewGetTaskParamsWithHTTPClient creates a new GetTaskParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetTaskParamsWithHTTPClient(client *http.Client) *GetTaskParams {
	var ()
	return &GetTaskParams{
		HTTPClient: client,
	}
}

/*GetTaskParams contains all the parameters to send to the API endpoint
for the get task operation typically these are written to a http.Request
*/
type GetTaskParams struct {

	/*Arn
	  ARN of the task to fetch

	*/
	Arn string
	/*Cluster
	  Cluster name of the task to fetch

	*/
	Cluster string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get task params
func (o *GetTaskParams) WithTimeout(timeout time.Duration) *GetTaskParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get task params
func (o *GetTaskParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get task params
func (o *GetTaskParams) WithContext(ctx context.Context) *GetTaskParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get task params
func (o *GetTaskParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get task params
func (o *GetTaskParams) WithHTTPClient(client *http.Client) *GetTaskParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get task params
func (o *GetTaskParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithArn adds the arn to the get task params
func (o *GetTaskParams) WithArn(arn string) *GetTaskParams {
	o.SetArn(arn)
	return o
}

// SetArn adds the arn to the get task params
func (o *GetTaskParams) SetArn(arn string) {
	o.Arn = arn
}

// WithCluster adds the cluster to the get task params
func (o *GetTaskParams) WithCluster(cluster string) *GetTaskParams {
	o.SetCluster(cluster)
	return o
}

// SetCluster adds the cluster to the get task params
func (o *GetTaskParams) SetCluster(cluster string) {
	o.Cluster = cluster
}

// WriteToRequest writes these params to a swagger request
func (o *GetTaskParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param arn
	if err := r.SetPathParam("arn", o.Arn); err != nil {
		return err
	}

	// path param cluster
	if err := r.SetPathParam("cluster", o.Cluster); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
