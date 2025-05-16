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

package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

const (
	// Prometheus API default timeout
	defaultPrometheusTimeout = 10 * time.Second
)

var _ Connector = &PrometheusConnector{}

// PrometheusConnector implements Connector interface for Prometheus connections
type PrometheusConnector struct {
	url       string
	username  string
	password  string
	token     string
	headers   map[string]string
	timeout   time.Duration
	client    *http.Client
	connected bool
}

// newPrometheusConnector creates a new PrometheusConnector
func newPrometheusConnector(vars map[string]any) *PrometheusConnector {
	pc := &PrometheusConnector{
		headers: make(map[string]string),
		timeout: defaultPrometheusTimeout,
	}

	// 修正变量名以避免导入遮蔽
	promURL, err := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorURL)
	if err != nil {
		klog.V(4).InfoS("get connector host failed use current hostname", "error", err)
	}

	pc.url = promURL

	username, err := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorUserName)
	if err != nil {
		klog.V(4).InfoS("get connector username failed use current username", "error", err)
	}

	pc.username = username

	password, err := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPassword)
	if err != nil {
		klog.V(4).InfoS("get connector password failed use current password", "error", err)
	}

	pc.password = password

	token, err := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorToken)
	if err != nil {
		klog.V(4).InfoS("get connector token failed use current token", "error", err)
	}

	pc.token = token

	prometheusVars, ok := vars["connector"].(map[string]any)
	if !ok {
		klog.V(4).InfoS("connector configuration is not a map")
		return nil
	}
	// Get custom headers from connector variables
	if headers, ok := prometheusVars["headers"].(map[string]any); ok {
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				pc.headers[k] = strVal
			}
		}
	}

	// Get timeout from connector variables
	if timeoutStr, ok := prometheusVars["timeout"].(string); ok {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			pc.timeout = timeout
		}
	}

	return pc
}

// Init initializes the Prometheus connection
func (pc *PrometheusConnector) Init(ctx context.Context) error {
	// Ensure URL is properly formatted
	if pc.url == "" {
		return errors.New("prometheus URL is required")
	}

	// Parse and normalize the URL
	parsedURL, err := url.Parse(pc.url)
	if err != nil {
		return errors.Wrapf(err, "invalid prometheus URL: %s", pc.url)
	}

	// If scheme is missing, default to http
	if parsedURL.Scheme == "" {
		klog.V(4).InfoS("No scheme specified in Prometheus URL, defaulting to HTTP", "url", pc.url)
		parsedURL.Scheme = "http"
	} else if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.Errorf("unsupported URL scheme: %s, only http and https are supported", parsedURL.Scheme)
	}

	// Ensure path ends with "/"
	if !strings.HasSuffix(parsedURL.Path, "/") {
		parsedURL.Path += "/"
	}

	// Update the URL with normalized version
	pc.url = parsedURL.String()
	klog.V(4).InfoS("Initializing Prometheus connector", "url", pc.url)

	// Create HTTP client with timeout
	pc.client = &http.Client{
		Timeout: pc.timeout,
	}

	// Test connection by sending a simple query
	testURL, err := url.Parse(pc.url + "api/v1/status/buildinfo")
	if err != nil {
		return errors.Wrap(err, "failed to parse URL for connection test")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL.String(), http.NoBody)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// Add auth headers if provided
	pc.addAuthHeaders(req)

	klog.V(4).InfoS("Testing connection to Prometheus server")
	resp, err := pc.client.Do(req)
	if err != nil {
		klog.ErrorS(err, "Failed to connect to Prometheus server", "url", pc.url)
		return errors.Wrap(err, "failed to connect to Prometheus")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		klog.ErrorS(err, "Prometheus server returned error status",
			"statusCode", resp.StatusCode,
			"response", string(bodyBytes))
		return errors.Errorf("failed to connect to Prometheus: status code %d", resp.StatusCode)
	}

	klog.V(2).InfoS("Successfully connected to Prometheus server", "url", pc.url)
	pc.connected = true
	return nil
}

