/*
 Copyright 2021 The KubeSphere Authors.

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

package utils

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	return newForClient(configCluster)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

func KubeConfigFormByte(data []byte) (*rest.Config, error) {
	ClientConfig, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, err
	}
	restConfig, err := ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return restConfig, nil
}

func NewClientForCluster(kubeConfig []byte) (*kubernetes.Clientset, error) {

	forClientConfig, err := KubeConfigFormByte(kubeConfig)
	if err != nil {
		return nil, err
	}

	return newForClient(forClientConfig)
}

func newForClient(config *rest.Config) (*kubernetes.Clientset, error) {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
