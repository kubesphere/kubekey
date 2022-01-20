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

package common

import (
	"bufio"
	"encoding/json"
	"fmt"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/version/kubesphere"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Loader interface {
	Load() (*kubekeyapiv1alpha2.Cluster, error)
}

type Options map[string]interface{}

func NewLoader(flag string, arg Argument) Loader {
	switch flag {
	case File:
		return NewFileLoader(arg)
	case Operator:
		return &ConfigMapLoader{}
	case AllInOne:
		return NewDefaultLoader(arg)
	default:
		return NewDefaultLoader(arg)
	}
}

type DefaultLoader struct {
	arg               Argument
	KubernetesVersion string
	KubeSphereVersion string
	KubeSphereEnable  bool
}

func NewDefaultLoader(arg Argument) *DefaultLoader {
	return &DefaultLoader{
		arg:               arg,
		KubernetesVersion: arg.KubernetesVersion,
		KubeSphereVersion: arg.KsVersion,
		KubeSphereEnable:  arg.KsEnable,
	}
}

func (d *DefaultLoader) Load() (*kubekeyapiv1alpha2.Cluster, error) {
	u, _ := user.Current()
	if u.Username != "root" {
		return nil, errors.New(fmt.Sprintf("Current user is %s. Please use root!", u.Username))
	}

	allInOne := kubekeyapiv1alpha2.Cluster{}
	if output, err := exec.Command("/bin/sh", "-c", "if [ ! -f \"$HOME/.ssh/id_rsa\" ]; then ssh-keygen -t rsa -P \"\" -f $HOME/.ssh/id_rsa && ls $HOME/.ssh;fi;").CombinedOutput(); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to generate public key: %v\n%s", err, string(output)))
	}
	if output, err := exec.Command("/bin/sh", "-c", "echo \"\n$(cat $HOME/.ssh/id_rsa.pub)\" >> $HOME/.ssh/authorized_keys && awk ' !x[$0]++{print > \"'$HOME'/.ssh/authorized_keys\"}' $HOME/.ssh/authorized_keys").CombinedOutput(); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to copy public key to authorized_keys: %v\n%s", err, string(output)))
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get hostname: %v\n", err))
	}

	allInOne.Spec.Hosts = append(allInOne.Spec.Hosts, kubekeyapiv1alpha2.HostCfg{
		Name:            hostname,
		Address:         util.LocalIP(),
		InternalAddress: util.LocalIP(),
		Port:            kubekeyapiv1alpha2.DefaultSSHPort,
		User:            u.Name,
		Password:        "",
		PrivateKeyPath:  fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir),
		Arch:            runtime.GOARCH,
	})

	allInOne.Spec.RoleGroups = map[string][]string{
		Master:   {hostname},
		ETCD:     {hostname},
		Worker:   {hostname},
		Registry: {hostname},
	}
	if d.KubernetesVersion != "" {
		s := strings.Split(d.KubernetesVersion, "-")
		if len(s) > 1 {
			allInOne.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: s[0],
				Type:    s[1],
			}
		} else {
			allInOne.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: d.KubernetesVersion,
			}
		}
	} else {
		allInOne.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
			Version: kubekeyapiv1alpha2.DefaultKubeVersion,
		}
	}

	if d.KubeSphereEnable {
		if err := defaultKSConfig(&allInOne.Spec.KubeSphere, d.KubeSphereVersion); err != nil {
			return nil, err
		}
	}

	if d.arg.ContainerManager != "" && d.arg.ContainerManager != Docker {
		allInOne.Spec.Kubernetes.ContainerManager = d.arg.ContainerManager
	}

	// must be a lower case
	allInOne.Name = "kubekey" + time.Now().Format("2006-01-02")
	return &allInOne, nil
}

type FileLoader struct {
	arg               Argument
	FilePath          string
	KubernetesVersion string
	KubeSphereVersion string
	KubeSphereEnable  bool
}

