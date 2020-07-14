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

package glusterfs

import (
	"encoding/base64"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"strings"
	"text/template"
)

var GlusterFSTempl = template.Must(template.New("glusterfs").Parse(
	dedent.Dedent(`---
apiVersion: v1
kind: Secret
metadata:
  name: heketi-secret
  namespace: kube-system
type: kubernetes.io/glusterfs
data:
  key: {{ .JwtAdminKey }}

---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{ .StorageClassName }}
  annotations:
    storageclass.kubesphere.io/supported-access-modes: '["ReadWriteOnce","ReadOnlyMany","ReadWriteMany"]'
    storageclass.beta.kubernetes.io/is-default-class: "{{ if .IsDefaultClass }}true{{ else }}false{{ end }}"
provisioner: kubernetes.io/glusterfs
parameters:
  resturl: "{{ .RestUrl }}"
  clusterid: "{{ .ClusterID }}"
  restauthenabled: "{{ if .RestAuthEnabled }}true{{ else }}false{{ end }}"
  restuser: "{{ .RestUser }}"
  secretNamespace: "kube-system"
  secretName: "{{ .SecretName }}"
  gidMin: "{{ .GidMin }}"
  gidMax: "{{ .GidMax }}"
  volumetype: "{{ .VolumeType }}"
allowVolumeExpansion: true

    `)))

func GenerateGlusterFSManifests(mgr *manager.Manager) (string, error) {

	return util.Render(GlusterFSTempl, util.Data{
		"IsDefaultClass":   mgr.Cluster.Storage.GlusterFS.IsDefaultClass,
		"StorageClassName": mgr.Cluster.Storage.GlusterFS.StorageClassName,
		"ClusterID":        mgr.Cluster.Storage.GlusterFS.ClusterID,
		"RestAuthEnabled":  mgr.Cluster.Storage.GlusterFS.RestAuthEnabled,
		"RestUrl":          mgr.Cluster.Storage.GlusterFS.RestUrl,
		"RestUser":         mgr.Cluster.Storage.GlusterFS.RestUser,
		"SecretName":       mgr.Cluster.Storage.GlusterFS.SecretName,
		"GidMin":           mgr.Cluster.Storage.GlusterFS.GidMin,
		"GidMax":           mgr.Cluster.Storage.GlusterFS.GidMax,
		"VolumeType":       mgr.Cluster.Storage.GlusterFS.VolumeType,
		"JwtAdminKey":      base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(mgr.Cluster.Storage.GlusterFS.JwtAdminKey))),
	})
}
