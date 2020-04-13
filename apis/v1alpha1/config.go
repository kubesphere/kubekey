package v1alpha1

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetClusterCfg(clusterCfgFile string) *ClusterCfg {
	if clusterCfgFile != "" {
		clusterInfo, err := ResolveClusterInfoFile(clusterCfgFile)
		if err != nil {
			log.Fatal("Failed to parse the configuration file: ", err)
		}
		return clusterInfo
	} else {
		clusterInfo := &ClusterCfg{
			Hosts: []HostCfg{{
				Role: []string{"etcd", "master", "worker"},
				User: "root",
			}},
			Network: NetworkConfig{
				Plugin:          DefaultNetworkPlugin,
				KubePodsCIDR:    DefaultPodsCIDR,
				KubeServiceCIDR: DefaultServiceCIDR,
			},
		}
		return clusterInfo
	}
}

func ResolveClusterInfoFile(configFile string) (*ClusterCfg, error) {
	fp, err := filepath.Abs(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup current directory name: %v", err)
	}
	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("can not find cluster info file: %v", err)
	}
	defer file.Close()

	clusterInfo, err := GetYamlFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return clusterInfo, nil
}

func GetYamlFile(filePath string) (*ClusterCfg, error) {
	result := ClusterCfg{}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	//var m HostJson
	err = yaml.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
