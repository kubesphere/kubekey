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
	"github.com/lithammer/dedent"
	"text/template"
)

var HaproxyManifest = template.Must(template.New("haproxy.yaml").Parse(
	dedent.Dedent(`
apiVersion: v1
kind: Pod
metadata:
  name: haproxy
  namespace: kube-system
  labels:
    addonmanager.kubernetes.io/mode: Reconcile
    k8s-app: kube-haproxy
  annotations:
    cfg-checksum: "{{ .Checksum }}"
spec:
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet
  nodeSelector:
    kubernetes.io/os: linux
  priorityClassName: system-node-critical
  containers:
  - name: haproxy
    image: {{ .HaproxyImage }}
    imagePullPolicy: Always
    resources:
      requests:
        cpu: 25m
        memory: 32M
    livenessProbe:
      httpGet:
        path: /healthz
        port: {{ .HealthCheckPort }}
    readinessProbe:
      httpGet:
        path: /healthz
        port: {{ .HealthCheckPort }}
    volumeMounts:
    - mountPath: /usr/local/etc/haproxy/
      name: etc-haproxy
      readOnly: true
  volumes:
  - name: etc-haproxy
    hostPath:
      path: /etc/kubekey/haproxy
`)))