// Close closes the Prometheus connection
func (pc *PrometheusConnector) Close(ctx context.Context) error {
	// HTTP client doesn't need explicit closing
	pc.connected = false
	return nil
}

// PutFile is not supported for Prometheus connector
func (pc *PrometheusConnector) PutFile(ctx context.Context, src []byte, dst string, mode fs.FileMode) error {
	return errors.New("putFile operation is not supported for Prometheus connector")
}

// FetchFile is not supported for Prometheus connector
func (pc *PrometheusConnector) FetchFile(ctx context.Context, src string, dst io.Writer) error {
	return errors.New("fetchFile operation is not supported for Prometheus connector")
}

// ExecuteCommand executes a PromQL query
// For Prometheus connector, the command is interpreted as a PromQL query
func (pc *PrometheusConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	if !pc.connected {
		return nil, errors.New("prometheus connector is not initialized, call Init() first")
	}

	// Parse the command
	queryParams := parseCommand(cmd)
	queryString := queryParams["query"]
	if queryString == "" {
		return nil, errors.New("query parameter is required for Prometheus queries")
	}

	klog.V(4).InfoS("Executing Prometheus query", "query", queryString)

	// Build query URL
	apiURL, err := url.Parse(pc.url + "api/v1/query")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse URL with base: %s", pc.url)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("query", queryString)

	// Add time parameter if provided
	if timeParam := queryParams["time"]; timeParam != "" {
		klog.V(4).InfoS("Using custom time parameter", "time", timeParam)
		params.Add("time", timeParam)
	}

	apiURL.RawQuery = params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}

	// Add auth headers
	pc.addAuthHeaders(req)

	// Execute request
	klog.V(4).InfoS("Sending request to Prometheus", "url", req.URL.String())
	resp, err := pc.client.Do(req)
	if err != nil {
		klog.ErrorS(err, "Failed to execute Prometheus query", "query", queryString)
		return nil, errors.Wrap(err, "failed to execute prometheus query")
	}
	defer resp.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.ErrorS(err, "Failed to read response body")
		return nil, errors.Wrap(err, "failed to read response body")
	}

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		klog.ErrorS(err, "Prometheus query failed",
			"statusCode", resp.StatusCode,
			"response", string(bodyBytes),
			"query", queryString)
		return nil, errors.Errorf("prometheus query failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Format the response based on the format parameter
	format := queryParams["format"]
	if format != "" {
		klog.V(4).InfoS("Formatting response", "format", format)
		return pc.formatResponse(bodyBytes, format)
	}

	// Default to prettified JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
		klog.V(4).InfoS("Failed to prettify JSON response, returning raw response")
		// If prettifying fails, return the original response
		return bodyBytes, nil
	}
	klog.V(4).InfoS("Prometheus query executed successfully")
	return prettyJSON.Bytes(), nil
}

// addAuthHeaders adds authentication headers to the request
func (pc *PrometheusConnector) addAuthHeaders(req *http.Request) {
	// Add basic auth if username and password are provided
	if pc.username != "" && pc.password != "" {
		req.SetBasicAuth(pc.username, pc.password)
		klog.V(4).InfoS("Added basic auth header to request", "username", pc.username)
	}

	// Add token auth if token is provided
	if pc.token != "" {
		req.Header.Set("Authorization", "Bearer "+pc.token)
		klog.V(4).InfoS("Added token auth header to request")
	}

	// Add content type for API requests
	req.Header.Set("Accept", "application/json")

	// Add custom headers
	for k, v := range pc.headers {
		req.Header.Set(k, v)
		klog.V(4).InfoS("Added custom header to request", "key", k)
	}
}

