package config

import (
	"bufio"
	"encoding/base64"
	"fmt"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/config/templates"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/version/kubesphere"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateKubeKeyConfig is used to generate cluster configuration file
func GenerateKubeKeyConfig(arg common.Argument, name string) error {
	if arg.FromCluster {
		err := GenerateConfigFromCluster(arg.FilePath, arg.KubeConfig, name)
		if err != nil {
			return err
		}
		return nil
	}

	opt := new(templates.Options)

	if name != "" {
		output := strings.Split(name, ".")
		opt.Name = output[0]
	} else {
		opt.Name = "sample"
	}
	if len(arg.KubernetesVersion) == 0 {
		opt.KubeVersion = kubekeyapiv1alpha2.DefaultKubeVersion
	} else {
		opt.KubeVersion = arg.KubernetesVersion
	}
	opt.KubeSphereEnabled = arg.KsEnable

	if arg.KsEnable {
		version := strings.TrimSpace(arg.KsVersion)
		ksInstaller, ok := kubesphere.VersionMap[version]
		if ok {
			opt.KubeSphereConfigMap = ksInstaller.CCToString()
		} else {
			if kubesphere.PreRelease(version) {
				opt.KubeSphereConfigMap = kubesphere.Latest().CCToString()
			} else {
				return errors.New(fmt.Sprintf("Unsupported KubeSphere version: %s", version))
			}
		}
	}

	ClusterObjStr, err := templates.GenerateCluster(opt)
	if err != nil {
		return errors.Wrap(err, "Failed to generate cluster config")
	}
	ClusterObjStrBase64 := base64.StdEncoding.EncodeToString([]byte(ClusterObjStr))

	if arg.FilePath != "" {
		CheckConfigFileStatus(arg.FilePath)
		cmdStr := fmt.Sprintf("echo %s | base64 -d > %s", ClusterObjStrBase64, arg.FilePath)
		output, err := exec.Command("/bin/sh", "-c", cmdStr).CombinedOutput()
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to write config to %s: %s", arg.FilePath, strings.TrimSpace(string(output))))
		}
	} else {
		currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return errors.Wrap(err, "Failed to get current dir")
		}
		CheckConfigFileStatus(fmt.Sprintf("%s/config-%s.yaml", currentDir, opt.Name))
		cmd := fmt.Sprintf("echo %s | base64 -d > %s/config-%s.yaml", ClusterObjStrBase64, currentDir, opt.Name)
		err1 := exec.Command("/bin/sh", "-c", cmd).Run()
		if err1 != nil {
			return err1
		}
	}

	return nil
}

// CheckConfigFileStatus is used to check the status of cluster configuration file.
func CheckConfigFileStatus(path string) {
	if util.IsExist(path) {
		reader := bufio.NewReader(os.Stdin)
	Loop:
		for {
			fmt.Printf("%s already exists. Are you sure you want to overwrite this config file? [yes/no]: ", path)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input != "" {
				switch input {
				case "yes":
					break Loop
				case "no":
					os.Exit(0)
				}
			}
		}
	}
}
