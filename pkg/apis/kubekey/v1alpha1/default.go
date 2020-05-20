package v1alpha1

import (
	"strconv"
)

const (
	DefaultPreDir         = "kubekey"
	DefaultSSHPort        = "22"
	DefaultDockerSockPath = "/var/run/docker.sock"
	DefaultLBPort         = "6443"
	DefaultLBDomain       = "lb.kubesphere.local"
	DefaultNetworkPlugin  = "calico"
	DefaultPodsCIDR       = "10.233.64.0/18"
	DefaultServiceCIDR    = "10.233.0.0/18"
	DefaultKubeImageRepo  = "kubesphere"
	DefaultClusterName    = "cluster.local"
	DefaultArch           = "amd64"
	DefaultHostName       = "allinone"
	DefaultEtcdRepo       = "kubesphere"
	DefaultEtcdVersion    = "v3.3.12"
	DefaultEtcdPort       = "2379"
	DefaultKubeVersion    = "v1.17.5"
	DefaultCalicoVersion  = "v3.13.0"
	DefaultFlannelVersion = "v0.11.0"
	DefaultCniVersion     = "v0.8.6"
	DefaultHelmVersion    = "v3.2.1"
	ETCDRole              = "etcd"
	MasterRole            = "master"
	WorkerRole            = "worker"
)

func (cfg *ClusterSpec) SetDefaultClusterSpec() (*ClusterSpec, *HostGroups) {
	clusterCfg := ClusterSpec{}

	clusterCfg.Hosts = SetDefaultHostsCfg(cfg)
	clusterCfg.RoleGroups = cfg.RoleGroups
	hostGroups := clusterCfg.GroupHosts()

	clusterCfg.ControlPlaneEndpoint = SetDefaultLBCfg(cfg, hostGroups.Master)
	clusterCfg.Network = SetDefaultNetworkCfg(cfg)
	clusterCfg.Kubernetes = SetDefaultClusterCfg(cfg)
	clusterCfg.Registry = cfg.Registry
	clusterCfg.Storage = SetDefaultStorageCfg(cfg)
	clusterCfg.KubeSphere = cfg.KubeSphere
	if cfg.Kubernetes.ImageRepo == "" {
		clusterCfg.Kubernetes.ImageRepo = DefaultKubeImageRepo
	}
	if cfg.Kubernetes.ClusterName == "" {
		clusterCfg.Kubernetes.ClusterName = DefaultClusterName
	}
	if cfg.Kubernetes.Version == "" {
		clusterCfg.Kubernetes.Version = DefaultKubeVersion
	}

	return &clusterCfg, hostGroups
}

func SetDefaultHostsCfg(cfg *ClusterSpec) []HostCfg {
	var hostscfg []HostCfg
	if len(cfg.Hosts) == 0 {
		return nil
	}
	for _, host := range cfg.Hosts {
		if len(host.Address) == 0 && len(host.InternalAddress) > 0 {
			host.Address = host.InternalAddress
		}
		if len(host.InternalAddress) == 0 && len(host.Address) > 0 {
			host.InternalAddress = host.Address
		}
		if host.User == "" {
			host.User = "root"
		}
		if host.Port == "" {
			host.Port = strconv.Itoa(22)
		}

		hostscfg = append(hostscfg, host)
	}
	return hostscfg
}

func SetDefaultLBCfg(cfg *ClusterSpec, masterGroup []HostCfg) ControlPlaneEndpoint {

	if cfg.ControlPlaneEndpoint.Address == "" {
		cfg.ControlPlaneEndpoint.Address = masterGroup[0].InternalAddress
	}
	if cfg.ControlPlaneEndpoint.Domain == "" {
		cfg.ControlPlaneEndpoint.Domain = DefaultLBDomain
	}
	if cfg.ControlPlaneEndpoint.Port == "" {
		cfg.ControlPlaneEndpoint.Port = DefaultLBPort
	}
	defaultLbCfg := cfg.ControlPlaneEndpoint
	return defaultLbCfg
}

func SetDefaultNetworkCfg(cfg *ClusterSpec) NetworkConfig {
	if cfg.Network.Plugin == "" {
		cfg.Network.Plugin = DefaultNetworkPlugin
	}
	if cfg.Network.KubePodsCIDR == "" {
		cfg.Network.KubePodsCIDR = DefaultPodsCIDR
	}
	if cfg.Network.KubeServiceCIDR == "" {
		cfg.Network.KubeServiceCIDR = DefaultServiceCIDR
	}

	defaultNetworkCfg := cfg.Network

	return defaultNetworkCfg
}

func SetDefaultClusterCfg(cfg *ClusterSpec) Kubernetes {
	if cfg.Kubernetes.Version == "" {
		cfg.Kubernetes.Version = DefaultKubeVersion
	}
	if cfg.Kubernetes.ImageRepo == "" {
		cfg.Kubernetes.ImageRepo = DefaultKubeImageRepo
	}
	if cfg.Kubernetes.ClusterName == "" {
		cfg.Kubernetes.ClusterName = DefaultClusterName
	}

	defaultClusterCfg := cfg.Kubernetes

	return defaultClusterCfg
}

func SetDefaultStorageCfg(cfg *ClusterSpec) Storage {
	if cfg.Storage.LocalVolume.StorageClassName != "" {
		cfg.Storage.LocalVolume.Enabled = true
	}
	if cfg.Storage.NfsClient.StorageClassName != "" {
		cfg.Storage.NfsClient.Enabled = true
	}
	if cfg.Storage.CephRBD.StorageClassName != "" {
		cfg.Storage.CephRBD.Enabled = true
	}
	if cfg.Storage.GlusterFS.StorageClassName != "" {
		cfg.Storage.GlusterFS.Enabled = true
	}

	if cfg.Storage.DefaultStorageClass != "" {
		switch cfg.Storage.DefaultStorageClass {
		case "local":
			cfg.Storage.LocalVolume.IsDefaultClass = true
		case "nfs":
			cfg.Storage.NfsClient.IsDefaultClass = true
		case "rbd":
			cfg.Storage.CephRBD.IsDefaultClass = true
		case "glusterfs":
			cfg.Storage.GlusterFS.IsDefaultClass = true
		}
	}

	defaultStorageCfg := cfg.Storage

	return defaultStorageCfg
}
