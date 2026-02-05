/*
Copyright 2024 The KubeSphere Authors.

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

package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsWatchRequest(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "watch query parameter true",
			url:      "/api/v1/pods?watch=true",
			expected: true,
		},
		{
			name:     "watch query parameter false",
			url:      "/api/v1/pods?watch=false",
			expected: false,
		},
		{
			name:     "watch query parameter empty",
			url:      "/api/v1/pods?watch=",
			expected: false,
		},
		{
			name:     "path ends with /watchx (last 7 chars match)",
			url:      "/api/v1/pods/watch",
			expected: false,
		},
		{
			name:     "normal path without watch",
			url:      "/api/v1/pods",
			expected: false,
		},
		{
			name:     "root path",
			url:      "/",
			expected: false,
		},
		{
			name:     "apis path",
			url:      "/apis/kubekey.kubesphere.io/v1alpha1/tasks",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			result := isWatchRequest(req)
			if result != tc.expected {
				t.Errorf("isWatchRequest(%q) = %v, expected %v", tc.url, result, tc.expected)
			}
		})
	}
}

func TestIsWatchRequest_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "watch in path but not at end",
			url:      "/api/v1/pods/watcher",
			expected: false,
		},
		{
			name:     "similar suffix",
			url:      "/api/v1/pods/atch",
			expected: false,
		},
		{
			name:     "short path /watch (length 6, less than 7)",
			url:      "/watch",
			expected: false,
		},
		{
			name:     "very short path",
			url:      "/",
			expected: false,
		},
		{
			name:     "exactly 7 chars path /watchx",
			url:      "/watchx",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			result := isWatchRequest(req)
			if result != tc.expected {
				t.Errorf("isWatchRequest(%q) = %v, expected %v", tc.url, result, tc.expected)
			}
		})
	}
}

func TestBufferedResponseWriter_Header(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewBufferedResponseWriter(req, handler, nil)

	header := writer.Header()
	if header == nil {
		t.Error("Header() should not return nil")
	}

	// Verify it's a valid http.Header
	header.Set("Content-Type", "application/json")
	if writer.header.Get("Content-Type") != "application/json" {
		t.Error("Header() should allow setting headers")
	}
}

func TestBufferedResponseWriter_Write(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewBufferedResponseWriter(req, handler, nil)

	data := []byte("test data")
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Write() unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write() returned %d, expected %d", n, len(data))
	}
}

func TestBufferedResponseWriter_WriteHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewBufferedResponseWriter(req, handler, nil)

	writer.WriteHeader(http.StatusOK)

	if writer.resp.StatusCode != http.StatusOK {
		t.Errorf("WriteHeader() set status to %d, expected %d", writer.resp.StatusCode, http.StatusOK)
	}
}

func TestBufferedResponseWriter_Finalize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler response"))
	})

	writer := NewBufferedResponseWriter(req, handler, nil)
	writer.Write([]byte("test data"))

	writer.Finalize()

	if writer.resp.StatusCode == 0 {
		t.Error("Finalize() should set default status code")
	}

	if writer.resp.StatusCode != http.StatusOK {
		t.Errorf("Finalize() set status to %d, expected %d", writer.resp.StatusCode, http.StatusOK)
	}
}

func TestBufferedResponseWriter_Response(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewBufferedResponseWriter(req, handler, nil)

	resp := writer.Response()
	if resp == nil {
		t.Error("Response() should not return nil")
	}

	if resp != writer.resp {
		t.Error("Response() should return the underlying response")
	}
}

func TestStreamingResponseWriter_Header(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods/watch", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewStreamingResponseWriter(req, handler, nil)

	header := writer.Header()
	if header == nil {
		t.Error("Header() should not return nil")
	}

	// Verify watch headers are set
	if header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be set to application/json for streaming")
	}
	if header.Get("Transfer-Encoding") != "chunked" {
		t.Error("Transfer-Encoding should be set to chunked for streaming")
	}
}

func TestStreamingResponseWriter_WriteHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods/watch", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewStreamingResponseWriter(req, handler, nil)

	writer.WriteHeader(http.StatusOK)

	if writer.resp.StatusCode != http.StatusOK {
		t.Errorf("WriteHeader() set status to %d, expected %d", writer.resp.StatusCode, http.StatusOK)
	}
}

func TestStreamingResponseWriter_Finalize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods/watch", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewStreamingResponseWriter(req, handler, nil)

	// Finalize should be a no-op for streaming
	writer.Finalize()

	// Should not panic, that's the test
}

func TestStreamingResponseWriter_Response(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods/watch", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewStreamingResponseWriter(req, handler, nil)

	resp := writer.Response()
	if resp == nil {
		t.Error("Response() should not return nil")
	}

	if resp != writer.resp {
		t.Error("Response() should return the underlying response")
	}
}

func TestNewResponseWriter_WatchRequest(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		isWatch bool
	}{
		{
			name:    "normal request",
			url:     "/api/v1/pods",
			isWatch: false,
		},
		{
			name:    "watch query parameter",
			url:     "/api/v1/pods?watch=true",
			isWatch: true,
		},
		{
			name:    "request without watch",
			url:     "/api/v1/pods/watch",
			isWatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			writer := NewResponseWriter(req, handler, nil)

			if tc.isWatch {
				if _, ok := writer.(*streamingResponseWriter); !ok {
					t.Error("expected streamingResponseWriter for watch request")
				}
			} else {
				if _, ok := writer.(*bufferedResponseWriter); !ok {
					t.Error("expected bufferedResponseWriter for normal request")
				}
			}
		})
	}
}

func TestBufferedResponseWriter_WriteMultiple(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewBufferedResponseWriter(req, handler, nil)

	// Multiple writes
	writer.Write([]byte("first"))
	writer.Write([]byte("second"))
	writer.Write([]byte("third"))

	expected := "firstsecondthird"
	if writer.buf.String() != expected {
		t.Errorf("buf.String() = %q, expected %q", writer.buf.String(), expected)
	}
}

func TestStreamingResponseWriter_Flush(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pods/watch", nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	writer := NewStreamingResponseWriter(req, handler, nil)

	// Flush should not panic for streaming
	writer.Flush()
}
