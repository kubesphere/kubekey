/*
 Copyright 2021 The KubeSphere Authors.

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

package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var ContainerdConfig = template.Must(template.New("config.toml").Parse(
	dedent.Dedent(`version = 2
root = "/var/lib/containerd"
state = "/run/containerd"

[grpc]
  address = "/run/containerd/containerd.sock"
  uid = 0
  gid = 0
  max_recv_message_size = 16777216
  max_send_message_size = 16777216

[ttrpc]
  address = ""
  uid = 0
  gid = 0

[debug]
  address = ""
  uid = 0
  gid = 0
  level = ""

[metrics]
  address = ""
  grpc_histogram = false

[cgroup]
  path = ""

[timeouts]
  "io.containerd.timeout.shim.cleanup" = "5s"
  "io.containerd.timeout.shim.load" = "5s"
  "io.containerd.timeout.shim.shutdown" = "3s"
  "io.containerd.timeout.task.state" = "2s"

[plugins]
  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
    runtime_type = "io.containerd.runc.v2"
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
      SystemdCgroup = true
  [plugins."io.containerd.grpc.v1.cri"]
    sandbox_image = "{{ .SandBoxImage }}"
    [plugins."io.containerd.grpc.v1.cri".cni]
      bin_dir = "/opt/cni/bin"
      conf_dir = "/etc/cni/net.d"
      max_conf_num = 1
      conf_template = ""
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        {{- if .Mirrors }}
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = [{{ .Mirrors }}, "https://registry-1.docker.io"]
        {{ else }}
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = ["https://registry-1.docker.io"]
        {{- end}}
        {{- range $value := .InsecureRegistries }}
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."{{$value}}"]
          endpoint = ["http://{{$value}}"]
        {{- end}}
      
        {{- if .Auths }}
        [plugins."io.containerd.grpc.v1.cri".registry.configs]
          {{- range $repo, $entry := .Auths }}
          [plugins."io.containerd.grpc.v1.cri".registry.configs."{{$repo}}".auth]
            username = "{{$entry.Username}}"
            password = "{{$entry.Password}}"
            [plugins."io.containerd.grpc.v1.cri".registry.configs."{{$repo}}".tls]
              ca_file = "{{$entry.CAFile}}"
              cert_file = "{{$entry.CertFile}}"
              key_file = "{{$entry.KeyFile}}"
              insecure_skip_verify = {{$entry.SkipTLSVerify}}
          {{- end}}
        {{- end}}
    `)))
