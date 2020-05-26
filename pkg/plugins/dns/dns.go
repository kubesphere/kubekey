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

package dns

import (
	"encoding/base64"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
)

func OverrideCorednsService(mgr *manager.Manager) error {
	corednsSvc, err := GenerateCorednsService(mgr)
	if err != nil {
		return err
	}
	corednsSvcgBase64 := base64.StdEncoding.EncodeToString([]byte(corednsSvc))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/coredns-svc.yaml\"", corednsSvcgBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate kubeadm config")
	}
	deleteKubednsSvcCmd := "/usr/local/bin/kubectl delete -n kube-system svc kube-dns"
	_, err2 := mgr.Runner.RunCmd(deleteKubednsSvcCmd)
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to delete kubeadm Kube-DNS service")
	}
	_, err3 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/coredns-svc.yaml")
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to create coredns service")
	}
	return nil
}

func DeployNodelocaldns(mgr *manager.Manager) error {
	nodelocaldns, err := GenerateNodelocaldnsService(mgr)
	if err != nil {
		return err
	}
	nodelocaldnsBase64 := base64.StdEncoding.EncodeToString([]byte(nodelocaldns))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/nodelocaldns.yaml\"", nodelocaldnsBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate nodelocaldns manifests")
	}
	_, err2 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/nodelocaldns.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to create nodelocaldns")
	}
	return nil
}
