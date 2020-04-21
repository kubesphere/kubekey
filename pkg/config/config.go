package config

import (
	"github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"strconv"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme, serializer.EnableStrict)

func ParseK2ClusterObj(clusterCfgPath string, logger *log.Logger) (*v1alpha1.K2Cluster, error) {
	if len(clusterCfgPath) == 0 {
		return nil, errors.New("cluster configuration path not provided")
	}

	cluster, err := ioutil.ReadFile(clusterCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read the given cluster configuration file")
	}

	return ParseK2ClusterCfg(cluster)
}

func ParseK2ClusterCfg(cluster []byte) (*v1alpha1.K2Cluster, error) {
	initCfg := &v1alpha1.K2Cluster{}
	if err := runtime.DecodeInto(Codecs.UniversalDecoder(), cluster, initCfg); err != nil {
		return nil, err
	}

	return GetDefaultClusterCfg(initCfg)
}

func GetDefaultClusterCfg(cfg *v1alpha1.K2Cluster) (*v1alpha1.K2Cluster, error) {
	internalCfg := &v1alpha1.K2Cluster{}

	// Default and convert to the internal API type
	Scheme.Default(cfg)
	if err := Scheme.Convert(cfg, internalCfg, nil); err != nil {
		return nil, errors.Wrap(err, "unable to convert versioned to internal cluster object")
	}

	return internalCfg, nil
}

func SetDefaultClusterCfg(cfg *v1alpha1.K2ClusterSpec) *v1alpha1.K2ClusterSpec {
	clusterCfg := &v1alpha1.K2ClusterSpec{}

	cfg.Hosts = SetDefaultHostsCfg(cfg)
	cfg.LBKubeApiserver = SetDefaultLBCfg(cfg)
	cfg.Network = SetDefaultNetworkCfg(cfg)

	if cfg.KubeImageRepo == "" {
		cfg.KubeImageRepo = v1alpha1.DefaultKubeImageRepo
	}
	if cfg.KubeClusterName == "" {
		cfg.KubeClusterName = v1alpha1.DefaultClusterName
	}
	if cfg.KubeVersion == "" {
		cfg.KubeVersion = v1alpha1.DefaultKubeVersion
	}
	clusterCfg = cfg
	return clusterCfg
}

func SetDefaultHostsCfg(cfg *v1alpha1.K2ClusterSpec) []v1alpha1.HostCfg {
	var hostscfg []v1alpha1.HostCfg
	if len(cfg.Hosts) == 0 {
		return nil
	}
	for index, host := range cfg.Hosts {
		host.ID = index

		if len(host.SSHAddress) == 0 && len(host.InternalAddress) > 0 {
			host.SSHAddress = host.InternalAddress
		}
		if len(host.InternalAddress) == 0 && len(host.SSHAddress) > 0 {
			host.InternalAddress = host.SSHAddress
		}
		if host.User == "" {
			host.User = "root"
		}
		if host.Port == "" {
			host.Port = strconv.Itoa(22)
		}

		for _, role := range host.Role {
			if role == "etcd" {
				host.IsEtcd = true
			}
			if role == "master" {
				host.IsMaster = true
			}
			if role == "worker" {
				host.IsWorker = true
			}
		}

		hostscfg = append(hostscfg, host)
	}

	return hostscfg
}

func SetDefaultLBCfg(cfg *v1alpha1.K2ClusterSpec) v1alpha1.LBKubeApiserverCfg {
	masterHosts := []v1alpha1.HostCfg{}
	hosts := SetDefaultHostsCfg(cfg)
	for _, host := range hosts {
		for _, role := range host.Role {
			if role == "etcd" {
				host.IsEtcd = true
			}
			if role == "master" {
				host.IsMaster = true
			}
			if role == "worker" {
				host.IsWorker = true
			}
		}
		if host.IsMaster {
			masterHosts = append(masterHosts, host)
		}
	}

	if cfg.LBKubeApiserver.Address == "" {
		cfg.LBKubeApiserver.Address = masterHosts[0].InternalAddress
	}
	if cfg.LBKubeApiserver.Domain == "" {
		cfg.LBKubeApiserver.Domain = v1alpha1.DefaultLBDomain
	}
	if cfg.LBKubeApiserver.Port == "" {
		cfg.LBKubeApiserver.Port = v1alpha1.DefaultLBPort
	}
	defaultLbCfg := cfg.LBKubeApiserver
	return defaultLbCfg
}

func SetDefaultNetworkCfg(cfg *v1alpha1.K2ClusterSpec) v1alpha1.NetworkConfig {
	if cfg.Network.Plugin == "" {
		cfg.Network.Plugin = v1alpha1.DefaultNetworkPlugin
	}
	if cfg.Network.KubePodsCIDR == "" {
		cfg.Network.KubePodsCIDR = v1alpha1.DefaultPodsCIDR
	}
	if cfg.Network.KubeServiceCIDR == "" {
		cfg.Network.KubeServiceCIDR = v1alpha1.DefaultServiceCIDR
	}

	defaultNetworkCfg := cfg.Network

	return defaultNetworkCfg
}
