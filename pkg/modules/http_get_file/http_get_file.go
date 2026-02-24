/*
Copyright 2023 The KubeSphere Authors.

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

package http_get_file

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

const (
	// Default timeout for http API
	defaultHttpTimeout = 10 * time.Second
)

type httpArgs struct {
	url      string
	username string
	password string
	token    string
	headers  map[string]string
	timeout  time.Duration
	client   *http.Client
}

func (hc *httpArgs) Init(ctx context.Context) error {
	// Ensure URL is provided
	if hc.url == "" {
		return errors.New("http URL is required")
	}

	// Parse and normalize the URL
	parsedURL, err := url.Parse(hc.url)
	if err != nil {
		return errors.Wrapf(err, "invalid http URL: %s", hc.url)
	}

	// Default to http if scheme is missing
	if parsedURL.Scheme == "" {
		klog.V(4).InfoS("No scheme specified in http URL, defaulting to HTTP", "url", hc.url)
		parsedURL.Scheme = "http"
	} else if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.Errorf("unsupported URL scheme: %s, only http and https are supported", parsedURL.Scheme)
	}

	// Update the URL with normalized version
	hc.url = parsedURL.String()
	klog.V(4).InfoS("Initializing http connector", "url", hc.url)

	// Create HTTP client with timeout
	hc.client = &http.Client{
		Timeout: hc.timeout,
	}
	return nil
}

// FetchFile from http file server.  dst is the local writer.
func (hc *httpArgs) FetchFile(ctx context.Context, dst io.Writer) error {
	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hc.url, http.NoBody)
	if err != nil {
		return errors.Wrap(err, "failed to create request for server info")
	}

	// Add authentication headers
	hc.addAuthHeaders(req)

	// Execute request
	resp, err := hc.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to get http server info")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if dst != nil {
			_, err = io.Copy(dst, resp.Body)
		}
		return err
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	klog.ErrorS(err, "http server info request failed",
		"statusCode", resp.StatusCode,
		"response", string(bodyBytes))
	return errors.Errorf("http server info request failed with status %d", resp.StatusCode)
}

// addAuthHeaders adds authentication headers to the request
func (hc *httpArgs) addAuthHeaders(req *http.Request) {
	// Add basic auth if username and password are provided
	if hc.username != "" && hc.password != "" {
		req.SetBasicAuth(hc.username, hc.password)
		klog.V(4).InfoS("Added basic auth header to request", "username", hc.username)
	}

	// Add token auth if token is provided
	if hc.token != "" {
		req.Header.Set("Authorization", "Bearer "+hc.token)
		klog.V(4).InfoS("Added token auth header to request")
	}

	// Add custom headers
	for k, v := range hc.headers {
		req.Header.Set(k, v)
		klog.V(4).InfoS("Added custom header to request", "key", k)
	}
}

func newHttpArgs(ctx context.Context, args map[string]any, vars map[string]any) (httpArg *httpArgs, err error) {

	httpArg = &httpArgs{
		headers: make(map[string]string),
		timeout: defaultHttpTimeout,
	}

	// Retrieve http URL
	httpURL, err := variable.StringVar(vars, args, "url")
	if err != nil {
		klog.V(4).InfoS("Failed to get http url, using current url", "error", err)
	}
	httpArg.url = httpURL

	// Retrieve username
	username, err := variable.StringVar(vars, args, _const.VariableConnectorUserName)
	if err != nil {
		klog.V(4).InfoS("Failed to get http username, using current username", "error", err)
	}
	httpArg.username = username

	// Retrieve password
	password, err := variable.StringVar(vars, args, _const.VariableConnectorPassword)
	if err != nil {
		klog.V(4).InfoS("Failed to get http password, using current password", "error", err)
	}
	httpArg.password = password

	// Retrieve token
	token, err := variable.StringVar(vars, args, _const.VariableConnectorToken)
	if err != nil {
		klog.V(4).InfoS("Failed to get connector token, using current token", "error", err)
	}
	httpArg.token = token

	if headers, ok := args["headers"].(map[string]any); ok {
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				httpArg.headers[k] = strVal
			}
		}
	}
	if timeoutStr, ok := args["timeout"].(string); ok {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			httpArg.timeout = timeout
		}
	}
	return httpArg, httpArg.Init(ctx)
}

func ModuleHttpGetFile(ctx context.Context, opts internal.ExecOptions) (string, string, error) {
	// get host variable
	ha, err := opts.GetAllVariables()
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetHostVariable, err
	}

	// check args
	args := variable.Extension2Variables(opts.Args)
	// Parse module arguments.
	httpArg, err := newHttpArgs(ctx, args, ha)
	if err != nil {
		return internal.StdoutFailed, internal.StderrParseArgument, err
	}

	destParam, err := variable.StringVar(ha, args, "dest")
	if err != nil {
		return internal.StdoutFailed, "\"dest\" in args should be string", err
	}

	// fetch file
	parentDir := filepath.Dir(destParam)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
			return internal.StdoutFailed, "failed to create dest dir", err
		}
	}

	tmpFile, err := os.CreateTemp(parentDir, "*.tmp")
	if err != nil {
		return internal.StdoutFailed, "failed to create dest file", err
	}

	tmpFilePath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpFilePath)
	}()

	err = httpArg.FetchFile(ctx, tmpFile)
	errC := tmpFile.Close()
	if err != nil {
		return internal.StdoutFailed, "failed to get http file", err
	}
	if errC != nil {
		return internal.StdoutFailed, "failed to close tmp file", errC
	}

	err = os.Rename(tmpFilePath, destParam)
	if err != nil {
		return internal.StdoutFailed, "failed to rename file", err
	}

	return internal.StdoutSuccess, "", nil
}
