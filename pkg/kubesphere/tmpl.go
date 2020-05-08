package kubesphere

import (
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"text/template"
)

var (
	KubeSphereTempl = template.Must(template.New("KubeSphere").Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  ks-config.yaml: |
    ---

    persistence:
      storageClass: ""

    etcd:
      monitoring: false
      endpointIps: ""
      port: 2379
      tlsEnable: true

    common:
      mysqlVolumeSize: {{ .Options.Common.MysqlVolumeSize }}
      minioVolumeSize: {{ .Options.Common.MinioVolumeSize }}
      etcdVolumeSize: {{ .Options.Common.EtcdVolumeSize }}
      openldapVolumeSize: {{ .Options.Common.OpenldapVolumeSize }}
      redisVolumSize: {{ .Options.Common.RedisVolumSize }}

    metrics_server:
      enabled: {{ .Options.MetricsServer.Enabled }}

    console:
      enableMultiLogin: {{ .Options.Console.EnableMultiLogin }}  # enable/disable multi login
      port: {{ .Options.Console.Port }}

    monitoring:
      prometheusReplicas: {{ .Options.Monitoring.PrometheusReplicas }}
      prometheusMemoryRequest: {{ .Options.Monitoring.PrometheusMemoryRequest }}
      prometheusVolumeSize: {{ .Options.Monitoring.PrometheusVolumeSize }}
      grafana:
        enabled: {{ .Options.Monitoring.Grafana.Enabled }}

    logging:
      enabled: {{ .Options.Logging.Enabled }}
      elasticsearchMasterReplicas: {{ .Options.Logging.ElasticsearchMasterReplicas }}
      elasticsearchDataReplicas: {{ .Options.Logging.ElasticsearchDataReplicas }}
      logsidecarReplicas: {{ .Options.Logging.LogsidecarReplicas }}
      elasticsearchMasterVolumeSize: {{ .Options.Logging.ElasticsearchMasterVolumeSize }}
      elasticsearchDataVolumeSize: {{ .Options.Logging.ElasticsearchDataVolumeSize }}
      logMaxAge: {{ .Options.Logging.LogMaxAge }}
      elkPrefix: {{ .Options.Logging.ElkPrefix }}
      kibana:
        enabled: {{ .Options.Logging.Kibana.Enabled }}

    openpitrix:
      enabled: {{ .Options.Openpitrix.Enabled }}

    devops:
      enabled: {{ .Options.Devops.Enabled }}
      jenkinsMemoryLim: {{ .Options.Devops.JenkinsMemoryLim }}
      jenkinsMemoryReq: {{ .Options.Devops.JenkinsMemoryReq }}
      jenkinsVolumeSize: {{ .Options.Devops.JenkinsVolumeSize }}
      jenkinsJavaOpts_Xms: {{ .Options.Devops.JenkinsJavaOptsXms}}
      jenkinsJavaOpts_Xmx: {{ .Options.Devops.JenkinsJavaOptsXmx }}
      jenkinsJavaOpts_MaxRAM: {{ .Options.Devops.JenkinsJavaOptsMaxRAM }}
      sonarqube:
        enabled: {{ .Options.Devops.Sonarqube.Enabled }}
        postgresqlVolumeSize: {{ .Options.Devops.Sonarqube.PostgresqlVolumeSize }}

    servicemesh:
      enabled: {{ .Options.ServiceMesh.Enabled }}

    notification:
      enabled: {{ .Options.Notification.Enabled }}

    alerting:
      enabled: {{ .Options.Alerting.Enabled }}

kind: ConfigMap
metadata:
  name: ks-installer
  namespace: kubesphere-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ks-installer
  namespace: kubesphere-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: ks-installer
rules:
- apiGroups:
  - ""
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - apps
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - extensions
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - apiregistration.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - tenant.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - certificates.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - devops.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - monitoring.coreos.com
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - logging.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - jaegertracing.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - storage.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - '*'
  verbs:
  - '*'

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ks-installer
subjects:
- kind: ServiceAccount
  name: ks-installer
  namespace: kubesphere-system
roleRef:
  kind: ClusterRole
  name: ks-installer
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ks-installer
  namespace: kubesphere-system
  labels:
    app: ks-install
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ks-install
  template:
    metadata:
      labels:
        app: ks-install
    spec:
      serviceAccountName: ks-installer
      containers:
      - name: installer
        image: kubespheredev/ks-installer:helm3-dev
        imagePullPolicy: IfNotPresent
    `)))
)

func GenerateKubeSphereYaml(mgr *manager.Manager) (string, error) {
	return util.Render(KubeSphereTempl, util.Data{
		"Options": mgr.Cluster.KubeSphere,
	})
}
