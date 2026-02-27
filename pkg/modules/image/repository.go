// Package image provides functionality for managing container image operations in KubeKey.
// This file contains repository-related code for handling both remote OCI registries and local
// directory storage of container images.
//
// The repository module supports:
// - Remote registry connections with authentication
// - Local directory-based image storage (OCI layout)
// - HTTP transport layer for local image operations
package image

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// repoCache caches *remote.Repository instances per local directory path.
// This improves performance by reusing repository connections for the same local directory.
var repoCache sync.Map

// newRepository creates a remote repository handle for image operations.
// It determines whether the URL refers to a local directory or remote registry
// and creates the appropriate repository instance with authentication.
func newRepository(url, img string, auths []imageAuth) (*remote.Repository, error) {
	if strings.HasPrefix(url, "local://") {
		// Source is local directory
		return newLocalRepository(img, strings.TrimPrefix(url, "local://"))
	} else if strings.HasPrefix(url, "oci://") {
		// Source is remote registry
		return newRemoteRepository(img, auths)
	} else {
		return nil, errors.New("invalid source image reference")
	}

}

// newRemoteRepository creates a remote repository connection to an OCI registry.
// It configures the repository with authentication credentials, TLS settings,
// and HTTP options based on the provided auth configuration.
func newRemoteRepository(reference string, auths []imageAuth) (*remote.Repository, error) {
	repository, err := remote.NewRepository(reference)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get remote repository %q", reference)
	}
	// Find matching auth for this host
	var matchedAuth *imageAuth
	for i := range auths {
		if strings.Split(auths[i].Registry, "/")[0] == reference {
			matchedAuth = &auths[i]
			break
		}
	}

	// Get TLS and HTTP settings from matched auth, use defaults if not found
	skipTLSVerify := false
	plainHTTP := false
	if matchedAuth != nil {
		if matchedAuth.SkipTLSVerify != nil {
			skipTLSVerify = *matchedAuth.SkipTLSVerify
		}
		if matchedAuth.PlainHTTP != nil {
			plainHTTP = *matchedAuth.PlainHTTP
		}
	}

	repository.Client = &auth.Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipTLSVerify,
				},
			},
		},
		Cache: auth.NewCache(),
		Credential: func(_ context.Context, hostport string) (auth.Credential, error) {
			if matchedAuth != nil {
				return auth.Credential{
					Username: matchedAuth.Username,
					Password: matchedAuth.Password,
				}, nil
			}
			return auth.Credential{}, nil
		},
	}
	repository.PlainHTTP = plainHTTP
	return repository, nil
}

// newLocalRepository creates a repository handle for local directory-based image storage.
// It implements an OCI layout-compatible storage backend using a custom HTTP transport.
// The repository is cached per local directory to improve performance.
func newLocalRepository(reference, localDir string) (*remote.Repository, error) {
	// Try to get from cache
	if cached, ok := repoCache.Load(localDir); ok {
		repo := cached.(*remote.Repository)
		// Update reference in case it changed
		ref, err := registry.ParseReference(reference)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse reference %q", reference)
		}
		repo.Reference = ref
		return repo, nil
	}

	// Create new repository
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse reference %q", reference)
	}

	repo := &remote.Repository{
		Reference: ref,
		Client:    &http.Client{Transport: &imageTransport{baseDir: localDir}},
	}

	// Store in cache
	repoCache.Store(localDir, repo)

	return repo, nil
}

var responseNotFound = &http.Response{Proto: "Local", StatusCode: http.StatusNotFound}
var responseNotAllowed = &http.Response{Proto: "Local", StatusCode: http.StatusMethodNotAllowed}
var responseServerError = &http.Response{Proto: "Local", StatusCode: http.StatusInternalServerError}
var responseCreated = &http.Response{Proto: "Local", StatusCode: http.StatusCreated}
var responseOK = &http.Response{Proto: "Local", StatusCode: http.StatusOK}

// const domain = "internal"
const apiPrefix = "/v2/"

type imageTransport struct {
	baseDir string
}

