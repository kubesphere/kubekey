package common

const (
	LocalHost = "localhost"

	AllInOne = "allInOne"
	File     = "file"
	Operator = "operator"

	Master = "master"
	Worker = "worker"
	ETCD   = "etcd"
	K8s    = "k8s"

	TmpDir                       = "/tmp/kubekey/"
	BinDir                       = "/usr/local/bin"
	KubeConfigDir                = "/etc/kubernetes"
	KubeCertDir                  = "/etc/kubernetes/pki"
	KubeManifestDir              = "/etc/kubernetes/manifests"
	KubeScriptDir                = "/usr/local/bin/kube-scripts"
	KubeletFlexvolumesPluginsDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"

	ETCDCertDir = "/etc/ssl/etcd/ssl"
)
