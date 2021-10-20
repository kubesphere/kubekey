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
