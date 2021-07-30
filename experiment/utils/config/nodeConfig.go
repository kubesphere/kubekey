package config

import (
	"fmt"
	"sync"
)

type NodeConfig struct {
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	Address         string `yaml:"address,omitempty" json:"address,omitempty"`
	InternalAddress string `yaml:"internalAddress,omitempty" json:"internalAddress,omitempty"`
	Port            int    `yaml:"port,omitempty" json:"port,omitempty"`
	User            string `yaml:"user,omitempty" json:"user,omitempty"`
	Password        string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey      string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath  string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Arch            string `yaml:"arch,omitempty" json:"arch,omitempty"`

	Labels   map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	ID       string            `yaml:"id,omitempty" json:"id,omitempty"`
	Index    int               `json:"-"`
	IsEtcd   bool              `json:"-"`
	IsMaster bool              `json:"-"`
	IsWorker bool              `json:"-"`
}

const (
	NodeHostName       = "NodeHostName"
	NodeIpAddress      = "NodeIpAddress"
	NodePort           = "NodePort"
	NodeUser           = "NodeUser"
	NodePassword       = "NodePassword"
	NodePrivateKey     = "NodePrivateKey"
	NodePrivateKeyPath = "NodePrivateKeyPath"
	NodeArch           = "NodeArch"
)

var (
	nodeConfigMap       = sync.Map{}
	nodeConfigSingleton sync.Once
)

//GetNodeMap kk will make a cache to store the node config map
func GetNodeMap(nodeName string) (map[string]interface{}, error) {
	nodeConfigSingleton.Do(func() {
		for _, node := range GetConfig().K8sNodes {
			// make struct to map, and the sync.map store the map
			m := map[string]interface{}{
				NodeHostName:       node.Name,
				NodeIpAddress:      node.InternalAddress,
				NodePort:           node.Port,
				NodeUser:           node.User,
				NodePassword:       node.Password,
				NodePrivateKey:     node.PrivateKey,
				NodePrivateKeyPath: node.PrivateKeyPath,
				NodeArch:           node.Arch,
			}
			nodeConfigMap.Store(node.Name, m)
		}
	})
	m, ok := nodeConfigMap.Load(nodeName)
	if !ok {
		return nil, fmt.Errorf("node [%s] get config failed", nodeName)
	}
	return m.(map[string]interface{}), nil
}
