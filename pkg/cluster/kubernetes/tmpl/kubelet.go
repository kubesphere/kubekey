package tmpl

import (
	"github.com/lithammer/dedent"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util"
	"text/template"
)

var (
	KubeletServiceTempl = template.Must(template.New("kubeletService").Parse(
		dedent.Dedent(`[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/

[Service]
ExecStart=/usr/local/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
    `)))

	KubeletEnvTempl = template.Must(template.New("kubeletEnv").Parse(
		dedent.Dedent(`# Note: This dropin only works with kubeadm and kubelet v1.11+
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
# This is a file that "kubeadm init" and "kubeadm join" generate at runtime, populating the KUBELET_KUBEADM_ARGS variable dynamically
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
# This is a file that the user can use for overrides of the kubelet args as a last resort. Preferably, the user should use
# the .NodeRegistration.KubeletExtraArgs object in the configuration files instead. KUBELET_EXTRA_ARGS should be sourced from this file.
EnvironmentFile=-/etc/default/kubelet
ExecStart=
ExecStart=/usr/local/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS $KUBELET_EXTRA_ARGS
    `)))
)

func GenerateKubeletService(cfg *kubekeyapi.K2ClusterSpec) (string, error) {
	return util.Render(KubeletServiceTempl, util.Data{})
}

func GenerateKubeletEnv(cfg *kubekeyapi.K2ClusterSpec) (string, error) {
	return util.Render(KubeletEnvTempl, util.Data{})
}
