package config

import (
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

func ParseClusterCfg(clusterCfgPath string, logger *log.Logger) (*kubekeyapi.K2Cluster, error) {
	clusterCfg := kubekeyapi.K2Cluster{}

	if len(clusterCfgPath) == 0 {
		user, _ := user.Current()
		if user.Name != "root" {
			return nil, errors.New(fmt.Sprintf("Current user is %s, Please use root !", user.Name))
		}
		clusterCfg = AllinoneHost(user)
	} else {
		fp, err := filepath.Abs(clusterCfgPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to lookup current directory")
		}
		content, err := ioutil.ReadFile(fp)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read the given cluster configuration file")
		}

		if err := yaml.Unmarshal(content, &clusterCfg); err != nil {
			return nil, errors.Wrap(err, "unable to convert file to yaml")
		}
	}

	defaultK2Cluster := SetDefaultK2Cluster(&clusterCfg)
	return defaultK2Cluster, nil

}

func SetDefaultK2Cluster(obj *kubekeyapi.K2Cluster) *kubekeyapi.K2Cluster {
	defaultCluster := &kubekeyapi.K2Cluster{}
	defaultCluster.APIVersion = obj.APIVersion
	defaultCluster.Kind = obj.APIVersion
	defaultCluster.Spec = SetDefaultK2ClusterSpec(&obj.Spec)
	return defaultCluster
}

func SetDefaultK2ClusterSpec(cfg *kubekeyapi.K2ClusterSpec) kubekeyapi.K2ClusterSpec {
	clusterCfg := kubekeyapi.K2ClusterSpec{}

	clusterCfg.Hosts = SetDefaultHostsCfg(cfg)
	clusterCfg.LBKubeApiserver = SetDefaultLBCfg(cfg)
	clusterCfg.Network = SetDefaultNetworkCfg(cfg)
	clusterCfg.KubeCluster = SetDefaultClusterCfg(cfg)
	clusterCfg.Registry = cfg.Registry
	clusterCfg.Plugins = cfg.Plugins
	if cfg.KubeCluster.ImageRepo == "" {
		clusterCfg.KubeCluster.ImageRepo = kubekeyapi.DefaultKubeImageRepo
	}
	if cfg.KubeCluster.ClusterName == "" {
		clusterCfg.KubeCluster.ClusterName = kubekeyapi.DefaultClusterName
	}
	if cfg.KubeCluster.Version == "" {
		clusterCfg.KubeCluster.Version = kubekeyapi.DefaultKubeVersion
	}
	return clusterCfg
}

func SetDefaultHostsCfg(cfg *kubekeyapi.K2ClusterSpec) []kubekeyapi.HostCfg {
	var hostscfg []kubekeyapi.HostCfg
	if len(cfg.Hosts) == 0 {
		return nil
	}
	clinetNum := 0
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
			if role == "client" {
				clinetNum++
			}
		}
		hostscfg = append(hostscfg, host)
	}

	if clinetNum == 0 {
		for index := range hostscfg {
			if hostscfg[index].IsMaster {
				hostscfg[index].IsClient = true
				hostscfg[index].Role = append(hostscfg[index].Role, "client")
				break
			}
		}
	}
	return hostscfg
}

func SetDefaultLBCfg(cfg *kubekeyapi.K2ClusterSpec) kubekeyapi.LBKubeApiserverCfg {
	masterHosts := []kubekeyapi.HostCfg{}
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
		cfg.LBKubeApiserver.Domain = kubekeyapi.DefaultLBDomain
	}
	if cfg.LBKubeApiserver.Port == "" {
		cfg.LBKubeApiserver.Port = kubekeyapi.DefaultLBPort
	}
	defaultLbCfg := cfg.LBKubeApiserver
	return defaultLbCfg
}

func SetDefaultNetworkCfg(cfg *kubekeyapi.K2ClusterSpec) kubekeyapi.NetworkConfig {
	if cfg.Network.Plugin == "" {
		cfg.Network.Plugin = kubekeyapi.DefaultNetworkPlugin
	}
	if cfg.Network.KubePodsCIDR == "" {
		cfg.Network.KubePodsCIDR = kubekeyapi.DefaultPodsCIDR
	}
	if cfg.Network.KubeServiceCIDR == "" {
		cfg.Network.KubeServiceCIDR = kubekeyapi.DefaultServiceCIDR
	}

	defaultNetworkCfg := cfg.Network

	return defaultNetworkCfg
}

func SetDefaultClusterCfg(cfg *kubekeyapi.K2ClusterSpec) kubekeyapi.KubeCluster {
	if cfg.KubeCluster.Version == "" {
		cfg.KubeCluster.Version = kubekeyapi.DefaultKubeVersion
	}
	if cfg.KubeCluster.ImageRepo == "" {
		cfg.KubeCluster.ImageRepo = kubekeyapi.DefaultKubeImageRepo
	}
	if cfg.KubeCluster.ClusterName == "" {
		cfg.KubeCluster.ClusterName = kubekeyapi.DefaultClusterName
	}

	defaultClusterCfg := cfg.KubeCluster

	return defaultClusterCfg
}

func AllinoneHost(user *user.User) kubekeyapi.K2Cluster {
	allinoneCfg := kubekeyapi.K2Cluster{}
	if err := exec.Command("/bin/sh", "-c", "if [ ! -f \"$HOME/.ssh/id_rsa\" ]; then ssh-keygen -t rsa -P \"\" -f $HOME/.ssh/id_rsa && ls $HOME/.ssh;fi;").Run(); err != nil {
		log.Fatalf("Failed to generate public key: %v", err)
	}
	if out, err := exec.Command("/bin/sh", "-c", "echo \"$(cat $HOME/.ssh/id_rsa.pub)\" >> $HOME/.ssh/authorized_keys").CombinedOutput(); err != nil {
		log.Fatalf("Failed to copy public key to authorized_keys: %v\n%s", err, string(out))
	}

	allinoneCfg.Spec.Hosts = append(allinoneCfg.Spec.Hosts, kubekeyapi.HostCfg{
		HostName:        "ks-allinone",
		SSHAddress:      "",
		InternalAddress: util.LocalIP(),
		Port:            "",
		User:            user.Name,
		Password:        "",
		SSHKeyPath:      fmt.Sprintf("%s/.ssh/id_rsa", user.HomeDir),
		Role:            []string{"master", "worker", "etcd"},
	})
	return allinoneCfg
}
