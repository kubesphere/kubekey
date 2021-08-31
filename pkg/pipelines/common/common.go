package common

const (
	AllInOne = "allInOne"
	File     = "file"
	Operator = "operator"

	Master = "master"
	Worker = "worker"
	Etcd   = "etcd"
	K8s    = "k8s"

	BinDir                       = "/usr/local/bin"
	KubeConfigDir                = "/etc/kubernetes"
	KubeCertDir                  = "/etc/kubernetes/pki"
	KubeManifestDir              = "/etc/kubernetes/manifests"
	KubeScriptDir                = "/usr/local/bin/kube-scripts"
	KubeletFlexvolumesPluginsDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
)
