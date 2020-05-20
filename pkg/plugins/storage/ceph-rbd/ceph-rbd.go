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
kind: Secret
apiVersion: v1
metadata:
  name: ceph-rbd-{{ .AdminID }}-secret
  namespace: kube-system
type: "kubernetes.io/rbd" 
data:
  secret: {{ .AdminSecret }}

---
kind: Secret
apiVersion: v1
metadata:
  name: ceph-rbd-user-secret
  namespace: kube-system
type: "kubernetes.io/rbd" 
data:
  secret: {{ .UserSecret }}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rbd-provisioner
  namespace: kube-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: rbd-provisioner
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
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rbd-provisioner
subjects:
- kind: ServiceAccount
  name: rbd-provisioner
  namespace: kube-system

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-provisioner
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
    resources: ["endpoints"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
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
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rbd-provisioner
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: rbd-provisioner
    spec:
      containers:
      - name: rbd-provisioner
        image: {{ .RBDProvisionerImage }}
        env:
        - name: PROVISIONER_NAME
          value: ceph.com/rbd
      serviceAccount: rbd-provisioner

---
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: {{ .StorageClassName }}
  annotations:
    storageclass.kubesphere.io/supported_access_modes: '["ReadWriteOnce","ReadOnlyMany"]'
    storageclass.beta.kubernetes.io/is-default-class: "{{ if .IsDefaultClass }}true{{ else }}false{{ end }}"
provisioner: ceph.com/rbd
parameters:
  monitors: {{ .Monitors}}
  adminId: {{ .AdminID }}
  adminSecretName: ceph-rbd-{{ .AdminID }}-secret
  adminSecretNamespace: kube-system
  pool: {{ .Pool }}
  userId: {{ .UserID }}
  userSecretName: ceph-rbd-user-secret
  userSecretNamespace: kube-system
  fsType: {{ .FsType }}
  imageFormat: "{{ .ImageFormat }}"
  {{- if eq .ImageFormat 2 }}imageFeatures: "{{ .ImageFeatures }}"{{ end }}
allowVolumeExpansion: true
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
		"AdminSecret":         base64.URLEncoding.EncodeToString([]byte(strings.TrimSpace(mgr.Cluster.Storage.CephRBD.AdminSecret))),
		"UserSecret":          base64.URLEncoding.EncodeToString([]byte(strings.TrimSpace(mgr.Cluster.Storage.CephRBD.UserSecret))),
	})
}
