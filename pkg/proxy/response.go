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
	"bytes"
	"io"
	"net/http"
)

// ResponseWriter is an interface for writing HTTP responses.
// It abstracts the difference between buffered (normal) and streaming (watch) responses.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Handler
	// Finalize prepares the response for return.
	// - Buffered responses: sets the response body
	// - Streaming responses: no-op (pipe remains open)
	Finalize()
	// Response returns the underlying http.Response
	Response() *http.Response
}

// handlerChainFunc wraps a handler with request info
type handlerChainFunc func(handler http.Handler) http.Handler

// bufferedResponseWriter writes responses to an internal buffer.
// Used for normal (non-watch) HTTP requests.
type bufferedResponseWriter struct {
	resp           *http.Response
	buf            bytes.Buffer
	header         http.Header
	request        *http.Request
	handler        http.Handler
	handlerChainFn handlerChainFunc
}

// NewBufferedResponseWriter creates a new buffered response writer
func NewBufferedResponseWriter(req *http.Request, handler http.Handler, handlerChainFn handlerChainFunc) *bufferedResponseWriter {
	return &bufferedResponseWriter{
		resp: &http.Response{
			Proto:  "local",
			Header: make(http.Header),
		},
		header:         make(http.Header),
		request:        req,
		handler:        handler,
		handlerChainFn: handlerChainFn,
	}
}

// Header returns the response headers
func (b *bufferedResponseWriter) Header() http.Header {
	return b.header
}

// Write writes data to the buffer
func (b *bufferedResponseWriter) Write(bs []byte) (int, error) {
	return b.buf.Write(bs)
}

// Flush implements http.Flusher for buffered responses.
// It flushes the current buffer to the response body.
func (b *bufferedResponseWriter) Flush() {
	if b.buf.Len() > 0 {
		b.resp.Body = io.NopCloser(bytes.NewReader(b.buf.Bytes()))
		b.buf.Reset()
	}
}

// WriteHeader sets the HTTP status code
func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	b.resp.StatusCode = statusCode
}

// Finalize sets the response body from the buffer and defaults status code to 200
func (b *bufferedResponseWriter) Finalize() {
	// Copy all headers
	for k, v := range b.header {
		for _, vv := range v {
			b.resp.Header.Add(k, vv)
		}
	}
	b.resp.Body = io.NopCloser(bytes.NewReader(b.buf.Bytes()))
	if b.resp.StatusCode == 0 {
		b.resp.StatusCode = http.StatusOK
	}
}

// Response returns the underlying http.Response
func (b *bufferedResponseWriter) Response() *http.Response {
	return b.resp
}

// ServeHTTP implements http.Handler for buffered responses.
// It wraps the handler with handlerChainFn and executes synchronously.
func (b *bufferedResponseWriter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	wrappedHandler := b.handlerChainFn(b.handler)
	wrappedHandler.ServeHTTP(rw, r)
	b.Finalize()
}

// streamingResponseWriter writes responses to a pipe for streaming.
// Used for watch requests.
type streamingResponseWriter struct {
	resp           *http.Response
	pw             *io.PipeWriter
	header         http.Header
	request        *http.Request
	handler        http.Handler
	handlerChainFn handlerChainFunc
}

// NewStreamingResponseWriter creates a new streaming response writer
func NewStreamingResponseWriter(req *http.Request, handler http.Handler, handlerChainFn handlerChainFunc) *streamingResponseWriter {
	pr, pw := io.Pipe()

	resp := &http.Response{
		StatusCode:    http.StatusOK,
		Proto:         "local",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          pr,
		ContentLength: -1, // chunked transfer encoding
	}

	// Set standard watch response headers
	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set("Transfer-Encoding", "chunked")

	rw := &streamingResponseWriter{
		resp:           resp,
		pw:             pw,
		header:         resp.Header,
		request:        req,
		handler:        handler,
		handlerChainFn: handlerChainFn,
	}

	// Start context cancellation listener
	// Close the pipe when the request context is cancelled
	go func() {
		<-req.Context().Done()
		pw.CloseWithError(req.Context().Err())
	}()

	return rw
}

// Header returns the response headers
func (s *streamingResponseWriter) Header() http.Header {
	return s.header
}

// Write writes data to the pipe for streaming
func (s *streamingResponseWriter) Write(bs []byte) (int, error) {
	return s.pw.Write(bs)
}

// Flush implements http.Flusher for streaming responses.
// For pipe-based streaming, data is immediately available to the reader,
// so this is a no-op but kept for interface compliance.
func (s *streamingResponseWriter) Flush() {
	// No-op for pipe-based streaming; data is immediately available
}

// WriteHeader sets the HTTP status code
func (s *streamingResponseWriter) WriteHeader(statusCode int) {
	s.resp.StatusCode = statusCode
}

// Finalize is a no-op for streaming responses
// The pipe remains open for continuous streaming
func (s *streamingResponseWriter) Finalize() {
	// No-op; body is already set to the pipe reader
}

// Response returns the underlying http.Response
func (s *streamingResponseWriter) Response() *http.Response {
	return s.resp
}

// ServeHTTP implements http.Handler for streaming responses.
// It wraps the handler with handlerChainFn and executes in a goroutine.
func (s *streamingResponseWriter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	go func() {
		wrappedHandler := s.handlerChainFn(s.handler)
		wrappedHandler.ServeHTTP(s, s.request)
		s.Finalize()
	}()
}

// Close closes the pipe writer
func (s *streamingResponseWriter) Close() {
	s.pw.Close()
}

// isWatchRequest checks if the request is a watch request.
// Watch request characteristics:
// 1. Query parameter watch=true
// 2. Path ends with /watch
func isWatchRequest(req *http.Request) bool {
	if req.URL.Query().Get("watch") == "true" {
		return true
	}
	path := req.URL.Path
	return len(path) > 7 && path[len(path)-7:] == "/watch"
}

// NewResponseWriter creates the appropriate ResponseWriter based on request type
// - Watch requests: create streaming response writer
// - Normal requests: create buffered response writer
func NewResponseWriter(req *http.Request, handler http.Handler, handlerChainFn handlerChainFunc) ResponseWriter {
	if isWatchRequest(req) {
		return NewStreamingResponseWriter(req, handler, handlerChainFn)
	}
	return NewBufferedResponseWriter(req, handler, handlerChainFn)
}
