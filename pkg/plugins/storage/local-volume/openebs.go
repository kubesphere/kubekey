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

package local_volume

import (
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"text/template"
)

var OpenebsTempl = template.Must(template.New("openebs").Parse(
	dedent.Dedent(`---
#Sample storage classes for OpenEBS Local PV
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{ .LocalVolume.StorageClassName }}
  annotations:
    storageclass.kubesphere.io/supported-access-modes: '["ReadWriteOnce"]'
    storageclass.beta.kubernetes.io/is-default-class: "{{ if .LocalVolume.IsDefaultClass }}true{{ else }}false{{ end }}"
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
        value: "hostpath"
      - name: BasePath
        value: "/var/openebs/local/"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
---
# Create Maya Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: openebs-maya-operator
  namespace: kube-system
---
# Define Role that allows operations on K8s pods/deployments
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: openebs-maya-operator
rules:
- apiGroups: ["*"]
  resources: ["nodes", "nodes/proxy"]
  verbs: ["*"]
- apiGroups: ["*"]
  resources: ["namespaces", "services", "pods", "pods/exec", "deployments", "deployments/finalizers", "replicationcontrollers", "replicasets", "events", "endpoints", "configmaps", "secrets", "jobs", "cronjobs"]
  verbs: ["*"]
- apiGroups: ["*"]
  resources: ["statefulsets", "daemonsets"]
  verbs: ["*"]
- apiGroups: ["*"]
  resources: ["resourcequotas", "limitranges"]
  verbs: ["list", "watch"]
- apiGroups: ["*"]
  resources: ["ingresses", "horizontalpodautoscalers", "verticalpodautoscalers", "poddisruptionbudgets", "certificatesigningrequests"]
  verbs: ["list", "watch"]
- apiGroups: ["*"]
  resources: ["storageclasses", "persistentvolumeclaims", "persistentvolumes"]
  verbs: ["*"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: [ "get", "list", "create", "update", "delete", "patch"]
- apiGroups: ["openebs.io"]
  resources: [ "*"]
  verbs: ["*"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
---
# Bind the Service Account with the Role Privileges.
# TODO: Check if default account also needs to be there
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: openebs-maya-operator
subjects:
- kind: ServiceAccount
  name: openebs-maya-operator
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: openebs-maya-operator
  apiGroup: rbac.authorization.k8s.io
---
# This is the node-disk-manager related config.
# It can be used to customize the disks probes and filters
apiVersion: v1
kind: ConfigMap
metadata:
  name: openebs-ndm-config
  namespace: kube-system
  labels:
    openebs.io/component-name: ndm-config
data:
  # udev-probe is default or primary probe which should be enabled to run ndm
  # filterconfigs contails configs of filters - in their form fo include
  # and exclude comma separated strings
  node-disk-manager.config: |
    probeconfigs:
      - key: udev-probe
        name: udev probe
        state: true
      - key: seachest-probe
        name: seachest probe
        state: false
      - key: smart-probe
        name: smart probe
        state: true
    filterconfigs:
      - key: os-disk-exclude-filter
        name: os disk exclude filter
        state: true
        exclude: "/,/etc/hosts,/boot"
      - key: vendor-filter
        name: vendor filter
        state: true
        include: ""
        exclude: "CLOUDBYT,OpenEBS"
      - key: path-filter
        name: path filter
        state: true
        include: ""
        exclude: "loop,/dev/fd0,/dev/sr0,/dev/ram,/dev/dm-,/dev/md"
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: openebs-ndm
  namespace: kube-system
  labels:
    name: openebs-ndm
    openebs.io/component-name: ndm
    openebs.io/version: 1.10.0
spec:
  selector:
    matchLabels:
      name: openebs-ndm
      openebs.io/component-name: ndm
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        name: openebs-ndm
        openebs.io/component-name: ndm
        openebs.io/version: 1.10.0
    spec:
      # By default the node-disk-manager will be run on all kubernetes nodes
      # If you would like to limit this to only some nodes, say the nodes
      # that have storage attached, you could label those node and use
      # nodeSelector.
      #
      # e.g. label the storage nodes with - "openebs.io/nodegroup"="storage-node"
      # kubectl label node <node-name> "openebs.io/nodegroup"="storage-node"
      #nodeSelector:
      #  "openebs.io/nodegroup": "storage-node"
      serviceAccountName: openebs-maya-operator
      hostNetwork: true
      containers:
      - name: node-disk-manager
        image: {{ .NodeDiskManagerImage }}
        args:
          - -v=4
        # The feature-gate is used to enable the new UUID algorithm.
        # This is a feature currently in Alpha state
        #  - --feature-gates="GPTBasedUUID"
        imagePullPolicy: Always
        securityContext:
          privileged: true
        volumeMounts:
        - name: config
          mountPath: /host/node-disk-manager.config
          subPath: node-disk-manager.config
          readOnly: true
        - name: udev
          mountPath: /run/udev
        - name: procmount
          mountPath: /host/proc
          readOnly: true
        - name: basepath
          mountPath: /var/openebs/ndm
        - name: sparsepath
          mountPath: /var/openebs/sparse
        env:
        # namespace in which NDM is installed will be passed to NDM Daemonset
        # as environment variable
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        # pass hostname as env variable using downward API to the NDM container
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        # specify the directory where the sparse files need to be created.
        # if not specified, then sparse files will not be created.
        - name: SPARSE_FILE_DIR
          value: "/var/openebs/sparse"
        # Size(bytes) of the sparse file to be created.
        - name: SPARSE_FILE_SIZE
          value: "10737418240"
        # Specify the number of sparse files to be created
        - name: SPARSE_FILE_COUNT
          value: "0"
        livenessProbe:
          exec:
            command:
            - pgrep
            - "ndm"
          initialDelaySeconds: 30
          periodSeconds: 60
      volumes:
      - name: config
        configMap:
          name: openebs-ndm-config
      - name: udev
        hostPath:
          path: /run/udev
          type: Directory
      # mount /proc (to access mount file of process 1 of host) inside container
      # to read mount-point of disks and partitions
      - name: procmount
        hostPath:
          path: /proc
          type: Directory
      - name: basepath
        hostPath:
          path: /var/openebs/ndm
          type: DirectoryOrCreate
      - name: sparsepath
        hostPath:
          path: /var/openebs/sparse
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openebs-ndm-operator
  namespace: kube-system
  labels:
    name: openebs-ndm-operator
    openebs.io/component-name: ndm-operator
    openebs.io/version: 1.10.0
spec:
  selector:
    matchLabels:
      name: openebs-ndm-operator
      openebs.io/component-name: ndm-operator
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        name: openebs-ndm-operator
        openebs.io/component-name: ndm-operator
        openebs.io/version: 1.10.0
    spec:
      serviceAccountName: openebs-maya-operator
      containers:
        - name: node-disk-operator
          image: {{ .NodeDiskOperatorImage }}
          imagePullPolicy: Always
          readinessProbe:
            exec:
              command:
                - stat
                - /tmp/operator-sdk-ready
            initialDelaySeconds: 4
            periodSeconds: 10
            failureThreshold: 1
          env:
            - name: WATCH_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            # the service account of the ndm-operator pod
            - name: SERVICE_ACCOUNT
              valueFrom:
                fieldRef:
                  fieldPath: spec.serviceAccountName
            - name: OPERATOR_NAME
              value: "node-disk-operator"
            - name: CLEANUP_JOB_IMAGE
              value: "{{ .LinuxUtilsImage }}"
            # OPENEBS_IO_INSTALL_CRD environment variable is used to enable/disable CRD installation
            # from NDM operator. By default the CRDs will be installed
            #- name: OPENEBS_IO_INSTALL_CRD
            #  value: "true"
          # Process name used for matching is limited to the 15 characters
          # present in the pgrep output.
          # So fullname can be used here with pgrep (cmd is < 15 chars).
          livenessProbe:
            exec:
              command:
              - pgrep
              - "ndo"
            initialDelaySeconds: 30
            periodSeconds: 60
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openebs-localpv-provisioner
  namespace: kube-system
  labels:
    name: openebs-localpv-provisioner
    openebs.io/component-name: openebs-localpv-provisioner
    openebs.io/version: 1.10.0
spec:
  selector:
    matchLabels:
      name: openebs-localpv-provisioner
      openebs.io/component-name: openebs-localpv-provisioner
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        name: openebs-localpv-provisioner
        openebs.io/component-name: openebs-localpv-provisioner
        openebs.io/version: 1.10.0
    spec:
      serviceAccountName: openebs-maya-operator
      containers:
      - name: openebs-provisioner-hostpath
        imagePullPolicy: Always
        image: {{ .ProvisionerLocalPVImage }}
        env:
        # OPENEBS_IO_K8S_MASTER enables openebs provisioner to connect to K8s
        # based on this address. This is ignored if empty.
        # This is supported for openebs provisioner version 0.5.2 onwards
        #- name: OPENEBS_IO_K8S_MASTER
        #  value: "http://10.128.0.12:8080"
        # OPENEBS_IO_KUBE_CONFIG enables openebs provisioner to connect to K8s
        # based on this config. This is ignored if empty.
        # This is supported for openebs provisioner version 0.5.2 onwards
        #- name: OPENEBS_IO_KUBE_CONFIG
        #  value: "/home/ubuntu/.kube/config"
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: OPENEBS_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        # OPENEBS_SERVICE_ACCOUNT provides the service account of this pod as
        # environment variable
        - name: OPENEBS_SERVICE_ACCOUNT
          valueFrom:
            fieldRef:
              fieldPath: spec.serviceAccountName
        - name: OPENEBS_IO_ENABLE_ANALYTICS
          value: "true"
        - name: OPENEBS_IO_INSTALLER_TYPE
          value: "openebs-operator-lite"
        - name: OPENEBS_IO_HELPER_IMAGE
          value: "{{ .LinuxUtilsImage }}"
        livenessProbe:
          exec:
            command:
            - sh
            - -c
            - test $(pgrep -c "^provisioner-loc.*") = 1
          initialDelaySeconds: 30
          periodSeconds: 60
    `)))

func GenerateOpenebsManifests(mgr *manager.Manager) (string, error) {
	return util.Render(OpenebsTempl, util.Data{
		"LocalVolume":             mgr.Cluster.Storage.LocalVolume,
		"ProvisionerLocalPVImage": preinstall.GetImage(mgr, "provisioner-localpv").ImageName(),
		"OpenebsToolsImage":       preinstall.GetImage(mgr, "openebs-tools").ImageName(),
		"NodeDiskManagerImage":    preinstall.GetImage(mgr, "node-disk-manager").ImageName(),
		"NodeDiskOperatorImage":   preinstall.GetImage(mgr, "node-disk-operator").ImageName(),
		"LinuxUtilsImage":         preinstall.GetImage(mgr, "linux-utils").ImageName(),
	})
}
