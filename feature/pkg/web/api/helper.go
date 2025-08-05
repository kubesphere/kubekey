/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"errors"
	"net/http"
	"runtime"
	"strings"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

// sanitizer is a string replacer that converts HTML special characters to their escaped equivalents
// to prevent XSS attacks in error messages. It replaces &, <, and > with their HTML entities.
var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

// HandleInternalError handles internal server errors (500) by logging the error and sending an appropriate response
func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusInternalServerError, response, req, err)
}

// HandleBadRequest handles bad request errors (400) by logging the error and sending an appropriate response
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusBadRequest, response, req, err)
}

// HandleNotFound handles not found errors (404) by logging the error and sending an appropriate response
func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusNotFound, response, req, err)
}

// HandleForbidden handles forbidden errors (403) by logging the error and sending an appropriate response
func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusForbidden, response, req, err)
}

// HandleUnauthorized handles unauthorized errors (401) by logging the error and sending an appropriate response
func HandleUnauthorized(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusUnauthorized, response, req, err)
}

// HandleTooManyRequests handles rate limiting errors (429) by logging the error and sending an appropriate response
func HandleTooManyRequests(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusTooManyRequests, response, req, err)
}

// HandleConflict handles conflict errors (409) by logging the error and sending an appropriate response
func HandleConflict(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusConflict, response, req, err)
}

// HandleError handles various types of errors by determining the appropriate status code
// and sending a response with the corresponding HTTP status code
func HandleError(response *restful.Response, req *restful.Request, err error) {
	var statusCode int
	var apiStatus apierrors.APIStatus
	var serviceError restful.ServiceError

	switch {
	case errors.As(err, &apiStatus):
		statusCode = int(apiStatus.Status().Code)
	case errors.As(err, &serviceError):
		statusCode = serviceError.Code
	default:
		statusCode = http.StatusInternalServerError
	}
	handle(statusCode, response, req, err)
}

// handle is an internal helper function that logs the error and sends an HTTP error response
// with the sanitized error message and specified status code
func handle(statusCode int, response *restful.Response, _ *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(2)
	klog.Errorf("%s:%d %v", fn, line, err)
	http.Error(response, sanitizer.Replace(err.Error()), statusCode)
}
