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

package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var RookCrds = template.Must(template.New("crds.yaml").Parse(
	dedent.Dedent(`
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephblockpoolradosnamespaces.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephBlockPoolRadosNamespace
    listKind: CephBlockPoolRadosNamespaceList
    plural: cephblockpoolradosnamespaces
    singular: cephblockpoolradosnamespace
  scope: Namespaced
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: CephBlockPoolRadosNamespace represents a Ceph BlockPool Rados Namespace
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: Spec represents the specification of a Ceph BlockPool Rados Namespace
              properties:
                blockPoolName:
                  description: BlockPoolName is the name of Ceph BlockPool. Typically it's the name of the CephBlockPool CR.
                  type: string
              required:
                - blockPoolName
              type: object
            status:
              description: Status represents the status of a CephBlockPool Rados Namespace
              properties:
                info:
                  additionalProperties:
                    type: string
                  nullable: true
                  type: object
                phase:
                  description: ConditionType represent a resource's status
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephblockpools.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephBlockPool
    listKind: CephBlockPoolList
    plural: cephblockpools
    singular: cephblockpool
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephBlockPool represents a Ceph Storage Pool
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: NamedBlockPoolSpec allows a block pool to be created with a non-default name. This is more specific than the NamedPoolSpec so we get schema validation on the allowed pool names that can be specified.
              properties:
                compressionMode:
                  description: 'DEPRECATED: use Parameters instead, e.g., Parameters["compression_mode"] = "force" The inline compression mode in Bluestore OSD to set to (options are: none, passive, aggressive, force) Do NOT set a default value for kubebuilder as this will override the Parameters'
                  enum:
                    - none
                    - passive
                    - aggressive
                    - force
                    - ""
                  nullable: true
                  type: string
                crushRoot:
                  description: The root of the crush hierarchy utilized by the pool
                  nullable: true
                  type: string
                deviceClass:
                  description: The device class the OSD should set to for use in the pool
                  nullable: true
                  type: string
                enableRBDStats:
                  description: EnableRBDStats is used to enable gathering of statistics for all RBD images in the pool
                  type: boolean
                erasureCoded:
                  description: The erasure code settings
                  properties:
                    algorithm:
                      description: The algorithm for erasure coding
                      type: string
                    codingChunks:
                      description: Number of coding chunks per object in an erasure coded storage pool (required for erasure-coded pool type). This is the number of OSDs that can be lost simultaneously before data cannot be recovered.
                      minimum: 0
                      type: integer
                    dataChunks:
                      description: Number of data chunks per object in an erasure coded storage pool (required for erasure-coded pool type). The number of chunks required to recover an object when any single OSD is lost is the same as dataChunks so be aware that the larger the number of data chunks, the higher the cost of recovery.
                      minimum: 0
                      type: integer
                  required:
                    - codingChunks
                    - dataChunks
                  type: object
                failureDomain:
                  description: 'The failure domain: osd/host/(region or zone if available) - technically also any type in the crush map'
                  type: string
                mirroring:
                  description: The mirroring settings
                  properties:
                    enabled:
                      description: Enabled whether this pool is mirrored or not
                      type: boolean
                    mode:
                      description: 'Mode is the mirroring mode: either pool or image'
                      type: string
                    peers:
                      description: Peers represents the peers spec
                      nullable: true
                      properties:
                        secretNames:
                          description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                          items:
                            type: string
                          type: array
                      type: object
                    snapshotSchedules:
                      description: SnapshotSchedules is the scheduling of snapshot for mirrored images/pools
                      items:
                        description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                        properties:
                          interval:
                            description: Interval represent the periodicity of the snapshot.
                            type: string
                          path:
                            description: Path is the path to snapshot, only valid for CephFS
                            type: string
                          startTime:
                            description: StartTime indicates when to start the snapshot
                            type: string
                        type: object
                      type: array
                  type: object
                name:
                  description: The desired name of the pool if different from the CephBlockPool CR name.
                  enum:
                    - device_health_metrics
                    - .nfs
                    - .mgr
                  type: string
                parameters:
                  additionalProperties:
                    type: string
                  description: Parameters is a list of properties to enable on a given pool
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                quotas:
                  description: The quota settings
                  nullable: true
                  properties:
                    maxBytes:
                      description: MaxBytes represents the quota in bytes Deprecated in favor of MaxSize
                      format: int64
                      type: integer
                    maxObjects:
                      description: MaxObjects represents the quota in objects
                      format: int64
                      type: integer
                    maxSize:
                      description: MaxSize represents the quota in bytes as a string
                      pattern: ^[0-9]+[\.]?[0-9]*([KMGTPE]i|[kMGTPE])?$
                      type: string
                  type: object
                replicated:
                  description: The replication settings
                  properties:
                    hybridStorage:
                      description: HybridStorage represents hybrid storage tier settings
                      nullable: true
                      properties:
                        primaryDeviceClass:
                          description: PrimaryDeviceClass represents high performance tier (for example SSD or NVME) for Primary OSD
                          minLength: 1
                          type: string
                        secondaryDeviceClass:
                          description: SecondaryDeviceClass represents low performance tier (for example HDDs) for remaining OSDs
                          minLength: 1
                          type: string
                      required:
                        - primaryDeviceClass
                        - secondaryDeviceClass
                      type: object
                    replicasPerFailureDomain:
                      description: ReplicasPerFailureDomain the number of replica in the specified failure domain
                      minimum: 1
                      type: integer
                    requireSafeReplicaSize:
                      description: RequireSafeReplicaSize if false allows you to set replica 1
                      type: boolean
                    size:
                      description: Size - Number of copies per object in a replicated storage pool, including the object itself (required for replicated pool type)
                      minimum: 0
                      type: integer
                    subFailureDomain:
                      description: SubFailureDomain the name of the sub-failure domain
                      type: string
                    targetSizeRatio:
                      description: TargetSizeRatio gives a hint (%) to Ceph in terms of expected consumption of the total cluster capacity
                      type: number
                  required:
                    - size
                  type: object
                statusCheck:
                  description: The mirroring statusCheck
                  properties:
                    mirror:
                      description: HealthCheckSpec represents the health check of an object store bucket
                      nullable: true
                      properties:
                        disabled:
                          type: boolean
                        interval:
                          description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                          type: string
                        timeout:
                          type: string
                      type: object
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
              type: object
            status:
              description: CephBlockPoolStatus represents the mirroring status of Ceph Storage Pool
              properties:
                conditions:
                  items:
                    description: Condition represents a status condition on any Rook-Ceph Custom Resource.
                    properties:
                      lastHeartbeatTime:
                        format: date-time
                        type: string
                      lastTransitionTime:
                        format: date-time
                        type: string
                      message:
                        type: string
                      reason:
                        description: ConditionReason is a reason for a condition
                        type: string
                      status:
                        type: string
                      type:
                        description: ConditionType represent a resource's status
                        type: string
                    type: object
                  type: array
                info:
                  additionalProperties:
                    type: string
                  nullable: true
                  type: object
                mirroringInfo:
                  description: MirroringInfoSpec is the status of the pool mirroring
                  properties:
                    details:
                      type: string
                    lastChanged:
                      type: string
                    lastChecked:
                      type: string
                    mode:
                      description: Mode is the mirroring mode
                      type: string
                    peers:
                      description: Peers are the list of peer sites connected to that cluster
                      items:
                        description: PeersSpec contains peer details
                        properties:
                          client_name:
                            description: ClientName is the CephX user used to connect to the peer
                            type: string
                          direction:
                            description: Direction is the peer mirroring direction
                            type: string
                          mirror_uuid:
                            description: MirrorUUID is the mirror UUID
                            type: string
                          site_name:
                            description: SiteName is the current site name
                            type: string
                          uuid:
                            description: UUID is the peer UUID
                            type: string
                        type: object
                      type: array
                    site_name:
                      description: SiteName is the current site name
                      type: string
                  type: object
                mirroringStatus:
                  description: MirroringStatusSpec is the status of the pool mirroring
                  properties:
                    details:
                      description: Details contains potential status errors
                      type: string
                    lastChanged:
                      description: LastChanged is the last time time the status last changed
                      type: string
                    lastChecked:
                      description: LastChecked is the last time time the status was checked
                      type: string
                    summary:
                      description: Summary is the mirroring status summary
                      properties:
                        daemon_health:
                          description: DaemonHealth is the health of the mirroring daemon
                          type: string
                        health:
                          description: Health is the mirroring health
                          type: string
                        image_health:
                          description: ImageHealth is the health of the mirrored image
                          type: string
                        states:
                          description: States is the various state for all mirrored images
                          nullable: true
                          properties:
                            error:
                              description: Error is when the mirroring state is errored
                              type: integer
                            replaying:
                              description: Replaying is when the replay of the mirroring journal is on-going
                              type: integer
                            starting_replay:
                              description: StartingReplay is when the replay of the mirroring journal starts
                              type: integer
                            stopped:
                              description: Stopped is when the mirroring state is stopped
                              type: integer
                            stopping_replay:
                              description: StopReplaying is when the replay of the mirroring journal stops
                              type: integer
                            syncing:
                              description: Syncing is when the image is syncing
                              type: integer
                            unknown:
                              description: Unknown is when the mirroring state is unknown
                              type: integer
                          type: object
                      type: object
                  type: object
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  description: ConditionType represent a resource's status
                  type: string
                snapshotScheduleStatus:
                  description: SnapshotScheduleStatusSpec is the status of the snapshot schedule
                  properties:
                    details:
                      description: Details contains potential status errors
                      type: string
                    lastChanged:
                      description: LastChanged is the last time time the status last changed
                      type: string
                    lastChecked:
                      description: LastChecked is the last time time the status was checked
                      type: string
                    snapshotSchedules:
                      description: SnapshotSchedules is the list of snapshots scheduled
                      items:
                        description: SnapshotSchedulesSpec is the list of snapshot scheduled for images in a pool
                        properties:
                          image:
                            description: Image is the mirrored image
                            type: string
                          items:
                            description: Items is the list schedules times for a given snapshot
                            items:
                              description: SnapshotSchedule is a schedule
                              properties:
                                interval:
                                  description: Interval is the interval in which snapshots will be taken
                                  type: string
                                start_time:
                                  description: StartTime is the snapshot starting time
                                  type: string
                              type: object
                            type: array
                          namespace:
                            description: Namespace is the RADOS namespace the image is part of
                            type: string
                          pool:
                            description: Pool is the pool name
                            type: string
                        type: object
                      nullable: true
                      type: array
                  type: object
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephbucketnotifications.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephBucketNotification
    listKind: CephBucketNotificationList
    plural: cephbucketnotifications
    singular: cephbucketnotification
  scope: Namespaced
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: CephBucketNotification represents a Bucket Notifications
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: BucketNotificationSpec represent the spec of a Bucket Notification
              properties:
                events:
                  description: List of events that should trigger the notification
                  items:
                    description: BucketNotificationSpec represent the event type of the bucket notification
                    enum:
                      - s3:ObjectCreated:*
                      - s3:ObjectCreated:Put
                      - s3:ObjectCreated:Post
                      - s3:ObjectCreated:Copy
                      - s3:ObjectCreated:CompleteMultipartUpload
                      - s3:ObjectRemoved:*
                      - s3:ObjectRemoved:Delete
                      - s3:ObjectRemoved:DeleteMarkerCreated
                    type: string
                  type: array
                filter:
                  description: Spec of notification filter
                  properties:
                    keyFilters:
                      description: Filters based on the object's key
                      items:
                        description: NotificationKeyFilterRule represent a single key rule in the Notification Filter spec
                        properties:
                          name:
                            description: Name of the filter - prefix/suffix/regex
                            enum:
                              - prefix
                              - suffix
                              - regex
                            type: string
                          value:
                            description: Value to filter on
                            type: string
                        required:
                          - name
                          - value
                        type: object
                      type: array
                    metadataFilters:
                      description: Filters based on the object's metadata
                      items:
                        description: NotificationFilterRule represent a single rule in the Notification Filter spec
                        properties:
                          name:
                            description: Name of the metadata or tag
                            minLength: 1
                            type: string
                          value:
                            description: Value to filter on
                            type: string
                        required:
                          - name
                          - value
                        type: object
                      type: array
                    tagFilters:
                      description: Filters based on the object's tags
                      items:
                        description: NotificationFilterRule represent a single rule in the Notification Filter spec
                        properties:
                          name:
                            description: Name of the metadata or tag
                            minLength: 1
                            type: string
                          value:
                            description: Value to filter on
                            type: string
                        required:
                          - name
                          - value
                        type: object
                      type: array
                  type: object
                topic:
                  description: The name of the topic associated with this notification
                  minLength: 1
                  type: string
              required:
                - topic
              type: object
            status:
              description: Status represents the status of an object
              properties:
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephbuckettopics.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephBucketTopic
    listKind: CephBucketTopicList
    plural: cephbuckettopics
    singular: cephbuckettopic
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephBucketTopic represents a Ceph Object Topic for Bucket Notifications
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: BucketTopicSpec represent the spec of a Bucket Topic
              properties:
                endpoint:
                  description: Contains the endpoint spec of the topic
                  properties:
                    amqp:
                      description: Spec of AMQP endpoint
                      properties:
                        ackLevel:
                          default: broker
                          description: The ack level required for this topic (none/broker/routeable)
                          enum:
                            - none
                            - broker
                            - routeable
                          type: string
                        disableVerifySSL:
                          description: Indicate whether the server certificate is validated by the client or not
                          type: boolean
                        exchange:
                          description: Name of the exchange that is used to route messages based on topics
                          minLength: 1
                          type: string
                        uri:
                          description: The URI of the AMQP endpoint to push notification to
                          minLength: 1
                          type: string
                      required:
                        - exchange
                        - uri
                      type: object
                    http:
                      description: Spec of HTTP endpoint
                      properties:
                        disableVerifySSL:
                          description: Indicate whether the server certificate is validated by the client or not
                          type: boolean
                        sendCloudEvents:
                          description: 'Send the notifications with the CloudEvents header: https://github.com/cloudevents/spec/blob/main/cloudevents/adapters/aws-s3.md Supported for Ceph Quincy (v17) or newer.'
                          type: boolean
                        uri:
                          description: The URI of the HTTP endpoint to push notification to
                          minLength: 1
                          type: string
                      required:
                        - uri
                      type: object
                    kafka:
                      description: Spec of Kafka endpoint
                      properties:
                        ackLevel:
                          default: broker
                          description: The ack level required for this topic (none/broker)
                          enum:
                            - none
                            - broker
                          type: string
                        disableVerifySSL:
                          description: Indicate whether the server certificate is validated by the client or not
                          type: boolean
                        uri:
                          description: The URI of the Kafka endpoint to push notification to
                          minLength: 1
                          type: string
                        useSSL:
                          description: Indicate whether to use SSL when communicating with the broker
                          type: boolean
                      required:
                        - uri
                      type: object
                  type: object
                objectStoreName:
                  description: The name of the object store on which to define the topic
                  minLength: 1
                  type: string
                objectStoreNamespace:
                  description: The namespace of the object store on which to define the topic
                  minLength: 1
                  type: string
                opaqueData:
                  description: Data which is sent in each event
                  type: string
                persistent:
                  description: Indication whether notifications to this endpoint are persistent or not
                  type: boolean
              required:
                - endpoint
                - objectStoreName
                - objectStoreNamespace
              type: object
            status:
              description: BucketTopicStatus represents the Status of a CephBucketTopic
              properties:
                ARN:
                  description: The ARN of the topic generated by the RGW
                  nullable: true
                  type: string
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephclients.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephClient
    listKind: CephClientList
    plural: cephclients
    singular: cephclient
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephClient represents a Ceph Client
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: Spec represents the specification of a Ceph Client
              properties:
                caps:
                  additionalProperties:
                    type: string
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                name:
                  type: string
              required:
                - caps
              type: object
            status:
              description: Status represents the status of a Ceph Client
              properties:
                info:
                  additionalProperties:
                    type: string
                  nullable: true
                  type: object
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  description: ConditionType represent a resource's status
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephclusters.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephCluster
    listKind: CephClusterList
    plural: cephclusters
    singular: cephcluster
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - description: Directory used on the K8s nodes
          jsonPath: .spec.dataDirHostPath
          name: DataDirHostPath
          type: string
        - description: Number of MONs
          jsonPath: .spec.mon.count
          name: MonCount
          type: string
        - jsonPath: .metadata.creationTimestamp
          name: Age
          type: date
        - jsonPath: .status.phase
          name: Phase
          type: string
        - description: Message
          jsonPath: .status.message
          name: Message
          type: string
        - description: Ceph Health
          jsonPath: .status.ceph.health
          name: Health
          type: string
        - jsonPath: .spec.external.enable
          name: External
          type: boolean
      name: v1
      schema:
        openAPIV3Schema:
          description: CephCluster is a Ceph storage cluster
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: ClusterSpec represents the specification of Ceph Cluster
              properties:
                annotations:
                  additionalProperties:
                    additionalProperties:
                      type: string
                    description: Annotations are annotations
                    type: object
                  description: The annotations-related configuration to add/set on each Pod related object.
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                cephVersion:
                  description: The version information that instructs Rook to orchestrate a particular version of Ceph.
                  nullable: true
                  properties:
                    allowUnsupported:
                      description: Whether to allow unsupported versions (do not set to true in production)
                      type: boolean
                    image:
                      description: Image is the container image used to launch the ceph daemons, such as quay.io/ceph/ceph:<tag> The full list of images can be found at https://quay.io/repository/ceph/ceph?tab=tags
                      type: string
                  type: object
                cleanupPolicy:
                  description: Indicates user intent when deleting a cluster; blocks orchestration and should not be set if cluster deletion is not imminent.
                  nullable: true
                  properties:
                    allowUninstallWithVolumes:
                      description: AllowUninstallWithVolumes defines whether we can proceed with the uninstall if they are RBD images still present
                      type: boolean
                    confirmation:
                      description: Confirmation represents the cleanup confirmation
                      nullable: true
                      pattern: ^$|^yes-really-destroy-data$
                      type: string
                    sanitizeDisks:
                      description: SanitizeDisks represents way we sanitize disks
                      nullable: true
                      properties:
                        dataSource:
                          description: DataSource is the data source to use to sanitize the disk with
                          enum:
                            - zero
                            - random
                          type: string
                        iteration:
                          description: Iteration is the number of pass to apply the sanitizing
                          format: int32
                          type: integer
                        method:
                          description: Method is the method we use to sanitize disks
                          enum:
                            - complete
                            - quick
                          type: string
                      type: object
                  type: object
                continueUpgradeAfterChecksEvenIfNotHealthy:
                  description: ContinueUpgradeAfterChecksEvenIfNotHealthy defines if an upgrade should continue even if PGs are not clean
                  type: boolean
                crashCollector:
                  description: A spec for the crash controller
                  nullable: true
                  properties:
                    daysToRetain:
                      description: DaysToRetain represents the number of days to retain crash until they get pruned
                      type: integer
                    disable:
                      description: Disable determines whether we should enable the crash collector
                      type: boolean
                  type: object
                dashboard:
                  description: Dashboard settings
                  nullable: true
                  properties:
                    enabled:
                      description: Enabled determines whether to enable the dashboard
                      type: boolean
                    port:
                      description: Port is the dashboard webserver port
                      maximum: 65535
                      minimum: 0
                      type: integer
                    ssl:
                      description: SSL determines whether SSL should be used
                      type: boolean
                    urlPrefix:
                      description: URLPrefix is a prefix for all URLs to use the dashboard with a reverse proxy
                      type: string
                  type: object
                dataDirHostPath:
                  description: The path on the host where config and data can be persisted
                  pattern: ^/(\S+)
                  type: string
                disruptionManagement:
                  description: A spec for configuring disruption management.
                  nullable: true
                  properties:
                    machineDisruptionBudgetNamespace:
                      description: Namespace to look for MDBs by the machineDisruptionBudgetController
                      type: string
                    manageMachineDisruptionBudgets:
                      description: This enables management of machinedisruptionbudgets
                      type: boolean
                    managePodBudgets:
                      description: This enables management of poddisruptionbudgets
                      type: boolean
                    osdMaintenanceTimeout:
                      description: OSDMaintenanceTimeout sets how many additional minutes the DOWN/OUT interval is for drained failure domains it only works if managePodBudgets is true. the default is 30 minutes
                      format: int64
                      type: integer
                    pgHealthCheckTimeout:
                      description: PGHealthCheckTimeout is the time (in minutes) that the operator will wait for the placement groups to become healthy (active+clean) after a drain was completed and OSDs came back up. Rook will continue with the next drain if the timeout exceeds. It only works if managePodBudgets is true. No values or 0 means that the operator will wait until the placement groups are healthy before unblocking the next drain.
                      format: int64
                      type: integer
                  type: object
                external:
                  description: Whether the Ceph Cluster is running external to this Kubernetes cluster mon, mgr, osd, mds, and discover daemons will not be created for external clusters.
                  nullable: true
                  properties:
                    enable:
                      description: Enable determines whether external mode is enabled or not
                      type: boolean
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                healthCheck:
                  description: Internal daemon healthchecks and liveness probe
                  nullable: true
                  properties:
                    daemonHealth:
                      description: DaemonHealth is the health check for a given daemon
                      nullable: true
                      properties:
                        mon:
                          description: Monitor represents the health check settings for the Ceph monitor
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                        osd:
                          description: ObjectStorageDaemon represents the health check settings for the Ceph OSDs
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                        status:
                          description: Status represents the health check settings for the Ceph health
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                      type: object
                    livenessProbe:
                      additionalProperties:
                        description: ProbeSpec is a wrapper around Probe so it can be enabled or disabled for a Ceph daemon
                        properties:
                          disabled:
                            description: Disabled determines whether probe is disable or not
                            type: boolean
                          probe:
                            description: Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
                            properties:
                              exec:
                                description: Exec specifies the action to take.
                                properties:
                                  command:
                                    description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                    items:
                                      type: string
                                    type: array
                                type: object
                              failureThreshold:
                                description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                format: int32
                                type: integer
                              grpc:
                                description: GRPC specifies an action involving a GRPC port. This is an alpha field and requires enabling GRPCContainerProbe feature gate.
                                properties:
                                  port:
                                    description: Port number of the gRPC service. Number must be in the range 1 to 65535.
                                    format: int32
                                    type: integer
                                  service:
                                    description: "Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). \n If this is not specified, the default behavior is defined by gRPC."
                                    type: string
                                required:
                                  - port
                                type: object
                              httpGet:
                                description: HTTPGet specifies the http request to perform.
                                properties:
                                  host:
                                    description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                    type: string
                                  httpHeaders:
                                    description: Custom headers to set in the request. HTTP allows repeated headers.
                                    items:
                                      description: HTTPHeader describes a custom header to be used in HTTP probes
                                      properties:
                                        name:
                                          description: The header field name
                                          type: string
                                        value:
                                          description: The header field value
                                          type: string
                                      required:
                                        - name
                                        - value
                                      type: object
                                    type: array
                                  path:
                                    description: Path to access on the HTTP server.
                                    type: string
                                  port:
                                    anyOf:
                                      - type: integer
                                      - type: string
                                    description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                  scheme:
                                    description: Scheme to use for connecting to the host. Defaults to HTTP.
                                    type: string
                                required:
                                  - port
                                type: object
                              initialDelaySeconds:
                                description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                              periodSeconds:
                                description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                format: int32
                                type: integer
                              successThreshold:
                                description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
                                format: int32
                                type: integer
                              tcpSocket:
                                description: TCPSocket specifies an action involving a TCP port.
                                properties:
                                  host:
                                    description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                    type: string
                                  port:
                                    anyOf:
                                      - type: integer
                                      - type: string
                                    description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                required:
                                  - port
                                type: object
                              terminationGracePeriodSeconds:
                                description: Optional duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this value overrides the value provided by the pod spec. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate. Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
                                format: int64
                                type: integer
                              timeoutSeconds:
                                description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                            type: object
                        type: object
                      description: LivenessProbe allows changing the livenessProbe configuration for a given daemon
                      type: object
                    startupProbe:
                      additionalProperties:
                        description: ProbeSpec is a wrapper around Probe so it can be enabled or disabled for a Ceph daemon
                        properties:
                          disabled:
                            description: Disabled determines whether probe is disable or not
                            type: boolean
                          probe:
                            description: Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
                            properties:
                              exec:
                                description: Exec specifies the action to take.
                                properties:
                                  command:
                                    description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                    items:
                                      type: string
                                    type: array
                                type: object
                              failureThreshold:
                                description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                                format: int32
                                type: integer
                              grpc:
                                description: GRPC specifies an action involving a GRPC port. This is an alpha field and requires enabling GRPCContainerProbe feature gate.
                                properties:
                                  port:
                                    description: Port number of the gRPC service. Number must be in the range 1 to 65535.
                                    format: int32
                                    type: integer
                                  service:
                                    description: "Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). \n If this is not specified, the default behavior is defined by gRPC."
                                    type: string
                                required:
                                  - port
                                type: object
                              httpGet:
                                description: HTTPGet specifies the http request to perform.
                                properties:
                                  host:
                                    description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                    type: string
                                  httpHeaders:
                                    description: Custom headers to set in the request. HTTP allows repeated headers.
                                    items:
                                      description: HTTPHeader describes a custom header to be used in HTTP probes
                                      properties:
                                        name:
                                          description: The header field name
                                          type: string
                                        value:
                                          description: The header field value
                                          type: string
                                      required:
                                        - name
                                        - value
                                      type: object
                                    type: array
                                  path:
                                    description: Path to access on the HTTP server.
                                    type: string
                                  port:
                                    anyOf:
                                      - type: integer
                                      - type: string
                                    description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                  scheme:
                                    description: Scheme to use for connecting to the host. Defaults to HTTP.
                                    type: string
                                required:
                                  - port
                                type: object
                              initialDelaySeconds:
                                description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                              periodSeconds:
                                description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                                format: int32
                                type: integer
                              successThreshold:
                                description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
                                format: int32
                                type: integer
                              tcpSocket:
                                description: TCPSocket specifies an action involving a TCP port.
                                properties:
                                  host:
                                    description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                    type: string
                                  port:
                                    anyOf:
                                      - type: integer
                                      - type: string
                                    description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                    x-kubernetes-int-or-string: true
                                required:
                                  - port
                                type: object
                              terminationGracePeriodSeconds:
                                description: Optional duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this value overrides the value provided by the pod spec. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate. Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
                                format: int64
                                type: integer
                              timeoutSeconds:
                                description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                format: int32
                                type: integer
                            type: object
                        type: object
                      description: StartupProbe allows changing the startupProbe configuration for a given daemon
                      type: object
                  type: object
                labels:
                  additionalProperties:
                    additionalProperties:
                      type: string
                    description: Labels are label for a given daemons
                    type: object
                  description: The labels-related configuration to add/set on each Pod related object.
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                logCollector:
                  description: Logging represents loggings settings
                  nullable: true
                  properties:
                    enabled:
                      description: Enabled represents whether the log collector is enabled
                      type: boolean
                    maxLogSize:
                      anyOf:
                        - type: integer
                        - type: string
                      description: MaxLogSize is the maximum size of the log per ceph daemons. Must be at least 1M.
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                    periodicity:
                      description: Periodicity is the periodicity of the log rotation.
                      pattern: ^$|^(hourly|daily|weekly|monthly|1h|24h|1d)$
                      type: string
                  type: object
                mgr:
                  description: A spec for mgr related options
                  nullable: true
                  properties:
                    allowMultiplePerNode:
                      description: AllowMultiplePerNode allows to run multiple managers on the same node (not recommended)
                      type: boolean
                    count:
                      description: Count is the number of manager to run
                      maximum: 2
                      minimum: 0
                      type: integer
                    modules:
                      description: Modules is the list of ceph manager modules to enable/disable
                      items:
                        description: Module represents mgr modules that the user wants to enable or disable
                        properties:
                          enabled:
                            description: Enabled determines whether a module should be enabled or not
                            type: boolean
                          name:
                            description: Name is the name of the ceph manager module
                            type: string
                        type: object
                      nullable: true
                      type: array
                  type: object
                mon:
                  description: A spec for mon related options
                  nullable: true
                  properties:
                    allowMultiplePerNode:
                      description: AllowMultiplePerNode determines if we can run multiple monitors on the same node (not recommended)
                      type: boolean
                    count:
                      description: Count is the number of Ceph monitors
                      maximum: 9
                      minimum: 0
                      type: integer
                    stretchCluster:
                      description: StretchCluster is the stretch cluster specification
                      properties:
                        failureDomainLabel:
                          description: 'FailureDomainLabel the failure domain name (e,g: zone)'
                          type: string
                        subFailureDomain:
                          description: SubFailureDomain is the failure domain within a zone
                          type: string
                        zones:
                          description: Zones is the list of zones
                          items:
                            description: StretchClusterZoneSpec represents the specification of a stretched zone in a Ceph Cluster
                            properties:
                              arbiter:
                                description: Arbiter determines if the zone contains the arbiter
                                type: boolean
                              name:
                                description: Name is the name of the zone
                                type: string
                              volumeClaimTemplate:
                                description: VolumeClaimTemplate is the PVC template
                                properties:
                                  apiVersion:
                                    description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
                                    type: string
                                  kind:
                                    description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                                    type: string
                                  metadata:
                                    description: 'Standard object''s metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata'
                                    properties:
                                      annotations:
                                        additionalProperties:
                                          type: string
                                        type: object
                                      finalizers:
                                        items:
                                          type: string
                                        type: array
                                      labels:
                                        additionalProperties:
                                          type: string
                                        type: object
                                      name:
                                        type: string
                                      namespace:
                                        type: string
                                    type: object
                                  spec:
                                    description: 'Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                                    properties:
                                      accessModes:
                                        description: 'AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                        items:
                                          type: string
                                        type: array
                                      dataSource:
                                        description: 'This field can be used to specify either: * An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An existing PVC (PersistentVolumeClaim) If the provisioner or an external controller can support the specified data source, it will create a new volume based on the contents of the specified data source. If the AnyVolumeDataSource feature gate is enabled, this field will always have the same contents as the DataSourceRef field.'
                                        properties:
                                          apiGroup:
                                            description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                            type: string
                                          kind:
                                            description: Kind is the type of resource being referenced
                                            type: string
                                          name:
                                            description: Name is the name of resource being referenced
                                            type: string
                                        required:
                                          - kind
                                          - name
                                        type: object
                                      dataSourceRef:
                                        description: 'Specifies the object from which to populate the volume with data, if a non-empty volume is desired. This may be any local object from a non-empty API group (non core object) or a PersistentVolumeClaim object. When this field is specified, volume binding will only succeed if the type of the specified object matches some installed volume populator or dynamic provisioner. This field will replace the functionality of the DataSource field and as such if both fields are non-empty, they must have the same value. For backwards compatibility, both fields (DataSource and DataSourceRef) will be set to the same value automatically if one of them is empty and the other is non-empty. There are two important differences between DataSource and DataSourceRef: * While DataSource only allows two specific types of objects, DataSourceRef   allows any non-core object, as well as PersistentVolumeClaim objects. * While DataSource ignores disallowed values (dropping them), DataSourceRef   preserves all values, and generates an error if a disallowed value is   specified. (Alpha) Using this field requires the AnyVolumeDataSource feature gate to be enabled.'
                                        properties:
                                          apiGroup:
                                            description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                            type: string
                                          kind:
                                            description: Kind is the type of resource being referenced
                                            type: string
                                          name:
                                            description: Name is the name of resource being referenced
                                            type: string
                                        required:
                                          - kind
                                          - name
                                        type: object
                                      resources:
                                        description: 'Resources represents the minimum resources the volume should have. If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements that are lower than previous value but must still be higher than capacity recorded in the status field of the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources'
                                        properties:
                                          limits:
                                            additionalProperties:
                                              anyOf:
                                                - type: integer
                                                - type: string
                                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                              x-kubernetes-int-or-string: true
                                            description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                            type: object
                                          requests:
                                            additionalProperties:
                                              anyOf:
                                                - type: integer
                                                - type: string
                                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                              x-kubernetes-int-or-string: true
                                            description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                            type: object
                                        type: object
                                      selector:
                                        description: A label query over volumes to consider for binding.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      storageClassName:
                                        description: 'Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1'
                                        type: string
                                      volumeMode:
                                        description: volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.
                                        type: string
                                      volumeName:
                                        description: VolumeName is the binding reference to the PersistentVolume backing this claim.
                                        type: string
                                    type: object
                                  status:
                                    description: 'Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                                    properties:
                                      accessModes:
                                        description: 'AccessModes contains the actual access modes the volume backing the PVC has. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                        items:
                                          type: string
                                        type: array
                                      allocatedResources:
                                        additionalProperties:
                                          anyOf:
                                            - type: integer
                                            - type: string
                                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                          x-kubernetes-int-or-string: true
                                        description: The storage resource within AllocatedResources tracks the capacity allocated to a PVC. It may be larger than the actual capacity when a volume expansion operation is requested. For storage quota, the larger value from allocatedResources and PVC.spec.resources is used. If allocatedResources is not set, PVC.spec.resources alone is used for quota calculation. If a volume expansion capacity request is lowered, allocatedResources is only lowered if there are no expansion operations in progress and if the actual volume capacity is equal or lower than the requested capacity. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                        type: object
                                      capacity:
                                        additionalProperties:
                                          anyOf:
                                            - type: integer
                                            - type: string
                                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                          x-kubernetes-int-or-string: true
                                        description: Represents the actual resources of the underlying volume.
                                        type: object
                                      conditions:
                                        description: Current Condition of persistent volume claim. If underlying persistent volume is being resized then the Condition will be set to 'ResizeStarted'.
                                        items:
                                          description: PersistentVolumeClaimCondition contails details about state of pvc
                                          properties:
                                            lastProbeTime:
                                              description: Last time we probed the condition.
                                              format: date-time
                                              type: string
                                            lastTransitionTime:
                                              description: Last time the condition transitioned from one status to another.
                                              format: date-time
                                              type: string
                                            message:
                                              description: Human-readable message indicating details about last transition.
                                              type: string
                                            reason:
                                              description: Unique, this should be a short, machine understandable string that gives the reason for condition's last transition. If it reports "ResizeStarted" that means the underlying persistent volume is being resized.
                                              type: string
                                            status:
                                              type: string
                                            type:
                                              description: PersistentVolumeClaimConditionType is a valid value of PersistentVolumeClaimCondition.Type
                                              type: string
                                          required:
                                            - status
                                            - type
                                          type: object
                                        type: array
                                      phase:
                                        description: Phase represents the current phase of PersistentVolumeClaim.
                                        type: string
                                      resizeStatus:
                                        description: ResizeStatus stores status of resize operation. ResizeStatus is not set by default but when expansion is complete resizeStatus is set to empty string by resize controller or kubelet. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                        type: string
                                    type: object
                                type: object
                                x-kubernetes-preserve-unknown-fields: true
                            type: object
                          nullable: true
                          type: array
                      type: object
                    volumeClaimTemplate:
                      description: VolumeClaimTemplate is the PVC definition
                      properties:
                        apiVersion:
                          description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
                          type: string
                        kind:
                          description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                          type: string
                        metadata:
                          description: 'Standard object''s metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata'
                          properties:
                            annotations:
                              additionalProperties:
                                type: string
                              type: object
                            finalizers:
                              items:
                                type: string
                              type: array
                            labels:
                              additionalProperties:
                                type: string
                              type: object
                            name:
                              type: string
                            namespace:
                              type: string
                          type: object
                        spec:
                          description: 'Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                          properties:
                            accessModes:
                              description: 'AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                              items:
                                type: string
                              type: array
                            dataSource:
                              description: 'This field can be used to specify either: * An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An existing PVC (PersistentVolumeClaim) If the provisioner or an external controller can support the specified data source, it will create a new volume based on the contents of the specified data source. If the AnyVolumeDataSource feature gate is enabled, this field will always have the same contents as the DataSourceRef field.'
                              properties:
                                apiGroup:
                                  description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                  type: string
                                kind:
                                  description: Kind is the type of resource being referenced
                                  type: string
                                name:
                                  description: Name is the name of resource being referenced
                                  type: string
                              required:
                                - kind
                                - name
                              type: object
                            dataSourceRef:
                              description: 'Specifies the object from which to populate the volume with data, if a non-empty volume is desired. This may be any local object from a non-empty API group (non core object) or a PersistentVolumeClaim object. When this field is specified, volume binding will only succeed if the type of the specified object matches some installed volume populator or dynamic provisioner. This field will replace the functionality of the DataSource field and as such if both fields are non-empty, they must have the same value. For backwards compatibility, both fields (DataSource and DataSourceRef) will be set to the same value automatically if one of them is empty and the other is non-empty. There are two important differences between DataSource and DataSourceRef: * While DataSource only allows two specific types of objects, DataSourceRef   allows any non-core object, as well as PersistentVolumeClaim objects. * While DataSource ignores disallowed values (dropping them), DataSourceRef   preserves all values, and generates an error if a disallowed value is   specified. (Alpha) Using this field requires the AnyVolumeDataSource feature gate to be enabled.'
                              properties:
                                apiGroup:
                                  description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                  type: string
                                kind:
                                  description: Kind is the type of resource being referenced
                                  type: string
                                name:
                                  description: Name is the name of resource being referenced
                                  type: string
                              required:
                                - kind
                                - name
                              type: object
                            resources:
                              description: 'Resources represents the minimum resources the volume should have. If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements that are lower than previous value but must still be higher than capacity recorded in the status field of the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources'
                              properties:
                                limits:
                                  additionalProperties:
                                    anyOf:
                                      - type: integer
                                      - type: string
                                    pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                    x-kubernetes-int-or-string: true
                                  description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                  type: object
                                requests:
                                  additionalProperties:
                                    anyOf:
                                      - type: integer
                                      - type: string
                                    pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                    x-kubernetes-int-or-string: true
                                  description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                  type: object
                              type: object
                            selector:
                              description: A label query over volumes to consider for binding.
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                  items:
                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                    properties:
                                      key:
                                        description: key is the label key that the selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                        items:
                                          type: string
                                        type: array
                                    required:
                                      - key
                                      - operator
                                    type: object
                                  type: array
                                matchLabels:
                                  additionalProperties:
                                    type: string
                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                  type: object
                              type: object
                            storageClassName:
                              description: 'Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1'
                              type: string
                            volumeMode:
                              description: volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.
                              type: string
                            volumeName:
                              description: VolumeName is the binding reference to the PersistentVolume backing this claim.
                              type: string
                          type: object
                        status:
                          description: 'Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                          properties:
                            accessModes:
                              description: 'AccessModes contains the actual access modes the volume backing the PVC has. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                              items:
                                type: string
                              type: array
                            allocatedResources:
                              additionalProperties:
                                anyOf:
                                  - type: integer
                                  - type: string
                                pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                x-kubernetes-int-or-string: true
                              description: The storage resource within AllocatedResources tracks the capacity allocated to a PVC. It may be larger than the actual capacity when a volume expansion operation is requested. For storage quota, the larger value from allocatedResources and PVC.spec.resources is used. If allocatedResources is not set, PVC.spec.resources alone is used for quota calculation. If a volume expansion capacity request is lowered, allocatedResources is only lowered if there are no expansion operations in progress and if the actual volume capacity is equal or lower than the requested capacity. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                              type: object
                            capacity:
                              additionalProperties:
                                anyOf:
                                  - type: integer
                                  - type: string
                                pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                x-kubernetes-int-or-string: true
                              description: Represents the actual resources of the underlying volume.
                              type: object
                            conditions:
                              description: Current Condition of persistent volume claim. If underlying persistent volume is being resized then the Condition will be set to 'ResizeStarted'.
                              items:
                                description: PersistentVolumeClaimCondition contails details about state of pvc
                                properties:
                                  lastProbeTime:
                                    description: Last time we probed the condition.
                                    format: date-time
                                    type: string
                                  lastTransitionTime:
                                    description: Last time the condition transitioned from one status to another.
                                    format: date-time
                                    type: string
                                  message:
                                    description: Human-readable message indicating details about last transition.
                                    type: string
                                  reason:
                                    description: Unique, this should be a short, machine understandable string that gives the reason for condition's last transition. If it reports "ResizeStarted" that means the underlying persistent volume is being resized.
                                    type: string
                                  status:
                                    type: string
                                  type:
                                    description: PersistentVolumeClaimConditionType is a valid value of PersistentVolumeClaimCondition.Type
                                    type: string
                                required:
                                  - status
                                  - type
                                type: object
                              type: array
                            phase:
                              description: Phase represents the current phase of PersistentVolumeClaim.
                              type: string
                            resizeStatus:
                              description: ResizeStatus stores status of resize operation. ResizeStatus is not set by default but when expansion is complete resizeStatus is set to empty string by resize controller or kubelet. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                              type: string
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                  type: object
                monitoring:
                  description: Prometheus based Monitoring settings
                  nullable: true
                  properties:
                    enabled:
                      description: Enabled determines whether to create the prometheus rules for the ceph cluster. If true, the prometheus types must exist or the creation will fail.
                      type: boolean
                    externalMgrEndpoints:
                      description: ExternalMgrEndpoints points to an existing Ceph prometheus exporter endpoint
                      items:
                        description: EndpointAddress is a tuple that describes single IP address.
                        properties:
                          hostname:
                            description: The Hostname of this endpoint
                            type: string
                          ip:
                            description: 'The IP of this endpoint. May not be loopback (127.0.0.0/8), link-local (169.254.0.0/16), or link-local multicast ((224.0.0.0/24). IPv6 is also accepted but not fully supported on all platforms. Also, certain kubernetes components, like kube-proxy, are not IPv6 ready. TODO: This should allow hostname or IP, See #4447.'
                            type: string
                          nodeName:
                            description: 'Optional: Node hosting this endpoint. This can be used to determine endpoints local to a node.'
                            type: string
                          targetRef:
                            description: Reference to object providing the endpoint.
                            properties:
                              apiVersion:
                                description: API version of the referent.
                                type: string
                              fieldPath:
                                description: 'If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.'
                                type: string
                              kind:
                                description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                                type: string
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names'
                                type: string
                              namespace:
                                description: 'Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/'
                                type: string
                              resourceVersion:
                                description: 'Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency'
                                type: string
                              uid:
                                description: 'UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids'
                                type: string
                            type: object
                        required:
                          - ip
                        type: object
                      nullable: true
                      type: array
                    externalMgrPrometheusPort:
                      description: ExternalMgrPrometheusPort Prometheus exporter port
                      maximum: 65535
                      minimum: 0
                      type: integer
                  type: object
                network:
                  description: Network related configuration
                  nullable: true
                  properties:
                    connections:
                      description: Settings for network connections such as compression and encryption across the wire.
                      nullable: true
                      properties:
                        compression:
                          description: Compression settings for the network connections.
                          nullable: true
                          properties:
                            enabled:
                              description: Whether to compress the data in transit across the wire. The default is not set. Requires Ceph Quincy (v17) or newer.
                              type: boolean
                          type: object
                        encryption:
                          description: Encryption settings for the network connections.
                          nullable: true
                          properties:
                            enabled:
                              description: Whether to encrypt the data in transit across the wire to prevent eavesdropping the data on the network. The default is not set. Even if encryption is not enabled, clients still establish a strong initial authentication for the connection and data integrity is still validated with a crc check. When encryption is enabled, all communication between clients and Ceph daemons, or between Ceph daemons will be encrypted.
                              type: boolean
                          type: object
                      type: object
                    dualStack:
                      description: DualStack determines whether Ceph daemons should listen on both IPv4 and IPv6
                      type: boolean
                    hostNetwork:
                      description: HostNetwork to enable host network
                      type: boolean
                    ipFamily:
                      description: IPFamily is the single stack IPv6 or IPv4 protocol
                      enum:
                        - IPv4
                        - IPv6
                      nullable: true
                      type: string
                    provider:
                      description: Provider is what provides network connectivity to the cluster e.g. "host" or "multus"
                      nullable: true
                      type: string
                    selectors:
                      additionalProperties:
                        type: string
                      description: Selectors string values describe what networks will be used to connect the cluster. Meanwhile the keys describe each network respective responsibilities or any metadata storage provider decide.
                      nullable: true
                      type: object
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                placement:
                  additionalProperties:
                    description: Placement is the placement for an object
                    properties:
                      nodeAffinity:
                        description: NodeAffinity is a group of node affinity scheduling rules
                        properties:
                          preferredDuringSchedulingIgnoredDuringExecution:
                            description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                            items:
                              description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                              properties:
                                preference:
                                  description: A node selector term, associated with the corresponding weight.
                                  properties:
                                    matchExpressions:
                                      description: A list of node selector requirements by node's labels.
                                      items:
                                        description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: The label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                            type: string
                                          values:
                                            description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                    matchFields:
                                      description: A list of node selector requirements by node's fields.
                                      items:
                                        description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: The label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                            type: string
                                          values:
                                            description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                  type: object
                                weight:
                                  description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                  format: int32
                                  type: integer
                              required:
                                - preference
                                - weight
                              type: object
                            type: array
                          requiredDuringSchedulingIgnoredDuringExecution:
                            description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                            properties:
                              nodeSelectorTerms:
                                description: Required. A list of node selector terms. The terms are ORed.
                                items:
                                  description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                  properties:
                                    matchExpressions:
                                      description: A list of node selector requirements by node's labels.
                                      items:
                                        description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: The label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                            type: string
                                          values:
                                            description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                    matchFields:
                                      description: A list of node selector requirements by node's fields.
                                      items:
                                        description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: The label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                            type: string
                                          values:
                                            description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                  type: object
                                type: array
                            required:
                              - nodeSelectorTerms
                            type: object
                        type: object
                      podAffinity:
                        description: PodAffinity is a group of inter pod affinity scheduling rules
                        properties:
                          preferredDuringSchedulingIgnoredDuringExecution:
                            description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                            items:
                              description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                              properties:
                                podAffinityTerm:
                                  description: Required. A pod affinity term, associated with the corresponding weight.
                                  properties:
                                    labelSelector:
                                      description: A label query over a set of resources, in this case pods.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    namespaceSelector:
                                      description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    namespaces:
                                      description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                      items:
                                        type: string
                                      type: array
                                    topologyKey:
                                      description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                      type: string
                                  required:
                                    - topologyKey
                                  type: object
                                weight:
                                  description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                  format: int32
                                  type: integer
                              required:
                                - podAffinityTerm
                                - weight
                              type: object
                            type: array
                          requiredDuringSchedulingIgnoredDuringExecution:
                            description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                            items:
                              description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                              properties:
                                labelSelector:
                                  description: A label query over a set of resources, in this case pods.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                      items:
                                        description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                namespaceSelector:
                                  description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                      items:
                                        description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                namespaces:
                                  description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                  items:
                                    type: string
                                  type: array
                                topologyKey:
                                  description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                  type: string
                              required:
                                - topologyKey
                              type: object
                            type: array
                        type: object
                      podAntiAffinity:
                        description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                        properties:
                          preferredDuringSchedulingIgnoredDuringExecution:
                            description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                            items:
                              description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                              properties:
                                podAffinityTerm:
                                  description: Required. A pod affinity term, associated with the corresponding weight.
                                  properties:
                                    labelSelector:
                                      description: A label query over a set of resources, in this case pods.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    namespaceSelector:
                                      description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    namespaces:
                                      description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                      items:
                                        type: string
                                      type: array
                                    topologyKey:
                                      description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                      type: string
                                  required:
                                    - topologyKey
                                  type: object
                                weight:
                                  description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                  format: int32
                                  type: integer
                              required:
                                - podAffinityTerm
                                - weight
                              type: object
                            type: array
                          requiredDuringSchedulingIgnoredDuringExecution:
                            description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                            items:
                              description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                              properties:
                                labelSelector:
                                  description: A label query over a set of resources, in this case pods.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                      items:
                                        description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                namespaceSelector:
                                  description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                      items:
                                        description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                        properties:
                                          key:
                                            description: key is the label key that the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                            items:
                                              type: string
                                            type: array
                                        required:
                                          - key
                                          - operator
                                        type: object
                                      type: array
                                    matchLabels:
                                      additionalProperties:
                                        type: string
                                      description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                      type: object
                                  type: object
                                namespaces:
                                  description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                  items:
                                    type: string
                                  type: array
                                topologyKey:
                                  description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                  type: string
                              required:
                                - topologyKey
                              type: object
                            type: array
                        type: object
                      tolerations:
                        description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                        items:
                          description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                          properties:
                            effect:
                              description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                              type: string
                            key:
                              description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                              type: string
                            operator:
                              description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                              type: string
                            tolerationSeconds:
                              description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                              format: int64
                              type: integer
                            value:
                              description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                              type: string
                          type: object
                        type: array
                      topologySpreadConstraints:
                        description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                        items:
                          description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                          properties:
                            labelSelector:
                              description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                  items:
                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                    properties:
                                      key:
                                        description: key is the label key that the selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                        items:
                                          type: string
                                        type: array
                                    required:
                                      - key
                                      - operator
                                    type: object
                                  type: array
                                matchLabels:
                                  additionalProperties:
                                    type: string
                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                  type: object
                              type: object
                            maxSkew:
                              description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                              format: int32
                              type: integer
                            topologyKey:
                              description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                              type: string
                            whenUnsatisfiable:
                              description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                              type: string
                          required:
                            - maxSkew
                            - topologyKey
                            - whenUnsatisfiable
                          type: object
                        type: array
                    type: object
                  description: The placement-related configuration to pass to kubernetes (affinity, node selector, tolerations).
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                priorityClassNames:
                  additionalProperties:
                    type: string
                  description: PriorityClassNames sets priority classes on components
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                removeOSDsIfOutAndSafeToRemove:
                  description: Remove the OSD that is out and safe to remove only if this option is true
                  type: boolean
                resources:
                  additionalProperties:
                    description: ResourceRequirements describes the compute resource requirements.
                    properties:
                      limits:
                        additionalProperties:
                          anyOf:
                            - type: integer
                            - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                        type: object
                      requests:
                        additionalProperties:
                          anyOf:
                            - type: integer
                            - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                        type: object
                    type: object
                  description: Resources set resource requests and limits
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                security:
                  description: Security represents security settings
                  nullable: true
                  properties:
                    kms:
                      description: KeyManagementService is the main Key Management option
                      nullable: true
                      properties:
                        connectionDetails:
                          additionalProperties:
                            type: string
                          description: ConnectionDetails contains the KMS connection details (address, port etc)
                          nullable: true
                          type: object
                          x-kubernetes-preserve-unknown-fields: true
                        tokenSecretName:
                          description: TokenSecretName is the kubernetes secret containing the KMS token
                          type: string
                      type: object
                  type: object
                skipUpgradeChecks:
                  description: SkipUpgradeChecks defines if an upgrade should be forced even if one of the check fails
                  type: boolean
                storage:
                  description: A spec for available storage in the cluster and how it should be used
                  nullable: true
                  properties:
                    config:
                      additionalProperties:
                        type: string
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    deviceFilter:
                      description: A regular expression to allow more fine-grained selection of devices on nodes across the cluster
                      type: string
                    devicePathFilter:
                      description: A regular expression to allow more fine-grained selection of devices with path names
                      type: string
                    devices:
                      description: List of devices to use as storage devices
                      items:
                        description: Device represents a disk to use in the cluster
                        properties:
                          config:
                            additionalProperties:
                              type: string
                            nullable: true
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                          fullpath:
                            type: string
                          name:
                            type: string
                        type: object
                      nullable: true
                      type: array
                      x-kubernetes-preserve-unknown-fields: true
                    nodes:
                      items:
                        description: Node is a storage nodes
                        properties:
                          config:
                            additionalProperties:
                              type: string
                            nullable: true
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                          deviceFilter:
                            description: A regular expression to allow more fine-grained selection of devices on nodes across the cluster
                            type: string
                          devicePathFilter:
                            description: A regular expression to allow more fine-grained selection of devices with path names
                            type: string
                          devices:
                            description: List of devices to use as storage devices
                            items:
                              description: Device represents a disk to use in the cluster
                              properties:
                                config:
                                  additionalProperties:
                                    type: string
                                  nullable: true
                                  type: object
                                  x-kubernetes-preserve-unknown-fields: true
                                fullpath:
                                  type: string
                                name:
                                  type: string
                              type: object
                            nullable: true
                            type: array
                            x-kubernetes-preserve-unknown-fields: true
                          name:
                            type: string
                          resources:
                            description: ResourceRequirements describes the compute resource requirements.
                            nullable: true
                            properties:
                              limits:
                                additionalProperties:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                              requests:
                                additionalProperties:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                          useAllDevices:
                            description: Whether to consume all the storage devices found on a machine
                            type: boolean
                          volumeClaimTemplates:
                            description: PersistentVolumeClaims to use as storage
                            items:
                              description: PersistentVolumeClaim is a user's request for and claim to a persistent volume
                              properties:
                                apiVersion:
                                  description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
                                  type: string
                                kind:
                                  description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                                  type: string
                                metadata:
                                  description: 'Standard object''s metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata'
                                  properties:
                                    annotations:
                                      additionalProperties:
                                        type: string
                                      type: object
                                    finalizers:
                                      items:
                                        type: string
                                      type: array
                                    labels:
                                      additionalProperties:
                                        type: string
                                      type: object
                                    name:
                                      type: string
                                    namespace:
                                      type: string
                                  type: object
                                spec:
                                  description: 'Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                                  properties:
                                    accessModes:
                                      description: 'AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                      items:
                                        type: string
                                      type: array
                                    dataSource:
                                      description: 'This field can be used to specify either: * An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An existing PVC (PersistentVolumeClaim) If the provisioner or an external controller can support the specified data source, it will create a new volume based on the contents of the specified data source. If the AnyVolumeDataSource feature gate is enabled, this field will always have the same contents as the DataSourceRef field.'
                                      properties:
                                        apiGroup:
                                          description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                          type: string
                                        kind:
                                          description: Kind is the type of resource being referenced
                                          type: string
                                        name:
                                          description: Name is the name of resource being referenced
                                          type: string
                                      required:
                                        - kind
                                        - name
                                      type: object
                                    dataSourceRef:
                                      description: 'Specifies the object from which to populate the volume with data, if a non-empty volume is desired. This may be any local object from a non-empty API group (non core object) or a PersistentVolumeClaim object. When this field is specified, volume binding will only succeed if the type of the specified object matches some installed volume populator or dynamic provisioner. This field will replace the functionality of the DataSource field and as such if both fields are non-empty, they must have the same value. For backwards compatibility, both fields (DataSource and DataSourceRef) will be set to the same value automatically if one of them is empty and the other is non-empty. There are two important differences between DataSource and DataSourceRef: * While DataSource only allows two specific types of objects, DataSourceRef   allows any non-core object, as well as PersistentVolumeClaim objects. * While DataSource ignores disallowed values (dropping them), DataSourceRef   preserves all values, and generates an error if a disallowed value is   specified. (Alpha) Using this field requires the AnyVolumeDataSource feature gate to be enabled.'
                                      properties:
                                        apiGroup:
                                          description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                          type: string
                                        kind:
                                          description: Kind is the type of resource being referenced
                                          type: string
                                        name:
                                          description: Name is the name of resource being referenced
                                          type: string
                                      required:
                                        - kind
                                        - name
                                      type: object
                                    resources:
                                      description: 'Resources represents the minimum resources the volume should have. If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements that are lower than previous value but must still be higher than capacity recorded in the status field of the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources'
                                      properties:
                                        limits:
                                          additionalProperties:
                                            anyOf:
                                              - type: integer
                                              - type: string
                                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                            x-kubernetes-int-or-string: true
                                          description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                          type: object
                                        requests:
                                          additionalProperties:
                                            anyOf:
                                              - type: integer
                                              - type: string
                                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                            x-kubernetes-int-or-string: true
                                          description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                          type: object
                                      type: object
                                    selector:
                                      description: A label query over volumes to consider for binding.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    storageClassName:
                                      description: 'Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1'
                                      type: string
                                    volumeMode:
                                      description: volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.
                                      type: string
                                    volumeName:
                                      description: VolumeName is the binding reference to the PersistentVolume backing this claim.
                                      type: string
                                  type: object
                                status:
                                  description: 'Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                                  properties:
                                    accessModes:
                                      description: 'AccessModes contains the actual access modes the volume backing the PVC has. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                      items:
                                        type: string
                                      type: array
                                    allocatedResources:
                                      additionalProperties:
                                        anyOf:
                                          - type: integer
                                          - type: string
                                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                        x-kubernetes-int-or-string: true
                                      description: The storage resource within AllocatedResources tracks the capacity allocated to a PVC. It may be larger than the actual capacity when a volume expansion operation is requested. For storage quota, the larger value from allocatedResources and PVC.spec.resources is used. If allocatedResources is not set, PVC.spec.resources alone is used for quota calculation. If a volume expansion capacity request is lowered, allocatedResources is only lowered if there are no expansion operations in progress and if the actual volume capacity is equal or lower than the requested capacity. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                      type: object
                                    capacity:
                                      additionalProperties:
                                        anyOf:
                                          - type: integer
                                          - type: string
                                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                        x-kubernetes-int-or-string: true
                                      description: Represents the actual resources of the underlying volume.
                                      type: object
                                    conditions:
                                      description: Current Condition of persistent volume claim. If underlying persistent volume is being resized then the Condition will be set to 'ResizeStarted'.
                                      items:
                                        description: PersistentVolumeClaimCondition contails details about state of pvc
                                        properties:
                                          lastProbeTime:
                                            description: Last time we probed the condition.
                                            format: date-time
                                            type: string
                                          lastTransitionTime:
                                            description: Last time the condition transitioned from one status to another.
                                            format: date-time
                                            type: string
                                          message:
                                            description: Human-readable message indicating details about last transition.
                                            type: string
                                          reason:
                                            description: Unique, this should be a short, machine understandable string that gives the reason for condition's last transition. If it reports "ResizeStarted" that means the underlying persistent volume is being resized.
                                            type: string
                                          status:
                                            type: string
                                          type:
                                            description: PersistentVolumeClaimConditionType is a valid value of PersistentVolumeClaimCondition.Type
                                            type: string
                                        required:
                                          - status
                                          - type
                                        type: object
                                      type: array
                                    phase:
                                      description: Phase represents the current phase of PersistentVolumeClaim.
                                      type: string
                                    resizeStatus:
                                      description: ResizeStatus stores status of resize operation. ResizeStatus is not set by default but when expansion is complete resizeStatus is set to empty string by resize controller or kubelet. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                      type: string
                                  type: object
                              type: object
                            type: array
                        type: object
                      nullable: true
                      type: array
                    onlyApplyOSDPlacement:
                      type: boolean
                    storageClassDeviceSets:
                      items:
                        description: StorageClassDeviceSet is a storage class device set
                        properties:
                          config:
                            additionalProperties:
                              type: string
                            description: Provider-specific device configuration
                            nullable: true
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                          count:
                            description: Count is the number of devices in this set
                            minimum: 1
                            type: integer
                          encrypted:
                            description: Whether to encrypt the deviceSet
                            type: boolean
                          name:
                            description: Name is a unique identifier for the set
                            type: string
                          placement:
                            description: Placement is the placement for an object
                            nullable: true
                            properties:
                              nodeAffinity:
                                description: NodeAffinity is a group of node affinity scheduling rules
                                properties:
                                  preferredDuringSchedulingIgnoredDuringExecution:
                                    description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                                    items:
                                      description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                                      properties:
                                        preference:
                                          description: A node selector term, associated with the corresponding weight.
                                          properties:
                                            matchExpressions:
                                              description: A list of node selector requirements by node's labels.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchFields:
                                              description: A list of node selector requirements by node's fields.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                          type: object
                                        weight:
                                          description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                          format: int32
                                          type: integer
                                      required:
                                        - preference
                                        - weight
                                      type: object
                                    type: array
                                  requiredDuringSchedulingIgnoredDuringExecution:
                                    description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                                    properties:
                                      nodeSelectorTerms:
                                        description: Required. A list of node selector terms. The terms are ORed.
                                        items:
                                          description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                          properties:
                                            matchExpressions:
                                              description: A list of node selector requirements by node's labels.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchFields:
                                              description: A list of node selector requirements by node's fields.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                          type: object
                                        type: array
                                    required:
                                      - nodeSelectorTerms
                                    type: object
                                type: object
                              podAffinity:
                                description: PodAffinity is a group of inter pod affinity scheduling rules
                                properties:
                                  preferredDuringSchedulingIgnoredDuringExecution:
                                    description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                                    items:
                                      description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                      properties:
                                        podAffinityTerm:
                                          description: Required. A pod affinity term, associated with the corresponding weight.
                                          properties:
                                            labelSelector:
                                              description: A label query over a set of resources, in this case pods.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaceSelector:
                                              description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaces:
                                              description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                              items:
                                                type: string
                                              type: array
                                            topologyKey:
                                              description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                              type: string
                                          required:
                                            - topologyKey
                                          type: object
                                        weight:
                                          description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                          format: int32
                                          type: integer
                                      required:
                                        - podAffinityTerm
                                        - weight
                                      type: object
                                    type: array
                                  requiredDuringSchedulingIgnoredDuringExecution:
                                    description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                                    items:
                                      description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                      properties:
                                        labelSelector:
                                          description: A label query over a set of resources, in this case pods.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaceSelector:
                                          description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaces:
                                          description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                          items:
                                            type: string
                                          type: array
                                        topologyKey:
                                          description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                          type: string
                                      required:
                                        - topologyKey
                                      type: object
                                    type: array
                                type: object
                              podAntiAffinity:
                                description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                                properties:
                                  preferredDuringSchedulingIgnoredDuringExecution:
                                    description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                                    items:
                                      description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                      properties:
                                        podAffinityTerm:
                                          description: Required. A pod affinity term, associated with the corresponding weight.
                                          properties:
                                            labelSelector:
                                              description: A label query over a set of resources, in this case pods.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaceSelector:
                                              description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaces:
                                              description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                              items:
                                                type: string
                                              type: array
                                            topologyKey:
                                              description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                              type: string
                                          required:
                                            - topologyKey
                                          type: object
                                        weight:
                                          description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                          format: int32
                                          type: integer
                                      required:
                                        - podAffinityTerm
                                        - weight
                                      type: object
                                    type: array
                                  requiredDuringSchedulingIgnoredDuringExecution:
                                    description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                                    items:
                                      description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                      properties:
                                        labelSelector:
                                          description: A label query over a set of resources, in this case pods.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaceSelector:
                                          description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaces:
                                          description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                          items:
                                            type: string
                                          type: array
                                        topologyKey:
                                          description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                          type: string
                                      required:
                                        - topologyKey
                                      type: object
                                    type: array
                                type: object
                              tolerations:
                                description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                                items:
                                  description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                                  properties:
                                    effect:
                                      description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                                      type: string
                                    key:
                                      description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                                      type: string
                                    operator:
                                      description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                                      type: string
                                    tolerationSeconds:
                                      description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                                      format: int64
                                      type: integer
                                    value:
                                      description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                                      type: string
                                  type: object
                                type: array
                              topologySpreadConstraints:
                                description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                                items:
                                  description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                                  properties:
                                    labelSelector:
                                      description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    maxSkew:
                                      description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                                      format: int32
                                      type: integer
                                    topologyKey:
                                      description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                                      type: string
                                    whenUnsatisfiable:
                                      description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                                      type: string
                                  required:
                                    - maxSkew
                                    - topologyKey
                                    - whenUnsatisfiable
                                  type: object
                                type: array
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                          portable:
                            description: Portable represents OSD portability across the hosts
                            type: boolean
                          preparePlacement:
                            description: Placement is the placement for an object
                            nullable: true
                            properties:
                              nodeAffinity:
                                description: NodeAffinity is a group of node affinity scheduling rules
                                properties:
                                  preferredDuringSchedulingIgnoredDuringExecution:
                                    description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                                    items:
                                      description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                                      properties:
                                        preference:
                                          description: A node selector term, associated with the corresponding weight.
                                          properties:
                                            matchExpressions:
                                              description: A list of node selector requirements by node's labels.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchFields:
                                              description: A list of node selector requirements by node's fields.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                          type: object
                                        weight:
                                          description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                          format: int32
                                          type: integer
                                      required:
                                        - preference
                                        - weight
                                      type: object
                                    type: array
                                  requiredDuringSchedulingIgnoredDuringExecution:
                                    description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                                    properties:
                                      nodeSelectorTerms:
                                        description: Required. A list of node selector terms. The terms are ORed.
                                        items:
                                          description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                          properties:
                                            matchExpressions:
                                              description: A list of node selector requirements by node's labels.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchFields:
                                              description: A list of node selector requirements by node's fields.
                                              items:
                                                description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: The label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                                    type: string
                                                  values:
                                                    description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                          type: object
                                        type: array
                                    required:
                                      - nodeSelectorTerms
                                    type: object
                                type: object
                              podAffinity:
                                description: PodAffinity is a group of inter pod affinity scheduling rules
                                properties:
                                  preferredDuringSchedulingIgnoredDuringExecution:
                                    description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                                    items:
                                      description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                      properties:
                                        podAffinityTerm:
                                          description: Required. A pod affinity term, associated with the corresponding weight.
                                          properties:
                                            labelSelector:
                                              description: A label query over a set of resources, in this case pods.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaceSelector:
                                              description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaces:
                                              description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                              items:
                                                type: string
                                              type: array
                                            topologyKey:
                                              description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                              type: string
                                          required:
                                            - topologyKey
                                          type: object
                                        weight:
                                          description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                          format: int32
                                          type: integer
                                      required:
                                        - podAffinityTerm
                                        - weight
                                      type: object
                                    type: array
                                  requiredDuringSchedulingIgnoredDuringExecution:
                                    description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                                    items:
                                      description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                      properties:
                                        labelSelector:
                                          description: A label query over a set of resources, in this case pods.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaceSelector:
                                          description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaces:
                                          description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                          items:
                                            type: string
                                          type: array
                                        topologyKey:
                                          description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                          type: string
                                      required:
                                        - topologyKey
                                      type: object
                                    type: array
                                type: object
                              podAntiAffinity:
                                description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                                properties:
                                  preferredDuringSchedulingIgnoredDuringExecution:
                                    description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                                    items:
                                      description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                      properties:
                                        podAffinityTerm:
                                          description: Required. A pod affinity term, associated with the corresponding weight.
                                          properties:
                                            labelSelector:
                                              description: A label query over a set of resources, in this case pods.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaceSelector:
                                              description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                              properties:
                                                matchExpressions:
                                                  description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                                  items:
                                                    description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                    properties:
                                                      key:
                                                        description: key is the label key that the selector applies to.
                                                        type: string
                                                      operator:
                                                        description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                        type: string
                                                      values:
                                                        description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                        items:
                                                          type: string
                                                        type: array
                                                    required:
                                                      - key
                                                      - operator
                                                    type: object
                                                  type: array
                                                matchLabels:
                                                  additionalProperties:
                                                    type: string
                                                  description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                                  type: object
                                              type: object
                                            namespaces:
                                              description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                              items:
                                                type: string
                                              type: array
                                            topologyKey:
                                              description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                              type: string
                                          required:
                                            - topologyKey
                                          type: object
                                        weight:
                                          description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                          format: int32
                                          type: integer
                                      required:
                                        - podAffinityTerm
                                        - weight
                                      type: object
                                    type: array
                                  requiredDuringSchedulingIgnoredDuringExecution:
                                    description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                                    items:
                                      description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                      properties:
                                        labelSelector:
                                          description: A label query over a set of resources, in this case pods.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaceSelector:
                                          description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                          properties:
                                            matchExpressions:
                                              description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                              items:
                                                description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                                properties:
                                                  key:
                                                    description: key is the label key that the selector applies to.
                                                    type: string
                                                  operator:
                                                    description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                    type: string
                                                  values:
                                                    description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                    items:
                                                      type: string
                                                    type: array
                                                required:
                                                  - key
                                                  - operator
                                                type: object
                                              type: array
                                            matchLabels:
                                              additionalProperties:
                                                type: string
                                              description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                              type: object
                                          type: object
                                        namespaces:
                                          description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                          items:
                                            type: string
                                          type: array
                                        topologyKey:
                                          description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                          type: string
                                      required:
                                        - topologyKey
                                      type: object
                                    type: array
                                type: object
                              tolerations:
                                description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                                items:
                                  description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                                  properties:
                                    effect:
                                      description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                                      type: string
                                    key:
                                      description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                                      type: string
                                    operator:
                                      description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                                      type: string
                                    tolerationSeconds:
                                      description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                                      format: int64
                                      type: integer
                                    value:
                                      description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                                      type: string
                                  type: object
                                type: array
                              topologySpreadConstraints:
                                description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                                items:
                                  description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                                  properties:
                                    labelSelector:
                                      description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    maxSkew:
                                      description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                                      format: int32
                                      type: integer
                                    topologyKey:
                                      description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                                      type: string
                                    whenUnsatisfiable:
                                      description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                                      type: string
                                  required:
                                    - maxSkew
                                    - topologyKey
                                    - whenUnsatisfiable
                                  type: object
                                type: array
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                          resources:
                            description: ResourceRequirements describes the compute resource requirements.
                            nullable: true
                            properties:
                              limits:
                                additionalProperties:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                              requests:
                                additionalProperties:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                          schedulerName:
                            description: Scheduler name for OSD pod placement
                            type: string
                          tuneDeviceClass:
                            description: TuneSlowDeviceClass Tune the OSD when running on a slow Device Class
                            type: boolean
                          tuneFastDeviceClass:
                            description: TuneFastDeviceClass Tune the OSD when running on a fast Device Class
                            type: boolean
                          volumeClaimTemplates:
                            description: VolumeClaimTemplates is a list of PVC templates for the underlying storage devices
                            items:
                              description: PersistentVolumeClaim is a user's request for and claim to a persistent volume
                              properties:
                                apiVersion:
                                  description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
                                  type: string
                                kind:
                                  description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                                  type: string
                                metadata:
                                  description: 'Standard object''s metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata'
                                  properties:
                                    annotations:
                                      additionalProperties:
                                        type: string
                                      type: object
                                      x-kubernetes-preserve-unknown-fields: true
                                    finalizers:
                                      items:
                                        type: string
                                      type: array
                                    labels:
                                      additionalProperties:
                                        type: string
                                      type: object
                                    name:
                                      type: string
                                    namespace:
                                      type: string
                                  type: object
                                spec:
                                  description: 'Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                                  properties:
                                    accessModes:
                                      description: 'AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                      items:
                                        type: string
                                      type: array
                                    dataSource:
                                      description: 'This field can be used to specify either: * An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An existing PVC (PersistentVolumeClaim) If the provisioner or an external controller can support the specified data source, it will create a new volume based on the contents of the specified data source. If the AnyVolumeDataSource feature gate is enabled, this field will always have the same contents as the DataSourceRef field.'
                                      properties:
                                        apiGroup:
                                          description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                          type: string
                                        kind:
                                          description: Kind is the type of resource being referenced
                                          type: string
                                        name:
                                          description: Name is the name of resource being referenced
                                          type: string
                                      required:
                                        - kind
                                        - name
                                      type: object
                                    dataSourceRef:
                                      description: 'Specifies the object from which to populate the volume with data, if a non-empty volume is desired. This may be any local object from a non-empty API group (non core object) or a PersistentVolumeClaim object. When this field is specified, volume binding will only succeed if the type of the specified object matches some installed volume populator or dynamic provisioner. This field will replace the functionality of the DataSource field and as such if both fields are non-empty, they must have the same value. For backwards compatibility, both fields (DataSource and DataSourceRef) will be set to the same value automatically if one of them is empty and the other is non-empty. There are two important differences between DataSource and DataSourceRef: * While DataSource only allows two specific types of objects, DataSourceRef   allows any non-core object, as well as PersistentVolumeClaim objects. * While DataSource ignores disallowed values (dropping them), DataSourceRef   preserves all values, and generates an error if a disallowed value is   specified. (Alpha) Using this field requires the AnyVolumeDataSource feature gate to be enabled.'
                                      properties:
                                        apiGroup:
                                          description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                          type: string
                                        kind:
                                          description: Kind is the type of resource being referenced
                                          type: string
                                        name:
                                          description: Name is the name of resource being referenced
                                          type: string
                                      required:
                                        - kind
                                        - name
                                      type: object
                                    resources:
                                      description: 'Resources represents the minimum resources the volume should have. If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements that are lower than previous value but must still be higher than capacity recorded in the status field of the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources'
                                      properties:
                                        limits:
                                          additionalProperties:
                                            anyOf:
                                              - type: integer
                                              - type: string
                                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                            x-kubernetes-int-or-string: true
                                          description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                          type: object
                                        requests:
                                          additionalProperties:
                                            anyOf:
                                              - type: integer
                                              - type: string
                                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                            x-kubernetes-int-or-string: true
                                          description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                          type: object
                                      type: object
                                    selector:
                                      description: A label query over volumes to consider for binding.
                                      properties:
                                        matchExpressions:
                                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                          items:
                                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                            properties:
                                              key:
                                                description: key is the label key that the selector applies to.
                                                type: string
                                              operator:
                                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                type: string
                                              values:
                                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                items:
                                                  type: string
                                                type: array
                                            required:
                                              - key
                                              - operator
                                            type: object
                                          type: array
                                        matchLabels:
                                          additionalProperties:
                                            type: string
                                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                          type: object
                                      type: object
                                    storageClassName:
                                      description: 'Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1'
                                      type: string
                                    volumeMode:
                                      description: volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.
                                      type: string
                                    volumeName:
                                      description: VolumeName is the binding reference to the PersistentVolume backing this claim.
                                      type: string
                                  type: object
                                status:
                                  description: 'Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                                  properties:
                                    accessModes:
                                      description: 'AccessModes contains the actual access modes the volume backing the PVC has. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                      items:
                                        type: string
                                      type: array
                                    allocatedResources:
                                      additionalProperties:
                                        anyOf:
                                          - type: integer
                                          - type: string
                                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                        x-kubernetes-int-or-string: true
                                      description: The storage resource within AllocatedResources tracks the capacity allocated to a PVC. It may be larger than the actual capacity when a volume expansion operation is requested. For storage quota, the larger value from allocatedResources and PVC.spec.resources is used. If allocatedResources is not set, PVC.spec.resources alone is used for quota calculation. If a volume expansion capacity request is lowered, allocatedResources is only lowered if there are no expansion operations in progress and if the actual volume capacity is equal or lower than the requested capacity. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                      type: object
                                    capacity:
                                      additionalProperties:
                                        anyOf:
                                          - type: integer
                                          - type: string
                                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                        x-kubernetes-int-or-string: true
                                      description: Represents the actual resources of the underlying volume.
                                      type: object
                                    conditions:
                                      description: Current Condition of persistent volume claim. If underlying persistent volume is being resized then the Condition will be set to 'ResizeStarted'.
                                      items:
                                        description: PersistentVolumeClaimCondition contails details about state of pvc
                                        properties:
                                          lastProbeTime:
                                            description: Last time we probed the condition.
                                            format: date-time
                                            type: string
                                          lastTransitionTime:
                                            description: Last time the condition transitioned from one status to another.
                                            format: date-time
                                            type: string
                                          message:
                                            description: Human-readable message indicating details about last transition.
                                            type: string
                                          reason:
                                            description: Unique, this should be a short, machine understandable string that gives the reason for condition's last transition. If it reports "ResizeStarted" that means the underlying persistent volume is being resized.
                                            type: string
                                          status:
                                            type: string
                                          type:
                                            description: PersistentVolumeClaimConditionType is a valid value of PersistentVolumeClaimCondition.Type
                                            type: string
                                        required:
                                          - status
                                          - type
                                        type: object
                                      type: array
                                    phase:
                                      description: Phase represents the current phase of PersistentVolumeClaim.
                                      type: string
                                    resizeStatus:
                                      description: ResizeStatus stores status of resize operation. ResizeStatus is not set by default but when expansion is complete resizeStatus is set to empty string by resize controller or kubelet. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                      type: string
                                  type: object
                              type: object
                            type: array
                        required:
                          - count
                          - name
                          - volumeClaimTemplates
                        type: object
                      nullable: true
                      type: array
                    useAllDevices:
                      description: Whether to consume all the storage devices found on a machine
                      type: boolean
                    useAllNodes:
                      type: boolean
                    volumeClaimTemplates:
                      description: PersistentVolumeClaims to use as storage
                      items:
                        description: PersistentVolumeClaim is a user's request for and claim to a persistent volume
                        properties:
                          apiVersion:
                            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
                            type: string
                          kind:
                            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                            type: string
                          metadata:
                            description: 'Standard object''s metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata'
                            properties:
                              annotations:
                                additionalProperties:
                                  type: string
                                type: object
                              finalizers:
                                items:
                                  type: string
                                type: array
                              labels:
                                additionalProperties:
                                  type: string
                                type: object
                              name:
                                type: string
                              namespace:
                                type: string
                            type: object
                          spec:
                            description: 'Spec defines the desired characteristics of a volume requested by a pod author. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                            properties:
                              accessModes:
                                description: 'AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                items:
                                  type: string
                                type: array
                              dataSource:
                                description: 'This field can be used to specify either: * An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An existing PVC (PersistentVolumeClaim) If the provisioner or an external controller can support the specified data source, it will create a new volume based on the contents of the specified data source. If the AnyVolumeDataSource feature gate is enabled, this field will always have the same contents as the DataSourceRef field.'
                                properties:
                                  apiGroup:
                                    description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                    type: string
                                  kind:
                                    description: Kind is the type of resource being referenced
                                    type: string
                                  name:
                                    description: Name is the name of resource being referenced
                                    type: string
                                required:
                                  - kind
                                  - name
                                type: object
                              dataSourceRef:
                                description: 'Specifies the object from which to populate the volume with data, if a non-empty volume is desired. This may be any local object from a non-empty API group (non core object) or a PersistentVolumeClaim object. When this field is specified, volume binding will only succeed if the type of the specified object matches some installed volume populator or dynamic provisioner. This field will replace the functionality of the DataSource field and as such if both fields are non-empty, they must have the same value. For backwards compatibility, both fields (DataSource and DataSourceRef) will be set to the same value automatically if one of them is empty and the other is non-empty. There are two important differences between DataSource and DataSourceRef: * While DataSource only allows two specific types of objects, DataSourceRef   allows any non-core object, as well as PersistentVolumeClaim objects. * While DataSource ignores disallowed values (dropping them), DataSourceRef   preserves all values, and generates an error if a disallowed value is   specified. (Alpha) Using this field requires the AnyVolumeDataSource feature gate to be enabled.'
                                properties:
                                  apiGroup:
                                    description: APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.
                                    type: string
                                  kind:
                                    description: Kind is the type of resource being referenced
                                    type: string
                                  name:
                                    description: Name is the name of resource being referenced
                                    type: string
                                required:
                                  - kind
                                  - name
                                type: object
                              resources:
                                description: 'Resources represents the minimum resources the volume should have. If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements that are lower than previous value but must still be higher than capacity recorded in the status field of the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources'
                                properties:
                                  limits:
                                    additionalProperties:
                                      anyOf:
                                        - type: integer
                                        - type: string
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                    type: object
                                  requests:
                                    additionalProperties:
                                      anyOf:
                                        - type: integer
                                        - type: string
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                    type: object
                                type: object
                              selector:
                                description: A label query over volumes to consider for binding.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              storageClassName:
                                description: 'Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1'
                                type: string
                              volumeMode:
                                description: volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.
                                type: string
                              volumeName:
                                description: VolumeName is the binding reference to the PersistentVolume backing this claim.
                                type: string
                            type: object
                          status:
                            description: 'Status represents the current information/status of a persistent volume claim. Read-only. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims'
                            properties:
                              accessModes:
                                description: 'AccessModes contains the actual access modes the volume backing the PVC has. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1'
                                items:
                                  type: string
                                type: array
                              allocatedResources:
                                additionalProperties:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: The storage resource within AllocatedResources tracks the capacity allocated to a PVC. It may be larger than the actual capacity when a volume expansion operation is requested. For storage quota, the larger value from allocatedResources and PVC.spec.resources is used. If allocatedResources is not set, PVC.spec.resources alone is used for quota calculation. If a volume expansion capacity request is lowered, allocatedResources is only lowered if there are no expansion operations in progress and if the actual volume capacity is equal or lower than the requested capacity. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                type: object
                              capacity:
                                additionalProperties:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: Represents the actual resources of the underlying volume.
                                type: object
                              conditions:
                                description: Current Condition of persistent volume claim. If underlying persistent volume is being resized then the Condition will be set to 'ResizeStarted'.
                                items:
                                  description: PersistentVolumeClaimCondition contails details about state of pvc
                                  properties:
                                    lastProbeTime:
                                      description: Last time we probed the condition.
                                      format: date-time
                                      type: string
                                    lastTransitionTime:
                                      description: Last time the condition transitioned from one status to another.
                                      format: date-time
                                      type: string
                                    message:
                                      description: Human-readable message indicating details about last transition.
                                      type: string
                                    reason:
                                      description: Unique, this should be a short, machine understandable string that gives the reason for condition's last transition. If it reports "ResizeStarted" that means the underlying persistent volume is being resized.
                                      type: string
                                    status:
                                      type: string
                                    type:
                                      description: PersistentVolumeClaimConditionType is a valid value of PersistentVolumeClaimCondition.Type
                                      type: string
                                  required:
                                    - status
                                    - type
                                  type: object
                                type: array
                              phase:
                                description: Phase represents the current phase of PersistentVolumeClaim.
                                type: string
                              resizeStatus:
                                description: ResizeStatus stores status of resize operation. ResizeStatus is not set by default but when expansion is complete resizeStatus is set to empty string by resize controller or kubelet. This is an alpha field and requires enabling RecoverVolumeExpansionFailure feature.
                                type: string
                            type: object
                        type: object
                      type: array
                  type: object
                waitTimeoutForHealthyOSDInMinutes:
                  description: WaitTimeoutForHealthyOSDInMinutes defines the time the operator would wait before an OSD can be stopped for upgrade or restart. If the timeout exceeds and OSD is not ok to stop, then the operator would skip upgrade for the current OSD and proceed with the next one if continueUpgradeAfterChecksEvenIfNotHealthy is false. If continueUpgradeAfterChecksEvenIfNotHealthy is true, then operator would continue with the upgrade of an OSD even if its not ok to stop after the timeout. This timeout won't be applied if skipUpgradeChecks is true. The default wait timeout is 10 minutes.
                  format: int64
                  type: integer
              type: object
            status:
              description: ClusterStatus represents the status of a Ceph cluster
              nullable: true
              properties:
                ceph:
                  description: CephStatus is the details health of a Ceph Cluster
                  properties:
                    capacity:
                      description: Capacity is the capacity information of a Ceph Cluster
                      properties:
                        bytesAvailable:
                          format: int64
                          type: integer
                        bytesTotal:
                          format: int64
                          type: integer
                        bytesUsed:
                          format: int64
                          type: integer
                        lastUpdated:
                          type: string
                      type: object
                    details:
                      additionalProperties:
                        description: CephHealthMessage represents the health message of a Ceph Cluster
                        properties:
                          message:
                            type: string
                          severity:
                            type: string
                        required:
                          - message
                          - severity
                        type: object
                      type: object
                    fsid:
                      type: string
                    health:
                      type: string
                    lastChanged:
                      type: string
                    lastChecked:
                      type: string
                    previousHealth:
                      type: string
                    versions:
                      description: CephDaemonsVersions show the current ceph version for different ceph daemons
                      properties:
                        cephfs-mirror:
                          additionalProperties:
                            type: integer
                          description: CephFSMirror shows CephFSMirror Ceph version
                          type: object
                        mds:
                          additionalProperties:
                            type: integer
                          description: Mds shows Mds Ceph version
                          type: object
                        mgr:
                          additionalProperties:
                            type: integer
                          description: Mgr shows Mgr Ceph version
                          type: object
                        mon:
                          additionalProperties:
                            type: integer
                          description: Mon shows Mon Ceph version
                          type: object
                        osd:
                          additionalProperties:
                            type: integer
                          description: Osd shows Osd Ceph version
                          type: object
                        overall:
                          additionalProperties:
                            type: integer
                          description: Overall shows overall Ceph version
                          type: object
                        rbd-mirror:
                          additionalProperties:
                            type: integer
                          description: RbdMirror shows RbdMirror Ceph version
                          type: object
                        rgw:
                          additionalProperties:
                            type: integer
                          description: Rgw shows Rgw Ceph version
                          type: object
                      type: object
                  type: object
                conditions:
                  items:
                    description: Condition represents a status condition on any Rook-Ceph Custom Resource.
                    properties:
                      lastHeartbeatTime:
                        format: date-time
                        type: string
                      lastTransitionTime:
                        format: date-time
                        type: string
                      message:
                        type: string
                      reason:
                        description: ConditionReason is a reason for a condition
                        type: string
                      status:
                        type: string
                      type:
                        description: ConditionType represent a resource's status
                        type: string
                    type: object
                  type: array
                message:
                  type: string
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  description: ConditionType represent a resource's status
                  type: string
                state:
                  description: ClusterState represents the state of a Ceph Cluster
                  type: string
                storage:
                  description: CephStorage represents flavors of Ceph Cluster Storage
                  properties:
                    deviceClasses:
                      items:
                        description: DeviceClasses represents device classes of a Ceph Cluster
                        properties:
                          name:
                            type: string
                        type: object
                      type: array
                  type: object
                version:
                  description: ClusterVersion represents the version of a Ceph Cluster
                  properties:
                    image:
                      type: string
                    version:
                      type: string
                  type: object
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephfilesystemmirrors.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephFilesystemMirror
    listKind: CephFilesystemMirrorList
    plural: cephfilesystemmirrors
    singular: cephfilesystemmirror
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephFilesystemMirror is the Ceph Filesystem Mirror object definition
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: FilesystemMirroringSpec is the filesystem mirroring specification
              properties:
                annotations:
                  additionalProperties:
                    type: string
                  description: The annotations-related configuration to add/set on each Pod related object.
                  nullable: true
                  type: object
                labels:
                  additionalProperties:
                    type: string
                  description: The labels-related configuration to add/set on each Pod related object.
                  nullable: true
                  type: object
                placement:
                  description: The affinity to place the rgw pods (default is to place on any available node)
                  nullable: true
                  properties:
                    nodeAffinity:
                      description: NodeAffinity is a group of node affinity scheduling rules
                      properties:
                        preferredDuringSchedulingIgnoredDuringExecution:
                          description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                          items:
                            description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                            properties:
                              preference:
                                description: A node selector term, associated with the corresponding weight.
                                properties:
                                  matchExpressions:
                                    description: A list of node selector requirements by node's labels.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchFields:
                                    description: A list of node selector requirements by node's fields.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                type: object
                              weight:
                                description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                format: int32
                                type: integer
                            required:
                              - preference
                              - weight
                            type: object
                          type: array
                        requiredDuringSchedulingIgnoredDuringExecution:
                          description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                          properties:
                            nodeSelectorTerms:
                              description: Required. A list of node selector terms. The terms are ORed.
                              items:
                                description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                properties:
                                  matchExpressions:
                                    description: A list of node selector requirements by node's labels.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchFields:
                                    description: A list of node selector requirements by node's fields.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                type: object
                              type: array
                          required:
                            - nodeSelectorTerms
                          type: object
                      type: object
                    podAffinity:
                      description: PodAffinity is a group of inter pod affinity scheduling rules
                      properties:
                        preferredDuringSchedulingIgnoredDuringExecution:
                          description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                          items:
                            description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                            properties:
                              podAffinityTerm:
                                description: Required. A pod affinity term, associated with the corresponding weight.
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              weight:
                                description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                format: int32
                                type: integer
                            required:
                              - podAffinityTerm
                              - weight
                            type: object
                          type: array
                        requiredDuringSchedulingIgnoredDuringExecution:
                          description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                          items:
                            description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                            properties:
                              labelSelector:
                                description: A label query over a set of resources, in this case pods.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaceSelector:
                                description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaces:
                                description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                items:
                                  type: string
                                type: array
                              topologyKey:
                                description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                type: string
                            required:
                              - topologyKey
                            type: object
                          type: array
                      type: object
                    podAntiAffinity:
                      description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                      properties:
                        preferredDuringSchedulingIgnoredDuringExecution:
                          description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                          items:
                            description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                            properties:
                              podAffinityTerm:
                                description: Required. A pod affinity term, associated with the corresponding weight.
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              weight:
                                description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                format: int32
                                type: integer
                            required:
                              - podAffinityTerm
                              - weight
                            type: object
                          type: array
                        requiredDuringSchedulingIgnoredDuringExecution:
                          description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                          items:
                            description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                            properties:
                              labelSelector:
                                description: A label query over a set of resources, in this case pods.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaceSelector:
                                description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaces:
                                description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                items:
                                  type: string
                                type: array
                              topologyKey:
                                description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                type: string
                            required:
                              - topologyKey
                            type: object
                          type: array
                      type: object
                    tolerations:
                      description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                      items:
                        description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                        properties:
                          effect:
                            description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                            type: string
                          key:
                            description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                            type: string
                          operator:
                            description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                            type: string
                          tolerationSeconds:
                            description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                            format: int64
                            type: integer
                          value:
                            description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                            type: string
                        type: object
                      type: array
                    topologySpreadConstraints:
                      description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                      items:
                        description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                        properties:
                          labelSelector:
                            description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                            properties:
                              matchExpressions:
                                description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                items:
                                  description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                  properties:
                                    key:
                                      description: key is the label key that the selector applies to.
                                      type: string
                                    operator:
                                      description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                      type: string
                                    values:
                                      description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                      items:
                                        type: string
                                      type: array
                                  required:
                                    - key
                                    - operator
                                  type: object
                                type: array
                              matchLabels:
                                additionalProperties:
                                  type: string
                                description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                type: object
                            type: object
                          maxSkew:
                            description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                            format: int32
                            type: integer
                          topologyKey:
                            description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                            type: string
                          whenUnsatisfiable:
                            description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                            type: string
                        required:
                          - maxSkew
                          - topologyKey
                          - whenUnsatisfiable
                        type: object
                      type: array
                  type: object
                priorityClassName:
                  description: PriorityClassName sets priority class on the cephfs-mirror pods
                  type: string
                resources:
                  description: The resource requirements for the cephfs-mirror pods
                  nullable: true
                  properties:
                    limits:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    requests:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                  type: object
              type: object
            status:
              description: Status represents the status of an object
              properties:
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephfilesystems.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephFilesystem
    listKind: CephFilesystemList
    plural: cephfilesystems
    singular: cephfilesystem
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - description: Number of desired active MDS daemons
          jsonPath: .spec.metadataServer.activeCount
          name: ActiveMDS
          type: string
        - jsonPath: .metadata.creationTimestamp
          name: Age
          type: date
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephFilesystem represents a Ceph Filesystem
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: FilesystemSpec represents the spec of a file system
              properties:
                dataPools:
                  description: The data pool settings, with optional predefined pool name.
                  items:
                    description: NamedPoolSpec represents the named ceph pool spec
                    properties:
                      compressionMode:
                        description: 'DEPRECATED: use Parameters instead, e.g., Parameters["compression_mode"] = "force" The inline compression mode in Bluestore OSD to set to (options are: none, passive, aggressive, force) Do NOT set a default value for kubebuilder as this will override the Parameters'
                        enum:
                          - none
                          - passive
                          - aggressive
                          - force
                          - ""
                        nullable: true
                        type: string
                      crushRoot:
                        description: The root of the crush hierarchy utilized by the pool
                        nullable: true
                        type: string
                      deviceClass:
                        description: The device class the OSD should set to for use in the pool
                        nullable: true
                        type: string
                      enableRBDStats:
                        description: EnableRBDStats is used to enable gathering of statistics for all RBD images in the pool
                        type: boolean
                      erasureCoded:
                        description: The erasure code settings
                        properties:
                          algorithm:
                            description: The algorithm for erasure coding
                            type: string
                          codingChunks:
                            description: Number of coding chunks per object in an erasure coded storage pool (required for erasure-coded pool type). This is the number of OSDs that can be lost simultaneously before data cannot be recovered.
                            minimum: 0
                            type: integer
                          dataChunks:
                            description: Number of data chunks per object in an erasure coded storage pool (required for erasure-coded pool type). The number of chunks required to recover an object when any single OSD is lost is the same as dataChunks so be aware that the larger the number of data chunks, the higher the cost of recovery.
                            minimum: 0
                            type: integer
                        required:
                          - codingChunks
                          - dataChunks
                        type: object
                      failureDomain:
                        description: 'The failure domain: osd/host/(region or zone if available) - technically also any type in the crush map'
                        type: string
                      mirroring:
                        description: The mirroring settings
                        properties:
                          enabled:
                            description: Enabled whether this pool is mirrored or not
                            type: boolean
                          mode:
                            description: 'Mode is the mirroring mode: either pool or image'
                            type: string
                          peers:
                            description: Peers represents the peers spec
                            nullable: true
                            properties:
                              secretNames:
                                description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                                items:
                                  type: string
                                type: array
                            type: object
                          snapshotSchedules:
                            description: SnapshotSchedules is the scheduling of snapshot for mirrored images/pools
                            items:
                              description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                              properties:
                                interval:
                                  description: Interval represent the periodicity of the snapshot.
                                  type: string
                                path:
                                  description: Path is the path to snapshot, only valid for CephFS
                                  type: string
                                startTime:
                                  description: StartTime indicates when to start the snapshot
                                  type: string
                              type: object
                            type: array
                        type: object
                      name:
                        description: Name of the pool
                        type: string
                      parameters:
                        additionalProperties:
                          type: string
                        description: Parameters is a list of properties to enable on a given pool
                        nullable: true
                        type: object
                        x-kubernetes-preserve-unknown-fields: true
                      quotas:
                        description: The quota settings
                        nullable: true
                        properties:
                          maxBytes:
                            description: MaxBytes represents the quota in bytes Deprecated in favor of MaxSize
                            format: int64
                            type: integer
                          maxObjects:
                            description: MaxObjects represents the quota in objects
                            format: int64
                            type: integer
                          maxSize:
                            description: MaxSize represents the quota in bytes as a string
                            pattern: ^[0-9]+[\.]?[0-9]*([KMGTPE]i|[kMGTPE])?$
                            type: string
                        type: object
                      replicated:
                        description: The replication settings
                        properties:
                          hybridStorage:
                            description: HybridStorage represents hybrid storage tier settings
                            nullable: true
                            properties:
                              primaryDeviceClass:
                                description: PrimaryDeviceClass represents high performance tier (for example SSD or NVME) for Primary OSD
                                minLength: 1
                                type: string
                              secondaryDeviceClass:
                                description: SecondaryDeviceClass represents low performance tier (for example HDDs) for remaining OSDs
                                minLength: 1
                                type: string
                            required:
                              - primaryDeviceClass
                              - secondaryDeviceClass
                            type: object
                          replicasPerFailureDomain:
                            description: ReplicasPerFailureDomain the number of replica in the specified failure domain
                            minimum: 1
                            type: integer
                          requireSafeReplicaSize:
                            description: RequireSafeReplicaSize if false allows you to set replica 1
                            type: boolean
                          size:
                            description: Size - Number of copies per object in a replicated storage pool, including the object itself (required for replicated pool type)
                            minimum: 0
                            type: integer
                          subFailureDomain:
                            description: SubFailureDomain the name of the sub-failure domain
                            type: string
                          targetSizeRatio:
                            description: TargetSizeRatio gives a hint (%) to Ceph in terms of expected consumption of the total cluster capacity
                            type: number
                        required:
                          - size
                        type: object
                      statusCheck:
                        description: The mirroring statusCheck
                        properties:
                          mirror:
                            description: HealthCheckSpec represents the health check of an object store bucket
                            nullable: true
                            properties:
                              disabled:
                                type: boolean
                              interval:
                                description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                                type: string
                              timeout:
                                type: string
                            type: object
                        type: object
                        x-kubernetes-preserve-unknown-fields: true
                    type: object
                  nullable: true
                  type: array
                metadataPool:
                  description: The metadata pool settings
                  nullable: true
                  properties:
                    compressionMode:
                      description: 'DEPRECATED: use Parameters instead, e.g., Parameters["compression_mode"] = "force" The inline compression mode in Bluestore OSD to set to (options are: none, passive, aggressive, force) Do NOT set a default value for kubebuilder as this will override the Parameters'
                      enum:
                        - none
                        - passive
                        - aggressive
                        - force
                        - ""
                      nullable: true
                      type: string
                    crushRoot:
                      description: The root of the crush hierarchy utilized by the pool
                      nullable: true
                      type: string
                    deviceClass:
                      description: The device class the OSD should set to for use in the pool
                      nullable: true
                      type: string
                    enableRBDStats:
                      description: EnableRBDStats is used to enable gathering of statistics for all RBD images in the pool
                      type: boolean
                    erasureCoded:
                      description: The erasure code settings
                      properties:
                        algorithm:
                          description: The algorithm for erasure coding
                          type: string
                        codingChunks:
                          description: Number of coding chunks per object in an erasure coded storage pool (required for erasure-coded pool type). This is the number of OSDs that can be lost simultaneously before data cannot be recovered.
                          minimum: 0
                          type: integer
                        dataChunks:
                          description: Number of data chunks per object in an erasure coded storage pool (required for erasure-coded pool type). The number of chunks required to recover an object when any single OSD is lost is the same as dataChunks so be aware that the larger the number of data chunks, the higher the cost of recovery.
                          minimum: 0
                          type: integer
                      required:
                        - codingChunks
                        - dataChunks
                      type: object
                    failureDomain:
                      description: 'The failure domain: osd/host/(region or zone if available) - technically also any type in the crush map'
                      type: string
                    mirroring:
                      description: The mirroring settings
                      properties:
                        enabled:
                          description: Enabled whether this pool is mirrored or not
                          type: boolean
                        mode:
                          description: 'Mode is the mirroring mode: either pool or image'
                          type: string
                        peers:
                          description: Peers represents the peers spec
                          nullable: true
                          properties:
                            secretNames:
                              description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                              items:
                                type: string
                              type: array
                          type: object
                        snapshotSchedules:
                          description: SnapshotSchedules is the scheduling of snapshot for mirrored images/pools
                          items:
                            description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                            properties:
                              interval:
                                description: Interval represent the periodicity of the snapshot.
                                type: string
                              path:
                                description: Path is the path to snapshot, only valid for CephFS
                                type: string
                              startTime:
                                description: StartTime indicates when to start the snapshot
                                type: string
                            type: object
                          type: array
                      type: object
                    parameters:
                      additionalProperties:
                        type: string
                      description: Parameters is a list of properties to enable on a given pool
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    quotas:
                      description: The quota settings
                      nullable: true
                      properties:
                        maxBytes:
                          description: MaxBytes represents the quota in bytes Deprecated in favor of MaxSize
                          format: int64
                          type: integer
                        maxObjects:
                          description: MaxObjects represents the quota in objects
                          format: int64
                          type: integer
                        maxSize:
                          description: MaxSize represents the quota in bytes as a string
                          pattern: ^[0-9]+[\.]?[0-9]*([KMGTPE]i|[kMGTPE])?$
                          type: string
                      type: object
                    replicated:
                      description: The replication settings
                      properties:
                        hybridStorage:
                          description: HybridStorage represents hybrid storage tier settings
                          nullable: true
                          properties:
                            primaryDeviceClass:
                              description: PrimaryDeviceClass represents high performance tier (for example SSD or NVME) for Primary OSD
                              minLength: 1
                              type: string
                            secondaryDeviceClass:
                              description: SecondaryDeviceClass represents low performance tier (for example HDDs) for remaining OSDs
                              minLength: 1
                              type: string
                          required:
                            - primaryDeviceClass
                            - secondaryDeviceClass
                          type: object
                        replicasPerFailureDomain:
                          description: ReplicasPerFailureDomain the number of replica in the specified failure domain
                          minimum: 1
                          type: integer
                        requireSafeReplicaSize:
                          description: RequireSafeReplicaSize if false allows you to set replica 1
                          type: boolean
                        size:
                          description: Size - Number of copies per object in a replicated storage pool, including the object itself (required for replicated pool type)
                          minimum: 0
                          type: integer
                        subFailureDomain:
                          description: SubFailureDomain the name of the sub-failure domain
                          type: string
                        targetSizeRatio:
                          description: TargetSizeRatio gives a hint (%) to Ceph in terms of expected consumption of the total cluster capacity
                          type: number
                      required:
                        - size
                      type: object
                    statusCheck:
                      description: The mirroring statusCheck
                      properties:
                        mirror:
                          description: HealthCheckSpec represents the health check of an object store bucket
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                  type: object
                metadataServer:
                  description: The mds pod info
                  properties:
                    activeCount:
                      description: The number of metadata servers that are active. The remaining servers in the cluster will be in standby mode.
                      format: int32
                      maximum: 10
                      minimum: 1
                      type: integer
                    activeStandby:
                      description: Whether each active MDS instance will have an active standby with a warm metadata cache for faster failover. If false, standbys will still be available, but will not have a warm metadata cache.
                      type: boolean
                    annotations:
                      additionalProperties:
                        type: string
                      description: The annotations-related configuration to add/set on each Pod related object.
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    labels:
                      additionalProperties:
                        type: string
                      description: The labels-related configuration to add/set on each Pod related object.
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    livenessProbe:
                      description: ProbeSpec is a wrapper around Probe so it can be enabled or disabled for a Ceph daemon
                      properties:
                        disabled:
                          description: Disabled determines whether probe is disable or not
                          type: boolean
                        probe:
                          description: Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
                          properties:
                            exec:
                              description: Exec specifies the action to take.
                              properties:
                                command:
                                  description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            failureThreshold:
                              description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                              format: int32
                              type: integer
                            grpc:
                              description: GRPC specifies an action involving a GRPC port. This is an alpha field and requires enabling GRPCContainerProbe feature gate.
                              properties:
                                port:
                                  description: Port number of the gRPC service. Number must be in the range 1 to 65535.
                                  format: int32
                                  type: integer
                                service:
                                  description: "Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). \n If this is not specified, the default behavior is defined by gRPC."
                                  type: string
                              required:
                                - port
                              type: object
                            httpGet:
                              description: HTTPGet specifies the http request to perform.
                              properties:
                                host:
                                  description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                  type: string
                                httpHeaders:
                                  description: Custom headers to set in the request. HTTP allows repeated headers.
                                  items:
                                    description: HTTPHeader describes a custom header to be used in HTTP probes
                                    properties:
                                      name:
                                        description: The header field name
                                        type: string
                                      value:
                                        description: The header field value
                                        type: string
                                    required:
                                      - name
                                      - value
                                    type: object
                                  type: array
                                path:
                                  description: Path to access on the HTTP server.
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                                scheme:
                                  description: Scheme to use for connecting to the host. Defaults to HTTP.
                                  type: string
                              required:
                                - port
                              type: object
                            initialDelaySeconds:
                              description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                            periodSeconds:
                              description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                              format: int32
                              type: integer
                            successThreshold:
                              description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
                              format: int32
                              type: integer
                            tcpSocket:
                              description: TCPSocket specifies an action involving a TCP port.
                              properties:
                                host:
                                  description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                              required:
                                - port
                              type: object
                            terminationGracePeriodSeconds:
                              description: Optional duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this value overrides the value provided by the pod spec. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate. Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
                              format: int64
                              type: integer
                            timeoutSeconds:
                              description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                          type: object
                      type: object
                    placement:
                      description: The affinity to place the mds pods (default is to place on all available node) with a daemonset
                      nullable: true
                      properties:
                        nodeAffinity:
                          description: NodeAffinity is a group of node affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                              items:
                                description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                                properties:
                                  preference:
                                    description: A node selector term, associated with the corresponding weight.
                                    properties:
                                      matchExpressions:
                                        description: A list of node selector requirements by node's labels.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchFields:
                                        description: A list of node selector requirements by node's fields.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                    type: object
                                  weight:
                                    description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - preference
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                              properties:
                                nodeSelectorTerms:
                                  description: Required. A list of node selector terms. The terms are ORed.
                                  items:
                                    description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                    properties:
                                      matchExpressions:
                                        description: A list of node selector requirements by node's labels.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchFields:
                                        description: A list of node selector requirements by node's fields.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                    type: object
                                  type: array
                              required:
                                - nodeSelectorTerms
                              type: object
                          type: object
                        podAffinity:
                          description: PodAffinity is a group of inter pod affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                              items:
                                description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                properties:
                                  podAffinityTerm:
                                    description: Required. A pod affinity term, associated with the corresponding weight.
                                    properties:
                                      labelSelector:
                                        description: A label query over a set of resources, in this case pods.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaceSelector:
                                        description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaces:
                                        description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                        items:
                                          type: string
                                        type: array
                                      topologyKey:
                                        description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                        type: string
                                    required:
                                      - topologyKey
                                    type: object
                                  weight:
                                    description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - podAffinityTerm
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                              items:
                                description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              type: array
                          type: object
                        podAntiAffinity:
                          description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                              items:
                                description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                properties:
                                  podAffinityTerm:
                                    description: Required. A pod affinity term, associated with the corresponding weight.
                                    properties:
                                      labelSelector:
                                        description: A label query over a set of resources, in this case pods.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaceSelector:
                                        description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaces:
                                        description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                        items:
                                          type: string
                                        type: array
                                      topologyKey:
                                        description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                        type: string
                                    required:
                                      - topologyKey
                                    type: object
                                  weight:
                                    description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - podAffinityTerm
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                              items:
                                description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              type: array
                          type: object
                        tolerations:
                          description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                          items:
                            description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                            properties:
                              effect:
                                description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                                type: string
                              key:
                                description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                                type: string
                              operator:
                                description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                                type: string
                              tolerationSeconds:
                                description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                                format: int64
                                type: integer
                              value:
                                description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                                type: string
                            type: object
                          type: array
                        topologySpreadConstraints:
                          description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                          items:
                            description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                            properties:
                              labelSelector:
                                description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              maxSkew:
                                description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                                format: int32
                                type: integer
                              topologyKey:
                                description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                                type: string
                              whenUnsatisfiable:
                                description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                                type: string
                            required:
                              - maxSkew
                              - topologyKey
                              - whenUnsatisfiable
                            type: object
                          type: array
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    priorityClassName:
                      description: PriorityClassName sets priority classes on components
                      type: string
                    resources:
                      description: The resource requirements for the rgw pods
                      nullable: true
                      properties:
                        limits:
                          additionalProperties:
                            anyOf:
                              - type: integer
                              - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                        requests:
                          additionalProperties:
                            anyOf:
                              - type: integer
                              - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    startupProbe:
                      description: ProbeSpec is a wrapper around Probe so it can be enabled or disabled for a Ceph daemon
                      properties:
                        disabled:
                          description: Disabled determines whether probe is disable or not
                          type: boolean
                        probe:
                          description: Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
                          properties:
                            exec:
                              description: Exec specifies the action to take.
                              properties:
                                command:
                                  description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            failureThreshold:
                              description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                              format: int32
                              type: integer
                            grpc:
                              description: GRPC specifies an action involving a GRPC port. This is an alpha field and requires enabling GRPCContainerProbe feature gate.
                              properties:
                                port:
                                  description: Port number of the gRPC service. Number must be in the range 1 to 65535.
                                  format: int32
                                  type: integer
                                service:
                                  description: "Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). \n If this is not specified, the default behavior is defined by gRPC."
                                  type: string
                              required:
                                - port
                              type: object
                            httpGet:
                              description: HTTPGet specifies the http request to perform.
                              properties:
                                host:
                                  description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                  type: string
                                httpHeaders:
                                  description: Custom headers to set in the request. HTTP allows repeated headers.
                                  items:
                                    description: HTTPHeader describes a custom header to be used in HTTP probes
                                    properties:
                                      name:
                                        description: The header field name
                                        type: string
                                      value:
                                        description: The header field value
                                        type: string
                                    required:
                                      - name
                                      - value
                                    type: object
                                  type: array
                                path:
                                  description: Path to access on the HTTP server.
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                                scheme:
                                  description: Scheme to use for connecting to the host. Defaults to HTTP.
                                  type: string
                              required:
                                - port
                              type: object
                            initialDelaySeconds:
                              description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                            periodSeconds:
                              description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                              format: int32
                              type: integer
                            successThreshold:
                              description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
                              format: int32
                              type: integer
                            tcpSocket:
                              description: TCPSocket specifies an action involving a TCP port.
                              properties:
                                host:
                                  description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                              required:
                                - port
                              type: object
                            terminationGracePeriodSeconds:
                              description: Optional duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this value overrides the value provided by the pod spec. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate. Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
                              format: int64
                              type: integer
                            timeoutSeconds:
                              description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                          type: object
                      type: object
                  required:
                    - activeCount
                  type: object
                mirroring:
                  description: The mirroring settings
                  nullable: true
                  properties:
                    enabled:
                      description: Enabled whether this filesystem is mirrored or not
                      type: boolean
                    peers:
                      description: Peers represents the peers spec
                      nullable: true
                      properties:
                        secretNames:
                          description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                          items:
                            type: string
                          type: array
                      type: object
                    snapshotRetention:
                      description: Retention is the retention policy for a snapshot schedule One path has exactly one retention policy. A policy can however contain multiple count-time period pairs in order to specify complex retention policies
                      items:
                        description: SnapshotScheduleRetentionSpec is a retention policy
                        properties:
                          duration:
                            description: Duration represents the retention duration for a snapshot
                            type: string
                          path:
                            description: Path is the path to snapshot
                            type: string
                        type: object
                      type: array
                    snapshotSchedules:
                      description: SnapshotSchedules is the scheduling of snapshot for mirrored filesystems
                      items:
                        description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                        properties:
                          interval:
                            description: Interval represent the periodicity of the snapshot.
                            type: string
                          path:
                            description: Path is the path to snapshot, only valid for CephFS
                            type: string
                          startTime:
                            description: StartTime indicates when to start the snapshot
                            type: string
                        type: object
                      type: array
                  type: object
                preserveFilesystemOnDelete:
                  description: Preserve the fs in the cluster on CephFilesystem CR deletion. Setting this to true automatically implies PreservePoolsOnDelete is true.
                  type: boolean
                preservePoolsOnDelete:
                  description: Preserve pools on filesystem deletion
                  type: boolean
                statusCheck:
                  description: The mirroring statusCheck
                  properties:
                    mirror:
                      description: HealthCheckSpec represents the health check of an object store bucket
                      nullable: true
                      properties:
                        disabled:
                          type: boolean
                        interval:
                          description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                          type: string
                        timeout:
                          type: string
                      type: object
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
              required:
                - dataPools
                - metadataPool
                - metadataServer
              type: object
            status:
              description: CephFilesystemStatus represents the status of a Ceph Filesystem
              properties:
                conditions:
                  items:
                    description: Condition represents a status condition on any Rook-Ceph Custom Resource.
                    properties:
                      lastHeartbeatTime:
                        format: date-time
                        type: string
                      lastTransitionTime:
                        format: date-time
                        type: string
                      message:
                        type: string
                      reason:
                        description: ConditionReason is a reason for a condition
                        type: string
                      status:
                        type: string
                      type:
                        description: ConditionType represent a resource's status
                        type: string
                    type: object
                  type: array
                info:
                  additionalProperties:
                    type: string
                  description: Use only info and put mirroringStatus in it?
                  nullable: true
                  type: object
                mirroringStatus:
                  description: MirroringStatus is the filesystem mirroring status
                  properties:
                    daemonsStatus:
                      description: PoolMirroringStatus is the mirroring status of a filesystem
                      items:
                        description: FilesystemMirrorInfoSpec is the filesystem mirror status of a given filesystem
                        properties:
                          daemon_id:
                            description: DaemonID is the cephfs-mirror name
                            type: integer
                          filesystems:
                            description: Filesystems is the list of filesystems managed by a given cephfs-mirror daemon
                            items:
                              description: FilesystemsSpec is spec for the mirrored filesystem
                              properties:
                                directory_count:
                                  description: DirectoryCount is the number of directories in the filesystem
                                  type: integer
                                filesystem_id:
                                  description: FilesystemID is the filesystem identifier
                                  type: integer
                                name:
                                  description: Name is name of the filesystem
                                  type: string
                                peers:
                                  description: Peers represents the mirroring peers
                                  items:
                                    description: FilesystemMirrorInfoPeerSpec is the specification of a filesystem peer mirror
                                    properties:
                                      remote:
                                        description: Remote are the remote cluster information
                                        properties:
                                          client_name:
                                            description: ClientName is cephx name
                                            type: string
                                          cluster_name:
                                            description: ClusterName is the name of the cluster
                                            type: string
                                          fs_name:
                                            description: FsName is the filesystem name
                                            type: string
                                        type: object
                                      stats:
                                        description: Stats are the stat a peer mirror
                                        properties:
                                          failure_count:
                                            description: FailureCount is the number of mirroring failure
                                            type: integer
                                          recovery_count:
                                            description: RecoveryCount is the number of recovery attempted after failures
                                            type: integer
                                        type: object
                                      uuid:
                                        description: UUID is the peer unique identifier
                                        type: string
                                    type: object
                                  type: array
                              type: object
                            type: array
                        type: object
                      nullable: true
                      type: array
                    details:
                      description: Details contains potential status errors
                      type: string
                    lastChanged:
                      description: LastChanged is the last time time the status last changed
                      type: string
                    lastChecked:
                      description: LastChecked is the last time time the status was checked
                      type: string
                  type: object
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  description: ConditionType represent a resource's status
                  type: string
                snapshotScheduleStatus:
                  description: FilesystemSnapshotScheduleStatusSpec is the status of the snapshot schedule
                  properties:
                    details:
                      description: Details contains potential status errors
                      type: string
                    lastChanged:
                      description: LastChanged is the last time time the status last changed
                      type: string
                    lastChecked:
                      description: LastChecked is the last time time the status was checked
                      type: string
                    snapshotSchedules:
                      description: SnapshotSchedules is the list of snapshots scheduled
                      items:
                        description: FilesystemSnapshotSchedulesSpec is the list of snapshot scheduled for images in a pool
                        properties:
                          fs:
                            description: Fs is the name of the Ceph Filesystem
                            type: string
                          path:
                            description: Path is the path on the filesystem
                            type: string
                          rel_path:
                            type: string
                          retention:
                            description: FilesystemSnapshotScheduleStatusRetention is the retention specification for a filesystem snapshot schedule
                            properties:
                              active:
                                description: Active is whether the scheduled is active or not
                                type: boolean
                              created:
                                description: Created is when the snapshot schedule was created
                                type: string
                              created_count:
                                description: CreatedCount is total amount of snapshots
                                type: integer
                              first:
                                description: First is when the first snapshot schedule was taken
                                type: string
                              last:
                                description: Last is when the last snapshot schedule was taken
                                type: string
                              last_pruned:
                                description: LastPruned is when the last snapshot schedule was pruned
                                type: string
                              pruned_count:
                                description: PrunedCount is total amount of pruned snapshots
                                type: integer
                              start:
                                description: Start is when the snapshot schedule starts
                                type: string
                            type: object
                          schedule:
                            type: string
                          subvol:
                            description: Subvol is the name of the sub volume
                            type: string
                        type: object
                      nullable: true
                      type: array
                  type: object
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephfilesystemsubvolumegroups.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephFilesystemSubVolumeGroup
    listKind: CephFilesystemSubVolumeGroupList
    plural: cephfilesystemsubvolumegroups
    singular: cephfilesystemsubvolumegroup
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephFilesystemSubVolumeGroup represents a Ceph Filesystem SubVolumeGroup
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: Spec represents the specification of a Ceph Filesystem SubVolumeGroup
              properties:
                filesystemName:
                  description: FilesystemName is the name of Ceph Filesystem SubVolumeGroup volume name. Typically it's the name of the CephFilesystem CR. If not coming from the CephFilesystem CR, it can be retrieved from the list of Ceph Filesystem volumes with ceph fs volume ls. To learn more about Ceph Filesystem abstractions see https://docs.ceph.com/en/latest/cephfs/fs-volumes/#fs-volumes-and-subvolumes
                  type: string
              required:
                - filesystemName
              type: object
            status:
              description: Status represents the status of a CephFilesystem SubvolumeGroup
              properties:
                info:
                  additionalProperties:
                    type: string
                  nullable: true
                  type: object
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  description: ConditionType represent a resource's status
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephnfses.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephNFS
    listKind: CephNFSList
    plural: cephnfses
    shortNames:
      - nfs
    singular: cephnfs
  scope: Namespaced
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: CephNFS represents a Ceph NFS
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: NFSGaneshaSpec represents the spec of an nfs ganesha server
              properties:
                rados:
                  description: RADOS is the Ganesha RADOS specification
                  nullable: true
                  properties:
                    namespace:
                      description: The namespace inside the Ceph pool (set by 'pool') where shared NFS-Ganesha config is stored. This setting is required for Ceph v15 and ignored for Ceph v16. As of Ceph Pacific v16+, this is internally set to the name of the CephNFS.
                      type: string
                    pool:
                      description: The Ceph pool used store the shared configuration for NFS-Ganesha daemons. This setting is required for Ceph v15 and ignored for Ceph v16. As of Ceph Pacific 16.2.7+, this is internally hardcoded to ".nfs".
                      type: string
                  type: object
                server:
                  description: Server is the Ganesha Server specification
                  properties:
                    active:
                      description: The number of active Ganesha servers
                      type: integer
                    annotations:
                      additionalProperties:
                        type: string
                      description: The annotations-related configuration to add/set on each Pod related object.
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    labels:
                      additionalProperties:
                        type: string
                      description: The labels-related configuration to add/set on each Pod related object.
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    logLevel:
                      description: LogLevel set logging level
                      type: string
                    placement:
                      description: The affinity to place the ganesha pods
                      nullable: true
                      properties:
                        nodeAffinity:
                          description: NodeAffinity is a group of node affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                              items:
                                description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                                properties:
                                  preference:
                                    description: A node selector term, associated with the corresponding weight.
                                    properties:
                                      matchExpressions:
                                        description: A list of node selector requirements by node's labels.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchFields:
                                        description: A list of node selector requirements by node's fields.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                    type: object
                                  weight:
                                    description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - preference
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                              properties:
                                nodeSelectorTerms:
                                  description: Required. A list of node selector terms. The terms are ORed.
                                  items:
                                    description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                    properties:
                                      matchExpressions:
                                        description: A list of node selector requirements by node's labels.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchFields:
                                        description: A list of node selector requirements by node's fields.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                    type: object
                                  type: array
                              required:
                                - nodeSelectorTerms
                              type: object
                          type: object
                        podAffinity:
                          description: PodAffinity is a group of inter pod affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                              items:
                                description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                properties:
                                  podAffinityTerm:
                                    description: Required. A pod affinity term, associated with the corresponding weight.
                                    properties:
                                      labelSelector:
                                        description: A label query over a set of resources, in this case pods.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaceSelector:
                                        description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaces:
                                        description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                        items:
                                          type: string
                                        type: array
                                      topologyKey:
                                        description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                        type: string
                                    required:
                                      - topologyKey
                                    type: object
                                  weight:
                                    description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - podAffinityTerm
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                              items:
                                description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              type: array
                          type: object
                        podAntiAffinity:
                          description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                              items:
                                description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                properties:
                                  podAffinityTerm:
                                    description: Required. A pod affinity term, associated with the corresponding weight.
                                    properties:
                                      labelSelector:
                                        description: A label query over a set of resources, in this case pods.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaceSelector:
                                        description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaces:
                                        description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                        items:
                                          type: string
                                        type: array
                                      topologyKey:
                                        description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                        type: string
                                    required:
                                      - topologyKey
                                    type: object
                                  weight:
                                    description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - podAffinityTerm
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                              items:
                                description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              type: array
                          type: object
                        tolerations:
                          description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                          items:
                            description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                            properties:
                              effect:
                                description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                                type: string
                              key:
                                description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                                type: string
                              operator:
                                description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                                type: string
                              tolerationSeconds:
                                description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                                format: int64
                                type: integer
                              value:
                                description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                                type: string
                            type: object
                          type: array
                        topologySpreadConstraints:
                          description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                          items:
                            description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                            properties:
                              labelSelector:
                                description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              maxSkew:
                                description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                                format: int32
                                type: integer
                              topologyKey:
                                description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                                type: string
                              whenUnsatisfiable:
                                description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                                type: string
                            required:
                              - maxSkew
                              - topologyKey
                              - whenUnsatisfiable
                            type: object
                          type: array
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    priorityClassName:
                      description: PriorityClassName sets the priority class on the pods
                      type: string
                    resources:
                      description: Resources set resource requests and limits
                      nullable: true
                      properties:
                        limits:
                          additionalProperties:
                            anyOf:
                              - type: integer
                              - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                        requests:
                          additionalProperties:
                            anyOf:
                              - type: integer
                              - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                  required:
                    - active
                  type: object
              required:
                - server
              type: object
            status:
              description: Status represents the status of an object
              properties:
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephobjectrealms.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectRealm
    listKind: CephObjectRealmList
    plural: cephobjectrealms
    singular: cephobjectrealm
  scope: Namespaced
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: CephObjectRealm represents a Ceph Object Store Gateway Realm
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: ObjectRealmSpec represent the spec of an ObjectRealm
              nullable: true
              properties:
                pull:
                  description: PullSpec represents the pulling specification of a Ceph Object Storage Gateway Realm
                  properties:
                    endpoint:
                      pattern: ^https*://
                      type: string
                  type: object
              type: object
            status:
              description: Status represents the status of an object
              properties:
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephobjectstores.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectStore
    listKind: CephObjectStoreList
    plural: cephobjectstores
    singular: cephobjectstore
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephObjectStore represents a Ceph Object Store Gateway
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: ObjectStoreSpec represent the spec of a pool
              properties:
                dataPool:
                  description: The data pool settings
                  nullable: true
                  properties:
                    compressionMode:
                      description: 'DEPRECATED: use Parameters instead, e.g., Parameters["compression_mode"] = "force" The inline compression mode in Bluestore OSD to set to (options are: none, passive, aggressive, force) Do NOT set a default value for kubebuilder as this will override the Parameters'
                      enum:
                        - none
                        - passive
                        - aggressive
                        - force
                        - ""
                      nullable: true
                      type: string
                    crushRoot:
                      description: The root of the crush hierarchy utilized by the pool
                      nullable: true
                      type: string
                    deviceClass:
                      description: The device class the OSD should set to for use in the pool
                      nullable: true
                      type: string
                    enableRBDStats:
                      description: EnableRBDStats is used to enable gathering of statistics for all RBD images in the pool
                      type: boolean
                    erasureCoded:
                      description: The erasure code settings
                      properties:
                        algorithm:
                          description: The algorithm for erasure coding
                          type: string
                        codingChunks:
                          description: Number of coding chunks per object in an erasure coded storage pool (required for erasure-coded pool type). This is the number of OSDs that can be lost simultaneously before data cannot be recovered.
                          minimum: 0
                          type: integer
                        dataChunks:
                          description: Number of data chunks per object in an erasure coded storage pool (required for erasure-coded pool type). The number of chunks required to recover an object when any single OSD is lost is the same as dataChunks so be aware that the larger the number of data chunks, the higher the cost of recovery.
                          minimum: 0
                          type: integer
                      required:
                        - codingChunks
                        - dataChunks
                      type: object
                    failureDomain:
                      description: 'The failure domain: osd/host/(region or zone if available) - technically also any type in the crush map'
                      type: string
                    mirroring:
                      description: The mirroring settings
                      properties:
                        enabled:
                          description: Enabled whether this pool is mirrored or not
                          type: boolean
                        mode:
                          description: 'Mode is the mirroring mode: either pool or image'
                          type: string
                        peers:
                          description: Peers represents the peers spec
                          nullable: true
                          properties:
                            secretNames:
                              description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                              items:
                                type: string
                              type: array
                          type: object
                        snapshotSchedules:
                          description: SnapshotSchedules is the scheduling of snapshot for mirrored images/pools
                          items:
                            description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                            properties:
                              interval:
                                description: Interval represent the periodicity of the snapshot.
                                type: string
                              path:
                                description: Path is the path to snapshot, only valid for CephFS
                                type: string
                              startTime:
                                description: StartTime indicates when to start the snapshot
                                type: string
                            type: object
                          type: array
                      type: object
                    parameters:
                      additionalProperties:
                        type: string
                      description: Parameters is a list of properties to enable on a given pool
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    quotas:
                      description: The quota settings
                      nullable: true
                      properties:
                        maxBytes:
                          description: MaxBytes represents the quota in bytes Deprecated in favor of MaxSize
                          format: int64
                          type: integer
                        maxObjects:
                          description: MaxObjects represents the quota in objects
                          format: int64
                          type: integer
                        maxSize:
                          description: MaxSize represents the quota in bytes as a string
                          pattern: ^[0-9]+[\.]?[0-9]*([KMGTPE]i|[kMGTPE])?$
                          type: string
                      type: object
                    replicated:
                      description: The replication settings
                      properties:
                        hybridStorage:
                          description: HybridStorage represents hybrid storage tier settings
                          nullable: true
                          properties:
                            primaryDeviceClass:
                              description: PrimaryDeviceClass represents high performance tier (for example SSD or NVME) for Primary OSD
                              minLength: 1
                              type: string
                            secondaryDeviceClass:
                              description: SecondaryDeviceClass represents low performance tier (for example HDDs) for remaining OSDs
                              minLength: 1
                              type: string
                          required:
                            - primaryDeviceClass
                            - secondaryDeviceClass
                          type: object
                        replicasPerFailureDomain:
                          description: ReplicasPerFailureDomain the number of replica in the specified failure domain
                          minimum: 1
                          type: integer
                        requireSafeReplicaSize:
                          description: RequireSafeReplicaSize if false allows you to set replica 1
                          type: boolean
                        size:
                          description: Size - Number of copies per object in a replicated storage pool, including the object itself (required for replicated pool type)
                          minimum: 0
                          type: integer
                        subFailureDomain:
                          description: SubFailureDomain the name of the sub-failure domain
                          type: string
                        targetSizeRatio:
                          description: TargetSizeRatio gives a hint (%) to Ceph in terms of expected consumption of the total cluster capacity
                          type: number
                      required:
                        - size
                      type: object
                    statusCheck:
                      description: The mirroring statusCheck
                      properties:
                        mirror:
                          description: HealthCheckSpec represents the health check of an object store bucket
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                  type: object
                gateway:
                  description: The rgw pod info
                  nullable: true
                  properties:
                    annotations:
                      additionalProperties:
                        type: string
                      description: The annotations-related configuration to add/set on each Pod related object.
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    caBundleRef:
                      description: The name of the secret that stores custom ca-bundle with root and intermediate certificates.
                      nullable: true
                      type: string
                    externalRgwEndpoints:
                      description: ExternalRgwEndpoints points to external rgw endpoint(s)
                      items:
                        description: EndpointAddress is a tuple that describes single IP address.
                        properties:
                          hostname:
                            description: The Hostname of this endpoint
                            type: string
                          ip:
                            description: 'The IP of this endpoint. May not be loopback (127.0.0.0/8), link-local (169.254.0.0/16), or link-local multicast ((224.0.0.0/24). IPv6 is also accepted but not fully supported on all platforms. Also, certain kubernetes components, like kube-proxy, are not IPv6 ready. TODO: This should allow hostname or IP, See #4447.'
                            type: string
                          nodeName:
                            description: 'Optional: Node hosting this endpoint. This can be used to determine endpoints local to a node.'
                            type: string
                          targetRef:
                            description: Reference to object providing the endpoint.
                            properties:
                              apiVersion:
                                description: API version of the referent.
                                type: string
                              fieldPath:
                                description: 'If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.'
                                type: string
                              kind:
                                description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                                type: string
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names'
                                type: string
                              namespace:
                                description: 'Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/'
                                type: string
                              resourceVersion:
                                description: 'Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency'
                                type: string
                              uid:
                                description: 'UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids'
                                type: string
                            type: object
                        required:
                          - ip
                        type: object
                      nullable: true
                      type: array
                    hostNetwork:
                      description: Whether host networking is enabled for the rgw daemon. If not set, the network settings from the cluster CR will be applied.
                      nullable: true
                      type: boolean
                      x-kubernetes-preserve-unknown-fields: true
                    instances:
                      description: The number of pods in the rgw replicaset.
                      format: int32
                      nullable: true
                      type: integer
                    labels:
                      additionalProperties:
                        type: string
                      description: The labels-related configuration to add/set on each Pod related object.
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    placement:
                      description: The affinity to place the rgw pods (default is to place on any available node)
                      nullable: true
                      properties:
                        nodeAffinity:
                          description: NodeAffinity is a group of node affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                              items:
                                description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                                properties:
                                  preference:
                                    description: A node selector term, associated with the corresponding weight.
                                    properties:
                                      matchExpressions:
                                        description: A list of node selector requirements by node's labels.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchFields:
                                        description: A list of node selector requirements by node's fields.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                    type: object
                                  weight:
                                    description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - preference
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                              properties:
                                nodeSelectorTerms:
                                  description: Required. A list of node selector terms. The terms are ORed.
                                  items:
                                    description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                    properties:
                                      matchExpressions:
                                        description: A list of node selector requirements by node's labels.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchFields:
                                        description: A list of node selector requirements by node's fields.
                                        items:
                                          description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: The label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                              type: string
                                            values:
                                              description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                    type: object
                                  type: array
                              required:
                                - nodeSelectorTerms
                              type: object
                          type: object
                        podAffinity:
                          description: PodAffinity is a group of inter pod affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                              items:
                                description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                properties:
                                  podAffinityTerm:
                                    description: Required. A pod affinity term, associated with the corresponding weight.
                                    properties:
                                      labelSelector:
                                        description: A label query over a set of resources, in this case pods.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaceSelector:
                                        description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaces:
                                        description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                        items:
                                          type: string
                                        type: array
                                      topologyKey:
                                        description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                        type: string
                                    required:
                                      - topologyKey
                                    type: object
                                  weight:
                                    description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - podAffinityTerm
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                              items:
                                description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              type: array
                          type: object
                        podAntiAffinity:
                          description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                          properties:
                            preferredDuringSchedulingIgnoredDuringExecution:
                              description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                              items:
                                description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                                properties:
                                  podAffinityTerm:
                                    description: Required. A pod affinity term, associated with the corresponding weight.
                                    properties:
                                      labelSelector:
                                        description: A label query over a set of resources, in this case pods.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaceSelector:
                                        description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                        properties:
                                          matchExpressions:
                                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                            items:
                                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                              properties:
                                                key:
                                                  description: key is the label key that the selector applies to.
                                                  type: string
                                                operator:
                                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                                  type: string
                                                values:
                                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                                  items:
                                                    type: string
                                                  type: array
                                              required:
                                                - key
                                                - operator
                                              type: object
                                            type: array
                                          matchLabels:
                                            additionalProperties:
                                              type: string
                                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                            type: object
                                        type: object
                                      namespaces:
                                        description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                        items:
                                          type: string
                                        type: array
                                      topologyKey:
                                        description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                        type: string
                                    required:
                                      - topologyKey
                                    type: object
                                  weight:
                                    description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                    format: int32
                                    type: integer
                                required:
                                  - podAffinityTerm
                                  - weight
                                type: object
                              type: array
                            requiredDuringSchedulingIgnoredDuringExecution:
                              description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                              items:
                                description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              type: array
                          type: object
                        tolerations:
                          description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                          items:
                            description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                            properties:
                              effect:
                                description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                                type: string
                              key:
                                description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                                type: string
                              operator:
                                description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                                type: string
                              tolerationSeconds:
                                description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                                format: int64
                                type: integer
                              value:
                                description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                                type: string
                            type: object
                          type: array
                        topologySpreadConstraints:
                          description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                          items:
                            description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                            properties:
                              labelSelector:
                                description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              maxSkew:
                                description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                                format: int32
                                type: integer
                              topologyKey:
                                description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                                type: string
                              whenUnsatisfiable:
                                description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                                type: string
                            required:
                              - maxSkew
                              - topologyKey
                              - whenUnsatisfiable
                            type: object
                          type: array
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    port:
                      description: The port the rgw service will be listening on (http)
                      format: int32
                      type: integer
                    priorityClassName:
                      description: PriorityClassName sets priority classes on the rgw pods
                      type: string
                    resources:
                      description: The resource requirements for the rgw pods
                      nullable: true
                      properties:
                        limits:
                          additionalProperties:
                            anyOf:
                              - type: integer
                              - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                        requests:
                          additionalProperties:
                            anyOf:
                              - type: integer
                              - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    securePort:
                      description: The port the rgw service will be listening on (https)
                      format: int32
                      maximum: 65535
                      minimum: 0
                      nullable: true
                      type: integer
                    service:
                      description: The configuration related to add/set on each rgw service.
                      nullable: true
                      properties:
                        annotations:
                          additionalProperties:
                            type: string
                          description: The annotations-related configuration to add/set on each rgw service. nullable optional
                          type: object
                      type: object
                    sslCertificateRef:
                      description: The name of the secret that stores the ssl certificate for secure rgw connections
                      nullable: true
                      type: string
                  type: object
                healthCheck:
                  description: The rgw Bucket healthchecks and liveness probe
                  nullable: true
                  properties:
                    bucket:
                      description: HealthCheckSpec represents the health check of an object store bucket
                      properties:
                        disabled:
                          type: boolean
                        interval:
                          description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                          type: string
                        timeout:
                          type: string
                      type: object
                    livenessProbe:
                      description: ProbeSpec is a wrapper around Probe so it can be enabled or disabled for a Ceph daemon
                      properties:
                        disabled:
                          description: Disabled determines whether probe is disable or not
                          type: boolean
                        probe:
                          description: Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
                          properties:
                            exec:
                              description: Exec specifies the action to take.
                              properties:
                                command:
                                  description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            failureThreshold:
                              description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                              format: int32
                              type: integer
                            grpc:
                              description: GRPC specifies an action involving a GRPC port. This is an alpha field and requires enabling GRPCContainerProbe feature gate.
                              properties:
                                port:
                                  description: Port number of the gRPC service. Number must be in the range 1 to 65535.
                                  format: int32
                                  type: integer
                                service:
                                  description: "Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). \n If this is not specified, the default behavior is defined by gRPC."
                                  type: string
                              required:
                                - port
                              type: object
                            httpGet:
                              description: HTTPGet specifies the http request to perform.
                              properties:
                                host:
                                  description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                  type: string
                                httpHeaders:
                                  description: Custom headers to set in the request. HTTP allows repeated headers.
                                  items:
                                    description: HTTPHeader describes a custom header to be used in HTTP probes
                                    properties:
                                      name:
                                        description: The header field name
                                        type: string
                                      value:
                                        description: The header field value
                                        type: string
                                    required:
                                      - name
                                      - value
                                    type: object
                                  type: array
                                path:
                                  description: Path to access on the HTTP server.
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                                scheme:
                                  description: Scheme to use for connecting to the host. Defaults to HTTP.
                                  type: string
                              required:
                                - port
                              type: object
                            initialDelaySeconds:
                              description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                            periodSeconds:
                              description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                              format: int32
                              type: integer
                            successThreshold:
                              description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
                              format: int32
                              type: integer
                            tcpSocket:
                              description: TCPSocket specifies an action involving a TCP port.
                              properties:
                                host:
                                  description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                              required:
                                - port
                              type: object
                            terminationGracePeriodSeconds:
                              description: Optional duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this value overrides the value provided by the pod spec. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate. Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
                              format: int64
                              type: integer
                            timeoutSeconds:
                              description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                          type: object
                      type: object
                    readinessProbe:
                      description: ProbeSpec is a wrapper around Probe so it can be enabled or disabled for a Ceph daemon
                      properties:
                        disabled:
                          description: Disabled determines whether probe is disable or not
                          type: boolean
                        probe:
                          description: Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
                          properties:
                            exec:
                              description: Exec specifies the action to take.
                              properties:
                                command:
                                  description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            failureThreshold:
                              description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                              format: int32
                              type: integer
                            grpc:
                              description: GRPC specifies an action involving a GRPC port. This is an alpha field and requires enabling GRPCContainerProbe feature gate.
                              properties:
                                port:
                                  description: Port number of the gRPC service. Number must be in the range 1 to 65535.
                                  format: int32
                                  type: integer
                                service:
                                  description: "Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). \n If this is not specified, the default behavior is defined by gRPC."
                                  type: string
                              required:
                                - port
                              type: object
                            httpGet:
                              description: HTTPGet specifies the http request to perform.
                              properties:
                                host:
                                  description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                  type: string
                                httpHeaders:
                                  description: Custom headers to set in the request. HTTP allows repeated headers.
                                  items:
                                    description: HTTPHeader describes a custom header to be used in HTTP probes
                                    properties:
                                      name:
                                        description: The header field name
                                        type: string
                                      value:
                                        description: The header field value
                                        type: string
                                    required:
                                      - name
                                      - value
                                    type: object
                                  type: array
                                path:
                                  description: Path to access on the HTTP server.
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                                scheme:
                                  description: Scheme to use for connecting to the host. Defaults to HTTP.
                                  type: string
                              required:
                                - port
                              type: object
                            initialDelaySeconds:
                              description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                            periodSeconds:
                              description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                              format: int32
                              type: integer
                            successThreshold:
                              description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
                              format: int32
                              type: integer
                            tcpSocket:
                              description: TCPSocket specifies an action involving a TCP port.
                              properties:
                                host:
                                  description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                              required:
                                - port
                              type: object
                            terminationGracePeriodSeconds:
                              description: Optional duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this value overrides the value provided by the pod spec. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate. Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
                              format: int64
                              type: integer
                            timeoutSeconds:
                              description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                          type: object
                      type: object
                    startupProbe:
                      description: ProbeSpec is a wrapper around Probe so it can be enabled or disabled for a Ceph daemon
                      properties:
                        disabled:
                          description: Disabled determines whether probe is disable or not
                          type: boolean
                        probe:
                          description: Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
                          properties:
                            exec:
                              description: Exec specifies the action to take.
                              properties:
                                command:
                                  description: Command is the command line to execute inside the container, the working directory for the command  is root ('/') in the container's filesystem. The command is simply exec'd, it is not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use a shell, you need to explicitly call out to that shell. Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            failureThreshold:
                              description: Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.
                              format: int32
                              type: integer
                            grpc:
                              description: GRPC specifies an action involving a GRPC port. This is an alpha field and requires enabling GRPCContainerProbe feature gate.
                              properties:
                                port:
                                  description: Port number of the gRPC service. Number must be in the range 1 to 65535.
                                  format: int32
                                  type: integer
                                service:
                                  description: "Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). \n If this is not specified, the default behavior is defined by gRPC."
                                  type: string
                              required:
                                - port
                              type: object
                            httpGet:
                              description: HTTPGet specifies the http request to perform.
                              properties:
                                host:
                                  description: Host name to connect to, defaults to the pod IP. You probably want to set "Host" in httpHeaders instead.
                                  type: string
                                httpHeaders:
                                  description: Custom headers to set in the request. HTTP allows repeated headers.
                                  items:
                                    description: HTTPHeader describes a custom header to be used in HTTP probes
                                    properties:
                                      name:
                                        description: The header field name
                                        type: string
                                      value:
                                        description: The header field value
                                        type: string
                                    required:
                                      - name
                                      - value
                                    type: object
                                  type: array
                                path:
                                  description: Path to access on the HTTP server.
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Name or number of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                                scheme:
                                  description: Scheme to use for connecting to the host. Defaults to HTTP.
                                  type: string
                              required:
                                - port
                              type: object
                            initialDelaySeconds:
                              description: 'Number of seconds after the container has started before liveness probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                            periodSeconds:
                              description: How often (in seconds) to perform the probe. Default to 10 seconds. Minimum value is 1.
                              format: int32
                              type: integer
                            successThreshold:
                              description: Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
                              format: int32
                              type: integer
                            tcpSocket:
                              description: TCPSocket specifies an action involving a TCP port.
                              properties:
                                host:
                                  description: 'Optional: Host name to connect to, defaults to the pod IP.'
                                  type: string
                                port:
                                  anyOf:
                                    - type: integer
                                    - type: string
                                  description: Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                              required:
                                - port
                              type: object
                            terminationGracePeriodSeconds:
                              description: Optional duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this value overrides the value provided by the pod spec. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate. Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
                              format: int64
                              type: integer
                            timeoutSeconds:
                              description: 'Number of seconds after which the probe times out. Defaults to 1 second. Minimum value is 1. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                              format: int32
                              type: integer
                          type: object
                      type: object
                  type: object
                metadataPool:
                  description: The metadata pool settings
                  nullable: true
                  properties:
                    compressionMode:
                      description: 'DEPRECATED: use Parameters instead, e.g., Parameters["compression_mode"] = "force" The inline compression mode in Bluestore OSD to set to (options are: none, passive, aggressive, force) Do NOT set a default value for kubebuilder as this will override the Parameters'
                      enum:
                        - none
                        - passive
                        - aggressive
                        - force
                        - ""
                      nullable: true
                      type: string
                    crushRoot:
                      description: The root of the crush hierarchy utilized by the pool
                      nullable: true
                      type: string
                    deviceClass:
                      description: The device class the OSD should set to for use in the pool
                      nullable: true
                      type: string
                    enableRBDStats:
                      description: EnableRBDStats is used to enable gathering of statistics for all RBD images in the pool
                      type: boolean
                    erasureCoded:
                      description: The erasure code settings
                      properties:
                        algorithm:
                          description: The algorithm for erasure coding
                          type: string
                        codingChunks:
                          description: Number of coding chunks per object in an erasure coded storage pool (required for erasure-coded pool type). This is the number of OSDs that can be lost simultaneously before data cannot be recovered.
                          minimum: 0
                          type: integer
                        dataChunks:
                          description: Number of data chunks per object in an erasure coded storage pool (required for erasure-coded pool type). The number of chunks required to recover an object when any single OSD is lost is the same as dataChunks so be aware that the larger the number of data chunks, the higher the cost of recovery.
                          minimum: 0
                          type: integer
                      required:
                        - codingChunks
                        - dataChunks
                      type: object
                    failureDomain:
                      description: 'The failure domain: osd/host/(region or zone if available) - technically also any type in the crush map'
                      type: string
                    mirroring:
                      description: The mirroring settings
                      properties:
                        enabled:
                          description: Enabled whether this pool is mirrored or not
                          type: boolean
                        mode:
                          description: 'Mode is the mirroring mode: either pool or image'
                          type: string
                        peers:
                          description: Peers represents the peers spec
                          nullable: true
                          properties:
                            secretNames:
                              description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                              items:
                                type: string
                              type: array
                          type: object
                        snapshotSchedules:
                          description: SnapshotSchedules is the scheduling of snapshot for mirrored images/pools
                          items:
                            description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                            properties:
                              interval:
                                description: Interval represent the periodicity of the snapshot.
                                type: string
                              path:
                                description: Path is the path to snapshot, only valid for CephFS
                                type: string
                              startTime:
                                description: StartTime indicates when to start the snapshot
                                type: string
                            type: object
                          type: array
                      type: object
                    parameters:
                      additionalProperties:
                        type: string
                      description: Parameters is a list of properties to enable on a given pool
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    quotas:
                      description: The quota settings
                      nullable: true
                      properties:
                        maxBytes:
                          description: MaxBytes represents the quota in bytes Deprecated in favor of MaxSize
                          format: int64
                          type: integer
                        maxObjects:
                          description: MaxObjects represents the quota in objects
                          format: int64
                          type: integer
                        maxSize:
                          description: MaxSize represents the quota in bytes as a string
                          pattern: ^[0-9]+[\.]?[0-9]*([KMGTPE]i|[kMGTPE])?$
                          type: string
                      type: object
                    replicated:
                      description: The replication settings
                      properties:
                        hybridStorage:
                          description: HybridStorage represents hybrid storage tier settings
                          nullable: true
                          properties:
                            primaryDeviceClass:
                              description: PrimaryDeviceClass represents high performance tier (for example SSD or NVME) for Primary OSD
                              minLength: 1
                              type: string
                            secondaryDeviceClass:
                              description: SecondaryDeviceClass represents low performance tier (for example HDDs) for remaining OSDs
                              minLength: 1
                              type: string
                          required:
                            - primaryDeviceClass
                            - secondaryDeviceClass
                          type: object
                        replicasPerFailureDomain:
                          description: ReplicasPerFailureDomain the number of replica in the specified failure domain
                          minimum: 1
                          type: integer
                        requireSafeReplicaSize:
                          description: RequireSafeReplicaSize if false allows you to set replica 1
                          type: boolean
                        size:
                          description: Size - Number of copies per object in a replicated storage pool, including the object itself (required for replicated pool type)
                          minimum: 0
                          type: integer
                        subFailureDomain:
                          description: SubFailureDomain the name of the sub-failure domain
                          type: string
                        targetSizeRatio:
                          description: TargetSizeRatio gives a hint (%) to Ceph in terms of expected consumption of the total cluster capacity
                          type: number
                      required:
                        - size
                      type: object
                    statusCheck:
                      description: The mirroring statusCheck
                      properties:
                        mirror:
                          description: HealthCheckSpec represents the health check of an object store bucket
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                  type: object
                preservePoolsOnDelete:
                  description: Preserve pools on object store deletion
                  type: boolean
                security:
                  description: Security represents security settings
                  nullable: true
                  properties:
                    kms:
                      description: KeyManagementService is the main Key Management option
                      nullable: true
                      properties:
                        connectionDetails:
                          additionalProperties:
                            type: string
                          description: ConnectionDetails contains the KMS connection details (address, port etc)
                          nullable: true
                          type: object
                          x-kubernetes-preserve-unknown-fields: true
                        tokenSecretName:
                          description: TokenSecretName is the kubernetes secret containing the KMS token
                          type: string
                      type: object
                  type: object
                zone:
                  description: The multisite info
                  nullable: true
                  properties:
                    name:
                      description: RGW Zone the Object Store is in
                      type: string
                  required:
                    - name
                  type: object
              type: object
            status:
              description: ObjectStoreStatus represents the status of a Ceph Object Store resource
              properties:
                bucketStatus:
                  description: BucketStatus represents the status of a bucket
                  properties:
                    details:
                      type: string
                    health:
                      description: ConditionType represent a resource's status
                      type: string
                    lastChanged:
                      type: string
                    lastChecked:
                      type: string
                  type: object
                conditions:
                  items:
                    description: Condition represents a status condition on any Rook-Ceph Custom Resource.
                    properties:
                      lastHeartbeatTime:
                        format: date-time
                        type: string
                      lastTransitionTime:
                        format: date-time
                        type: string
                      message:
                        type: string
                      reason:
                        description: ConditionReason is a reason for a condition
                        type: string
                      status:
                        type: string
                      type:
                        description: ConditionType represent a resource's status
                        type: string
                    type: object
                  type: array
                info:
                  additionalProperties:
                    type: string
                  nullable: true
                  type: object
                message:
                  type: string
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  description: ConditionType represent a resource's status
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephobjectstoreusers.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectStoreUser
    listKind: CephObjectStoreUserList
    plural: cephobjectstoreusers
    shortNames:
      - rcou
      - objectuser
    singular: cephobjectstoreuser
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephObjectStoreUser represents a Ceph Object Store Gateway User
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: ObjectStoreUserSpec represent the spec of an Objectstoreuser
              properties:
                capabilities:
                  description: Additional admin-level capabilities for the Ceph object store user
                  nullable: true
                  properties:
                    bucket:
                      description: Admin capabilities to read/write Ceph object store buckets. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
                      enum:
                        - '*'
                        - read
                        - write
                        - read, write
                      type: string
                    metadata:
                      description: Admin capabilities to read/write Ceph object store metadata. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
                      enum:
                        - '*'
                        - read
                        - write
                        - read, write
                      type: string
                    usage:
                      description: Admin capabilities to read/write Ceph object store usage. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
                      enum:
                        - '*'
                        - read
                        - write
                        - read, write
                      type: string
                    user:
                      description: Admin capabilities to read/write Ceph object store users. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
                      enum:
                        - '*'
                        - read
                        - write
                        - read, write
                      type: string
                    zone:
                      description: Admin capabilities to read/write Ceph object store zones. Documented in https://docs.ceph.com/en/latest/radosgw/admin/?#add-remove-admin-capabilities
                      enum:
                        - '*'
                        - read
                        - write
                        - read, write
                      type: string
                  type: object
                displayName:
                  description: The display name for the ceph users
                  type: string
                quotas:
                  description: ObjectUserQuotaSpec can be used to set quotas for the object store user to limit their usage. See the [Ceph docs](https://docs.ceph.com/en/latest/radosgw/admin/?#quota-management) for more
                  nullable: true
                  properties:
                    maxBuckets:
                      description: Maximum bucket limit for the ceph user
                      nullable: true
                      type: integer
                    maxObjects:
                      description: Maximum number of objects across all the user's buckets
                      format: int64
                      nullable: true
                      type: integer
                    maxSize:
                      anyOf:
                        - type: integer
                        - type: string
                      description: Maximum size limit of all objects across all the user's buckets See https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Quantity for more info.
                      nullable: true
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      x-kubernetes-int-or-string: true
                  type: object
                store:
                  description: The store the user will be created in
                  type: string
              type: object
            status:
              description: ObjectStoreUserStatus represents the status Ceph Object Store Gateway User
              properties:
                info:
                  additionalProperties:
                    type: string
                  nullable: true
                  type: object
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephobjectzonegroups.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectZoneGroup
    listKind: CephObjectZoneGroupList
    plural: cephobjectzonegroups
    singular: cephobjectzonegroup
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephObjectZoneGroup represents a Ceph Object Store Gateway Zone Group
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: ObjectZoneGroupSpec represent the spec of an ObjectZoneGroup
              properties:
                realm:
                  description: The display name for the ceph users
                  type: string
              required:
                - realm
              type: object
            status:
              description: Status represents the status of an object
              properties:
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephobjectzones.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectZone
    listKind: CephObjectZoneList
    plural: cephobjectzones
    singular: cephobjectzone
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephObjectZone represents a Ceph Object Store Gateway Zone
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: ObjectZoneSpec represent the spec of an ObjectZone
              properties:
                dataPool:
                  description: The data pool settings
                  nullable: true
                  properties:
                    compressionMode:
                      description: 'DEPRECATED: use Parameters instead, e.g., Parameters["compression_mode"] = "force" The inline compression mode in Bluestore OSD to set to (options are: none, passive, aggressive, force) Do NOT set a default value for kubebuilder as this will override the Parameters'
                      enum:
                        - none
                        - passive
                        - aggressive
                        - force
                        - ""
                      nullable: true
                      type: string
                    crushRoot:
                      description: The root of the crush hierarchy utilized by the pool
                      nullable: true
                      type: string
                    deviceClass:
                      description: The device class the OSD should set to for use in the pool
                      nullable: true
                      type: string
                    enableRBDStats:
                      description: EnableRBDStats is used to enable gathering of statistics for all RBD images in the pool
                      type: boolean
                    erasureCoded:
                      description: The erasure code settings
                      properties:
                        algorithm:
                          description: The algorithm for erasure coding
                          type: string
                        codingChunks:
                          description: Number of coding chunks per object in an erasure coded storage pool (required for erasure-coded pool type). This is the number of OSDs that can be lost simultaneously before data cannot be recovered.
                          minimum: 0
                          type: integer
                        dataChunks:
                          description: Number of data chunks per object in an erasure coded storage pool (required for erasure-coded pool type). The number of chunks required to recover an object when any single OSD is lost is the same as dataChunks so be aware that the larger the number of data chunks, the higher the cost of recovery.
                          minimum: 0
                          type: integer
                      required:
                        - codingChunks
                        - dataChunks
                      type: object
                    failureDomain:
                      description: 'The failure domain: osd/host/(region or zone if available) - technically also any type in the crush map'
                      type: string
                    mirroring:
                      description: The mirroring settings
                      properties:
                        enabled:
                          description: Enabled whether this pool is mirrored or not
                          type: boolean
                        mode:
                          description: 'Mode is the mirroring mode: either pool or image'
                          type: string
                        peers:
                          description: Peers represents the peers spec
                          nullable: true
                          properties:
                            secretNames:
                              description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                              items:
                                type: string
                              type: array
                          type: object
                        snapshotSchedules:
                          description: SnapshotSchedules is the scheduling of snapshot for mirrored images/pools
                          items:
                            description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                            properties:
                              interval:
                                description: Interval represent the periodicity of the snapshot.
                                type: string
                              path:
                                description: Path is the path to snapshot, only valid for CephFS
                                type: string
                              startTime:
                                description: StartTime indicates when to start the snapshot
                                type: string
                            type: object
                          type: array
                      type: object
                    parameters:
                      additionalProperties:
                        type: string
                      description: Parameters is a list of properties to enable on a given pool
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    quotas:
                      description: The quota settings
                      nullable: true
                      properties:
                        maxBytes:
                          description: MaxBytes represents the quota in bytes Deprecated in favor of MaxSize
                          format: int64
                          type: integer
                        maxObjects:
                          description: MaxObjects represents the quota in objects
                          format: int64
                          type: integer
                        maxSize:
                          description: MaxSize represents the quota in bytes as a string
                          pattern: ^[0-9]+[\.]?[0-9]*([KMGTPE]i|[kMGTPE])?$
                          type: string
                      type: object
                    replicated:
                      description: The replication settings
                      properties:
                        hybridStorage:
                          description: HybridStorage represents hybrid storage tier settings
                          nullable: true
                          properties:
                            primaryDeviceClass:
                              description: PrimaryDeviceClass represents high performance tier (for example SSD or NVME) for Primary OSD
                              minLength: 1
                              type: string
                            secondaryDeviceClass:
                              description: SecondaryDeviceClass represents low performance tier (for example HDDs) for remaining OSDs
                              minLength: 1
                              type: string
                          required:
                            - primaryDeviceClass
                            - secondaryDeviceClass
                          type: object
                        replicasPerFailureDomain:
                          description: ReplicasPerFailureDomain the number of replica in the specified failure domain
                          minimum: 1
                          type: integer
                        requireSafeReplicaSize:
                          description: RequireSafeReplicaSize if false allows you to set replica 1
                          type: boolean
                        size:
                          description: Size - Number of copies per object in a replicated storage pool, including the object itself (required for replicated pool type)
                          minimum: 0
                          type: integer
                        subFailureDomain:
                          description: SubFailureDomain the name of the sub-failure domain
                          type: string
                        targetSizeRatio:
                          description: TargetSizeRatio gives a hint (%) to Ceph in terms of expected consumption of the total cluster capacity
                          type: number
                      required:
                        - size
                      type: object
                    statusCheck:
                      description: The mirroring statusCheck
                      properties:
                        mirror:
                          description: HealthCheckSpec represents the health check of an object store bucket
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                  type: object
                metadataPool:
                  description: The metadata pool settings
                  nullable: true
                  properties:
                    compressionMode:
                      description: 'DEPRECATED: use Parameters instead, e.g., Parameters["compression_mode"] = "force" The inline compression mode in Bluestore OSD to set to (options are: none, passive, aggressive, force) Do NOT set a default value for kubebuilder as this will override the Parameters'
                      enum:
                        - none
                        - passive
                        - aggressive
                        - force
                        - ""
                      nullable: true
                      type: string
                    crushRoot:
                      description: The root of the crush hierarchy utilized by the pool
                      nullable: true
                      type: string
                    deviceClass:
                      description: The device class the OSD should set to for use in the pool
                      nullable: true
                      type: string
                    enableRBDStats:
                      description: EnableRBDStats is used to enable gathering of statistics for all RBD images in the pool
                      type: boolean
                    erasureCoded:
                      description: The erasure code settings
                      properties:
                        algorithm:
                          description: The algorithm for erasure coding
                          type: string
                        codingChunks:
                          description: Number of coding chunks per object in an erasure coded storage pool (required for erasure-coded pool type). This is the number of OSDs that can be lost simultaneously before data cannot be recovered.
                          minimum: 0
                          type: integer
                        dataChunks:
                          description: Number of data chunks per object in an erasure coded storage pool (required for erasure-coded pool type). The number of chunks required to recover an object when any single OSD is lost is the same as dataChunks so be aware that the larger the number of data chunks, the higher the cost of recovery.
                          minimum: 0
                          type: integer
                      required:
                        - codingChunks
                        - dataChunks
                      type: object
                    failureDomain:
                      description: 'The failure domain: osd/host/(region or zone if available) - technically also any type in the crush map'
                      type: string
                    mirroring:
                      description: The mirroring settings
                      properties:
                        enabled:
                          description: Enabled whether this pool is mirrored or not
                          type: boolean
                        mode:
                          description: 'Mode is the mirroring mode: either pool or image'
                          type: string
                        peers:
                          description: Peers represents the peers spec
                          nullable: true
                          properties:
                            secretNames:
                              description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                              items:
                                type: string
                              type: array
                          type: object
                        snapshotSchedules:
                          description: SnapshotSchedules is the scheduling of snapshot for mirrored images/pools
                          items:
                            description: SnapshotScheduleSpec represents the snapshot scheduling settings of a mirrored pool
                            properties:
                              interval:
                                description: Interval represent the periodicity of the snapshot.
                                type: string
                              path:
                                description: Path is the path to snapshot, only valid for CephFS
                                type: string
                              startTime:
                                description: StartTime indicates when to start the snapshot
                                type: string
                            type: object
                          type: array
                      type: object
                    parameters:
                      additionalProperties:
                        type: string
                      description: Parameters is a list of properties to enable on a given pool
                      nullable: true
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    quotas:
                      description: The quota settings
                      nullable: true
                      properties:
                        maxBytes:
                          description: MaxBytes represents the quota in bytes Deprecated in favor of MaxSize
                          format: int64
                          type: integer
                        maxObjects:
                          description: MaxObjects represents the quota in objects
                          format: int64
                          type: integer
                        maxSize:
                          description: MaxSize represents the quota in bytes as a string
                          pattern: ^[0-9]+[\.]?[0-9]*([KMGTPE]i|[kMGTPE])?$
                          type: string
                      type: object
                    replicated:
                      description: The replication settings
                      properties:
                        hybridStorage:
                          description: HybridStorage represents hybrid storage tier settings
                          nullable: true
                          properties:
                            primaryDeviceClass:
                              description: PrimaryDeviceClass represents high performance tier (for example SSD or NVME) for Primary OSD
                              minLength: 1
                              type: string
                            secondaryDeviceClass:
                              description: SecondaryDeviceClass represents low performance tier (for example HDDs) for remaining OSDs
                              minLength: 1
                              type: string
                          required:
                            - primaryDeviceClass
                            - secondaryDeviceClass
                          type: object
                        replicasPerFailureDomain:
                          description: ReplicasPerFailureDomain the number of replica in the specified failure domain
                          minimum: 1
                          type: integer
                        requireSafeReplicaSize:
                          description: RequireSafeReplicaSize if false allows you to set replica 1
                          type: boolean
                        size:
                          description: Size - Number of copies per object in a replicated storage pool, including the object itself (required for replicated pool type)
                          minimum: 0
                          type: integer
                        subFailureDomain:
                          description: SubFailureDomain the name of the sub-failure domain
                          type: string
                        targetSizeRatio:
                          description: TargetSizeRatio gives a hint (%) to Ceph in terms of expected consumption of the total cluster capacity
                          type: number
                      required:
                        - size
                      type: object
                    statusCheck:
                      description: The mirroring statusCheck
                      properties:
                        mirror:
                          description: HealthCheckSpec represents the health check of an object store bucket
                          nullable: true
                          properties:
                            disabled:
                              type: boolean
                            interval:
                              description: Interval is the internal in second or minute for the health check to run like 60s for 60 seconds
                              type: string
                            timeout:
                              type: string
                          type: object
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                  type: object
                zoneGroup:
                  description: The display name for the ceph users
                  type: string
              required:
                - dataPool
                - metadataPool
                - zoneGroup
              type: object
            status:
              description: Status represents the status of an object
              properties:
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.1-0.20210420220833-f284e2e8098c
  creationTimestamp: null
  name: cephrbdmirrors.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephRBDMirror
    listKind: CephRBDMirrorList
    plural: cephrbdmirrors
    singular: cephrbdmirror
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.phase
          name: Phase
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: CephRBDMirror represents a Ceph RBD Mirror
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: RBDMirroringSpec represents the specification of an RBD mirror daemon
              properties:
                annotations:
                  additionalProperties:
                    type: string
                  description: The annotations-related configuration to add/set on each Pod related object.
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                count:
                  description: Count represents the number of rbd mirror instance to run
                  minimum: 1
                  type: integer
                labels:
                  additionalProperties:
                    type: string
                  description: The labels-related configuration to add/set on each Pod related object.
                  nullable: true
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                peers:
                  description: Peers represents the peers spec
                  nullable: true
                  properties:
                    secretNames:
                      description: SecretNames represents the Kubernetes Secret names to add rbd-mirror or cephfs-mirror peers
                      items:
                        type: string
                      type: array
                  type: object
                placement:
                  description: The affinity to place the rgw pods (default is to place on any available node)
                  nullable: true
                  properties:
                    nodeAffinity:
                      description: NodeAffinity is a group of node affinity scheduling rules
                      properties:
                        preferredDuringSchedulingIgnoredDuringExecution:
                          description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.
                          items:
                            description: An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).
                            properties:
                              preference:
                                description: A node selector term, associated with the corresponding weight.
                                properties:
                                  matchExpressions:
                                    description: A list of node selector requirements by node's labels.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchFields:
                                    description: A list of node selector requirements by node's fields.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                type: object
                              weight:
                                description: Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
                                format: int32
                                type: integer
                            required:
                              - preference
                              - weight
                            type: object
                          type: array
                        requiredDuringSchedulingIgnoredDuringExecution:
                          description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.
                          properties:
                            nodeSelectorTerms:
                              description: Required. A list of node selector terms. The terms are ORed.
                              items:
                                description: A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
                                properties:
                                  matchExpressions:
                                    description: A list of node selector requirements by node's labels.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchFields:
                                    description: A list of node selector requirements by node's fields.
                                    items:
                                      description: A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: The label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
                                          type: string
                                        values:
                                          description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                type: object
                              type: array
                          required:
                            - nodeSelectorTerms
                          type: object
                      type: object
                    podAffinity:
                      description: PodAffinity is a group of inter pod affinity scheduling rules
                      properties:
                        preferredDuringSchedulingIgnoredDuringExecution:
                          description: The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                          items:
                            description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                            properties:
                              podAffinityTerm:
                                description: Required. A pod affinity term, associated with the corresponding weight.
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              weight:
                                description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                format: int32
                                type: integer
                            required:
                              - podAffinityTerm
                              - weight
                            type: object
                          type: array
                        requiredDuringSchedulingIgnoredDuringExecution:
                          description: If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                          items:
                            description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                            properties:
                              labelSelector:
                                description: A label query over a set of resources, in this case pods.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaceSelector:
                                description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaces:
                                description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                items:
                                  type: string
                                type: array
                              topologyKey:
                                description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                type: string
                            required:
                              - topologyKey
                            type: object
                          type: array
                      type: object
                    podAntiAffinity:
                      description: PodAntiAffinity is a group of inter pod anti affinity scheduling rules
                      properties:
                        preferredDuringSchedulingIgnoredDuringExecution:
                          description: The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.
                          items:
                            description: The weights of all of the matched WeightedPodAffinityTerm fields are added per-node to find the most preferred node(s)
                            properties:
                              podAffinityTerm:
                                description: Required. A pod affinity term, associated with the corresponding weight.
                                properties:
                                  labelSelector:
                                    description: A label query over a set of resources, in this case pods.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaceSelector:
                                    description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                    properties:
                                      matchExpressions:
                                        description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                        items:
                                          description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                          properties:
                                            key:
                                              description: key is the label key that the selector applies to.
                                              type: string
                                            operator:
                                              description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                              type: string
                                            values:
                                              description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                              items:
                                                type: string
                                              type: array
                                          required:
                                            - key
                                            - operator
                                          type: object
                                        type: array
                                      matchLabels:
                                        additionalProperties:
                                          type: string
                                        description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                        type: object
                                    type: object
                                  namespaces:
                                    description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                    items:
                                      type: string
                                    type: array
                                  topologyKey:
                                    description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                    type: string
                                required:
                                  - topologyKey
                                type: object
                              weight:
                                description: weight associated with matching the corresponding podAffinityTerm, in the range 1-100.
                                format: int32
                                type: integer
                            required:
                              - podAffinityTerm
                              - weight
                            type: object
                          type: array
                        requiredDuringSchedulingIgnoredDuringExecution:
                          description: If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.
                          items:
                            description: Defines a set of pods (namely those matching the labelSelector relative to the given namespace(s)) that this pod should be co-located (affinity) or not co-located (anti-affinity) with, where co-located is defined as running on a node whose value of the label with key <topologyKey> matches that of any node on which a pod of the set of pods is running
                            properties:
                              labelSelector:
                                description: A label query over a set of resources, in this case pods.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaceSelector:
                                description: A label query over the set of namespaces that the term applies to. The term is applied to the union of the namespaces selected by this field and the ones listed in the namespaces field. null selector and null or empty namespaces list means "this pod's namespace". An empty selector ({}) matches all namespaces. This field is beta-level and is only honored when PodAffinityNamespaceSelector feature is enabled.
                                properties:
                                  matchExpressions:
                                    description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                    items:
                                      description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                      properties:
                                        key:
                                          description: key is the label key that the selector applies to.
                                          type: string
                                        operator:
                                          description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                          type: string
                                        values:
                                          description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                          items:
                                            type: string
                                          type: array
                                      required:
                                        - key
                                        - operator
                                      type: object
                                    type: array
                                  matchLabels:
                                    additionalProperties:
                                      type: string
                                    description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                    type: object
                                type: object
                              namespaces:
                                description: namespaces specifies a static list of namespace names that the term applies to. The term is applied to the union of the namespaces listed in this field and the ones selected by namespaceSelector. null or empty namespaces list and null namespaceSelector means "this pod's namespace"
                                items:
                                  type: string
                                type: array
                              topologyKey:
                                description: This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches that of any node on which any of the selected pods is running. Empty topologyKey is not allowed.
                                type: string
                            required:
                              - topologyKey
                            type: object
                          type: array
                      type: object
                    tolerations:
                      description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>
                      items:
                        description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                        properties:
                          effect:
                            description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                            type: string
                          key:
                            description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                            type: string
                          operator:
                            description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                            type: string
                          tolerationSeconds:
                            description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                            format: int64
                            type: integer
                          value:
                            description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                            type: string
                        type: object
                      type: array
                    topologySpreadConstraints:
                      description: TopologySpreadConstraint specifies how to spread matching pods among the given topology
                      items:
                        description: TopologySpreadConstraint specifies how to spread matching pods among the given topology.
                        properties:
                          labelSelector:
                            description: LabelSelector is used to find matching pods. Pods that match this label selector are counted to determine the number of pods in their corresponding topology domain.
                            properties:
                              matchExpressions:
                                description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                                items:
                                  description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                                  properties:
                                    key:
                                      description: key is the label key that the selector applies to.
                                      type: string
                                    operator:
                                      description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                      type: string
                                    values:
                                      description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                      items:
                                        type: string
                                      type: array
                                  required:
                                    - key
                                    - operator
                                  type: object
                                type: array
                              matchLabels:
                                additionalProperties:
                                  type: string
                                description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                                type: object
                            type: object
                          maxSkew:
                            description: 'MaxSkew describes the degree to which pods may be unevenly distributed. When whenUnsatisfiable=DoNotSchedule, it is the maximum permitted difference between the number of matching pods in the target topology and the global minimum. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 1/1/0: | zone1 | zone2 | zone3 | |   P   |   P   |       | - if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become 1/1/1; scheduling it onto zone1(zone2) would make the ActualSkew(2-0) on zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be scheduled onto any zone. When whenUnsatisfiable=ScheduleAnyway, it is used to give higher precedence to topologies that satisfy it. It''s a required field. Default value is 1 and 0 is not allowed.'
                            format: int32
                            type: integer
                          topologyKey:
                            description: TopologyKey is the key of node labels. Nodes that have a label with this key and identical values are considered to be in the same topology. We consider each <key, value> as a "bucket", and try to put balanced number of pods into each bucket. It's a required field.
                            type: string
                          whenUnsatisfiable:
                            description: 'WhenUnsatisfiable indicates how to deal with a pod if it doesn''t satisfy the spread constraint. - DoNotSchedule (default) tells the scheduler not to schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in any location,   but giving higher precedence to topologies that would help reduce the   skew. A constraint is considered "Unsatisfiable" for an incoming pod if and only if every possible node assignment for that pod would violate "MaxSkew" on some topology. For example, in a 3-zone cluster, MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1: | zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies MaxSkew(1). In other words, the cluster can still be imbalanced, but scheduler won''t make it *more* imbalanced. It''s a required field.'
                            type: string
                        required:
                          - maxSkew
                          - topologyKey
                          - whenUnsatisfiable
                        type: object
                      type: array
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                priorityClassName:
                  description: PriorityClassName sets priority class on the rbd mirror pods
                  type: string
                resources:
                  description: The resource requirements for the rbd mirror pods
                  nullable: true
                  properties:
                    limits:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                    requests:
                      additionalProperties:
                        anyOf:
                          - type: integer
                          - type: string
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        x-kubernetes-int-or-string: true
                      description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      type: object
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
              required:
                - count
              type: object
            status:
              description: Status represents the status of an object
              properties:
                observedGeneration:
                  description: ObservedGeneration is the latest generation observed by the controller.
                  format: int64
                  type: integer
                phase:
                  type: string
              type: object
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: objectbucketclaims.objectbucket.io
spec:
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
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                storageClassName:
                  type: string
                bucketName:
                  type: string
                generateBucketName:
                  type: string
                additionalConfig:
                  type: object
                  nullable: true
                  x-kubernetes-preserve-unknown-fields: true
                objectBucketName:
                  type: string
            status:
              type: object
              x-kubernetes-preserve-unknown-fields: true
      subresources:
        status: {}
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: objectbuckets.objectbucket.io
spec:
  group: objectbucket.io
  names:
    kind: ObjectBucket
    listKind: ObjectBucketList
    plural: objectbuckets
    singular: objectbucket
    shortNames:
      - ob
      - obs
  scope: Cluster
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                storageClassName:
                  type: string
                endpoint:
                  type: object
                  nullable: true
                  properties:
                    bucketHost:
                      type: string
                    bucketPort:
                      type: integer
                      format: int32
                    bucketName:
                      type: string
                    region:
                      type: string
                    subRegion:
                      type: string
                    additionalConfig:
                      type: object
                      nullable: true
                      x-kubernetes-preserve-unknown-fields: true
                authentication:
                  type: object
                  nullable: true
                  items:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                additionalState:
                  type: object
                  nullable: true
                  x-kubernetes-preserve-unknown-fields: true
                reclaimPolicy:
                  type: string
                claimRef:
                  type: object
                  nullable: true
                  x-kubernetes-preserve-unknown-fields: true
            status:
              type: object
              x-kubernetes-preserve-unknown-fields: true
      subresources:
        status: {}
`)))