func NewFileLoader(arg Argument) *FileLoader {
	return &FileLoader{
		arg:               arg,
		FilePath:          arg.FilePath,
		KubernetesVersion: arg.KubernetesVersion,
		KubeSphereVersion: arg.KsVersion,
		KubeSphereEnable:  arg.KsEnable,
	}
}

func (f FileLoader) Load() (*kubekeyapiv1alpha2.Cluster, error) {
	var objName string

	clusterCfg := kubekeyapiv1alpha2.Cluster{}
	fp, err := filepath.Abs(f.FilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to look up current directory")
	}
	// fixme: It will lead to nil pointer err
	//if len(f.KubernetesVersion) != 0 {
	//	_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("sed -i \"/version/s/\\:.*/\\: %s/g\" %s", f.KubernetesVersion, fp)).Run()
	//}
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
			contentToJson, err := k8syaml.ToJSON(content)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to convert configuration to json")
			}
			if err := json.Unmarshal(contentToJson, &clusterCfg); err != nil {
				return nil, errors.Wrap(err, "Failed to unmarshal configuration")
			}
			metadata := result["metadata"].(map[interface{}]interface{})
			objName = metadata["name"].(string)
		}

		if result["kind"] == "ConfigMap" || result["kind"] == "ClusterConfiguration" {
			metadata := result["metadata"].(map[interface{}]interface{})
			labels := metadata["labels"].(map[interface{}]interface{})
			clusterCfg.Spec.KubeSphere.Enabled = true

			v, ok := labels["version"]
			if !ok {
				return nil, errors.New("Unknown version")
			}

			version := v.(string)
			_, stable := kubesphere.StabledVersionSupport(version)
			_, latest := kubesphere.LatestRelease(version)
			_, dev := kubesphere.DevRelease(version)
			if stable || latest || dev {
				clusterCfg.Spec.KubeSphere.Configurations = "---\n" + string(content)
				clusterCfg.Spec.KubeSphere.Version = version
			} else {
				return nil, errors.New(fmt.Sprintf("Unsupported KubeSphere version: %s", version))
			}
		}
	}

	if f.KubeSphereEnable {
		if err := defaultKSConfig(&clusterCfg.Spec.KubeSphere, f.KubeSphereVersion); err != nil {
			return nil, err
		}
	}

	if f.KubernetesVersion != "" {
		s := strings.Split(f.KubernetesVersion, "-")
		if len(s) > 1 {
			clusterCfg.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: s[0],
				Type:    s[1],
			}
		} else {
			clusterCfg.Spec.Kubernetes = kubekeyapiv1alpha2.Kubernetes{
				Version: f.KubernetesVersion,
			}
		}
	}

	if f.arg.ContainerManager != "" && f.arg.ContainerManager != Docker {
		clusterCfg.Spec.Kubernetes.ContainerManager = f.arg.ContainerManager
	}

	clusterCfg.Name = objName
	return &clusterCfg, nil
}

type ConfigMapLoader struct {
}

func (c ConfigMapLoader) Load() (*kubekeyapiv1alpha2.Cluster, error) {
	return nil, nil
}

func defaultKSConfig(ks *kubekeyapiv1alpha2.KubeSphere, version string) error {
	ks.Enabled = true
	version = strings.TrimSpace(version)
	ksInstaller, ok := kubesphere.StabledVersionSupport(version)
	if ok {
		ks.Version = ksInstaller.Version
		ks.Configurations = ksInstaller.CCToString()
	} else if latest, ok := kubesphere.LatestRelease(version); ok {
		ks.Version = version
		ks.Configurations = latest.CCToString()
	} else if dev, ok := kubesphere.DevRelease(version); ok {
		ks.Version = version
		ks.Configurations = dev.CCToString()
	} else {
		return errors.New(fmt.Sprintf("Unsupported KubeSphere version: %s", version))
	}
	return nil
}
