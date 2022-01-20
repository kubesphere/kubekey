/*
 Copyright 2022 The KubeSphere Authors.

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

package artifact

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/config"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/logger"
)

// PushTracker returns a new InMemoryTracker which tracks the ref status
var PushTracker = docker.NewInMemoryTracker()

func GetResolver(ctx context.Context, auth registryAuth) remotes.Resolver {
	username := auth.Username
	secret := auth.Password

	options := docker.ResolverOptions{
		Tracker: PushTracker,
	}

	hostOptions := config.HostOptions{}
	hostOptions.Credentials = func(host string) (string, string, error) {
		return username, secret, nil
	}

	if auth.PlainHTTP {
		hostOptions.DefaultScheme = "http"
	}

	defaultConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	hostOptions.DefaultTLS = defaultConfig

	options.Hosts = config.ConfigureHosts(ctx, hostOptions)
	return docker.NewResolver(options)
}

type registryAuth struct {
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	PlainHTTP bool   `json:"plainHTTP,omitempty"`
}

func Auths(manifest *common.ArtifactManifest) (auths map[string]registryAuth) {
	if len(manifest.Spec.ManifestRegistry.Auths.Raw) == 0 {
		return
	}

	err := json.Unmarshal(manifest.Spec.ManifestRegistry.Auths.Raw, &auths)
	if err != nil {
		logger.Log.Fatal("Failed to Parse Registry Auths configuration: %v", manifest.Spec.ManifestRegistry.Auths.Raw)
		return
	}

	return
}
