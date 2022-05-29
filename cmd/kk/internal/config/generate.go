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

package config

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
	"github.com/kubesphere/kubekey/cmd/kk/internal/config/templates"
	"github.com/kubesphere/kubekey/cmd/kk/internal/version/kubesphere"
	"github.com/kubesphere/kubekey/util/workflow/util"

	"github.com/pkg/errors"
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
		ksInstaller, ok := kubesphere.StabledVersionSupport(version)
		if ok {
			opt.KubeSphereConfigMap = ksInstaller.CCToString()
		} else if latest, ok := kubesphere.LatestRelease(version); ok {
			latest.Version = version
			opt.KubeSphereConfigMap = latest.CCToString()
		} else if dev, ok := kubesphere.DevRelease(version); ok {
			dev.Version = version
			opt.KubeSphereConfigMap = dev.CCToString()
		} else {
			return errors.New(fmt.Sprintf("Unsupported KubeSphere version: %s", version))
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
		if err := exec.Command("/bin/sh", "-c", cmd).Run(); err != nil {
			return err
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
