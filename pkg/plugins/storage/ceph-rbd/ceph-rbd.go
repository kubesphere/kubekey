/*
Copyright 2020 The KubeSphere Authors.

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

package ceph_rbd

import (
	"encoding/base64"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"strings"
	"text/template"
)

var RBDProvisionerTempl = template.Must(template.New("rbd-provisioner").Parse(
	dedent.Dedent(`---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: rbd-provisioner
  namespace: kube-system
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "update", "patch"]
  - apiGroups: [""]
    resources: ["services"]
    resourceNames: ["kube-dns","coredns"]
    verbs: ["list", "get"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "create", "delete"]
  - apiGroups: ["policy"]
    resourceNames: ["rbd-provisioner"]
    resources: ["podsecuritypolicies"]
    verbs: ["use"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rbd-provisioner
subjects:
  - kind: ServiceAccount
    name: rbd-provisioner
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: rbd-provisioner
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rbd-provisioner
  namespace: kube-system
  labels:
    app: rbd-provisioner
    version: v2.1.1-k8s1.11
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: rbd-provisioner
      version: v2.1.1-k8s1.11
  template:
    metadata:
      labels:
        app: rbd-provisioner
        version: v2.1.1-k8s1.11
    spec:
      priorityClassName: system-cluster-critical
      serviceAccount: rbd-provisioner
      containers:
        - name: rbd-provisioner
          image: {{ .RBDProvisionerImage }}
          imagePullPolicy: IfNotPresent
          env:
            - name: PROVISIONER_NAME
              value: ceph.com/rbd
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
          command:
            - "/usr/local/bin/rbd-provisioner"
          args:
            - "-id=${POD_NAME}"

---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: rbd-provisioner
  annotations:
    seccomp.security.alpha.kubernetes.io/defaultProfileName:  'docker/default'
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: 'docker/default'
    #apparmor.security.beta.kubernetes.io/defaultProfileName:  'runtime/default'
    #apparmor.security.beta.kubernetes.io/allowedProfileNames: 'runtime/default'
  labels:
    addonmanager.kubernetes.io/mode: Reconcile
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'MustRunAs'
    ranges:
      - min: 1
        max: 65535
  fsGroup:
    rule: 'MustRunAs'
    ranges:
      - min: 1
        max: 65535
  readOnlyRootFilesystem: false

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: rbd-provisioner
  namespace: kube-system
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get"]
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rbd-provisioner
  namespace: kube-system
subjects:
  - kind: ServiceAccount
    name: rbd-provisioner
    namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rbd-provisioner

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rbd-provisioner
  namespace: kube-system

---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{ .StorageClassName }}
  annotations:
    storageclass.kubesphere.io/supported_access_modes: '["ReadWriteOnce","ReadOnlyMany"]'
    storageclass.beta.kubernetes.io/is-default-class: "{{ if .IsDefaultClass }}true{{ else }}false{{ end }}"
provisioner: ceph.com/rbd
reclaimPolicy: Delete
parameters:
  monitors: {{ .Monitors}}
  adminId: {{ .AdminID }}
  adminSecretNamespace: kube-system
  adminSecretName: ceph-secret-admin
  pool: {{ .Pool }}
  userId: {{ .UserID }}
  userSecretNamespace: kube-system
  userSecretName: ceph-secret-user
  fsType: {{ .FsType }}
  imageFormat: "{{ .ImageFormat }}"
  {{ if eq .ImageFormat 2 }}imageFeatures: "{{ .ImageFeatures }}"{{ end }}
allowVolumeExpansion: true

---
kind: Secret
apiVersion: v1
metadata:
  name: ceph-secret-admin
  namespace: kube-system
type: kubernetes.io/rbd
data:
  secret: {{ .AdminSecret }}
---
kind: Secret
apiVersion: v1
metadata:
  name: ceph-secret-user
  namespace: kube-system
type: kubernetes.io/rbd
data:
  key: {{ .UserSecret }}

    `)))

func GenerateRBDProvisionerManifests(mgr *manager.Manager) (string, error) {
	return util.Render(RBDProvisionerTempl, util.Data{
		"RBDProvisionerImage": images.GetImage(mgr, "rbd-provisioner").ImageName(),
		"IsDefaultClass":      mgr.Cluster.Storage.CephRBD.IsDefaultClass,
		"StorageClassName":    mgr.Cluster.Storage.CephRBD.StorageClassName,
		"Monitors":            strings.Join(mgr.Cluster.Storage.CephRBD.Monitors, ","),
		"AdminID":             mgr.Cluster.Storage.CephRBD.AdminID,
		"Pool":                mgr.Cluster.Storage.CephRBD.Pool,
		"UserID":              mgr.Cluster.Storage.CephRBD.UserID,
		"FsType":              mgr.Cluster.Storage.CephRBD.FsType,
		"ImageFormat":         mgr.Cluster.Storage.CephRBD.ImageFormat,
		"ImageFeatures":       mgr.Cluster.Storage.CephRBD.ImageFeatures,
		"AdminSecret":         base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(mgr.Cluster.Storage.CephRBD.AdminSecret))),
		"UserSecret":          base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(mgr.Cluster.Storage.CephRBD.UserSecret))),
	})
}