// parseCommand parses the command string into query parameters
// The command can be either:
// - A simple PromQL query string
// - A JSON string with parameters (query, time, format, etc.)
func parseCommand(cmd string) map[string]string {
	result := make(map[string]string)

	// Check if the command is empty
	if cmd == "" {
		klog.V(4).InfoS("Empty command passed to Prometheus connector")
		return result
	}

	// Check if the command is a JSON string
	var jsonCmd map[string]any
	if err := json.Unmarshal([]byte(cmd), &jsonCmd); err == nil {
		// Extract parameters from JSON
		for k, v := range jsonCmd {
			if strVal, ok := v.(string); ok {
				result[k] = strVal
			} else if v != nil {
				// Try to convert non-string values to string
				result[k] = fmt.Sprintf("%v", v)
			}
		}
		klog.V(4).InfoS("Parsed JSON command", "params", result)
		return result
	}

	// If not JSON, treat the entire command as a query
	result["query"] = cmd
	klog.V(4).InfoS("Using command as raw query", "query", cmd)
	return result
}

// formatResponse formats the response according to the specified format
func (pc *PrometheusConnector) formatResponse(bodyBytes []byte, format string) ([]byte, error) {
	// Parse the response
	var response map[string]any
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return bodyBytes, nil
	}

	switch format {
	case "raw":
		// Return the original response
		return bodyBytes, nil
	case "value":
		// Extract single value if possible
		return pc.extractSimpleValue(response)
	case "table":
		// Format as table
		return pc.formatAsTable(response)
	default:
		// Default to prettified JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
			return bodyBytes, nil
		}
		return prettyJSON.Bytes(), nil
	}
}

// extractSimpleValue attempts to extract a simple value from the Prometheus response
func (pc *PrometheusConnector) extractSimpleValue(response map[string]any) ([]byte, error) {
	// 验证响应格式
	if err := validatePrometheusResponse(response); err != nil {
		return nil, err
	}

	data, ok := response["data"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid response format: data field missing")
	}

	resultType, ok := data["resultType"].(string)
	if !ok {
		return nil, errors.New("invalid response format: resultType field missing")
	}

	result, ok := data["result"]
	if !ok {
		return nil, errors.New("invalid response format: result field missing")
	}

	// 根据不同的结果类型处理
	switch resultType {
	case "vector":
		return extractVectorValue(result)
	case "scalar":
		return extractScalarValue(result)
	case "string":
		return extractStringValue(result)
	case "matrix":
		return extractMatrixValue(result)
	default:
		return nil, errors.Errorf("unsupported result type: %s", resultType)
	}
}

// validatePrometheusResponse 验证Prometheus响应的基本结构
func validatePrometheusResponse(response map[string]any) error {
	if status, ok := response["status"].(string); !ok || status != "success" {
		return errors.New("prometheus query failed")
	}
	return nil
}

// extractVectorValue 从向量结果中提取值
func extractVectorValue(result any) ([]byte, error) {
	samples, ok := result.([]any)
	if !ok || len(samples) == 0 {
		return []byte("No data"), nil
	}

	sample, ok := samples[0].(map[string]any)
	if !ok {
		return nil, errors.New("invalid response format: sample format invalid")
	}

	value, ok := sample["value"].([]any)
	if !ok || len(value) < 2 {
		return nil, errors.New("invalid response format: value format invalid")
	}

	return []byte(fmt.Sprintf("%v", value[1])), nil
}

// extractScalarValue 从标量结果中提取值
func extractScalarValue(result any) ([]byte, error) {
	value, ok := result.([]any)
	if !ok || len(value) < 2 {
		return nil, errors.New("invalid response format: scalar format invalid")
	}

	return []byte(fmt.Sprintf("%v", value[1])), nil
}

// extractStringValue 从字符串结果中提取值
func extractStringValue(result any) ([]byte, error) {
	value, ok := result.([]any)
	if !ok || len(value) < 2 {
		return nil, errors.New("invalid response format: string format invalid")
	}

	return []byte(fmt.Sprintf("%v", value[1])), nil
}

// extractMatrixValue 从矩阵结果中提取值
func extractMatrixValue(result any) ([]byte, error) {
	matrixData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return []byte(fmt.Sprintf("%v", result)), nil
	}
	return matrixData, nil
}

// formatAsTable 重构以降低认知复杂度
func (pc *PrometheusConnector) formatAsTable(response map[string]any) ([]byte, error) {
	// 验证响应格式并获取结果集
	result, err := getValidVectorResult(response)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return []byte("No data"), nil
	}

	// 构建表格
	return buildTableFromResult(result)
}

