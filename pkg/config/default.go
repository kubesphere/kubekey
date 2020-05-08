package config

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
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

	//defaultK2Cluster := SetDefaultK2Cluster(&clusterCfg)
	//return defaultK2Cluster, nil
	return &clusterCfg, nil
}

//func SetDefaultK2Cluster(obj *kubekeyapi.K2Cluster) *kubekeyapi.K2Cluster {
//	//fmt.Println(obj)
//	out, _ := json.MarshalIndent(obj, "", "  ")
//	fmt.Println(string(out))
//	defaultCluster := &kubekeyapi.K2Cluster{}
//	defaultCluster.APIVersion = obj.APIVersion
//	defaultCluster.Kind = obj.APIVersion
//	defaultCluster.Spec = SetDefaultK2ClusterSpec(&obj.Spec)
//	return defaultCluster
//}

func SetDefaultK2ClusterSpec(cfg *kubekeyapi.K2ClusterSpec, masterGroup []kubekeyapi.HostCfg) *kubekeyapi.K2ClusterSpec {
	clusterCfg := kubekeyapi.K2ClusterSpec{}

	clusterCfg.Hosts = SetDefaultHostsCfg(cfg)
	clusterCfg.ControlPlaneEndpoint = SetDefaultLBCfg(cfg, masterGroup)
	clusterCfg.Network = SetDefaultNetworkCfg(cfg)
	clusterCfg.Kubernetes = SetDefaultClusterCfg(cfg)
	clusterCfg.Registry = cfg.Registry
	clusterCfg.Storage = cfg.Storage
	clusterCfg.KubeSphere = cfg.KubeSphere
	if cfg.Kubernetes.ImageRepo == "" {
		clusterCfg.Kubernetes.ImageRepo = kubekeyapi.DefaultKubeImageRepo
	}
	if cfg.Kubernetes.ClusterName == "" {
		clusterCfg.Kubernetes.ClusterName = kubekeyapi.DefaultClusterName
	}
	if cfg.Kubernetes.Version == "" {
		clusterCfg.Kubernetes.Version = kubekeyapi.DefaultKubeVersion
	}
	return &clusterCfg
}

func SetDefaultHostsCfg(cfg *kubekeyapi.K2ClusterSpec) []kubekeyapi.HostCfg {
	var hostscfg []kubekeyapi.HostCfg
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

func SetDefaultLBCfg(cfg *kubekeyapi.K2ClusterSpec, masterGroup []kubekeyapi.HostCfg) kubekeyapi.ControlPlaneEndpoint {

	if cfg.ControlPlaneEndpoint.Address == "" {
		cfg.ControlPlaneEndpoint.Address = masterGroup[0].InternalAddress
	}
	if cfg.ControlPlaneEndpoint.Domain == "" {
		cfg.ControlPlaneEndpoint.Domain = kubekeyapi.DefaultLBDomain
	}
	if cfg.ControlPlaneEndpoint.Port == "" {
		cfg.ControlPlaneEndpoint.Port = kubekeyapi.DefaultLBPort
	}
	defaultLbCfg := cfg.ControlPlaneEndpoint
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

func SetDefaultClusterCfg(cfg *kubekeyapi.K2ClusterSpec) kubekeyapi.Kubernetes {
	if cfg.Kubernetes.Version == "" {
		cfg.Kubernetes.Version = kubekeyapi.DefaultKubeVersion
	}
	if cfg.Kubernetes.ImageRepo == "" {
		cfg.Kubernetes.ImageRepo = kubekeyapi.DefaultKubeImageRepo
	}
	if cfg.Kubernetes.ClusterName == "" {
		cfg.Kubernetes.ClusterName = kubekeyapi.DefaultClusterName
	}

	defaultClusterCfg := cfg.Kubernetes

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
		Name:            "ks-allinone",
		Address:         "",
		InternalAddress: util.LocalIP(),
		Port:            "",
		User:            user.Name,
		Password:        "",
		PrivateKeyPath:  fmt.Sprintf("%s/.ssh/id_rsa", user.HomeDir),
	})
	return allinoneCfg
}
