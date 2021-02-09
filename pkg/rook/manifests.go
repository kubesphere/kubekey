package rook

const (
	common = `---
###################################################################################################################
# Create the common resources that are necessary to start the operator and the ceph cluster.
# These resources *must* be created before the operator.yaml and cluster.yaml or their variants.
# The samples all assume that a single operator will manage a single cluster crd in the same "rook-ceph" namespace.
#
# If the operator needs to manage multiple clusters (in different namespaces), see the section below
# for "cluster-specific resources". The resources below that section will need to be created for each namespace
# where the operator needs to manage the cluster. The resources above that section do not be created again.
#
# Most of the sections are prefixed with a 'OLM' keyword which is used to build our CSV for an OLM (Operator Life Cycle manager)
###################################################################################################################

# Namespace where the operator and other rook resources are created
apiVersion: v1
kind: Namespace
metadata:
  name: rook-ceph
# OLM: BEGIN CEPH CRD
# The CRD declarations
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephclusters.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephCluster
    listKind: CephClusterList
    plural: cephclusters
    singular: cephcluster
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            annotations: {}
            cephVersion:
              properties:
                allowUnsupported:
                  type: boolean
                image:
                  type: string
            dashboard:
              properties:
                enabled:
                  type: boolean
                urlPrefix:
                  type: string
                port:
                  type: integer
                  minimum: 0
                  maximum: 65535
                ssl:
                  type: boolean
            dataDirHostPath:
              pattern: ^/(\S+)
              type: string
            disruptionManagement:
              properties:
                machineDisruptionBudgetNamespace:
                  type: string
                managePodBudgets:
                  type: boolean
                osdMaintenanceTimeout:
                  type: integer
                manageMachineDisruptionBudgets:
                  type: boolean
            skipUpgradeChecks:
              type: boolean
            continueUpgradeAfterChecksEvenIfNotHealthy:
              type: boolean
            mon:
              properties:
                allowMultiplePerNode:
                  type: boolean
                count:
                  maximum: 9
                  minimum: 0
                  type: integer
                volumeClaimTemplate: {}
            mgr:
              properties:
                modules:
                  items:
                    properties:
                      name:
                        type: string
                      enabled:
                        type: boolean
            network:
              properties:
                hostNetwork:
                  type: boolean
                provider:
                  type: string
                selectors: {}
            storage:
              properties:
                disruptionManagement:
                  properties:
                    machineDisruptionBudgetNamespace:
                      type: string
                    managePodBudgets:
                      type: boolean
                    osdMaintenanceTimeout:
                      type: integer
                    manageMachineDisruptionBudgets:
                      type: boolean
                useAllNodes:
                  type: boolean
                nodes:
                  items:
                    properties:
                      name:
                        type: string
                      config:
                        properties:
                          metadataDevice:
                            type: string
                          storeType:
                            type: string
                            pattern: ^(bluestore)$
                          databaseSizeMB:
                            type: string
                          walSizeMB:
                            type: string
                          journalSizeMB:
                            type: string
                          osdsPerDevice:
                            type: string
                          encryptedDevice:
                            type: string
                            pattern: ^(true|false)$
                      useAllDevices:
                        type: boolean
                      deviceFilter:
                        type: string
                      devicePathFilter:
                        type: string
                      devices:
                        type: array
                        items:
                          properties:
                            name:
                              type: string
                            config: {}
                      resources: {}
                  type: array
                useAllDevices:
                  type: boolean
                deviceFilter:
                  type: string
                devicePathFilter:
                  type: string
                config: {}
                storageClassDeviceSets: {}
            driveGroups:
              type: array
              items:
                properties:
                  name:
                    type: string
                  spec: {}
                  placement: {}
                required:
                - name
                - spec
            monitoring:
              properties:
                enabled:
                  type: boolean
                rulesNamespace:
                  type: string
                externalMgrEndpoints:
                  type: array
                  items:
                    properties:
                      ip:
                        type: string
            removeOSDsIfOutAndSafeToRemove:
              type: boolean
            external:
              properties:
                enable:
                  type: boolean
            cleanupPolicy:
              properties:
                confirmation:
                  type: string
                  pattern: ^$|^yes-really-destroy-data$
                sanitizeDisks:
                  properties:
                    method:
                      type: string
                      pattern: ^(complete|quick)$
                    dataSource:
                      type: string
                      pattern: ^(zero|random)$
                    iteration:
                      type: integer
                      format: int32
            placement: {}
            resources: {}
            healthCheck: {}
  subresources:
    status: {}
  additionalPrinterColumns:
    - name: DataDirHostPath
      type: string
      description: Directory used on the K8s nodes
      JSONPath: .spec.dataDirHostPath
    - name: MonCount
      type: string
      description: Number of MONs
      JSONPath: .spec.mon.count
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp
    - name: Phase
      type: string
      description: Phase
      JSONPath: .status.phase
    - name: Message
      type: string
      description: Message
      JSONPath: .status.message
    - name: Health
      type: string
      description: Ceph Health
      JSONPath: .status.ceph.health
# OLM: END CEPH CRD
# OLM: BEGIN CEPH CLIENT CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephclients.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephClient
    listKind: CephClientList
    plural: cephclients
    singular: cephclient
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            caps:
              type: object
  subresources:
    status: {}
# OLM: END CEPH CLIENT CRD
# OLM: BEGIN CEPH RBD MIRROR CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephrbdmirrors.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephRBDMirror
    listKind: CephRBDMirrorList
    plural: cephrbdmirrors
    singular: cephrbdmirror
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            count:
              type: integer
              minimum: 1
              maximum: 100
  subresources:
    status: {}
# OLM: END CEPH RBD MIRROR CRD
# OLM: BEGIN CEPH FS CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephfilesystems.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephFilesystem
    listKind: CephFilesystemList
    plural: cephfilesystems
    singular: cephfilesystem
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            metadataServer:
              properties:
                activeCount:
                  minimum: 1
                  maximum: 10
                  type: integer
                activeStandby:
                  type: boolean
                annotations: {}
                placement: {}
                resources: {}
            metadataPool:
              properties:
                failureDomain:
                  type: string
                crushRoot:
                  type: string
                replicated:
                  properties:
                    size:
                      minimum: 0
                      maximum: 10
                      type: integer
                    requireSafeReplicaSize:
                      type: boolean
                erasureCoded:
                  properties:
                    dataChunks:
                      minimum: 0
                      maximum: 10
                      type: integer
                    codingChunks:
                      minimum: 0
                      maximum: 10
                      type: integer
                compressionMode:
                  type: string
                  enum:
                  - ""
                  - none
                  - passive
                  - aggressive
                  - force
            dataPools:
              type: array
              items:
                properties:
                  failureDomain:
                    type: string
                  crushRoot:
                    type: string
                  replicated:
                    properties:
                      size:
                        minimum: 0
                        maximum: 10
                        type: integer
                      requireSafeReplicaSize:
                        type: boolean
                  erasureCoded:
                    properties:
                      dataChunks:
                        minimum: 0
                        maximum: 10
                        type: integer
                      codingChunks:
                        minimum: 0
                        maximum: 10
                        type: integer
                  compressionMode:
                    type: string
                    enum:
                    - ""
                    - none
                    - passive
                    - aggressive
                    - force
                  parameters:
                    type: object
            preservePoolsOnDelete:
              type: boolean
  additionalPrinterColumns:
    - name: ActiveMDS
      type: string
      description: Number of desired active MDS daemons
      JSONPath: .spec.metadataServer.activeCount
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp
  subresources:
    status: {}
# OLM: END CEPH FS CRD
# OLM: BEGIN CEPH NFS CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephnfses.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephNFS
    listKind: CephNFSList
    plural: cephnfses
    singular: cephnfs
    shortNames:
    - nfs
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            rados:
              properties:
                pool:
                  type: string
                namespace:
                  type: string
            server:
              properties:
                active:
                  type: integer
                annotations: {}
                placement: {}
                resources: {}
  subresources:
    status: {}
# OLM: END CEPH NFS CRD
# OLM: BEGIN CEPH OBJECT STORE CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephobjectstores.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectStore
    listKind: CephObjectStoreList
    plural: cephobjectstores
    singular: cephobjectstore
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            gateway:
              properties:
                type:
                  type: string
                sslCertificateRef: {}
                port:
                  type: integer
                  minimum: 1
                  maximum: 65535
                securePort: {}
                instances:
                  type: integer
                externalRgwEndpoints:
                  type: array
                  items:
                    properties:
                      ip:
                        type: string
                annotations: {}
                placement: {}
                resources: {}
            metadataPool:
              properties:
                failureDomain:
                  type: string
                crushRoot:
                  type: string
                replicated:
                  properties:
                    size:
                      type: integer
                    requireSafeReplicaSize:
                      type: boolean
                erasureCoded:
                  properties:
                    dataChunks:
                      type: integer
                    codingChunks:
                      type: integer
                compressionMode:
                  type: string
                  enum:
                  - ""
                  - none
                  - passive
                  - aggressive
                  - force
                parameters:
                  type: object
            dataPool:
              properties:
                failureDomain:
                  type: string
                crushRoot:
                  type: string
                replicated:
                  properties:
                    size:
                      type: integer
                    requireSafeReplicaSize:
                      type: boolean
                erasureCoded:
                  properties:
                    dataChunks:
                      type: integer
                    codingChunks:
                      type: integer
                compressionMode:
                  type: string
                  enum:
                  - ""
                  - none
                  - passive
                  - aggressive
                  - force
                parameters:
                  type: object
            preservePoolsOnDelete:
              type: boolean
            healthCheck:
              properties:
                bucket:
                  properties:
                    enabled:
                      type: boolean
                    interval:
                      type: string
  subresources:
    status: {}
# OLM: END CEPH OBJECT STORE CRD
# OLM: BEGIN CEPH OBJECT STORE USERS CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephobjectstoreusers.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectStoreUser
    listKind: CephObjectStoreUserList
    plural: cephobjectstoreusers
    singular: cephobjectstoreuser
    shortNames:
    - rcou
    - objectuser
  scope: Namespaced
  version: v1
  subresources:
    status: {}
# OLM: END CEPH OBJECT STORE USERS CRD
# OLM: BEGIN CEPH OBJECT REALM CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephobjectrealms.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectRealm
    listKind: CephObjectRealmList
    plural: cephobjectrealms
    singular: cephobjectrealm
  scope: Namespaced
  version: v1
  subresources:
    status: {}
# OLM: END CEPH OBJECT REALM CRD
# OLM: BEGIN CEPH OBJECT ZONEGROUP CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephobjectzonegroups.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectZoneGroup
    listKind: CephObjectZoneGroupList
    plural: cephobjectzonegroups
    singular: cephobjectzonegroup
  scope: Namespaced
  version: v1
  subresources:
    status: {}
# OLM: END CEPH OBJECT ZONEGROUP CRD
# OLM: BEGIN CEPH OBJECT ZONE CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephobjectzones.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectZone
    listKind: CephObjectZoneList
    plural: cephobjectzones
    singular: cephobjectzone
  scope: Namespaced
  version: v1
  subresources:
    status: {}
# OLM: END CEPH OBJECT ZONE CRD
# OLM: BEGIN CEPH BLOCK POOL CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephblockpools.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephBlockPool
    listKind: CephBlockPoolList
    plural: cephblockpools
    singular: cephblockpool
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            failureDomain:
                type: string
            crushRoot:
                type: string
            replicated:
              properties:
                size:
                  type: integer
                  minimum: 0
                  maximum: 9
                targetSizeRatio:
                  type: number
                requireSafeReplicaSize:
                  type: boolean
            erasureCoded:
              properties:
                dataChunks:
                  type: integer
                  minimum: 0
                  maximum: 9
                codingChunks:
                  type: integer
                  minimum: 0
                  maximum: 9
            compressionMode:
              type: string
              enum:
              - ""
              - none
              - passive
              - aggressive
              - force
            enableRBDStats:
              description: EnableRBDStats is used to enable gathering of statistics
                for all RBD images in the pool
              type: boolean
            parameters:
              type: object
  subresources:
    status: {}
# OLM: END CEPH BLOCK POOL CRD
# OLM: BEGIN CEPH VOLUME POOL CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: volumes.rook.io
spec:
  group: rook.io
  names:
    kind: Volume
    listKind: VolumeList
    plural: volumes
    singular: volume
    shortNames:
    - rv
  scope: Namespaced
  version: v1alpha2
  subresources:
    status: {}
# OLM: END CEPH VOLUME POOL CRD
# OLM: BEGIN OBJECTBUCKET CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbuckets.objectbucket.io
spec:
  group: objectbucket.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
  names:
    kind: ObjectBucket
    listKind: ObjectBucketList
    plural: objectbuckets
    singular: objectbucket
    shortNames:
      - ob
      - obs
  scope: Cluster
  subresources:
    status: {}
# OLM: END OBJECTBUCKET CRD
# OLM: BEGIN OBJECTBUCKETCLAIM CRD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbucketclaims.objectbucket.io
spec:
  versions:
    - name: v1alpha1
      served: true
      storage: true
  group: objectbucket.io
  names:
    kind: ObjectBucketClaim
    listKind: ObjectBucketClaimList
    plural: objectbucketclaims
    singular: objectbucketclaim
    shortNames:
      - obc
      - obcs
  scope: Namespaced
  subresources:
    status: {}
# OLM: END OBJECTBUCKETCLAIM CRD
# OLM: BEGIN OBJECTBUCKET ROLEBINDING
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-object-bucket
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-object-bucket
subjects:
  - kind: ServiceAccount
    name: rook-ceph-system
    namespace: rook-ceph
# OLM: END OBJECTBUCKET ROLEBINDING
# OLM: BEGIN OPERATOR ROLE
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-admission-controller
  namespace: rook-ceph
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rook-ceph-admission-controller-role
rules:
  - apiGroups: ["ceph.rook.io"]
    resources: ["*"]
    verbs: ["get", "watch", "list"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rook-ceph-admission-controller-rolebinding
subjects:
  - kind: ServiceAccount
    name: rook-ceph-admission-controller
    apiGroup: ""
    namespace: rook-ceph
roleRef:
  kind: ClusterRole
  name: rook-ceph-admission-controller-role
  apiGroup: rbac.authorization.k8s.io
---
# The cluster role for managing all the cluster-specific resources in a namespace
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: rook-ceph-cluster-mgmt
  labels:
    operator: rook
    storage-backend: ceph
rules:
- apiGroups:
  - ""
  - apps
  - extensions
  resources:
  - secrets
  - pods
  - pods/log
  - services
  - configmaps
  - deployments
  - daemonsets
  verbs:
  - get
  - list
  - watch
  - patch
  - create
  - update
  - delete
---
# The role for the operator to manage resources in its own namespace
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: rook-ceph-system
  namespace: rook-ceph
  labels:
    operator: rook
    storage-backend: ceph
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - configmaps
  - services
  verbs:
  - get
  - list
  - watch
  - patch
  - create
  - update
  - delete
- apiGroups:
  - apps
  - extensions
  resources:
  - daemonsets
  - statefulsets
  - deployments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - k8s.cni.cncf.io
  resources:
  - network-attachment-definitions
  verbs:
  - get
---
# The cluster role for managing the Rook CRDs
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: rook-ceph-global
  labels:
    operator: rook
    storage-backend: ceph
rules:
- apiGroups:
  - ""
  resources:
  # Pod access is needed for fencing
  - pods
  # Node access is needed for determining nodes where mons should run
  - nodes
  - nodes/proxy
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
    # PVs and PVCs are managed by the Rook provisioner
  - persistentvolumes
  - persistentvolumeclaims
  - endpoints
  verbs:
  - get
  - list
  - watch
  - patch
  - create
  - update
  - delete
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - ceph.rook.io
  resources:
  - "*"
  verbs:
  - "*"
- apiGroups:
  - rook.io
  resources:
  - "*"
  verbs:
  - "*"
- apiGroups:
  - policy
  - apps
  - extensions
  resources:
  # This is for the clusterdisruption controller
  - poddisruptionbudgets
  # This is for both clusterdisruption and nodedrain controllers
  - deployments
  - replicasets
  verbs:
  - "*"
- apiGroups:
  - healthchecking.openshift.io
  resources:
  - machinedisruptionbudgets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - machine.openshift.io
  resources:
  - machines
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - storage.k8s.io
  resources:
  - csidrivers
  verbs:
  - create
  - delete
---
# Aspects of ceph-mgr that require cluster-wide access
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-cluster
  labels:
    operator: rook
    storage-backend: ceph
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - nodes
  - nodes/proxy
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - list
  - get
  - watch
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-object-bucket
  labels:
    operator: rook
    storage-backend: ceph
rules:
- apiGroups:
  - ""
  verbs:
  - "*"
  resources:
  - secrets
  - configmaps
- apiGroups:
    - storage.k8s.io
  resources:
    - storageclasses
  verbs:
    - get
    - list
    - watch
- apiGroups:
  - "objectbucket.io"
  verbs:
  - "*"
  resources:
  - "*"
# OLM: END OPERATOR ROLE
# OLM: BEGIN SERVICE ACCOUNT SYSTEM
---
# The rook system service account used by the operator, agent, and discovery pods
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-system
  namespace: rook-ceph
  labels:
    operator: rook
    storage-backend: ceph
# imagePullSecrets:
# - name: my-registry-secret

# OLM: END SERVICE ACCOUNT SYSTEM
# OLM: BEGIN OPERATOR ROLEBINDING
---
# Grant the operator, agent, and discovery agents access to resources in the namespace
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-system
  namespace: rook-ceph
  labels:
    operator: rook
    storage-backend: ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-system
subjects:
- kind: ServiceAccount
  name: rook-ceph-system
  namespace: rook-ceph
---
# Grant the rook system daemons cluster-wide access to manage the Rook CRDs, PVCs, and storage classes
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-global
  labels:
    operator: rook
    storage-backend: ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-global
subjects:
- kind: ServiceAccount
  name: rook-ceph-system
  namespace: rook-ceph
# OLM: END OPERATOR ROLEBINDING
#################################################################################################################
# Beginning of cluster-specific resources. The example will assume the cluster will be created in the "rook-ceph"
# namespace. If you want to create the cluster in a different namespace, you will need to modify these roles
# and bindings accordingly.
#################################################################################################################
# Service account for the Ceph OSDs. Must exist and cannot be renamed.
# OLM: BEGIN SERVICE ACCOUNT OSD
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-osd
  namespace: rook-ceph
# imagePullSecrets:
# - name: my-registry-secret

# OLM: END SERVICE ACCOUNT OSD
# OLM: BEGIN SERVICE ACCOUNT MGR
---
# Service account for the Ceph Mgr. Must exist and cannot be renamed.
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-mgr
  namespace: rook-ceph
# imagePullSecrets:
# - name: my-registry-secret

# OLM: END SERVICE ACCOUNT MGR
# OLM: BEGIN CMD REPORTER SERVICE ACCOUNT
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-cmd-reporter
  namespace: rook-ceph
# OLM: END CMD REPORTER SERVICE ACCOUNT
# OLM: BEGIN CLUSTER ROLE
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
  namespace: rook-ceph
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: [ "get", "list", "watch", "create", "update", "delete" ]
- apiGroups: ["ceph.rook.io"]
  resources: ["cephclusters", "cephclusters/finalizers"]
  verbs: [ "get", "list", "create", "update", "delete" ]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
---
# Aspects of ceph-mgr that require access to the system namespace
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-system
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
---
# Aspects of ceph-mgr that operate within the cluster's namespace
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr
  namespace: rook-ceph
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - pods/log
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - ceph.rook.io
  resources:
  - "*"
  verbs:
  - "*"
# OLM: END CLUSTER ROLE
# OLM: BEGIN CMD REPORTER ROLE
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-cmd-reporter
  namespace: rook-ceph
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - configmaps
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
# OLM: END CMD REPORTER ROLE
# OLM: BEGIN CLUSTER ROLEBINDING
---
# Allow the operator to create resources in this cluster's namespace
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-cluster-mgmt
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-cluster-mgmt
subjects:
- kind: ServiceAccount
  name: rook-ceph-system
  namespace: rook-ceph
---
# Allow the osd pods in this namespace to work with configmaps
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-osd
subjects:
- kind: ServiceAccount
  name: rook-ceph-osd
  namespace: rook-ceph
---
# Allow the ceph mgr to access the cluster-specific resources necessary for the mgr modules
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-mgr
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: rook-ceph
---
# Allow the ceph mgr to access the rook system resources necessary for the mgr modules
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-system
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-mgr-system
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: rook-ceph
---
# Allow the ceph mgr to access cluster-wide resources necessary for the mgr modules
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-cluster
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-mgr-cluster
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: rook-ceph

---
# Allow the ceph osd to access cluster-wide resources necessary for determining their topology location
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-osd
subjects:
- kind: ServiceAccount
  name: rook-ceph-osd
  namespace: rook-ceph

# OLM: END CLUSTER ROLEBINDING
# OLM: BEGIN CMD REPORTER ROLEBINDING
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-cmd-reporter
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-cmd-reporter
subjects:
- kind: ServiceAccount
  name: rook-ceph-cmd-reporter
  namespace: rook-ceph
# OLM: END CMD REPORTER ROLEBINDING
#################################################################################################################
# Beginning of pod security policy resources. The example will assume the cluster will be created in the
# "rook-ceph" namespace. If you want to create the cluster in a different namespace, you will need to modify
# the roles and bindings accordingly.
#################################################################################################################
# OLM: BEGIN CLUSTER POD SECURITY POLICY
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  # Note: Kubernetes matches PSPs to deployments alphabetically. In some environments, this PSP may
  # need to be renamed with a value that will match before others.
  name: 00-rook-privileged
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: 'runtime/default'
    seccomp.security.alpha.kubernetes.io/defaultProfileName:  'runtime/default'
spec:
  privileged: true
  allowedCapabilities:
    # required by CSI
    - SYS_ADMIN
  # fsGroup - the flexVolume agent has fsGroup capabilities and could potentially be any group
  fsGroup:
    rule: RunAsAny
  # runAsUser, supplementalGroups - Rook needs to run some pods as root
  # Ceph pods could be run as the Ceph user, but that user isn't always known ahead of time
  runAsUser:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  # seLinux - seLinux context is unknown ahead of time; set if this is well-known
  seLinux:
    rule: RunAsAny
  volumes:
    # recommended minimum set
    - configMap
    - downwardAPI
    - emptyDir
    - persistentVolumeClaim
    - secret
    - projected
    # required for Rook
    - hostPath
    - flexVolume
  # allowedHostPaths can be set to Rook's known host volume mount points when they are fully-known
  # allowedHostPaths:
  #   - pathPrefix: "/run/udev"  # for OSD prep
  #     readOnly: false
  #   - pathPrefix: "/dev"  # for OSD prep
  #     readOnly: false
  #   - pathPrefix: "/var/lib/rook"  # or whatever the dataDirHostPath value is set to
  #     readOnly: false
  # Ceph requires host IPC for setting up encrypted devices
  hostIPC: true
  # Ceph OSDs need to share the same PID namespace
  hostPID: true
  # hostNetwork can be set to 'false' if host networking isn't used
  hostNetwork: true
  hostPorts:
    # Ceph messenger protocol v1
    - min: 6789
      max: 6790 # <- support old default port
    # Ceph messenger protocol v2
    - min: 3300
      max: 3300
    # Ceph RADOS ports for OSDs, MDSes
    - min: 6800
      max: 7300
    # # Ceph dashboard port HTTP (not recommended)
    # - min: 7000
    #   max: 7000
    # Ceph dashboard port HTTPS
    - min: 8443
      max: 8443
    # Ceph mgr Prometheus Metrics
    - min: 9283
      max: 9283
# OLM: END CLUSTER POD SECURITY POLICY
# OLM: BEGIN POD SECURITY POLICY BINDINGS
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: 'psp:rook'
rules:
  - apiGroups:
      - policy
    resources:
      - podsecuritypolicies
    resourceNames:
      - 00-rook-privileged
    verbs:
      - use
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-ceph-system-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-ceph-system
    namespace: rook-ceph
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-default-psp
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: default
  namespace: rook-ceph
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-osd-psp
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: rook-ceph-osd
  namespace: rook-ceph
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-mgr-psp
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: rook-ceph
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-cmd-reporter-psp
  namespace: rook-ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: rook-ceph-cmd-reporter
  namespace: rook-ceph
# OLM: END CLUSTER POD SECURITY POLICY BINDINGS
# OLM: BEGIN CSI CEPHFS SERVICE ACCOUNT
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-cephfs-plugin-sa
  namespace: rook-ceph
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-cephfs-provisioner-sa
  namespace: rook-ceph
# OLM: END CSI CEPHFS SERVICE ACCOUNT
# OLM: BEGIN CSI CEPHFS ROLE
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: rook-ceph
  name: cephfs-external-provisioner-cfg
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "create", "delete"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
# OLM: END CSI CEPHFS ROLE
# OLM: BEGIN CSI CEPHFS ROLEBINDING
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-provisioner-role-cfg
  namespace: rook-ceph
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-provisioner-sa
    namespace: rook-ceph
roleRef:
  kind: Role
  name: cephfs-external-provisioner-cfg
  apiGroup: rbac.authorization.k8s.io
# OLM: END CSI CEPHFS ROLEBINDING
# OLM: BEGIN CSI CEPHFS CLUSTER ROLE
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-nodeplugin
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-external-provisioner-runner
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotcontents"]
    verbs: ["create", "get", "list", "watch", "update", "delete"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotcontents/status"]
    verbs: ["update"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["create", "list", "watch", "delete", "get", "update"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots/status"]
    verbs: ["update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update", "patch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims/status"]
    verbs: ["update", "patch"]
# OLM: END CSI CEPHFS CLUSTER ROLE
# OLM: BEGIN CSI CEPHFS CLUSTER ROLEBINDING
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-cephfs-plugin-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-plugin-sa
    namespace: rook-ceph
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-cephfs-provisioner-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-provisioner-sa
    namespace: rook-ceph
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-nodeplugin
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-plugin-sa
    namespace: rook-ceph
roleRef:
  kind: ClusterRole
  name: cephfs-csi-nodeplugin
  apiGroup: rbac.authorization.k8s.io

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-provisioner-role
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-provisioner-sa
    namespace: rook-ceph
roleRef:
  kind: ClusterRole
  name: cephfs-external-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
# OLM: END CSI CEPHFS CLUSTER ROLEBINDING
# OLM: BEGIN CSI RBD SERVICE ACCOUNT
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-rbd-plugin-sa
  namespace: rook-ceph
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-rbd-provisioner-sa
  namespace: rook-ceph
# OLM: END CSI RBD SERVICE ACCOUNT
# OLM: BEGIN CSI RBD ROLE
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: rook-ceph
  name: rbd-external-provisioner-cfg
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
# OLM: END CSI RBD ROLE
# OLM: BEGIN CSI RBD ROLEBINDING
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-provisioner-role-cfg
  namespace: rook-ceph
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-provisioner-sa
    namespace: rook-ceph
roleRef:
  kind: Role
  name: rbd-external-provisioner-cfg
  apiGroup: rbac.authorization.k8s.io
# OLM: END CSI RBD ROLEBINDING
# OLM: BEGIN CSI RBD CLUSTER ROLE
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-nodeplugin
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-external-provisioner-runner
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update", "patch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotcontents"]
    verbs: ["create", "get", "list", "watch", "update", "delete"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotcontents/status"]
    verbs: ["update"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["create", "list", "watch", "delete", "get", "update"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots/status"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims/status"]
    verbs: ["update", "patch"]
# OLM: END CSI RBD CLUSTER ROLE
# OLM: BEGIN CSI RBD CLUSTER ROLEBINDING
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-rbd-plugin-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-plugin-sa
    namespace: rook-ceph
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-rbd-provisioner-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-provisioner-sa
    namespace: rook-ceph
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-nodeplugin
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-plugin-sa
    namespace: rook-ceph
roleRef:
  kind: ClusterRole
  name: rbd-csi-nodeplugin
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-provisioner-role
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-provisioner-sa
    namespace: rook-ceph
roleRef:
  kind: ClusterRole
  name: rbd-external-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
# OLM: END CSI RBD CLUSTER ROLEBINDING
`

	operate = `---
# Rook Ceph Operator Config ConfigMap
# Use this ConfigMap to override Rook-Ceph Operator configurations.
# NOTE! Precedence will be given to this config if the same Env Var config also exists in the
#       Operator Deployment.
# To move a configuration(s) from the Operator Deployment to this ConfigMap, add the config
# here. It is recommended to then remove it from the Deployment to eliminate any future confusion.
kind: ConfigMap
apiVersion: v1
metadata:
  name: rook-ceph-operator-config
  # should be in the namespace of the operator
  namespace: rook-ceph
data:
  # Enable the CSI driver.
  # To run the non-default version of the CSI driver, see the override-able image properties in operator.yaml
  ROOK_CSI_ENABLE_CEPHFS: "true"
  # Enable the default version of the CSI RBD driver. To start another version of the CSI driver, see image properties below.
  ROOK_CSI_ENABLE_RBD: "true"
  ROOK_CSI_ENABLE_GRPC_METRICS: "true"

  # Set logging level for csi containers.
  # Supported values from 0 to 5. 0 for general useful logs, 5 for trace level verbosity.
  # CSI_LOG_LEVEL: "0"

  # Enable cephfs kernel driver instead of ceph-fuse.
  # If you disable the kernel client, your application may be disrupted during upgrade.
  # See the upgrade guide: https://rook.io/docs/rook/master/ceph-upgrade.html
  # NOTE! cephfs quota is not supported in kernel version < 4.17
  CSI_FORCE_CEPHFS_KERNEL_CLIENT: "true"

  # (Optional) Allow starting unsupported ceph-csi image
  ROOK_CSI_ALLOW_UNSUPPORTED_VERSION: "false"
  # The default version of CSI supported by Rook will be started. To change the version
  # of the CSI driver to something other than what is officially supported, change
  # these images to the desired release of the CSI driver.
  # ROOK_CSI_CEPH_IMAGE: "quay.io/cephcsi/cephcsi:v3.1.1"
  # ROOK_CSI_REGISTRAR_IMAGE: "quay.io/k8scsi/csi-node-driver-registrar:v1.2.0"
  # ROOK_CSI_RESIZER_IMAGE: "quay.io/k8scsi/csi-resizer:v0.4.0"
  # ROOK_CSI_PROVISIONER_IMAGE: "quay.io/k8scsi/csi-provisioner:v1.6.0"
  # ROOK_CSI_SNAPSHOTTER_IMAGE: "quay.io/k8scsi/csi-snapshotter:v2.1.1"
  # ROOK_CSI_ATTACHER_IMAGE: "quay.io/k8scsi/csi-attacher:v2.1.0"

  # (Optional) set user created priorityclassName for csi plugin pods.
  # CSI_PLUGIN_PRIORITY_CLASSNAME: "system-node-critical"

  # (Optional) set user created priorityclassName for csi provisioner pods.
  # CSI_PROVISIONER_PRIORITY_CLASSNAME: "system-cluster-critical"

  # CSI CephFS plugin daemonset update strategy, supported values are OnDelete and RollingUpdate.
  # Default value is RollingUpdate.
  # CSI_CEPHFS_PLUGIN_UPDATE_STRATEGY: "OnDelete"
  # CSI RBD plugin daemonset update strategy, supported values are OnDelete and RollingUpdate.
  # Default value is RollingUpdate.
  # CSI_RBD_PLUGIN_UPDATE_STRATEGY: "OnDelete"

  # kubelet directory path, if kubelet configured to use other than /var/lib/kubelet path.
  # ROOK_CSI_KUBELET_DIR_PATH: "/var/lib/kubelet"

  # (Optional) Ceph Provisioner NodeAffinity.
  # CSI_PROVISIONER_NODE_AFFINITY: "role=storage-node; storage=rook, ceph"
  # (Optional) CEPH CSI provisioner tolerations list. Put here list of taints you want to tolerate in YAML format.
  # CSI provisioner would be best to start on the same nodes as other ceph daemons.
  # CSI_PROVISIONER_TOLERATIONS: |
  #   - effect: NoSchedule
  #     key: node-role.kubernetes.io/controlplane
  #     operator: Exists
  #   - effect: NoExecute
  #     key: node-role.kubernetes.io/etcd
  #     operator: Exists
  # (Optional) Ceph CSI plugin NodeAffinity.
  # CSI_PLUGIN_NODE_AFFINITY: "role=storage-node; storage=rook, ceph"
  # (Optional) CEPH CSI plugin tolerations list. Put here list of taints you want to tolerate in YAML format.
  # CSI plugins need to be started on all the nodes where the clients need to mount the storage.
  # CSI_PLUGIN_TOLERATIONS: |
  #   - effect: NoSchedule
  #     key: node-role.kubernetes.io/controlplane
  #     operator: Exists
  #   - effect: NoExecute
  #     key: node-role.kubernetes.io/etcd
  #     operator: Exists

  # (Optional) CEPH CSI RBD provisioner resource requirement list, Put here list of resource
  # requests and limits you want to apply for provisioner pod
  # CSI_RBD_PROVISIONER_RESOURCE: |
  #  - name : csi-provisioner
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 100m
  #      limits:
  #        memory: 256Mi
  #        cpu: 200m
  #  - name : csi-resizer
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 100m
  #      limits:
  #        memory: 256Mi
  #        cpu: 200m
  #  - name : csi-attacher
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 100m
  #      limits:
  #        memory: 256Mi
  #        cpu: 200m
  #  - name : csi-snapshotter
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 100m
  #      limits:
  #        memory: 256Mi
  #        cpu: 200m
  #  - name : csi-rbdplugin
  #    resource:
  #      requests:
  #        memory: 512Mi
  #        cpu: 250m
  #      limits:
  #        memory: 1Gi
  #        cpu: 500m
  #  - name : liveness-prometheus
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 50m
  #      limits:
  #        memory: 256Mi
  #        cpu: 100m
  # (Optional) CEPH CSI RBD plugin resource requirement list, Put here list of resource
  # requests and limits you want to apply for plugin pod
  # CSI_RBD_PLUGIN_RESOURCE: |
  #  - name : driver-registrar
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 50m
  #      limits:
  #        memory: 256Mi
  #        cpu: 100m
  #  - name : csi-rbdplugin
  #    resource:
  #      requests:
  #        memory: 512Mi
  #        cpu: 250m
  #      limits:
  #        memory: 1Gi
  #        cpu: 500m
  #  - name : liveness-prometheus
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 50m
  #      limits:
  #        memory: 256Mi
  #        cpu: 100m
  # (Optional) CEPH CSI CephFS provisioner resource requirement list, Put here list of resource
  # requests and limits you want to apply for provisioner pod
  # CSI_CEPHFS_PROVISIONER_RESOURCE: |
  #  - name : csi-provisioner
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 100m
  #      limits:
  #        memory: 256Mi
  #        cpu: 200m
  #  - name : csi-resizer
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 100m
  #      limits:
  #        memory: 256Mi
  #        cpu: 200m
  #  - name : csi-attacher
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 100m
  #      limits:
  #        memory: 256Mi
  #        cpu: 200m
  #  - name : csi-cephfsplugin
  #    resource:
  #      requests:
  #        memory: 512Mi
  #        cpu: 250m
  #      limits:
  #        memory: 1Gi
  #        cpu: 500m
  #  - name : liveness-prometheus
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 50m
  #      limits:
  #        memory: 256Mi
  #        cpu: 100m
  # (Optional) CEPH CSI CephFS plugin resource requirement list, Put here list of resource
  # requests and limits you want to apply for plugin pod
  # CSI_CEPHFS_PLUGIN_RESOURCE: |
  #  - name : driver-registrar
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 50m
  #      limits:
  #        memory: 256Mi
  #        cpu: 100m
  #  - name : csi-cephfsplugin
  #    resource:
  #      requests:
  #        memory: 512Mi
  #        cpu: 250m
  #      limits:
  #        memory: 1Gi
  #        cpu: 500m
  #  - name : liveness-prometheus
  #    resource:
  #      requests:
  #        memory: 128Mi
  #        cpu: 50m
  #      limits:
  #        memory: 256Mi
  #        cpu: 100m

  # Configure CSI CSI Ceph FS grpc and liveness metrics port
  # CSI_CEPHFS_GRPC_METRICS_PORT: "9091"
  # CSI_CEPHFS_LIVENESS_METRICS_PORT: "9081"
  # Configure CSI RBD grpc and liveness metrics port
  # CSI_RBD_GRPC_METRICS_PORT: "9090"
  # CSI_RBD_LIVENESS_METRICS_PORT: "9080"

  # Whether the OBC provisioner should watch on the operator namespace or not, if not the namespace of the cluster will be used
  ROOK_OBC_WATCH_OPERATOR_NAMESPACE: "true"
---
# OLM: BEGIN OPERATOR DEPLOYMENT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rook-ceph-operator
  namespace: rook-ceph
  labels:
    operator: rook
    storage-backend: ceph
spec:
  selector:
    matchLabels:
      app: rook-ceph-operator
  replicas: 1
  template:
    metadata:
      labels:
        app: rook-ceph-operator
    spec:
      serviceAccountName: rook-ceph-system
      containers:
      - name: rook-ceph-operator
        image: rook/ceph:v1.4.6
        args: ["ceph", "operator"]
        volumeMounts:
        - mountPath: /var/lib/rook
          name: rook-config
        - mountPath: /etc/ceph
          name: default-config-dir
        env:
        # If the operator should only watch for cluster CRDs in the same namespace, set this to "true".
        # If this is not set to true, the operator will watch for cluster CRDs in all namespaces.
        - name: ROOK_CURRENT_NAMESPACE_ONLY
          value: "false"
        # To disable RBAC, uncomment the following:
        # - name: RBAC_ENABLED
        #   value: "false"
        # Rook Agent toleration. Will tolerate all taints with all keys.
        # Choose between NoSchedule, PreferNoSchedule and NoExecute:
        # - name: AGENT_TOLERATION
        #   value: "NoSchedule"
        # (Optional) Rook Agent toleration key. Set this to the key of the taint you want to tolerate
        # - name: AGENT_TOLERATION_KEY
        #   value: "<KeyOfTheTaintToTolerate>"
        # (Optional) Rook Agent tolerations list. Put here list of taints you want to tolerate in YAML format.
        # - name: AGENT_TOLERATIONS
        #   value: |
        #     - effect: NoSchedule
        #       key: node-role.kubernetes.io/controlplane
        #       operator: Exists
        #     - effect: NoExecute
        #       key: node-role.kubernetes.io/etcd
        #       operator: Exists
        # (Optional) Rook Agent priority class name to set on the pod(s)
        # - name: AGENT_PRIORITY_CLASS_NAME
        #   value: "<PriorityClassName>"
        # (Optional) Rook Agent NodeAffinity.
        # - name: AGENT_NODE_AFFINITY
        #   value: "role=storage-node; storage=rook,ceph"
        # - name: AGENT_MOUNT_SECURITY_MODE
        #   value: "Any"
        # Set the path where the Rook agent can find the flex volumes
        # - name: FLEXVOLUME_DIR_PATH
        #   value: "<PathToFlexVolumes>"
        # Set the path where kernel modules can be found
        # - name: LIB_MODULES_DIR_PATH
        #   value: "<PathToLibModules>"
        # Mount any extra directories into the agent container
        # - name: AGENT_MOUNTS
        #   value: "somemount=/host/path:/container/path,someothermount=/host/path2:/container/path2"
        # Rook Discover toleration. Will tolerate all taints with all keys.
        # Choose between NoSchedule, PreferNoSchedule and NoExecute:
        # - name: DISCOVER_TOLERATION
        #   value: "NoSchedule"
        # (Optional) Rook Discover toleration key. Set this to the key of the taint you want to tolerate
        # - name: DISCOVER_TOLERATION_KEY
        #   value: "<KeyOfTheTaintToTolerate>"
        # (Optional) Rook Discover tolerations list. Put here list of taints you want to tolerate in YAML format.
        # - name: DISCOVER_TOLERATIONS
        #   value: |
        #     - effect: NoSchedule
        #       key: node-role.kubernetes.io/controlplane
        #       operator: Exists
        #     - effect: NoExecute
        #       key: node-role.kubernetes.io/etcd
        #       operator: Exists
        # (Optional) Rook Discover priority class name to set on the pod(s)
        # - name: DISCOVER_PRIORITY_CLASS_NAME
        #   value: "<PriorityClassName>"
        # (Optional) Discover Agent NodeAffinity.
        # - name: DISCOVER_AGENT_NODE_AFFINITY
        #   value: "role=storage-node; storage=rook, ceph"
        # Allow rook to create multiple file systems. Note: This is considered
        # an experimental feature in Ceph as described at
        # http://docs.ceph.com/docs/master/cephfs/experimental-features/#multiple-filesystems-within-a-ceph-cluster
        # which might cause mons to crash as seen in https://github.com/rook/rook/issues/1027
        - name: ROOK_ALLOW_MULTIPLE_FILESYSTEMS
          value: "false"

        # The logging level for the operator: INFO | DEBUG
        - name: ROOK_LOG_LEVEL
          value: "INFO"

        # The duration between discovering devices in the rook-discover daemonset.
        - name: ROOK_DISCOVER_DEVICES_INTERVAL
          value: "60m"

        # Whether to start pods as privileged that mount a host path, which includes the Ceph mon and osd pods.
        # Set this to true if SELinux is enabled (e.g. OpenShift) to workaround the anyuid issues.
        # For more details see https://github.com/rook/rook/issues/1314#issuecomment-355799641
        - name: ROOK_HOSTPATH_REQUIRES_PRIVILEGED
          value: "false"

        # In some situations SELinux relabelling breaks (times out) on large filesystems, and doesn't work with cephfs ReadWriteMany volumes (last relabel wins).
        # Disable it here if you have similar issues.
        # For more details see https://github.com/rook/rook/issues/2417
        - name: ROOK_ENABLE_SELINUX_RELABELING
          value: "true"

        # In large volumes it will take some time to chown all the files. Disable it here if you have performance issues.
        # For more details see https://github.com/rook/rook/issues/2254
        - name: ROOK_ENABLE_FSGROUP
          value: "true"

        # Disable automatic orchestration when new devices are discovered
        - name: ROOK_DISABLE_DEVICE_HOTPLUG
          value: "false"

        # Provide customised regex as the values using comma. For eg. regex for rbd based volume, value will be like "(?i)rbd[0-9]+".
        # In case of more than one regex, use comma to seperate between them.
        # Default regex will be "(?i)dm-[0-9]+,(?i)rbd[0-9]+,(?i)nbd[0-9]+"
        # Add regex expression after putting a comma to blacklist a disk
        # If value is empty, the default regex will be used.
        - name: DISCOVER_DAEMON_UDEV_BLACKLIST
          value: "(?i)dm-[0-9]+,(?i)rbd[0-9]+,(?i)nbd[0-9]+"

        # Whether to enable the flex driver. By default it is enabled and is fully supported, but will be deprecated in some future release
        # in favor of the CSI driver.
        - name: ROOK_ENABLE_FLEX_DRIVER
          value: "false"

        # Whether to start the discovery daemon to watch for raw storage devices on nodes in the cluster.
        # This daemon does not need to run if you are only going to create your OSDs based on StorageClassDeviceSets with PVCs.
        - name: ROOK_ENABLE_DISCOVERY_DAEMON
          value: "true"

        # Time to wait until the node controller will move Rook pods to other
        # nodes after detecting an unreachable node.
        # Pods affected by this setting are:
        # mgr, rbd, mds, rgw, nfs, PVC based mons and osds, and ceph toolbox
        # The value used in this variable replaces the default value of 300 secs
        # added automatically by k8s as Toleration for
        # <node.kubernetes.io/unreachable>
        # The total amount of time to reschedule Rook pods in healthy nodes
        # before detecting a <not ready node> condition will be the sum of:
        #  --> node-monitor-grace-period: 40 seconds (k8s kube-controller-manager flag)
        #  --> ROOK_UNREACHABLE_NODE_TOLERATION_SECONDS: 5 seconds
        - name: ROOK_UNREACHABLE_NODE_TOLERATION_SECONDS
          value: "5"

        # The name of the node to pass with the downward API
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        # The pod name to pass with the downward API
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        # The pod namespace to pass with the downward API
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace

        #  Uncomment it to run lib bucket provisioner in multithreaded mode
        #- name: LIB_BUCKET_PROVISIONER_THREADS
        #  value: "5"

      # Uncomment it to run rook operator on the host network
      #hostNetwork: true
      volumes:
      - name: rook-config
        emptyDir: {}
      - name: default-config-dir
        emptyDir: {}
`

	dashboard = `---
apiVersion: v1
kind: Service
metadata:
  name: rook-ceph-mgr-dashboard-external-https
  namespace: rook-ceph
  labels:
    app: rook-ceph-mgr
    rook_cluster: rook-ceph
spec:
  ports:
  - name: dashboard
    port: 8443
    protocol: TCP
    targetPort: 8443
  selector:
    app: rook-ceph-mgr
    rook_cluster: rook-ceph
  sessionAffinity: None
  type: NodePort
`

	//測試ceph自动寻找生成osd过程
	cluster = `---
apiVersion: ceph.rook.io/v1
kind: CephCluster
metadata:
  name: rook-ceph
  namespace: rook-ceph
spec:
  cephVersion:
    # The container image used to launch the Ceph daemon pods (mon, mgr, osd, mds, rgw).
    # v13 is mimic, v14 is nautilus, and v15 is octopus.
    # RECOMMENDATION: In production, use a specific version tag instead of the general v14 flag, which pulls the latest release and could result in different
    # versions running within the cluster. See tags available at https://hub.docker.com/r/ceph/ceph/tags/.
    # If you want to be more precise, you can always use a timestamp tag such ceph/ceph:v15.2.4-20200630
    # This tag might not contain a new Ceph version, just security fixes from the underlying operating system, which will reduce vulnerabilities
    image: ceph/ceph:v15.2.4
    allowUnsupported: false
  # The path on the host where configuration files will be persisted. Must be specified.
  # Important: if you reinstall the cluster, make sure you delete this directory from each host or else the mons will fail to start on the new cluster.
  # In Minikube, the '/data' directory is configured to persist across reboots. Use "/data/rook" in Minikube environment.
  dataDirHostPath: /var/lib/rook
  # Whether or not upgrade should continue even if a check fails
  # This means Ceph's status could be degraded and we don't recommend upgrading but you might decide otherwise
  # Use at your OWN risk
  # To understand Rook's upgrade process of Ceph, read https://rook.io/docs/rook/master/ceph-upgrade.html#ceph-version-upgrades
  skipUpgradeChecks: false
  # Whether or not continue if PGs are not clean during an upgrade
  continueUpgradeAfterChecksEvenIfNotHealthy: false
  # set the amount of mons to be started
  mon:
    count: 3
    allowMultiplePerNode: false
  mgr:
    modules:
    # Several modules should not need to be included in this list. The "dashboard" and "monitoring" modules
    # are already enabled by other settings in the cluster CR and the "rook" module is always enabled.
    - name: pg_autoscaler
      enabled: true
  # enable the ceph dashboard for viewing cluster status
  dashboard:
    enabled: true
    # serve the dashboard under a subpath (useful when you are accessing the dashboard via a reverse proxy)
    # urlPrefix: /ceph-dashboard
    # serve the dashboard at the given port.
    # port: 8443
    # serve the dashboard using SSL
    ssl: true
  # enable prometheus alerting for cluster
  monitoring:
    # requires Prometheus to be pre-installed
    enabled: false
    # namespace to deploy prometheusRule in. If empty, namespace of the cluster will be used.
    # Recommended:
    # If you have a single rook-ceph cluster, set the rulesNamespace to the same namespace as the cluster or keep it empty.
    # If you have multiple rook-ceph clusters in the same k8s cluster, choose the same namespace (ideally, namespace with prometheus
    # deployed) to set rulesNamespace for all the clusters. Otherwise, you will get duplicate alerts with multiple alert definitions.
    rulesNamespace: rook-ceph
  network:
    # enable host networking
    #provider: host
    # EXPERIMENTAL: enable the Multus network provider
    #provider: multus
    # Provide internet protocol version. IPv6, IPv4 or empty string are valid options. Empty string would mean IPv4
    #ipFamily: "IPv6"
  # enable the crash collector for ceph daemon crash collection
  crashCollector:
    disable: false
  cleanupPolicy:
    # cleanup should only be added to the cluster when the cluster is about to be deleted.
    # After any field of the cleanup policy is set, Rook will stop configuring the cluster as if the cluster is about
    # to be destroyed in order to prevent these settings from being deployed unintentionally.
    # To signify that automatic deletion is desired, use the value "yes-really-destroy-data". Only this and an empty
    # string are valid values for this field.
    confirmation: ""
    # sanitizeDisks represents settings for sanitizing OSD disks on cluster deletion
    sanitizeDisks:
      # method indicates if the entire disk should be sanitized or simply ceph's metadata
      # in both case, re-install is possible
      # possible choices are 'complete' or 'quick' (default)
      method: quick
      # dataSource indicate where to get random bytes from to write on the disk
      # possible choices are 'zero' (default) or 'random'
      # using random sources will consume entropy from the system and will take much more time then the zero source
      dataSource: zero
      # iteration overwrite N times instead of the default (1)
      # takes an integer value
      iteration: 1
    # allowUninstallWithVolumes defines how the uninstall should be performed
    # If set to true, cephCluster deletion does not wait for the PVs to be deleted.
    allowUninstallWithVolumes: false
  # To control where various services will be scheduled by kubernetes, use the placement configuration sections below.
  # The example under 'all' would have all services scheduled on kubernetes nodes labeled with 'role=storage-node' and
  # tolerate taints with a key of 'storage-node'.
#  placement:
#    all:
#      nodeAffinity:
#        requiredDuringSchedulingIgnoredDuringExecution:
#          nodeSelectorTerms:
#          - matchExpressions:
#            - key: role
#              operator: In
#              values:
#              - storage-node
#      podAffinity:
#      podAntiAffinity:
#      topologySpreadConstraints:
#      tolerations:
#      - key: storage-node
#        operator: Exists
# The above placement information can also be specified for mon, osd, and mgr components
#    mon:
# Monitor deployments may contain an anti-affinity rule for avoiding monitor
# collocation on the same node. This is a required rule when host network is used
# or when AllowMultiplePerNode is false. Otherwise this anti-affinity rule is a
# preferred rule with weight: 50.
#    osd:
#    mgr:
#    cleanup:
  annotations:
#    all:
#    mon:
#    osd:
#    cleanup:
#    prepareosd:
# If no mgr annotations are set, prometheus scrape annotations will be set by default.
#    mgr:
  labels:
#    all:
#    mon:
#    osd:
#    cleanup:
#    mgr:
#    prepareosd:
  resources:
# The requests and limits set here, allow the mgr pod to use half of one CPU core and 1 gigabyte of memory
#    mgr:
#      limits:
#        cpu: "500m"
#        memory: "1024Mi"
#      requests:
#        cpu: "500m"
#        memory: "1024Mi"
# The above example requests/limits can also be added to the mon and osd components
#    mon:
#    osd:
#    prepareosd:
#    crashcollector:
#    cleanup:
  # The option to automatically remove OSDs that are out and are safe to destroy.
  removeOSDsIfOutAndSafeToRemove: false
#  priorityClassNames:
#    all: rook-ceph-default-priority-class
#    mon: rook-ceph-mon-priority-class
#    osd: rook-ceph-osd-priority-class
#    mgr: rook-ceph-mgr-priority-class
  storage: # cluster level storage configuration and selection
    useAllNodes: false
    useAllDevices: false
    #deviceFilter:
    config:
      # metadataDevice: "md0" # specify a non-rotational storage so ceph-volume will use it as block db device of bluestore.
      # databaseSizeMB: "1024" # uncomment if the disks are smaller than 100 GB
      # journalSizeMB: "1024"  # uncomment if the disks are 20 GB or smaller
      # osdsPerDevice: "1" # this value can be overridden at the node or device level
      # encryptedDevice: "true" # the default value for this option is "false"
# Individual nodes and their config can be specified as well, but 'useAllNodes' above must be set to false. Then, only the named
# nodes below will be used as storage resources.  Each node's 'name' field should match their 'kubernetes.io/hostname' label.
#    nodes:
#    - name: "172.17.4.201"
#      devices: # specific devices to use for storage can be specified for each node
#      - name: "sdb"
#      - name: "nvme01" # multiple osds can be created on high performance devices
#        config:
#          osdsPerDevice: "5"
#      - name: "/dev/disk/by-id/ata-ST4000DM004-XXXX" # devices can be specified using full udev paths
#      config: # configuration can be specified at the node level which overrides the cluster level config
#        storeType: filestore
#    - name: "172.17.4.301"
#      deviceFilter: "^sd."
    nodes:
    - name: "node5"
      devices: # specific devices to use for storage can be specified for each node
      - name: "xvdb"
    - name: "node6"
      devices:
      - name: "xvdb"
    - name: "node7"
      devices:
      - name: "xvdb"
  # The section for configuring management of daemon disruptions during upgrade or fencing.
  disruptionManagement:
    # If true, the operator will create and manage PodDisruptionBudgets for OSD, Mon, RGW, and MDS daemons. OSD PDBs are managed dynamically
    # via the strategy outlined in the [design](https://github.com/rook/rook/blob/master/design/ceph/ceph-managed-disruptionbudgets.md). The operator will
    # block eviction of OSDs by default and unblock them safely when drains are detected.
    managePodBudgets: false
    osdMaintenanceTimeout: 30
    # If true, the operator will create and manage MachineDisruptionBudgets to ensure OSDs are only fenced when the cluster is healthy.
    # Only available on OpenShift.
    manageMachineDisruptionBudgets: false
    # Namespace in which to watch for the MachineDisruptionBudgets.
    machineDisruptionBudgetNamespace: openshift-machine-api

  # healthChecks
  # Valid values for daemons are 'mon', 'osd', 'status'
  healthCheck:
    daemonHealth:
      mon:
        disabled: false
        interval: 45s
      osd:
        disabled: false
        interval: 60s
      status:
        disabled: false
        interval: 60s
    # Change pod liveness probe, it works for all mon,mgr,osd daemons
    livenessProbe:
      mon:
        disabled: false
      mgr:
        disabled: false
      osd:
        disabled: false
`

	toolbox = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rook-ceph-tools
  namespace: rook-ceph
  labels:
    app: rook-ceph-tools
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rook-ceph-tools
  template:
    metadata:
      labels:
        app: rook-ceph-tools
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: rook-ceph-tools
        image: rook/ceph:v1.4.6
        command: ["/tini"]
        args: ["-g", "--", "/usr/local/bin/toolbox.sh"]
        imagePullPolicy: IfNotPresent
        env:
          - name: ROOK_CEPH_USERNAME
            valueFrom:
              secretKeyRef:
                name: rook-ceph-mon
                key: ceph-username
          - name: ROOK_CEPH_SECRET
            valueFrom:
              secretKeyRef:
                name: rook-ceph-mon
                key: ceph-secret
        volumeMounts:
          - mountPath: /etc/ceph
            name: ceph-config
          - name: mon-endpoint-volume
            mountPath: /etc/rook
      volumes:
        - name: mon-endpoint-volume
          configMap:
            name: rook-ceph-mon-endpoints
            items:
            - key: data
              path: mon-endpoints
        - name: ceph-config
          emptyDir: {}
      tolerations:
        - key: "node.kubernetes.io/unreachable"
          operator: "Exists"
          effect: "NoExecute"
          tolerationSeconds: 5
`
)

var (
	cephmaps = map[string]string{"common.yaml": common, "operate.yaml": operate, "dashboard-external-https.yaml": dashboard, "cluster.yaml": cluster, "toolbox": toolbox}

/*	KubeSphereTempl = template.Must(template.New("KubeSphere").Parse(
	dedent.Dedent()))*/
)
