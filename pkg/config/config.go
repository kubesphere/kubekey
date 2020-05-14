package config

import (
	"encoding/base64"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	K2ClusterObjTempl = template.Must(template.New("K2Cluster").Parse(
		dedent.Dedent(`apiVersion: kubekey.kubesphere.io/v1alpha1
kind: Cluster
metadata:
  name: {{ .Options.Name }}
spec:
  hosts:
  - {name: node1, address: 172.16.0.2, internalAddress: 172.16.0.2, user: ubuntu, password: Qcloud@123}
  - {name: node2, address: 172.16.0.2, internalAddress: 172.16.0.2, user: ubuntu, password: Qcloud@123}
  roleGroups:
    etcd:
    - node1
    master: 
    - node1
    worker:
    - node1
    - node2
  controlPlaneEndpoint:
    domain: lb.kubesphere.local
    address: ""
    port: "6443"
  kubernetes:
    version: v1.17.4
    imageRepo: kubesphere
    clusterName: cluster.local
  network:
    plugin: calico
    kube_pods_cidr: 10.233.64.0/18
    kube_service_cidr: 10.233.0.0/18
  registry:
    registryMirrors: []
    insecureRegistries: []
{{- if ne .StorageNum 0 }}
  stroage:
    defaultStorageClass: {{ .Options.DefaultStorageClass }}
    {{- if .Options.LocalVolumeEnabled }}
    localVolume:
      storageClassName: local
    {{- end }}
    {{- if .Options.NfsClientEnabled }}
    nfsClient:
      storageClassName: nfs-client
      nfsServer: ""
      nfsPath: ""
      nfsVrs3Enabled: false
      nfsArchiveOnDelete: false
    {{- end }}
{{- end }}
{{- if .Options.KubeSphereEnabled }}
  kubesphere:
    console:
      enableMultiLogin: False  # enable/disable multi login
      port: 30880
    common:
      mysqlVolumeSize: 20Gi
      minioVolumeSize: 20Gi
      etcdVolumeSize: 20Gi
      openldapVolumeSize: 2Gi
      redisVolumSize: 2Gi
    monitoring:
      prometheusReplicas: 1
      prometheusMemoryRequest: 400Mi
      prometheusVolumeSize: 20Gi
      grafana:
        enabled: false
    logging:
      enabled: false
      elasticsearchMasterReplicas: 1
      elasticsearchDataReplicas: 1
      logsidecarReplicas: 2
      elasticsearchMasterVolumeSize: 4Gi
      elasticsearchDataVolumeSize: 20Gi
      logMaxAge: 7
      elkPrefix: logstash
      kibana:
        enabled: false
    openpitrix:
      enabled: false
    devops:
      enabled: false
      jenkinsMemoryLim: 2Gi
      jenkinsMemoryReq: 1500Mi
      jenkinsVolumeSize: 8Gi
      jenkinsJavaOpts_Xms: 512m
      jenkinsJavaOpts_Xmx: 512m
      jenkinsJavaOpts_MaxRAM: 2g
      sonarqube:
        enabled: false
        postgresqlVolumeSize: 8Gi
    notification:
      enabled: false
    alerting:
      enabled: false
    serviceMesh:
      enabled: false
    metricsServer:
      enabled: false
{{- end }}
    `)))
)

type Options struct {
	Name                    string
	StorageNum              int
	DefaultStorageClass     string
	DefaultStorageClassName string
	LocalVolumeEnabled      bool
	NfsClientEnabled        bool
	KubeSphereEnabled       bool
}

func GenerateK2ClusterObjStr(opt *Options, storageNum int) (string, error) {
	return util.Render(K2ClusterObjTempl, util.Data{
		"StorageNum": storageNum,
		"Options":    opt,
	})
}

func GenerateK2ClusterObj(addons, name string) error {
	opt := Options{}
	if name != "" {
		out := strings.Split(name, ".")
		opt.Name = out[0]
	} else {
		opt.Name = "demo"
	}
	addonsList := strings.Split(addons, ",")
	for index, addon := range addonsList {
		switch strings.TrimSpace(addon) {
		case "localVolume":
			opt.LocalVolumeEnabled = true
			opt.StorageNum++
			if index == 0 {
				opt.DefaultStorageClass = "localVolume"
			}
		case "nfsClient":
			opt.NfsClientEnabled = true
			opt.StorageNum++
			if index == 0 {
				opt.DefaultStorageClass = "nfsClient"
			}
		case "kubesphere":
			opt.KubeSphereEnabled = true
		case "":
		default:
			return errors.New(fmt.Sprintf("This plugin is not supported: %s", strings.TrimSpace(addon)))
		}
	}

	K2ClusterObjStr, err := GenerateK2ClusterObjStr(&opt, opt.StorageNum)
	if err != nil {
		return errors.Wrap(err, "Faild to generate k2cluster config")
	}
	K2ClusterObjStrBase64 := base64.StdEncoding.EncodeToString([]byte(K2ClusterObjStr))

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "Failed to get current dir")
	}
	cmd := fmt.Sprintf("echo %s | base64 -d > %s/%s.yaml", K2ClusterObjStrBase64, currentDir, opt.Name)
	err1 := exec.Command("/bin/sh", "-c", cmd).Run()
	if err1 != nil {
		return err1
	}
	return nil
}
