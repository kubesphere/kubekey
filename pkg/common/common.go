package common

const (
	K3s        = "k3s"
	Kubernetes = "Kubernetes"

	LocalHost = "localhost"

	AllInOne = "allInOne"
	File     = "file"
	Operator = "operator"

	Master = "master"
	Worker = "worker"
	ETCD   = "etcd"
	K8s    = "k8s"

	KubeBinaries = "KubeBinaries"

	TmpDir                       = "/tmp/kubekey"
	BinDir                       = "/usr/local/bin"
	KubeConfigDir                = "/etc/kubernetes"
	KubeAddonsDir                = "/etc/kubernetes/addons"
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
	NodePreCheck      = "nodePreCheck"
	K8sVersion        = "k8sVersion"        // current k8s version
	KubeSphereVersion = "kubeSphereVersion" // current KubeSphere version
	ClusterNodeStatus = "clusterNodeStatus"
	DesiredK8sVersion = "desiredK8sVersion"
	PlanK8sVersion    = "planK8sVersion"
	NodeK8sVersion    = "NodeK8sVersion"

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