// RoundTrip handles HTTP requests for local directory-based image storage.
// It routes requests to appropriate handlers based on the HTTP method (HEAD, GET, POST, PUT).
func (i imageTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var resp *http.Response

	switch request.Method {
	case http.MethodHead:
		resp = i.head(request)
	case http.MethodPost:
		resp = i.post(request)
	case http.MethodPut:
		resp = i.put(request)
	case http.MethodGet:
		resp = i.get(request)
	default:
		resp = responseNotAllowed
	}

	if resp != nil {
		resp.Request = request
	}

	return resp, nil
}

// head handles HTTP HEAD requests for checking the existence of blobs and manifests
// in the local directory storage.
func (i imageTransport) head(request *http.Request) *http.Response {
	if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
		filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		return responseOK
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "manifests") { // manifests
		filename := filepath.Join(i.baseDir, request.Host, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			klog.V(4).ErrorS(err, "failed to unmarshal file", "filename", filename)

			return responseServerError
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			klog.V(4).ErrorS(nil, "unknown mediaType", "filename", filename)

			return responseServerError
		}

		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{mediaType},
			},
			ContentLength: int64(len(file)),
		}
	}

	return responseNotAllowed
}

// post handles HTTP POST requests for initiating blob uploads.
// Returns an accepted response with the upload location.
func (i imageTransport) post(request *http.Request) *http.Response {
	if strings.HasSuffix(request.URL.Path, "/uploads/") {
		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusAccepted,
			Header: http.Header{
				"Location": []string{filepath.Dir(request.URL.Path)},
			},
			Request: request,
		}
	}

	return responseNotAllowed
}

// put handles HTTP PUT requests for uploading blobs and manifests to local directory storage.
// It creates files in the blobs directory or manifests directory based on the request path.
func (i imageTransport) put(request *http.Request) *http.Response {
	if strings.HasSuffix(request.URL.Path, "/uploads") { // blobs
		defer request.Body.Close()

		filename := filepath.Join(i.baseDir, "blobs", request.URL.Query().Get("digest"))
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			klog.V(4).ErrorS(err, "failed to create dir", "dir", filepath.Dir(filename))

			return responseServerError
		}

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to create file", "filename", filename)
			return responseServerError
		}

		defer func() {
			if err = file.Sync(); err != nil {
				klog.V(4).ErrorS(err, "failed to sync file", "filename", filename)
			}
			if err = file.Close(); err != nil {
				klog.V(4).ErrorS(err, "failed to close file", "filename", filename)
			}
		}()

		if _, err = io.Copy(file, request.Body); err != nil {
			klog.V(4).ErrorS(err, "failed to write file", "filename", filename)

			return responseServerError
		}

		return responseCreated
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "/manifests") { // manifests
		body, err := io.ReadAll(request.Body)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read request")

			return responseServerError
		}
		defer request.Body.Close()

		filename := filepath.Join(i.baseDir, request.Host, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			klog.V(4).ErrorS(err, "failed to create dir", "dir", filepath.Dir(filename))

			return responseServerError
		}

		if err := os.WriteFile(filename, body, os.ModePerm); err != nil {
			klog.V(4).ErrorS(err, "failed to write file", "filename", filename)

			return responseServerError
		}

		return responseCreated
	}

	return responseNotAllowed
}

// get handles HTTP GET requests for retrieving blobs and manifests from local directory storage.
func (i imageTransport) get(request *http.Request) *http.Response {
	if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
		filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		file, err := os.Open(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		fStat, err := file.Stat()
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		return &http.Response{
			Proto:         "Local",
			StatusCode:    http.StatusOK,
			ContentLength: fStat.Size(),
			Body:          file,
		}
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "manifests") { // manifests
		filename := filepath.Join(i.baseDir, request.Host, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			klog.V(4).ErrorS(err, "failed to unmarshal file data", "filename", filename)

			return responseServerError
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			klog.V(4).ErrorS(nil, "unknown mediaType", "filename", filename)

			return responseServerError
		}

		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{mediaType},
			},
			ContentLength: int64(len(file)),
			Body:          io.NopCloser(bytes.NewReader(file)),
		}
	}

	return responseNotAllowed
}
