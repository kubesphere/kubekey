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

package nfs_client

import (
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"text/template"
)

var NfsClientTempl = template.Must(template.New("nfs-client").Parse(
	dedent.Dedent(`# Default values for nfs-client-provisioner.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1
strategyType: Recreate

image:
  repository: {{ .NfsClientProvisionerRepo }}
  tag: {{ .NfsClientProvisionerTag }}
  pullPolicy: IfNotPresent

nfs:
  server: {{ .NfsClient.NfsServer }}
  path: {{ .NfsClient.NfsPath }}
  mountOptions:
  {{- if .NfsClient.NfsVrs3Enabled }}
  - 'nfsvers=3'
  {{- end }}

# For creating the StorageClass automatically:
storageClass:
  create: true

  # Set a provisioner name. If unset, a name will be generated.
  # provisionerName:

  # Set StorageClass as the default StorageClass
  # Ignored if storageClass.create is false
  defaultClass: {{ .NfsClient.IsDefaultClass }}

  # Set a StorageClass name
  # Ignored if storageClass.create is false
  name: {{ .NfsClient.StorageClassName }}

  # Allow volume to be expanded dynamically
  allowVolumeExpansion: true

  # Method used to reclaim an obsoleted volume
  reclaimPolicy: Delete

  # When set to false your PVs will not be archived by the provisioner upon deletion of the PVC.
  archiveOnDelete: {{ .NfsClient.NfsArchiveOnDelete }}

## For RBAC support:
rbac:
  # Specifies whether RBAC resources should be created
  create: true

# If true, create & use Pod Security Policy resources
# https://kubernetes.io/docs/concepts/policy/pod-security-policy/
podSecurityPolicy:
  enabled: false

## Set pod priorityClassName
# priorityClassName: ""

serviceAccount:
  # Specifies whether a ServiceAccount should be created
  create: true

  # The name of the ServiceAccount to use.
  # If not set and create is true, a name is generated using the fullname template
  name:

resources: {}
  # limits:
  #  cpu: 100m
  #  memory: 128Mi
  # requests:
  #  cpu: 100m
  #  memory: 128Mi

nodeSelector: {}

tolerations: []

affinity: {}

    `)))

func GenerateNfsClientValuesFile(mgr *manager.Manager) (string, error) {
	return util.Render(NfsClientTempl, util.Data{
		"NfsClient":                mgr.Cluster.Storage.NfsClient,
		"NfsClientProvisionerRepo": preinstall.GetImage(mgr, "nfs-client-provisioner").ImageRepo(),
		"NfsClientProvisionerTag":  preinstall.GetImage(mgr, "nfs-client-provisioner").Tag,
	})
}
