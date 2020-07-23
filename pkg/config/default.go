/*
Copyright 2020 The KubeSphere Authors.

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

package config

import (
	"bufio"
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func ParseClusterCfg(clusterCfgPath, k8sVersion, ksVersion string, ksEnabled bool, logger *log.Logger) (*kubekeyapi.Cluster, error) {
	var clusterCfg *kubekeyapi.Cluster
	if len(clusterCfgPath) == 0 {
		user, _ := user.Current()
		if user.Name != "root" {
			return nil, errors.New(fmt.Sprintf("Current user is %s. Please use root!", user.Name))
		}
		clusterCfg = AllinoneCfg(user, k8sVersion, ksVersion, ksEnabled, logger)
	} else {
		cfg, err := ParseCfg(clusterCfgPath)
		if err != nil {
			return nil, err
		}
		clusterCfg = cfg
	}

	return clusterCfg, nil
}

func ParseCfg(clusterCfgPath string) (*kubekeyapi.Cluster, error) {
	clusterCfg := kubekeyapi.Cluster{}
	fp, err := filepath.Abs(clusterCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to look up current directory")
	}
	file, err := os.Open(fp)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open the given cluster configuration file")
	}
	defer file.Close()
	b1 := bufio.NewReader(file)

	for {
		result := make(map[string]interface{})
		content, err := k8syaml.NewYAMLReader(b1).Read()
		if len(content) == 0 {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read the given cluster configuration file")
		}
		err = yaml.Unmarshal(content, &result)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to unmarshal the given cluster configuration file")
		}
		if result["kind"] == "Cluster" {
			if err := yaml.Unmarshal(content, &clusterCfg); err != nil {
				return nil, errors.Wrap(err, "Unable to convert file to yaml")
			}
		}

		if result["kind"] == "ConfigMap" || result["kind"] == "ClusterConfiguration" {
			metadata := result["metadata"].(map[interface{}]interface{})
			labels := metadata["labels"].(map[interface{}]interface{})
			repo := clusterCfg.Spec.Registry.PrivateRegistry
			clusterCfg.Spec.KubeSphere.Enabled = true
			var configMapBase64, kubesphereYaml string
			_, ok := labels["version"]
			if ok {
				switch labels["version"] {
				case "v3.0.0":
					configMapBase64 = base64.StdEncoding.EncodeToString(content)
					kubesphereYaml, err = kubesphere.GenerateKubeSphereYaml(repo, "latest")
					if err != nil {
						return nil, err
					}
					clusterCfg.Spec.KubeSphere.Version = "v3.0.0"
				case "v2.1.1":
					configMapBase64 = base64.StdEncoding.EncodeToString(content)
					kubesphereYaml, err = kubesphere.GenerateKubeSphereYaml(repo, "v2.1.1")
					if err != nil {
						return nil, err
					}
					clusterCfg.Spec.KubeSphere.Version = "v2.1.1"
				default:
					return nil, errors.Wrap(err, fmt.Sprintf("Unsupported versions: %s", labels["version"]))
				}

				currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
				if err != nil {
					return nil, errors.Wrap(err, "Faild to get current dir")
				}
				_, err1 := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo %s | base64 -d > %s/kubekey/ks-installer-configmap.yaml", configMapBase64, currentDir)).CombinedOutput()
				if err1 != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("Failed to write to file: %s/kubekey/ks-installer-configmap.yaml", currentDir))
				}
				kubesphereYamlBase64 := base64.StdEncoding.EncodeToString([]byte(kubesphereYaml))
				_, err2 := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo %s | base64 -d > %s/kubekey/ks-installer-deployment.yaml", kubesphereYamlBase64, currentDir)).CombinedOutput()
				if err2 != nil {
					return nil, errors.Wrap(err2, fmt.Sprintf("Failed to generate KubeSphere manifests: %s/kubekey/ks-installer-deployment.yaml", currentDir))
				}
			}
		}
	}
	return &clusterCfg, nil
}

func AllinoneCfg(user *user.User, k8sVersion, ksVersion string, ksEnabled bool, logger *log.Logger) *kubekeyapi.Cluster {
	allinoneCfg := kubekeyapi.Cluster{}
	if output, err := exec.Command("/bin/sh", "-c", "if [ ! -f \"$HOME/.ssh/id_rsa\" ]; then ssh-keygen -t rsa -P \"\" -f $HOME/.ssh/id_rsa && ls $HOME/.ssh;fi;").CombinedOutput(); err != nil {
		log.Fatalf("Failed to generate public key: %v\n%s", err, string(output))
	}
	if output, err := exec.Command("/bin/sh", "-c", "echo \"\n$(cat $HOME/.ssh/id_rsa.pub)\" >> $HOME/.ssh/authorized_keys && awk ' !x[$0]++{print > \"'$HOME'/.ssh/authorized_keys\"}' $HOME/.ssh/authorized_keys").CombinedOutput(); err != nil {
		log.Fatalf("Failed to copy public key to authorized_keys: %v\n%s", err, string(output))
	}

	allinoneCfg.Spec.Hosts = append(allinoneCfg.Spec.Hosts, kubekeyapi.HostCfg{
		Name:            "ks-allinone",
		Address:         util.LocalIP(),
		InternalAddress: util.LocalIP(),
		Port:            "",
		User:            user.Name,
		Password:        "",
		PrivateKeyPath:  fmt.Sprintf("%s/.ssh/id_rsa", user.HomeDir),
		Arch:            runtime.GOARCH,
	})

	allinoneCfg.Spec.RoleGroups = kubekeyapi.RoleGroups{
		Etcd:   []string{"ks-allinone"},
		Master: []string{"ks-allinone"},
		Worker: []string{"ks-allinone"},
	}
	if k8sVersion != "" {
		allinoneCfg.Spec.Kubernetes = kubekeyapi.Kubernetes{
			Version: k8sVersion,
		}
	}
	allinoneCfg.Spec.RoleGroups = kubekeyapi.RoleGroups{
		Etcd:   []string{"ks-allinone"},
		Master: []string{"ks-allinone"},
		Worker: []string{"ks-allinone"},
	}

	if ksEnabled {
		allinoneCfg.Spec.Storage = kubekeyapi.Storage{
			DefaultStorageClass: "localVolume",
			LocalVolume:         kubekeyapi.LocalVolume{StorageClassName: "local"},
		}
		allinoneCfg.Spec.KubeSphere.Enabled = true
		var configMapBase64, kubesphereYaml string
		switch strings.TrimSpace(ksVersion) {
		case "":
			configMapBase64 = base64.StdEncoding.EncodeToString([]byte(kubesphere.V3_0_0))
			kubesphereYaml, _ = kubesphere.GenerateKubeSphereYaml("", "latest")
			allinoneCfg.Spec.KubeSphere.Version = "v3.0.0"
		case "v3.0.0":
			configMapBase64 = base64.StdEncoding.EncodeToString([]byte(kubesphere.V3_0_0))
			kubesphereYaml, _ = kubesphere.GenerateKubeSphereYaml("", "latest")
			allinoneCfg.Spec.KubeSphere.Version = "v3.0.0"
		case "v2.1.1":
			configMapBase64 = base64.StdEncoding.EncodeToString([]byte(kubesphere.V2_1_1))
			kubesphereYaml, _ = kubesphere.GenerateKubeSphereYaml("", "v2.1.1")
			allinoneCfg.Spec.KubeSphere.Version = "v2.1.1"
		default:
			logger.Fatalf("Unsupported version: %s", strings.TrimSpace(ksVersion))
		}

		currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(errors.Wrap(err, "Faild to get current dir"))
		}
		_, err1 := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo %s | base64 -d > %s/kubekey/ks-installer-configmap.yaml", configMapBase64, currentDir)).CombinedOutput()
		if err1 != nil {
			logger.Fatalf("Unsupported versions: %s", ksVersion)
		}
		kubesphereYamlBase64 := base64.StdEncoding.EncodeToString([]byte(kubesphereYaml))
		_, err2 := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo %s | base64 -d > %s/kubekey/ks-installer-deployment.yaml", kubesphereYamlBase64, currentDir)).CombinedOutput()
		if err2 != nil {
			logger.Fatal(errors.Wrap(err2, fmt.Sprintf("Failed to generate KubeSphere manifests: %s/kubekey/ks-installer-deployment.yaml", currentDir)))
		}
	}
	return &allinoneCfg
}
