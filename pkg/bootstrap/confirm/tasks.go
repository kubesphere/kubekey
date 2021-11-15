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

package confirm

import (
	"bufio"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/table"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"regexp"
	"strings"
)

// PreCheckResults defines the items to be checked.
type PreCheckResults struct {
	Name      string `table:"name"`
	Sudo      string `table:"sudo"`
	Curl      string `table:"curl"`
	Openssl   string `table:"openssl"`
	Ebtables  string `table:"ebtables"`
	Socat     string `table:"socat"`
	Ipset     string `table:"ipset"`
	Conntrack string `table:"conntrack"`
	Docker    string `table:"docker"`
	Nfs       string `table:"nfs client"`
	Ceph      string `table:"ceph client"`
	Glusterfs string `table:"glusterfs client"`
	Time      string `table:"time"`
}

type InstallationConfirm struct {
	common.KubeAction
}

func (i *InstallationConfirm) Execute(runtime connector.Runtime) error {
	var (
		results  []PreCheckResults
		stopFlag bool
	)

	pre := make([]map[string]string, 0, len(runtime.GetAllHosts()))
	for _, host := range runtime.GetAllHosts() {
		if v, ok := host.GetCache().Get(common.NodePreCheck); ok {
			pre = append(pre, v.(map[string]string))
		} else {
			return errors.New("get node check result failed by host cache")
		}
	}

	for node := range pre {
		var result PreCheckResults
		_ = mapstructure.Decode(pre[node], &result)
		results = append(results, result)
	}
	table.OutputA(results)
	reader := bufio.NewReader(os.Stdin)

	for _, host := range results {
		if host.Conntrack == "" {
			fmt.Printf("%s: conntrack is required. \n", host.Name)
			logger.Log.Errorf("%s: conntrack is required. \n", host.Name)
			stopFlag = true
		}
	}

	if stopFlag {
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("This is a simple check of your environment.")
	fmt.Println("Before installation, you should ensure that your machines meet all requirements specified at")
	fmt.Println("https://github.com/kubesphere/kubekey#requirements-and-recommendations")
	fmt.Println("")

	confirmOK := false
	for !confirmOK {
		fmt.Printf("Continue this installation? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			logger.Log.Fatal(err)
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "yes":
			confirmOK = true
		case "no":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}

type DeleteConfirm struct {
	common.KubeAction
	Content string
}

func (d *DeleteConfirm) Execute(runtime connector.Runtime) error {
	reader := bufio.NewReader(os.Stdin)

	var res string
	for {
		fmt.Printf("Are you sure to delete this %s? [yes/no]: ", d.Content)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)

		if input != "" && (input == "yes" || input == "no") {
			res = input
			break
		}
	}

	if res == "no" {
		os.Exit(0)
	}
	return nil
}

type UpgradeConfirm struct {
	common.KubeAction
}

func (u *UpgradeConfirm) Execute(runtime connector.Runtime) error {
	pre := make([]map[string]string, len(runtime.GetAllHosts()), len(runtime.GetAllHosts()))
	for i, host := range runtime.GetAllHosts() {
		if v, ok := host.GetCache().Get(common.NodePreCheck); ok {
			pre[i] = v.(map[string]string)
		} else {
			return errors.New("get node check result failed by host cache")
		}
	}

	results := make([]PreCheckResults, len(pre), len(pre))
	for i := range pre {
		var result PreCheckResults
		_ = mapstructure.Decode(pre[i], &result)
		results[i] = result
	}
	table.OutputA(results)
	fmt.Println()

	warningFlag := false
	cmp, err := versionutil.MustParseSemantic(u.KubeConf.Cluster.Kubernetes.Version).Compare("v1.19.0")
	if err != nil {
		logger.Log.Fatalf("Failed to compare kubernetes version: %v", err)
	}
	if cmp == 0 || cmp == 1 {
		for _, result := range results {
			dockerVersion, err := RefineDockerVersion(result.Docker)
			if err != nil {
				logger.Log.Fatalf("Failed to get docker version: %v", err)
			}
			cmp, err := versionutil.MustParseSemantic(dockerVersion).Compare("20.10.0")
			if err != nil {
				logger.Log.Fatalf("Failed to compare docker version: %v", err)
			}
			warningFlag = warningFlag || (cmp == -1)
		}

		if warningFlag {
			fmt.Println(`
Warning:

  An old Docker version may cause the failure of upgrade. It is recommended that you upgrade Docker to 20.10+ beforehand.

  Issue: https://github.com/kubernetes/kubernetes/issues/101056`)
			fmt.Print("\n")
		}
	}

	nodeStats, ok := u.PipelineCache.GetMustString(common.ClusterNodeStatus)
	if !ok {
		return errors.New("get cluster nodes status failed by pipeline cache")
	}
	fmt.Println("Cluster nodes status:")
	fmt.Println(nodeStats + "\n")

	fmt.Println("Upgrade Confirmation:")
	currentK8sVersion, ok := u.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current Kubernetes version failed by pipeline cache")
	}
	fmt.Printf("kubernetes version: %s to %s\n", currentK8sVersion, u.KubeConf.Cluster.Kubernetes.Version)

	if u.KubeConf.Cluster.KubeSphere.Enabled {
		currentKsVersion, ok := u.PipelineCache.GetMustString(common.KubeSphereVersion)
		if !ok {
			return errors.New("get current KubeSphere version failed by pipeline cache")
		}
		fmt.Printf("kubesphere version: %s to %s\n", currentKsVersion, u.KubeConf.Cluster.KubeSphere.Version)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	confirmOK := false
	for !confirmOK {
		fmt.Printf("Continue upgrading cluster? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(input)

		switch input {
		case "yes":
			confirmOK = true
		case "no":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}

func RefineDockerVersion(version string) (string, error) {
	var newVersionComponents []string
	versionMatchRE := regexp.MustCompile(`^\s*v?([0-9]+(?:\.[0-9]+)*)(.*)*$`)
	parts := versionMatchRE.FindStringSubmatch(version)
	if parts == nil {
		return "", fmt.Errorf("could not parse %q as version", version)
	}
	numbers, _ := parts[1], parts[2]
	components := strings.Split(numbers, ".")

	for index, c := range components {
		newVersion := strings.TrimPrefix(c, "0")
		if index == len(components)-1 && newVersion == "" {
			newVersion = "0"
		}
		newVersionComponents = append(newVersionComponents, newVersion)
	}
	return strings.Join(newVersionComponents, "."), nil
}
