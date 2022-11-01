/*
 Copyright 2022 The KubeSphere Authors.

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

package types

// DefaultK3sConfigLocation is the default location for the k3s config file.
const DefaultK3sConfigLocation = "/etc/rancher/k3s/config.yaml"

// K3sServerConfiguration is the configuration for the k3s server.
type K3sServerConfiguration struct {
	// Database
	DataStoreEndPoint string `json:"datastore-endpoint,omitempty"`
	DataStoreCAFile   string `json:"datastore-cafile,omitempty"`
	DataStoreCertFile string `json:"datastore-certfile,omitempty"`
	DataStoreKeyFile  string `json:"datastore-keyfile,omitempty"`

	// Cluster
	Token       string `json:"token,omitempty"`
	TokenFile   string `json:"token-file,omitempty"`
	Server      string `json:"server,omitempty"`
	ClusterInit bool   `json:"cluster-init,omitempty"`

	// Listener
	// BindAddress k3s bind address.
	BindAddress string `json:"bind-address,omitempty"`
	// HTTPSListenPort HTTPS listen port.
	HTTPSListenPort int `json:"https-listen-port,omitempty"`
	// AdvertiseAddress IP address that apiserver uses to advertise to members of the cluster.
	AdvertiseAddress string `json:"advertise-address,omitempty"`
	// AdvertisePort Port that apiserver uses to advertise to members of the cluster (default: listen-port).
	AdvertisePort int `json:"advertise-port,omitempty"`
	// TLSSan Add additional hostname or IP as a Subject Alternative Name in the TLS cert.
	TLSSan string `json:"tls-san,omitempty"`

	// Networking
	// ClusterCIDR Network CIDR to use for pod IPs.
	ClusterCIDR string `json:"cluster-cidr,omitempty"`
	// ServiceCIDR Network CIDR to use for services IPs.
	ServiceCIDR string `json:"service-cidr,omitempty"`
	// ServiceNodePortRange Port range to reserve for services with NodePort visibility.
	ServiceNodePortRange string `json:"service-node-port-range,omitempty"`
	// ClusterDNS cluster IP for coredns service. Should be in your service-cidr range.
	ClusterDNS string `json:"cluster-dns,omitempty"`
	// ClusterDomain cluster Domain.
	ClusterDomain string `json:"cluster-domain,omitempty"`
	// FlannelBackend One of ‘none’, ‘vxlan’, ‘ipsec’, ‘host-gw’, or ‘wireguard’. (default: vxlan)
	FlannelBackend string `json:"flannel-backend,omitempty"`

	// Kubernetes components
	// Disable do not deploy packaged components and delete any deployed components
	// (valid items: coredns, servicelb, traefik,local-storage, metrics-server).
	Disable string `json:"disable,omitempty"`
	// DisableKubeProxy disable running kube-proxy.
	DisableKubeProxy bool `json:"disable-kube-roxy,omitempty"`
	// DisableNetworkPolicy disable k3s default network policy controller.
	DisableNetworkPolicy bool `json:"disable-network-policy,omitempty"`
	// DisableHelmController disable Helm controller.
	DisableHelmController bool `json:"disable-helm-controller,omitempty"`

	// Kubernetes processes
	// DisableCloudController Disable k3s default cloud controller manager.
	DisableCloudController bool `json:"disable-cloud-controller,omitempty"`
	// KubeAPIServerArgs Customized flag for kube-apiserver process.
	KubeAPIServerArgs []string `json:"kube-apiserver-arg,omitempty"`
	// KubeControllerManagerArgs Customized flag for kube-controller-manager process.
	KubeControllerManagerArgs []string `json:"kube-controller-manager-arg,omitempty"`
	// KubeSchedulerArgs Customized flag for kube-scheduler process.
	KubeSchedulerArgs []string `json:"kube-scheduler-args,omitempty"`

	// Agent
	K3sAgentConfiguration `json:",inline"`
}

// K3sAgentConfiguration is the configuration for the k3s agent.
type K3sAgentConfiguration struct {
	// Cluster
	Token     string `json:"token,omitempty"`
	TokenFile string `json:"token-file,omitempty"`
	Server    string `json:"server,omitempty"`

	// NodeName k3s node name.
	NodeName string `json:"node-name,omitempty"`
	// NodeLabels registering and starting kubelet with set of labels.
	NodeLabels []string `json:"node-label,omitempty"`
	// NodeTaints registering and starting kubelet with set of taints.
	NodeTaints []string `json:"node-taint,omitempty"`
	// SeLinux Enable SELinux in containerd
	SeLinux bool `json:"selinux,omitempty"`
	// LBServerPort
	// Local port for supervisor client load-balancer.
	// If the supervisor and apiserver are not colocated an additional port 1 less than this port
	// will also be used for the apiserver client load-balancer. (default: 6444)
	LBServerPort int `json:"lb-server-port,omitempty"`
	// DataDir Folder to hold state.
	DataDir string `json:"data-dir,omitempty"`

	// Runtime
	// ContainerRuntimeEndpoint Disable embedded containerd and use alternative CRI implementation.
	ContainerRuntimeEndpoint string `json:"container-runtime-endpoint,omitempty"`
	// PauseImage Customized pause image for containerd or Docker sandbox.
	PauseImage string `json:"pause-image,omitempty"`
	// PrivateRegistry Path to a private registry configuration file.
	PrivateRegistry string `json:"private-registry,omitempty"`

	// Networking
	// NodeIP IP address to advertise for node.
	NodeIP string `json:"node-ip,omitempty"`
	// NodeExternalIP External IP address to advertise for node.
	NodeExternalIP string `json:"node-external-ip,omitempty"`
	// ResolvConf Path to Kubelet resolv.conf file.
	ResolvConf string `json:"resolv-conf,omitempty"`

	// Kubernetes
	// KubeletArgs Customized flag for kubelet process.
	KubeletArgs []string `json:"kubelet-arg,omitempty"`
	// KubeProxyArgs Customized flag for kube-proxy process.
	KubeProxyArgs []string `json:"kube-proxy-arg,omitempty"`
}
