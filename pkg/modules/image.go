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

package modules

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	imagev1 "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type imageArgs struct {
	pull *imagePullArgs
	push *imagePushArgs
}

type imagePullArgs struct {
	manifests     []string
	skipTLSVerify *bool
	username      string
	password      string
}

func (i imagePullArgs) pull(ctx context.Context) error {
	for _, img := range i.manifests {
		src, err := remote.NewRepository(img)
		if err != nil {
			return fmt.Errorf("failed to get remote image: %w", err)
		}
		src.Client = &auth.Client{
			Client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: *i.skipTLSVerify,
					},
				},
			},
			Cache: auth.NewCache(),
			Credential: auth.StaticCredential(src.Reference.Registry, auth.Credential{
				Username: i.username,
				Password: i.password,
			}),
		}

		dst, err := newLocalRepository(filepath.Join(domain, src.Reference.Repository)+":"+src.Reference.Reference,
			filepath.Join(_const.GetWorkDir(), _const.ArtifactDir, _const.ArtifactImagesDir))
		if err != nil {
			return fmt.Errorf("failed to get local image: %w", err)
		}

		if _, err = oras.Copy(ctx, src, src.Reference.Reference, dst, "", oras.DefaultCopyOptions); err != nil {
			return fmt.Errorf("failed to copy image: %w", err)
		}
	}

	return nil
}

type imagePushArgs struct {
	imagesDir     string
	skipTLSVerify *bool
	registry      string
	username      string
	password      string
	namespace     string
}

// push local dir images to remote registry
func (i imagePushArgs) push(ctx context.Context) error {
	manifests, err := findLocalImageManifests(i.imagesDir)
	klog.V(5).Info("manifests found", "manifests", manifests)
	if err != nil {
		return fmt.Errorf("failed to find local image manifests: %w", err)
	}

	for _, img := range manifests {
		src, err := newLocalRepository(filepath.Join(domain, img), i.imagesDir)
		if err != nil {
			return fmt.Errorf("failed to get local image: %w", err)
		}
		repo := src.Reference.Repository
		if i.namespace != "" {
			repo = filepath.Join(i.namespace, filepath.Base(repo))
		}

		dst, err := remote.NewRepository(filepath.Join(i.registry, repo) + ":" + src.Reference.Reference)
		if err != nil {
			return fmt.Errorf("failed to get remote repo: %w", err)
		}
		dst.Client = &auth.Client{
			Client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: *i.skipTLSVerify,
					},
				},
			},
			Cache: auth.NewCache(),
			Credential: auth.StaticCredential(i.registry, auth.Credential{
				Username: i.username,
				Password: i.password,
			}),
		}

		if _, err = oras.Copy(ctx, src, src.Reference.Reference, dst, "", oras.DefaultCopyOptions); err != nil {
			return fmt.Errorf("failed to copy image: %w", err)
		}
	}

	return nil
}

func newImageArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*imageArgs, error) {
	ia := &imageArgs{}
	// check args
	args := variable.Extension2Variables(raw)
	if pullArgs, ok := args["pull"]; ok {
		pull, ok := pullArgs.(map[string]any)
		if !ok {
			return nil, errors.New("\"pull\" should be map")
		}
		ipl := &imagePullArgs{}
		ipl.manifests, _ = variable.StringSliceVar(vars, pull, "manifests")
		ipl.username, _ = variable.StringVar(vars, pull, "username")
		ipl.password, _ = variable.StringVar(vars, pull, "password")
		ipl.skipTLSVerify, _ = variable.BoolVar(vars, pull, "skipTLSVerify")
		if ipl.skipTLSVerify == nil {
			ipl.skipTLSVerify = ptr.To(false)
		}
		// check args
		if len(ipl.manifests) == 0 {
			return nil, errors.New("\"pull.manifests\" is required")
		}
		ia.pull = ipl
	}
	// if namespace_override is not empty, it will override the image manifests namespace_override. (namespace maybe multi sub path)
	// push to private registry
	if pushArgs, ok := args["push"]; ok {
		push, ok := pushArgs.(map[string]any)
		if !ok {
			return nil, errors.New("\"push\" should be map")
		}

		ips := &imagePushArgs{}
		ips.registry, _ = variable.StringVar(vars, push, "registry")
		ips.username, _ = variable.StringVar(vars, push, "username")
		ips.password, _ = variable.StringVar(vars, push, "password")
		ips.namespace, _ = variable.StringVar(vars, push, "namespace_override")
		ips.imagesDir, _ = variable.StringVar(vars, push, "images_dir")
		ips.skipTLSVerify, _ = variable.BoolVar(vars, push, "skipTLSVerify")
		if ips.skipTLSVerify == nil {
			ips.skipTLSVerify = ptr.To(false)
		}
		// check args
		if ips.registry == "" {
			return nil, errors.New("\"push.registry\" is required")
		}
		if ips.imagesDir == "" {
			return nil, errors.New("\"push.images_dir\" is required")
		}
		ia.push = ips
	}

	return ia, nil
}

// ModuleImage deal "image" module
func ModuleImage(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}

	ia, err := newImageArgs(ctx, options.Args, ha)
	if err != nil {
		return "", err.Error()
	}

	// pull image manifests to local dir
	if ia.pull != nil {
		if err := ia.pull.pull(ctx); err != nil {
			return "", fmt.Sprintf("failed to pull image: %v", err)
		}
	}
	// push image to private registry
	if ia.push != nil {
		if err := ia.push.push(ctx); err != nil {
			return "", fmt.Sprintf("failed to push image: %v", err)
		}
	}

	return StdoutSuccess, ""
}

