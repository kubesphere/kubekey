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
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/config"
)

// PushTracker returns a new InMemoryTracker which tracks the ref status
var PushTracker = docker.NewInMemoryTracker()

func GetResolver(ctx context.Context) remotes.Resolver {
	options := docker.ResolverOptions{
		Tracker: PushTracker,
	}

	hostOptions := config.HostOptions{}

	defaultConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	hostOptions.DefaultTLS = defaultConfig

	options.Hosts = config.ConfigureHosts(ctx, hostOptions)
	return docker.NewResolver(options)
}
