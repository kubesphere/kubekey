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

package config

import (
	"bufio"
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
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
	ClusterObjTempl = template.Must(template.New("Cluster").Parse(
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
    version: {{ .Options.KubeVersion }}
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
  storage:
    defaultStorageClass: {{ .Options.DefaultStorageClass }}
    {{- if .Options.LocalVolumeEnabled }}
    localVolume:
      storageClassName: local
    {{- end }}
    {{- if .Options.NfsClientEnabled }}
    nfsClient:
      storageClassName: nfs-client
      # Hostname of the NFS server(ip or hostname)
      nfsServer: SHOULD_BE_REPLACED
      # Basepath of the mount point
      nfsPath: SHOULD_BE_REPLACED
      nfsVrs3Enabled: false
      nfsArchiveOnDelete: false
    {{- end }}
    {{- if .Options.CephRBDEnabled }}
    rbd:
      storageClassName: rbd
      # Ceph rbd monitor endpoints, for example
      # monitors:
      #   - 172.25.0.1:6789
      #   - 172.25.0.2:6789
      #   - 172.25.0.3:6789
      monitors:
      - SHOULD_BE_REPLACED
      adminID: admin
      # ceph admin secret, for example,
      # adminSecret: AQAnwihbXo+uDxAAD0HmWziVgTaAdai90IzZ6Q==
      adminSecret: TYPE_ADMIN_ACCOUNT_HERE
      userID: admin
      # ceph user secret, for example,
      # userSecret: AQAnwihbXo+uDxAAD0HmWziVgTaAdai90IzZ6Q==
      userSecret: TYPE_USER_SECRET_HERE
      pool: rbd
      fsType: ext4
      imageFormat: 2
      imageFeatures: layering
    {{- end }}
    {{- if .Options.GlusterFSEnabled }}
    glusterfs:
      storageClassName: glusterfs
      restAuthEnabled: true
      # e.g. glusterfs_provisioner_resturl: http://192.168.0.4:8080
      restUrl: SHOULD_BE_REPLACED
      # e.g. glusterfs_provisioner_clusterid: 6a6792ed25405eaa6302da99f2f5e24b
      clusterID: SHOULD_BE_REPLACED
      restUser: admin
      secretName: heketi-secret
      gidMin: 40000
      gidMax: 50000
      volumeType: replicate:2
      # e.g. jwt_admin_key: 123456
      jwtAdminKey: SHOULD_BE_REPLACED
    {{- end }}
{{- end }}


{{ if .Options.KubeSphereEnabled }}
{{ .Options.KubeSphereConfigMap }}
{{ end }}
    `)))
)

type Options struct {
	Name                    string
	KubeVersion             string
	StorageNum              int
	DefaultStorageClass     string
	DefaultStorageClassName string
	LocalVolumeEnabled      bool
	NfsClientEnabled        bool
	CephRBDEnabled          bool
	GlusterFSEnabled        bool
	KubeSphereEnabled       bool
	KubeSphereConfigMap     string
}

func GenerateClusterObjStr(opt *Options, storageNum int) (string, error) {
	return util.Render(ClusterObjTempl, util.Data{
		"StorageNum":  storageNum,
		"KubeVersion": kubekeyapi.DefaultKubeVersion,
		"Options":     opt,
	})
}

func GenerateClusterObj(addons, k8sVersion, ksVersion, name, clusterCfgPath string, ksEnabled bool) error {
	opt := Options{}
	if name != "" {
		output := strings.Split(name, ".")
		opt.Name = output[0]
	} else {
		opt.Name = "config-sample"
	}
	opt.KubeVersion = k8sVersion
	opt.KubeSphereEnabled = ksEnabled
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
		case "rbd":
			opt.CephRBDEnabled = true
			opt.StorageNum++
			if index == 0 {
				opt.DefaultStorageClass = "rbd"
			}
		case "glusterfs":
			opt.GlusterFSEnabled = true
			opt.StorageNum++
			if index == 0 {
				opt.DefaultStorageClass = "glusterfs"
			}
		case "":
		default:
			return errors.New(fmt.Sprintf("This storage plugin is not supported: %s", strings.TrimSpace(addon)))
		}
	}

	if ksEnabled {
		if opt.StorageNum == 0 {
			opt.LocalVolumeEnabled = true
			opt.StorageNum++
			opt.DefaultStorageClass = "localVolume"
		}
		switch strings.TrimSpace(ksVersion) {
		case "":
			opt.KubeSphereConfigMap = kubesphere.V3_0_0
		case "v3.0.0":
			opt.KubeSphereConfigMap = kubesphere.V3_0_0
		case "v2.1.1":
			opt.KubeSphereConfigMap = kubesphere.V2_1_1
		default:
			return errors.New(fmt.Sprintf("Unsupported version: %s", strings.TrimSpace(ksVersion)))
		}
	}

	ClusterObjStr, err := GenerateClusterObjStr(&opt, opt.StorageNum)
	if err != nil {
		return errors.Wrap(err, "Faild to generate cluster config")
	}
	ClusterObjStrBase64 := base64.StdEncoding.EncodeToString([]byte(ClusterObjStr))

	if clusterCfgPath != "" {
		CheckConfigFileStatus(clusterCfgPath)
		cmd := fmt.Sprintf("echo %s | base64 -d > %s", ClusterObjStrBase64, clusterCfgPath)
		output, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to write config to %s: %s", clusterCfgPath, strings.TrimSpace(string(output))))
		}
	} else {
		currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return errors.Wrap(err, "Failed to get current dir")
		}
		CheckConfigFileStatus(fmt.Sprintf("%s/%s.yaml", currentDir, opt.Name))
		cmd := fmt.Sprintf("echo %s | base64 -d > %s/%s.yaml", ClusterObjStrBase64, currentDir, opt.Name)
		err1 := exec.Command("/bin/sh", "-c", cmd).Run()
		if err1 != nil {
			return err1
		}
	}

	return nil
}

func CheckConfigFileStatus(path string) {
	if util.IsExist(path) {
		reader := bufio.NewReader(os.Stdin)
	Loop:
		for {
			fmt.Printf("%s already exists. Are you sure you want to overwrite this config file? [yes/no]: ", path)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input != "" {
				switch input {
				case "yes":
					break Loop
				case "no":
					os.Exit(0)
				}
			}
		}
	}
}
