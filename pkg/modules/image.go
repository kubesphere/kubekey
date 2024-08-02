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
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	imagev1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleImage(ctx context.Context, options ExecOptions) (stdout string, stderr string) {
	// get host variable
	ha, err := options.Variable.Get(variable.GetAllVariable(options.Host))
	if err != nil {
		return "", fmt.Sprintf("failed to get host variable: %v", err)
	}

	// check args
	args := variable.Extension2Variables(options.Args)
	pullParam, _ := variable.StringSliceVar(ha.(map[string]any), args, "pull")
	// if namespace_override is not empty, it will override the image manifests namespace_override. (namespace maybe multi sub path)
	// push to private registry
	pushParam := args["push"]

	// pull image manifests to local dir
	for _, img := range pullParam {
		src, err := remote.NewRepository(img)
		if err != nil {
			return "", fmt.Sprintf("failed to get remote image: %v", err)
		}
		dst, err := NewLocalRepository(filepath.Join(domain, src.Reference.Repository) + ":" + src.Reference.Reference)
		if err != nil {
			return "", fmt.Sprintf("failed to get local image: %v", err)
		}
		if _, err = oras.Copy(context.Background(), src, src.Reference.Reference, dst, "", oras.DefaultCopyOptions); err != nil {
			return "", fmt.Sprintf("failed to copy image: %v", err)
		}
	}

	// push image to private registry
	if pushParam != nil {
		registry, _ := variable.StringVar(ha.(map[string]any), pushParam.(map[string]any), "registry")
		username, _ := variable.StringVar(ha.(map[string]any), pushParam.(map[string]any), "username")
		password, _ := variable.StringVar(ha.(map[string]any), pushParam.(map[string]any), "password")
		namespace, _ := variable.StringVar(ha.(map[string]any), pushParam.(map[string]any), "namespace_override")

		manifests, err := findLocalImageManifests(filepath.Join(_const.GetWorkDir(), "kubekey", "images"))
		if err != nil {
			return "", fmt.Sprintf("failed to find local image manifests: %v", err)
		}
		for _, img := range manifests {
			src, err := NewLocalRepository(filepath.Join(domain, img))
			if err != nil {
				return "", fmt.Sprintf("failed to get local image: %v", err)
			}
			repo := src.Reference.Repository
			if namespace != "" {
				repo = filepath.Join(namespace, filepath.Base(repo))
			}
			dst, err := remote.NewRepository(filepath.Join(registry, repo) + ":" + src.Reference.Reference)
			if err != nil {
				return "", fmt.Sprintf("failed to get local image: %v", err)
			}
			dst.Client = &auth.Client{
				Client: retry.DefaultClient,
				Cache:  auth.NewCache(),
				Credential: auth.StaticCredential(registry, auth.Credential{
					Username: username,
					Password: password,
				}),
			}

			if _, err = oras.Copy(context.Background(), src, src.Reference.Reference, dst, "", oras.DefaultCopyOptions); err != nil {
				return "", fmt.Sprintf("failed to copy image: %v", err)
			}
		}
	}

	return stdoutSuccess, ""
}

func findLocalImageManifests(localDir string) ([]string, error) {
	if _, err := os.Stat(localDir); err != nil {
		// images is not exist, skip
		klog.V(4).ErrorS(err, "failed to stat local directory")
		return nil, nil
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
			return err
		}
		if data["mediaType"].(string) == imagev1.MediaTypeImageIndex {
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

func NewLocalRepository(reference string) (*remote.Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	return &remote.Repository{
		Reference: ref,
		Client:    &http.Client{Transport: &imageTransport{baseDir: filepath.Join(_const.GetWorkDir(), "kubekey", "images")}},
	}, nil
}

var ResponseNotFound = &http.Response{Proto: "Local", StatusCode: http.StatusNotFound}
var ResponseNotAllowed = &http.Response{Proto: "Local", StatusCode: http.StatusMethodNotAllowed}
var ResponseServerError = &http.Response{Proto: "Local", StatusCode: http.StatusInternalServerError}
var ResponseCreated = &http.Response{Proto: "Local", StatusCode: http.StatusCreated}
var ResponseOK = &http.Response{Proto: "Local", StatusCode: http.StatusOK}

const domain = "internal"
const apiPrefix = "/v2/"

type imageTransport struct {
	baseDir string
}

func (i imageTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	switch request.Method {
	case http.MethodHead: // check if file exist
		if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
			filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
			if _, err := os.Stat(filename); err != nil {
				return ResponseNotFound, nil
			}
			return ResponseOK, nil
		} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "manifests") { // manifests
			filename := filepath.Join(i.baseDir, strings.TrimPrefix(request.URL.Path, apiPrefix))
			if _, err := os.Stat(filename); err != nil {
				return ResponseNotFound, nil
			}
			file, err := os.ReadFile(filename)
			if err != nil {
				return ResponseServerError, err
			}
			var data map[string]any
			if err := json.Unmarshal(file, &data); err != nil {
				return ResponseServerError, err
			}

			return &http.Response{
				Proto:      "Local",
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{data["mediaType"].(string)},
				},
				ContentLength: int64(len(file)),
			}, nil
		}
		return ResponseNotAllowed, nil
	case http.MethodPost:
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
		return ResponseNotAllowed, nil
	case http.MethodPut:
		if strings.HasSuffix(request.URL.Path, "/uploads") { // blobs
			body, err := io.ReadAll(request.Body)
			if err != nil {
				return ResponseServerError, nil
			}
			defer request.Body.Close()

			filename := filepath.Join(i.baseDir, "blobs", request.URL.Query().Get("digest"))
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				return ResponseServerError, nil
			}
			if err := os.WriteFile(filename, body, os.ModePerm); err != nil {
				return ResponseServerError, nil
			}
			return ResponseCreated, nil
		} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "/manifests") { // manifest
			filename := filepath.Join(i.baseDir, strings.TrimPrefix(request.URL.Path, apiPrefix))
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				return ResponseServerError, nil
			}
			body, err := io.ReadAll(request.Body)
			if err != nil {
				return ResponseServerError, nil
			}
			defer request.Body.Close()
			if err := os.WriteFile(filename, body, os.ModePerm); err != nil {
				return ResponseServerError, nil
			}
			return ResponseCreated, nil
		}

		return ResponseNotAllowed, nil
	case http.MethodGet:
		if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
			filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
			if _, err := os.Stat(filename); err != nil {
				return ResponseNotFound, nil
			}
			file, err := os.ReadFile(filename)
			if err != nil {
				return ResponseServerError, err
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
				return ResponseNotFound, nil
			}
			file, err := os.ReadFile(filename)
			if err != nil {
				return ResponseServerError, err
			}
			var data map[string]any
			if err := json.Unmarshal(file, &data); err != nil {
				return ResponseServerError, err
			}

			return &http.Response{
				Proto:      "Local",
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{data["mediaType"].(string)},
				},
				ContentLength: int64(len(file)),
				Body:          io.NopCloser(bytes.NewReader(file)),
			}, nil
		}
		return ResponseNotAllowed, nil
	default:
		return ResponseNotAllowed, nil
	}
}
