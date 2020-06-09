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

package csi_neonsan

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"text/template"
	"time"
)

const neonsanPluginInstallation = `
#!/bin/bash

status=$(systemctl is-active neonsan-plugin)
if [[ ${status} = "active" ]]; then
  echo "neonsan-plugin is active, stop it "
  systemctl stop neonsan-plugin
fi

podName=$(kubectl -n kube-system get pod -l app=csi-neonsan,role=controller --field-selector=status.phase=Running -o name |  head -1 | awk -F'/' '{print $2}')
if test -z ${podName}; then
  echo "no controller pod, failed to start, please check"
  exit 1
fi

kubectl -n kube-system -c csi-neonsan cp ${podName}:neonsan-csi-driver /usr/bin/neonsan-plugin
chmod +x /usr/bin/neonsan-plugin

cat > /etc/systemd/system/neonsan-plugin.service  <<EOF
[Unit]
Description=NeonSAN CSI Plugin

[Service]
Type=simple
ExecStart=/usr/bin/neonsan-plugin  --endpoint=unix:///var/lib/kubelet/plugins/neonsan.csi.qingstor.com/csi.sock

Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable neonsan-plugin
systemctl start neonsan-plugin

status=$(systemctl is-active neonsan-plugin)
if [[ ${status} = "active" ]]; then
  echo "neonsan-plugin start successfully"
else
  echo "neonsan-plugin failed to start, please check"
fi`

var NeonsanCSIStorageClassTempl = template.Must(template.New("csi-neonsan").Parse(
	dedent.Dedent(`
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: csi-neonsan
provisioner: neonsan.csi.qingstor.com
parameters:
  fsType: {{ .NeonsanCSI.FsType }}
  replica: "{{ .NeonsanCSI.Replica }}"
  pool: {{ .NeonsanCSI.Pool }}
reclaimPolicy: Delete
allowVolumeExpansion: true
    `)))

func generateNeonsanCSIStorageClassValuesFile(mgr *manager.Manager) (string, error) {
	return util.Render(NeonsanCSIStorageClassTempl, util.Data{
		"NeonsanCSI": mgr.Cluster.Storage.NeonsanCSI,
	})
}

func DeployNeonsanCSI(mgr *manager.Manager) error {
	_, _ = mgr.Runner.RunCmdOutput("/usr/local/bin/helm repo add test https://charts.kubesphere.io/test")
	_, _ = mgr.Runner.RunCmdOutput("/usr/local/bin/helm delete csi-neonsan --namespace kube-system")
	_, err := mgr.Runner.RunCmdOutput("/usr/local/bin/helm install test/csi-neonsan --name-template csi-neonsan --namespace kube-system --devel")
	if err != nil {
		return err
	}

	scFile, err := generateNeonsanCSIStorageClassValuesFile(mgr)
	if err != nil {
		return err
	}
	scValuesFileBase64 := base64.StdEncoding.EncodeToString([]byte(scFile))
	_, err = mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/csi-neonsan-sc.yaml\"", scValuesFileBase64))
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate csi-neonsan storage class values file")
	}

	_, err = mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/csi-neonsan-sc.yaml")
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to deploy csi-neonsan-sc.yaml")
	}

	mgr.Logger.Info("waiting for csi-neonsan to be Running ")
	for i := 0; i < 300; i++ {
		output, err2 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl -n kube-system get pod -l app=csi-neonsan,role=controller | grep csi-neonsan-controller | awk '{print $3}'")
		if err2 != nil {
			return errors.Wrap(errors.WithStack(err2), "Failed to get pod of csi-neonsan")
		}
		if output == "Running" {
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("waiting for csi-neonsan to be Running time out")
}

func InstallNeonsanPlugin(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	_, _ = node, conn
	_, err := mgr.Runner.RunCmdOutput(neonsanPluginInstallation)
	return err
}
