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

package v1beta1

// ServerConfiguration defines the desired state of k3s server configuration.
type ServerConfiguration struct {
	// Database is the database configuration.
	Database Database `json:"database,omitempty"`

	// Listener is the listener configuration.
	Listener Listener `json:"listener,omitempty"`

	// Networking is the networking configuration.
	Networking Networking `json:"networking,omitempty"`

	// Agent is the agent configuration.
	Agent AgentConfiguration `json:"agent,omitempty"`
}

// AgentConfiguration defines the desired state of k3s agent configuration.
type AgentConfiguration struct {
	// Node defines the k3s agent node configuration.
	Node AgentNode `json:"node,omitempty"`

	// Runtime defines the k3s agent runtime configuration.
	Runtime AgentRuntime `json:"runtime,omitempty"`

	// Networking defines the k3s agent networking configuration.
	Networking AgentNetworking `json:"networking,omitempty"`
}

// Database defines the desired state of k3s database configuration.
type Database struct {
	// DataStoreEndPoint specify etcd, Mysql, Postgres, or Sqlite (default) data source name.
	DataStoreEndPoint string `json:"dataStoreEndPoint,omitempty"`

	// DataStoreCAFile TLS Certificate Authority file used to secure datastore backend communication.
	DataStoreCAFile string `json:"dataStoreCAFile,omitempty"`

	// DataStoreCertFile TLS certification file used to secure datastore backend communication.
	DataStoreCertFile string `json:"dataStoreCertFile,omitempty"`

	// DataStoreKeyFile TLS key file used to secure datastore backend communication.
	DataStoreKeyFile string `json:"dataStoreKeyFile,omitempty"`

	// ClusterInit initialize a new cluster using embedded Etcd.
	ClusterInit bool `json:"clusterInit,omitempty"`
}

// Cluster is the desired state of k3s cluster configuration.
type Cluster struct {
	// Token shared secret used to join a server or agent to a cluster.
	Token string `json:"token,omitempty"`

	// TokenFile file containing the cluster-secret/token.
	TokenFile string `json:"tokenFile,omitempty"`

	// Server which server to connect to, used to join a cluster.
	Server string `json:"server,omitempty"`
}

// Listener defines the desired state of k3s listener configuration.
type Listener struct {
	// BindAddress k3s bind address.
	BindAddress string `json:"bindAddress,omitempty"`

	// HTTPSListenPort HTTPS listen port.
	HTTPSListenPort int `json:"httpsListenPort,omitempty"`

	// AdvertiseAddress IP address that apiserver uses to advertise to members of the cluster.
	AdvertiseAddress string `json:"advertiseAddress,omitempty"`

	// AdvertisePort Port that apiserver uses to advertise to members of the cluster (default: listen-port).
	AdvertisePort int `json:"advertisePort,omitempty"`

	// TLSSan Add additional hostname or IP as a Subject Alternative Name in the TLS cert.
	TLSSan string `json:"tlsSan,omitempty"`
}

// Networking defines the desired state of k3s networking configuration.
type Networking struct {
	// ClusterCIDR Network CIDR to use for pod IPs.
	ClusterCIDR string `json:"clusterCIDR,omitempty"`

	// ServiceCIDR Network CIDR to use for services IPs.
	ServiceCIDR string `json:"serviceCIDR,omitempty"`

	// ServiceNodePortRange Port range to reserve for services with NodePort visibility.
	ServiceNodePortRange string `json:"serviceNodePortRange,omitempty"`

	// ClusterDNS cluster IP for coredns service. Should be in your service-cidr range.
	ClusterDNS string `json:"clusterDNS,omitempty"`

	// ClusterDomain cluster Domain.
	ClusterDomain string `json:"clusterDomain,omitempty"`

	// FlannelBackend One of ‘none’, ‘vxlan’, ‘ipsec’, ‘host-gw’, or ‘wireguard’. (default: vxlan)
	FlannelBackend string `json:"flannelBackend,omitempty"`
}

// AgentNode defines the desired state of k3s agent node configuration.
type AgentNode struct {
	// NodeName k3s node name.
	NodeName string `json:"nodeName,omitempty"`

	// NodeLabels registering and starting kubelet with set of labels.
	NodeLabels []string `json:"nodeLabels,omitempty"`

	// NodeTaints registering and starting kubelet with set of taints.
	NodeTaints []string `json:"nodeTaints,omitempty"`

	// SeLinux Enable SELinux in containerd
	SeLinux bool `json:"seLinux,omitempty"`

	// LBServerPort
	// Local port for supervisor client load-balancer.
	// If the supervisor and apiserver are not colocated an additional port 1 less than this port
	// will also be used for the apiserver client load-balancer. (default: 6444)
	LBServerPort int `json:"lbServerPort,omitempty"`

	// DataDir Folder to hold state.
	DataDir string `json:"dataDir,omitempty"`
}

// AgentRuntime defines the desired state of k3s agent runtime configuration.
type AgentRuntime struct {
	// ContainerRuntimeEndpoint Disable embedded containerd and use alternative CRI implementation.
	ContainerRuntimeEndpoint string `json:"containerRuntimeEndpoint,omitempty"`

	// PauseImage Customized pause image for containerd or Docker sandbox.
	PauseImage string `json:"pauseImage,omitempty"`

	// PrivateRegistry Path to a private registry configuration file.
	PrivateRegistry string `json:"privateRegistry,omitempty"`
}

// AgentNetworking defines the desired state of k3s agent networking configuration.
type AgentNetworking struct {
	// NodeIP IP address to advertise for node.
	NodeIP string `json:"nodeIP,omitempty"`

	// NodeExternalIP External IP address to advertise for node.
	NodeExternalIP string `json:"nodeExternalIP,omitempty"`

	// ResolvConf Path to Kubelet resolv.conf file.
	ResolvConf string `json:"resolvConf,omitempty"`
}
