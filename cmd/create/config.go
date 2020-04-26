package create

import (
	"encoding/base64"
	"fmt"
	"github.com/lithammer/dedent"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var (
	K2ClusterObjTempl = template.Must(template.New("etcdSslCfg").Parse(
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
    `)))
)

func GenerateK2ClusterObjStr(mgr *manager.Manager, index int) (string, error) {
	return util.Render(K2ClusterObjTempl, util.Data{})
}

func GenerateK2ClusterObj() error {
	K2ClusterObjStr, _ := GenerateK2ClusterObjStr(nil, 0)
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
