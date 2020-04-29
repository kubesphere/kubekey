package create

import (
	"encoding/base64"
	"fmt"
	"github.com/lithammer/dedent"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	K2ClusterObjTempl = template.Must(template.New("K2Cluster").Parse(
		dedent.Dedent(`apiVersion: kubekey.io/v1alpha1
kind: K2Cluster
metadata:
  name: demo
spec:
  hosts:
  - hostName: node1
    sshAddress: 172.16.0.2
    internalAddress: 172.16.0.2
    port: "22"
    user: ubuntu
    password: Qcloud@123
    sshKeyPath: ""
    role:
    - etcd
    - master
    - worker
  lbKubeapiserver:
    domain: lb.kubesphere.local
    address: ""
    port: "6443"
  kubeCluster:
    version: v1.17.4
    imageRepo: kubekey
    clusterName: cluster.local
  network:
    plugin: calico
    kube_pods_cidr: 10.233.64.0/18
    kube_service_cidr: 10.233.0.0/18
  registry:
    registryMirrors: []
    insecureRegistries: []
{{- if ne .PluginsNum 0 }}
  plugins:
    {{- if .Options.LocalVolumeEnable }}
    localVolume:
      isDefaultClass: {{ .Options.LocalVolumeIsDefault }}
    {{- end }}
    {{- if .Options.NfsClientEnable }}
    nfsClient:
      isDefaultClass: {{ .Options.NfsClientIsDefault }}
      nfsServer: ""
      nfsPath: ""
      nfsVrs3Enabled: false
      nfsArchiveOnDelete: false
    {{- end }}
{{- end }}
    `)))
)

type PluginOptions struct {
	LocalVolumeEnable    bool
	LocalVolumeIsDefault bool
	NfsClientEnable      bool
	NfsClientIsDefault   bool
}

func GenerateK2ClusterObjStr(opt *PluginOptions, plugins []string) (string, error) {
	return util.Render(K2ClusterObjTempl, util.Data{
		"PluginsNum": len(plugins),
		"Options":    opt,
	})
}

func GenerateK2ClusterObj(addons string) error {
	opt := PluginOptions{}
	addonsList := strings.Split(addons, ",")
	for index, addon := range addonsList {
		switch strings.TrimSpace(addon) {
		case "localVolume":
			opt.LocalVolumeEnable = true
			if index == 0 {
				opt.LocalVolumeIsDefault = true
			}
		case "nfsClient":
			opt.NfsClientEnable = true
			if index == 0 {
				opt.NfsClientIsDefault = true
			}
		default:
			return errors.New(fmt.Sprintf("This plugin is not supported: %s", strings.TrimSpace(addon)))
		}
	}

	K2ClusterObjStr, err := GenerateK2ClusterObjStr(&opt, addonsList)
	if err != nil {
		return errors.Wrap(err, "faild to generate k2cluster config")
	}
	fmt.Println(K2ClusterObjStr)
	K2ClusterObjStrBase64 := base64.StdEncoding.EncodeToString([]byte(K2ClusterObjStr))

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "faild get current dir")
	}
	cmd := fmt.Sprintf("echo %s | base64 -d > %s/k2cluster-demo.yaml", K2ClusterObjStrBase64, currentDir)
	err1 := exec.Command("/bin/sh", "-c", cmd).Run()
	if err1 != nil {
		return err1
	}
	return nil
}
