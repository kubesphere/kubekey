package utils

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func NewClient(config string) (*kubernetes.Clientset, error) {
	var kubeconfig string
	if config != "" {
		config, err := filepath.Abs(config)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to look up current directory")
		}
		kubeconfig = config
	} else {
		kubeconfig = filepath.Join(homeDir(), ".kube", "config")
	}
	// use the current context in kubeconfig
	configCluster, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(configCluster)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}