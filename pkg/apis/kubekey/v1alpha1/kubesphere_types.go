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

package v1alpha1

type KubeSphere struct {
	Console       Console       `yaml:"console" json:"console"`
	Common        Common        `yaml:"common" json:"common"`
	Openpitrix    Openpitrix    `yaml:"openpitrix" json:"openpitrix"`
	Monitoring    Monitoring    `yaml:"monitoring" json:"monitoring"`
	Logging       Logging       `yaml:"logging" json:"logging"`
	Devops        Devops        `yaml:"devops" json:"devops"`
	Notification  Notification  `yaml:"notification" json:"notification"`
	Alerting      Alerting      `yaml:"alerting" json:"alerting"`
	ServiceMesh   ServiceMesh   `yaml:"serviceMesh" json:"serviceMesh"`
	MetricsServer MetricsServer `yaml:"metricsServer" json:"metricsServer"`
}

type Console struct {
	EnableMultiLogin bool `yaml:"enableMultiLogin" json:"enableMultiLogin"`
	Port             int  `yaml:"port" json:"port"`
}

type Common struct {
	MysqlVolumeSize    string `yaml:"mysqlVolumeSize" json:"mysqlVolumeSize"`
	MinioVolumeSize    string `yaml:"minioVolumeSize" json:"minioVolumeSize"`
	EtcdVolumeSize     string `yaml:"etcdVolumeSize" json:"etcdVolumeSize"`
	OpenldapVolumeSize string `yaml:"openldapVolumeSize" json:"openldapVolumeSize"`
	RedisVolumSize     string `yaml:"redisVolumSize" json:"redisVolumSize"`
}

type Monitoring struct {
	PrometheusReplicas      int     `yaml:"prometheusReplicas" json:"prometheusReplicas"`
	PrometheusMemoryRequest string  `yaml:"prometheusMemoryRequest" json:"prometheusMemoryRequest"`
	PrometheusVolumeSize    string  `yaml:"prometheusVolumeSize" json:"prometheusVolumeSize"`
	Grafana                 Grafana `yaml:"grafana" json:"grafana"`
}

type Grafana struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Logging struct {
	Enabled                       bool   `yaml:"enabled" json:"enabled"`
	ElasticsearchMasterReplicas   int    `yaml:"elasticsearchMasterReplicas" json:"elasticsearchMasterReplicas"`
	ElasticsearchDataReplicas     int    `yaml:"elasticsearchDataReplicas" json:"elasticsearchDataReplicas"`
	LogsidecarReplicas            int    `yaml:"logsidecarReplicas" json:"logsidecarReplicas"`
	ElasticsearchMasterVolumeSize string `yaml:"elasticsearchMasterVolumeSize" json:"elasticsearchMasterVolumeSize"`
	ElasticsearchDataVolumeSize   string `yaml:"elasticsearchDataVolumeSize" json:"elasticsearchDataVolumeSize"`
	LogMaxAge                     int    `yaml:"logMaxAge" json:"logMaxAge"`
	ElkPrefix                     string `yaml:"elkPrefix" json:"elkPrefix"`
	Kibana                        Kibana `yaml:"kibana" json:"kibana"`
}

type Kibana struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Devops struct {
	Enabled               bool      `yaml:"enabled" json:"enabled"`
	JenkinsMemoryLim      string    `yaml:"jenkinsMemoryLim" json:"jenkinsMemoryLim"`
	JenkinsMemoryReq      string    `yaml:"jenkinsMemoryReq" json:"jenkinsMemoryReq"`
	JenkinsVolumeSize     string    `yaml:"jenkinsVolumeSize" json:"jenkinsVolumeSize"`
	JenkinsJavaOptsXms    string    `yaml:"jenkinsJavaOptsXms" json:"jenkinsJavaOptsXms"`
	JenkinsJavaOptsXmx    string    `yaml:"jenkinsJavaOptsXmx" json:"jenkinsJavaOptsXmx"`
	JenkinsJavaOptsMaxRAM string    `yaml:"jenkinsJavaOptsMaxRAM" json:"jenkinsJavaOptsMaxRAM"`
	Sonarqube             Sonarqube `yaml:"sonarqube" json:"sonarqube"`
}

type Sonarqube struct {
	Enabled              bool   `yaml:"enabled" json:"enabled"`
	PostgresqlVolumeSize string `yaml:"postgresqlVolumeSize" json:"postgresqlVolumeSize"`
}

type Openpitrix struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type ServiceMesh struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Notification struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Alerting struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type MetricsServer struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}