// findLocalImageManifests get image manifests with whole image's name.
func findLocalImageManifests(localDir string) ([]string, error) {
	if _, err := os.Stat(localDir); err != nil {
		klog.V(4).ErrorS(err, "failed to stat local directory", "image_dir", localDir)
		// images is not exist, skip
		return make([]string, 0), nil
	}

	var manifests []string
	if err := filepath.WalkDir(localDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == filepath.Join(localDir, "blobs") {
			return filepath.SkipDir
		}

		if d.IsDir() || filepath.Base(path) == "manifests" {
			return nil
		}

		file, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			// skip un except file (empty)
			klog.V(4).ErrorS(err, "unmarshal manifests file error", "file", path)

			return nil
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			return errors.New("invalid mediaType")
		}
		if mediaType == imagev1.MediaTypeImageIndex || mediaType == "application/vnd.docker.distribution.manifest.list.v2+json" {
			subpath, err := filepath.Rel(localDir, path)
			if err != nil {
				return err
			}
			// the last dir is manifests. should delete it
			manifests = append(manifests, filepath.Dir(filepath.Dir(subpath))+":"+filepath.Base(subpath))
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return manifests, nil
}

// newLocalRepository local dir images repository
func newLocalRepository(reference, localDir string) (*remote.Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	return &remote.Repository{
		Reference: ref,
		Client:    &http.Client{Transport: &imageTransport{baseDir: localDir}},
	}, nil
}

var responseNotFound = &http.Response{Proto: "Local", StatusCode: http.StatusNotFound}
var responseNotAllowed = &http.Response{Proto: "Local", StatusCode: http.StatusMethodNotAllowed}
var responseServerError = &http.Response{Proto: "Local", StatusCode: http.StatusInternalServerError}
var responseCreated = &http.Response{Proto: "Local", StatusCode: http.StatusCreated}
var responseOK = &http.Response{Proto: "Local", StatusCode: http.StatusOK}

const domain = "internal"
const apiPrefix = "/v2/"

type imageTransport struct {
	baseDir string
}

// RoundTrip deal http.Request in local dir images.
func (i imageTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	switch request.Method {
	case http.MethodHead: // check if file exist
		return i.head(request)
	case http.MethodPost:
		return i.post(request)
	case http.MethodPut:
		return i.put(request)
	case http.MethodGet:
		return i.get(request)
	default:
		return responseNotAllowed, nil
	}
}

// head method for http.MethodHead. check if file is exist in blobs dir or manifests dir
func (i imageTransport) head(request *http.Request) (*http.Response, error) {
	if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
		filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound, nil
		}

		return responseOK, nil
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "manifests") { // manifests
		filename := filepath.Join(i.baseDir, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound, nil
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError, nil
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			klog.V(4).ErrorS(err, "failed to unmarshal file", "filename", filename)

			return responseServerError, nil
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			klog.V(4).ErrorS(nil, "unknown mediaType", "filename", filename)

			return responseServerError, nil
		}

		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{mediaType},
			},
			ContentLength: int64(len(file)),
		}, nil
	}

	return responseNotAllowed, nil
}

// post method for http.MethodPost, accept request.
func (i imageTransport) post(request *http.Request) (*http.Response, error) {
	if strings.HasSuffix(request.URL.Path, "/uploads/") {
		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusAccepted,
			Header: http.Header{
				"Location": []string{filepath.Dir(request.URL.Path)},
			},
			Request: request,
		}, nil
	}

	return responseNotAllowed, nil
}

// put method for http.MethodPut, create file in blobs dir or manifests dir
func (i imageTransport) put(request *http.Request) (*http.Response, error) {
	if strings.HasSuffix(request.URL.Path, "/uploads") { // blobs
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return responseServerError, nil
		}
		defer request.Body.Close()

		filename := filepath.Join(i.baseDir, "blobs", request.URL.Query().Get("digest"))
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return responseServerError, nil
		}

		if err := os.WriteFile(filename, body, os.ModePerm); err != nil {
			return responseServerError, nil
		}

		return responseCreated, nil
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "/manifests") { // manifest
		filename := filepath.Join(i.baseDir, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return responseServerError, nil
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			return responseServerError, nil
		}
		defer request.Body.Close()

		if err := os.WriteFile(filename, body, os.ModePerm); err != nil {
			return responseServerError, nil
		}

		return responseCreated, nil
	}

	return responseNotAllowed, nil
}

// get method for http.MethodGet, get file in blobs dir or manifest dir
func (i imageTransport) get(request *http.Request) (*http.Response, error) {
	if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
		filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound, nil
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError, nil
		}

		return &http.Response{
			Proto:         "Local",
			StatusCode:    http.StatusOK,
			ContentLength: int64(len(file)),
			Body:          io.NopCloser(bytes.NewReader(file)),
		}, nil
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "manifests") { // manifests
		filename := filepath.Join(i.baseDir, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound, nil
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError, nil
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			return responseServerError, err
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			return responseServerError, nil
		}

		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{mediaType},
			},
			ContentLength: int64(len(file)),
			Body:          io.NopCloser(bytes.NewReader(file)),
		}, nil
	}

	return responseNotAllowed, nil
}
