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

package kubernetes

import (
	"encoding/base64"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"text/template"
)

var (
	// k8sCertsRenewServiceTempl defines the template of k8s-certs-renew service for systemd.
	k8sCertsRenewServiceTempl = template.Must(template.New("k8sCertsRenewService").Parse(
		dedent.Dedent(`[Unit]
Description=Renew K8S control plane certificates

[Service]
Type=oneshot
ExecStart=/usr/local/bin/kube-scripts/k8s-certs-renew.sh

    `)))
	// k8sCertsRenewTimerTempl defines the template of k8s-certs-renew timer for systemd.
	k8sCertsRenewTimerTempl = template.Must(template.New("k8sCertsRenewTimer").Parse(
		dedent.Dedent(`[Unit]
Description=Timer to renew K8S control plane certificates

[Timer]
OnCalendar=Mon *-*-1,2,3,4,5,6,7 03:00:00

[Install]
WantedBy=multi-user.target

    `)))

	// k8sCertsRenewTimerTempl defines the template of k8s-certs-renew timer for systemd.
	k8sCertsRenewScriptTempl = template.Must(template.New("k8sCertsRenewScript").Parse(
		dedent.Dedent(`#!/bin/bash

{{- if .IsKubeadmAlphaCerts }}
kubeadmCerts='/usr/local/bin/kubeadm alpha certs'
{{- else}}
kubeadmCerts='/usr/local/bin/kubeadm certs'
{{- end }}

getCertValidDays() {
  local earliestExpireDate; earliestExpireDate=$(${kubeadmCerts} check-expiration | grep -o "[A-Za-z]\{3,4\}\s\w\w,\s[0-9]\{4,\}\s\w*:\w*\s\w*\s*" | xargs -I {} date -d {} +%s | sort | head -n 1)
  local today; today="$(date +%s)"
  echo -n $(( ($earliestExpireDate - $today) / (24 * 60 * 60) ))
}

echo "## Expiration before renewal ##"
${kubeadmCerts} check-expiration

if [ $(getCertValidDays) -lt 30 ]; then
  echo "## Renewing certificates managed by kubeadm ##"
  ${kubeadmCerts} renew all

  echo "## Restarting control plane pods managed by kubeadm ##"
{{- if .IsDocker}}
  $(which docker | grep docker) ps -af 'name=k8s_POD_(kube-apiserver|kube-controller-manager|kube-scheduler|etcd)-*' -q | /usr/bin/xargs $(which docker | grep docker) rm -f
{{- else}}
  $(which crictl | grep crictl) pods --namespace kube-system --name 'kube-scheduler-*|kube-controller-manager-*|kube-apiserver-*|etcd-*' -q | /usr/bin/xargs $(which crictl | grep crictl) rmp -f
{{- end }}

  echo "## Updating /root/.kube/config ##"
  cp /etc/kubernetes/admin.conf /root/.kube/config
fi

echo "## Waiting for apiserver to be up again ##"
until printf "" 2>>/dev/null >>/dev/tcp/127.0.0.1/6443; do sleep 1; done

echo "## Expiration after renewal ##"
${kubeadmCerts} check-expiration

    `)))
)

// GenerateK8sCertsRenewService is used to generate k8sCertsRenew's service content for systemd.
func GenerateK8sCertsRenewService() (string, error) {
	return util.Render(k8sCertsRenewServiceTempl, util.Data{})
}

// GenerateK8sCertsRenewTimer is used to generate k8sCertsRenew's timer content for systemd.
func GenerateK8sCertsRenewTimer() (string, error) {
	return util.Render(k8sCertsRenewTimerTempl, util.Data{})
}

// GenerateK8sCertsRenewScript is used to generate k8sCertsRenew's script content.
func GenerateK8sCertsRenewScript(mgr *manager.Manager) (string, error) {
	var IsKubeadmAlphaCerts, IsDocker bool

	cmp, err := versionutil.MustParseSemantic(mgr.Cluster.Kubernetes.Version).Compare("v1.20.0")
	if err != nil {
		mgr.Logger.Fatal("Failed to compare version: %v", err)
	}
	if cmp == -1 {
		IsKubeadmAlphaCerts = true
	}

	if mgr.ContainerManager == "docker" || mgr.ContainerManager == "" {
		IsDocker = true
	}

	return util.Render(k8sCertsRenewScriptTempl, util.Data{
		"IsKubeadmAlphaCerts": IsKubeadmAlphaCerts,
		"IsDocker":            IsDocker,
	})
}

func InstallK8sCertsRenew(mgr *manager.Manager) error {
	// Generate k8s certs renew script
	k8sCertsRenewScript, err1 := GenerateK8sCertsRenewScript(mgr)
	if err1 != nil {
		return err1
	}
	k8sCertsRenewScriptBase64 := base64.StdEncoding.EncodeToString([]byte(k8sCertsRenewScript))
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /usr/local/bin/kube-scripts/k8s-certs-renew.sh && chmod +x /usr/local/bin/kube-scripts/k8s-certs-renew.sh\"", k8sCertsRenewScriptBase64), 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate k8s-certs-renew script")
	}

	// Generate k8s certs renew service
	k8sCertsRenewService, err1 := GenerateK8sCertsRenewService()
	if err1 != nil {
		return err1
	}
	k8sCertsRenewServiceBase64 := base64.StdEncoding.EncodeToString([]byte(k8sCertsRenewService))
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/k8s-certs-renew.service\"", k8sCertsRenewServiceBase64), 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate k8s-certs-renew service")
	}

	// Generate k8s certs renew timer
	k8sCertsRenewTimer, err1 := GenerateK8sCertsRenewTimer()
	if err1 != nil {
		return err1
	}
	k8sCertsRenewTimerBase64 := base64.StdEncoding.EncodeToString([]byte(k8sCertsRenewTimer))
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/k8s-certs-renew.timer\"", k8sCertsRenewTimerBase64), 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate k8s-certs-renew timer")
	}

	// Start k8s-certs-renew service.
	if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl enable --now k8s-certs-renew.timer\"", 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to enable k8s-certs-renew service")
	}
	return nil
}