// getValidVectorResult 验证响应并获取vector类型的结果集
func getValidVectorResult(response map[string]any) ([]any, error) {
	if status, ok := response["status"].(string); !ok || status != "success" {
		return nil, errors.New("prometheus query failed")
	}

	data, ok := response["data"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid response format: data field missing")
	}

	resultType, ok := data["resultType"].(string)
	if !ok {
		return nil, errors.New("invalid response format: resultType field missing")
	}

	if resultType != "vector" {
		return nil, errors.Errorf("table format only supported for vector results, got: %s", resultType)
	}

	result, ok := data["result"].([]any)
	if !ok {
		return nil, errors.New("invalid response format: result field missing or not an array")
	}

	return result, nil
}

// buildTableFromResult 从结果集构建表格
func buildTableFromResult(result []any) ([]byte, error) {
	var builder strings.Builder

	// 表格标题
	if _, err := builder.WriteString("METRIC\tVALUE\tTIMESTAMP\n"); err != nil {
		return nil, err
	}

	// 表格行
	for _, item := range result {
		sample, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// 获取指标名称
		metric := getMetricName(sample)

		// 添加值和时间戳
		if err := addValueAndTimestamp(&builder, sample, metric); err != nil {
			return nil, err
		}
	}

	return []byte(builder.String()), nil
}

// getMetricName 提取指标名称
func getMetricName(sample map[string]any) string {
	metric := "undefined"
	m, ok := sample["metric"].(map[string]any)
	if !ok {
		return metric
	}

	// 提取指标名称
	parts := []string{}
	for k, v := range m {
		if k == "__name__" {
			metric = fmt.Sprintf("%v", v)
		} else if strVal, ok := v.(string); ok {
			parts = append(parts, fmt.Sprintf("%s=%q", k, strVal))
		}
	}

	// 如果有标签，在指标名称中包含它们
	if len(parts) > 0 {
		metric = fmt.Sprintf("%s{%s}", metric, strings.Join(parts, ", "))
	}

	return metric
}

// addValueAndTimestamp 添加值和时间戳到表格行
func addValueAndTimestamp(builder *strings.Builder, sample map[string]any, metric string) error {
	value, ok := sample["value"].([]any)
	if !ok || len(value) < 2 {
		return nil // 跳过无效数据
	}

	timestamp := ""
	if ts, ok := value[0].(float64); ok {
		timestamp = fmt.Sprintf("%.0f", ts)
	}

	if _, err := fmt.Fprintf(builder, "%s\t%v\t%s\n", metric, value[1], timestamp); err != nil {
		return err
	}

	return nil
}

// GetServerInfo returns information about the Prometheus server
// This is useful for checking server version, uptime, and other details
func (pc *PrometheusConnector) GetServerInfo(ctx context.Context) (map[string]any, error) {
	if !pc.connected {
		return nil, errors.New("prometheus connector is not initialized, call Init() first")
	}

	klog.V(4).InfoS("Getting Prometheus server information")

	// Build query URL for server info
	infoURL, err := url.Parse(pc.url + "api/v1/status/buildinfo")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse URL for server info")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL.String(), http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request for server info")
	}

	// Add auth headers
	pc.addAuthHeaders(req)

	// Execute request
	resp, err := pc.client.Do(req)
	if err != nil {
		klog.ErrorS(err, "Failed to get Prometheus server info")
		return nil, errors.Wrap(err, "failed to get prometheus server info")
	}
	defer resp.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.ErrorS(err, "Failed to read server info response body")
		return nil, errors.Wrap(err, "failed to read server info response body")
	}

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		klog.ErrorS(err, "Prometheus server info request failed",
			"statusCode", resp.StatusCode,
			"response", string(bodyBytes))
		return nil, errors.Errorf("prometheus server info request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		klog.ErrorS(err, "Failed to parse server info response")
		return nil, errors.Wrap(err, "failed to parse server info response")
	}

	klog.V(4).InfoS("Successfully retrieved Prometheus server information")
	return result, nil
}
