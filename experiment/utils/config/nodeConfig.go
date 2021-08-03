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

	IsMaster = "IsMaster"
	IsETCD   = "IsETCD"
	IsWorker = "IsWorker"
)

type NodeSyncMap struct {
	sync.Map
}

var (
	nodeConfigMap       = make(map[string]*NodeSyncMap)
	nodeConfigSingleton sync.Once
)

//GetNodeMap kk will make a cache to store the node config map
func GetNodeMap(nodeName string) (*NodeSyncMap, error) {
	nodeConfigSingleton.Do(func() {
		for _, node := range GetManager().K8sNodes {
			m := new(NodeSyncMap)
			m.Store(NodeHostName, node.Name)
			m.Store(NodeIpAddress, node.InternalAddress)
			m.Store(NodePort, node.Port)
			m.Store(NodeUser, node.User)
			m.Store(NodePassword, node.Password)
			m.Store(NodePrivateKey, node.PrivateKey)
			m.Store(NodePrivateKeyPath, node.PrivateKeyPath)
			m.Store(NodeArch, node.Arch)
			m.Store(IsMaster, false)
			m.Store(IsETCD, false)

			nodeConfigMap[node.Name] = m
		}
	})
	m, ok := nodeConfigMap[nodeName]
	if !ok {
		return nil, fmt.Errorf("node [%s] get config failed", nodeName)
	}
	return m, nil
}

func (n *NodeSyncMap) TagMaster() {
	n.Store(IsMaster, true)
}

func (n *NodeSyncMap) TagWorker() {
	n.Store(IsWorker, true)
}

func (n *NodeSyncMap) TagETCD() {
	n.Store(IsETCD, true)
}

func (n *NodeSyncMap) IsMaster() bool {
	if v, ok := n.Load(IsMaster); ok {
		return v.(bool)
	} else {
		return false
	}
}

func (n *NodeSyncMap) IsWorker() bool {
	if v, ok := n.Load(IsWorker); ok {
		return v.(bool)
	} else {
		return false
	}
}

func (n *NodeSyncMap) IsETCD() bool {
	if v, ok := n.Load(IsETCD); ok {
		return v.(bool)
	} else {
		return false
	}
}
