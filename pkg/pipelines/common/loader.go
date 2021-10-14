package common

import (
	"bufio"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/version/kubesphere"
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
	Load() (*kubekeyapiv1alpha1.Cluster, error)
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
	KubernetesVersion string
	KubeSphereVersion string
	KubeSphereEnable  bool
}

func NewDefaultLoader(arg Argument) *DefaultLoader {
	return &DefaultLoader{
		KubernetesVersion: arg.KubernetesVersion,
		KubeSphereVersion: arg.KsVersion,
		KubeSphereEnable:  arg.KsEnable,
	}
}

func (d *DefaultLoader) Load() (*kubekeyapiv1alpha1.Cluster, error) {
	u, _ := user.Current()
	if u.Username != "root" {
		return nil, errors.New(fmt.Sprintf("Current user is %s. Please use root!", u.Username))
	}

	allInOne := kubekeyapiv1alpha1.Cluster{}
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

	allInOne.Spec.Hosts = append(allInOne.Spec.Hosts, kubekeyapiv1alpha1.HostCfg{
		Name:            hostname,
		Address:         util.LocalIP(),
		InternalAddress: util.LocalIP(),
		Port:            kubekeyapiv1alpha1.DefaultSSHPort,
		User:            u.Name,
		Password:        "",
		PrivateKeyPath:  fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir),
		Arch:            runtime.GOARCH,
	})

	allInOne.Spec.RoleGroups = kubekeyapiv1alpha1.RoleGroups{
		Etcd:   []string{hostname},
		Master: []string{hostname},
		Worker: []string{hostname},
	}
	if d.KubernetesVersion != "" {
		s := strings.Split(d.KubernetesVersion, "-")
		if len(s) > 1 {
			allInOne.Spec.Kubernetes = kubekeyapiv1alpha1.Kubernetes{
				Version: s[0],
				Type:    s[1],
			}
		} else {
			allInOne.Spec.Kubernetes = kubekeyapiv1alpha1.Kubernetes{
				Version: d.KubernetesVersion,
			}
		}
	} else {
		allInOne.Spec.Kubernetes = kubekeyapiv1alpha1.Kubernetes{
			Version: kubekeyapiv1alpha1.DefaultKubeVersion,
		}
	}

	if d.KubeSphereEnable {
		if err := defaultKSConfig(&allInOne.Spec.KubeSphere, d.KubeSphereVersion); err != nil {
			return nil, err
		}
	}

	allInOne.Name = AllInOne + time.Now().Format("2006-01-02")

	return &allInOne, nil
}

type FileLoader struct {
	FilePath          string
	KubernetesVersion string
	KubeSphereVersion string
	KubeSphereEnable  bool
}

func NewFileLoader(arg Argument) *FileLoader {
	return &FileLoader{
		FilePath:          arg.FilePath,
		KubernetesVersion: arg.KubernetesVersion,
		KubeSphereVersion: arg.KsVersion,
		KubeSphereEnable:  arg.KsEnable,
	}
}

func (f FileLoader) Load() (*kubekeyapiv1alpha1.Cluster, error) {
	var objName string

	clusterCfg := kubekeyapiv1alpha1.Cluster{}
	fp, err := filepath.Abs(f.FilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to look up current directory")
	}
	// It will lead to nil pointer err
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
			if err := yaml.Unmarshal(content, &clusterCfg); err != nil {
				return nil, errors.Wrap(err, "Unable to convert file to yaml")
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
			if kubesphere.VersionSupport(version) || kubesphere.PreRelease(version) {
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
	clusterCfg.Name = objName
	return &clusterCfg, nil
}

type ConfigMapLoader struct {
}

func (c ConfigMapLoader) Load() (*kubekeyapiv1alpha1.Cluster, error) {
	return nil, nil
}

func defaultKSConfig(ks *kubekeyapiv1alpha1.KubeSphere, version string) error {
	ks.Enabled = true
	version = strings.TrimSpace(version)
	ksInstaller, ok := kubesphere.VersionMap[version]
	if ok {
		ks.Version = ksInstaller.Version
		ks.Configurations = ksInstaller.CCToString()
	} else {
		if kubesphere.PreRelease(version) {
			ks.Version = version
			ks.Configurations = kubesphere.Latest().CCToString()
		} else {
			return errors.New(fmt.Sprintf("Unsupported KubeSphere version: %s", version))
		}
	}
	return nil
}
