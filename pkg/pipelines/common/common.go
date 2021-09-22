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

	KubeBinaries = "KubeBinaries"

	TmpDir                       = "/tmp/kubekey/"
	BinDir                       = "/usr/local/bin"
	KubeConfigDir                = "/etc/kubernetes"
	KubeCertDir                  = "/etc/kubernetes/pki"
	KubeManifestDir              = "/etc/kubernetes/manifests"
	KubeScriptDir                = "/usr/local/bin/kube-scripts"
	KubeletFlexvolumesPluginsDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"

	ETCDCertDir = "/etc/ssl/etcd/ssl"

	HaproxyDir = "/etc/kubekey/haproxy"

	IPv4Regexp = "[\\d]+\\.[\\d]+\\.[\\d]+\\.[\\d]+"
	IPv6Regexp = "[a-f0-9]{1,4}(:[a-f0-9]{1,4}){7}|[a-f0-9]{1,4}(:[a-f0-9]{1,4}){0,7}::[a-f0-9]{0,4}(:[a-f0-9]{1,4}){0,7}"

	Calico  = "calico"
	Flannel = "flannel"
	Cilium  = "cilium"
	Kubeovn = "kubeovn"

	Docker     = "docker"
	Crictl     = "crictl"
	Conatinerd = "containerd"
	Crio       = "crio"
	Isula      = "isula"

	// global cache key
	// PreCheckModule
	NodePreCheck = "nodePreCheck"

	// ETCDModule
	ETCDCluster = "etcdCluster"
	ETCDName    = "etcdName"
	ETCDExist   = "etcdExist"

	// KubernetesModule
	ClusterStatus = "clusterStatus"
	ClusterExist  = "clusterExist"

	// CertsModule
	Certificate   = "certificate"
	CaCertificate = "caCertificate"
)
