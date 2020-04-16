package v1alpha1

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"strconv"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme, serializer.EnableStrict)

func LoadClusterCfg(clusterCfgPath string, logger *log.Logger) (*ClusterCfg, error) {
	if len(clusterCfgPath) == 0 {
		return nil, errors.New("cluster configuration path not provided")
	}

	cluster, err := ioutil.ReadFile(clusterCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read the given cluster configuration file")
	}

	return ParseClusterCfg(cluster)
}

func ParseClusterCfg(cluster []byte) (*ClusterCfg, error) {
	initCfg := &ClusterCfg{}
	if err := runtime.DecodeInto(Codecs.UniversalDecoder(), cluster, initCfg); err != nil {
		return nil, err
	}

	return GetDefaultClusterCfg(initCfg)
}

func GetDefaultClusterCfg(cfg *ClusterCfg) (*ClusterCfg, error) {
	internalCfg := &ClusterCfg{}

	// Default and convert to the internal API type
	Scheme.Default(cfg)
	if err := Scheme.Convert(cfg, internalCfg, nil); err != nil {
		return nil, errors.Wrap(err, "unable to convert versioned to internal cluster object")
	}

	return internalCfg, nil
}

func SetDefaultClusterCfg(cfg *ClusterCfg) *ClusterCfg {
	clusterCfg := &ClusterCfg{}

	cfg.Hosts = SetDefaultHostsCfg(cfg)
	cfg.LBKubeApiserver = SetDefaultLBCfg(cfg)
	cfg.Network = SetDefaultNetworkCfg(cfg)

	if cfg.KubeImageRepo == "" {
		cfg.KubeImageRepo = DefaultKubeImageRepo
	}
	if cfg.KubeClusterName == "" {
		cfg.KubeClusterName = DefaultClusterName
	}
	if cfg.KubeVersion == "" {
		cfg.KubeVersion = DefaultKubeVersion
	}
	clusterCfg = cfg
	return clusterCfg
}

func SetDefaultHostsCfg(cfg *ClusterCfg) []HostCfg {
	var hostscfg []HostCfg
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

func SetDefaultLBCfg(cfg *ClusterCfg) LBKubeApiserverCfg {
	masterHosts := []HostCfg{}
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
		cfg.LBKubeApiserver.Domain = DefaultLBDomain
	}
	if cfg.LBKubeApiserver.Port == "" {
		cfg.LBKubeApiserver.Port = DefaultLBPort
	}
	defaultLbCfg := cfg.LBKubeApiserver
	return defaultLbCfg
}

func SetDefaultNetworkCfg(cfg *ClusterCfg) NetworkConfig {
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
