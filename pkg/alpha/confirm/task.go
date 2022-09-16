package confirm

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kubesphere/kubekey/pkg/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/table"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type UpgradeK8sConfirm struct {
	common.KubeAction
}

func (u *UpgradeK8sConfirm) Execute(runtime connector.Runtime) error {
	pre := make([]map[string]string, len(runtime.GetAllHosts()), len(runtime.GetAllHosts()))
	for i, host := range runtime.GetAllHosts() {
		if v, ok := host.GetCache().Get(common.NodePreCheck); ok {
			pre[i] = v.(map[string]string)
		} else {
			return errors.New("get node check result failed by host cache")
		}
	}

	results := make([]confirm.PreCheckResults, len(pre), len(pre))
	for i := range pre {
		var result confirm.PreCheckResults
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
			if len(result.Docker) != 0 {
				dockerVersion, err := confirm.RefineDockerVersion(result.Docker)
				if err != nil {
					logger.Log.Fatalf("Failed to get docker version: %v", err)
				}
				cmp, err := versionutil.MustParseSemantic(dockerVersion).Compare("20.10.0")
				if err != nil {
					logger.Log.Fatalf("Failed to compare docker version: %v", err)
				}
				warningFlag = warningFlag || (cmp == -1)
			}
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

	fmt.Println()

	if k8sVersion, err := versionutil.ParseGeneric(u.KubeConf.Cluster.Kubernetes.Version); err == nil {
		if cri, ok := u.PipelineCache.GetMustString(common.ClusterNodeCRIRuntimes); ok {
			k8sV124 := versionutil.MustParseSemantic("v1.24.0")
			if k8sVersion.AtLeast(k8sV124) && versionutil.MustParseSemantic(currentK8sVersion).LessThan(k8sV124) && strings.Contains(cri, "docker") {
				fmt.Println("[Notice]")
				fmt.Println("Pre-upgrade check failed. The container runtime of the current cluster is Docker.")
				fmt.Println("Kubernetes v1.24 and later no longer support dockershim and Docker.")
				fmt.Println("Make sure you have completed the migration from Docker to other container runtimes that are compatible with the Kubernetes CRI.")
				fmt.Println("For more information, see:")
				fmt.Println("https://kubernetes.io/docs/setup/production-environment/container-runtimes/#container-runtimes")
				fmt.Println("https://kubernetes.io/blog/2022/02/17/dockershim-faq/")
				fmt.Println("")
			}
		}
	}

	reader := bufio.NewReader(os.Stdin)
	confirmOK := false
	for !confirmOK {
		fmt.Printf("Continue upgrading kubernetes? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "yes", "y":
			confirmOK = true
		case "no", "n":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}

type UpgradeKsConfirm struct {
	common.KubeAction
}

func (u *UpgradeKsConfirm) Execute(runtime connector.Runtime) error {
	pre := make([]map[string]string, len(runtime.GetAllHosts()), len(runtime.GetAllHosts()))
	for i, host := range runtime.GetAllHosts() {
		if v, ok := host.GetCache().Get(common.NodePreCheck); ok {
			pre[i] = v.(map[string]string)
		} else {
			return errors.New("get node check result failed by host cache")
		}
	}

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
		fmt.Printf("Continue upgrading KubeSphere? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "yes", "y":
			confirmOK = true
		case "no", "n":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}

type CreateK8sConfirm struct {
	common.KubeAction
}

func (c *CreateK8sConfirm) Execute(runtime connector.Runtime) error {
	var (
		results  []confirm.PreCheckResults
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
		var result confirm.PreCheckResults
		_ = mapstructure.Decode(pre[node], &result)
		results = append(results, result)
	}
	table.OutputA(results)
	reader := bufio.NewReader(os.Stdin)

	if c.KubeConf.Arg.Artifact == "" {
		for _, host := range results {
			if host.Sudo == "" {
				logger.Log.Errorf("%s: sudo is required.", host.Name)
				stopFlag = true
			}

			if host.Conntrack == "" {
				logger.Log.Errorf("%s: conntrack is required.", host.Name)
				stopFlag = true
			}

			if host.Socat == "" {
				logger.Log.Errorf("%s: socat is required.", host.Name)
				stopFlag = true
			}
		}
	}

	fmt.Println("")
	fmt.Println("This is a simple check of your environment.")
	fmt.Println("Before installation, ensure that your machines meet all requirements specified at")
	fmt.Println("https://github.com/kubesphere/kubekey#requirements-and-recommendations")
	fmt.Println("")

	if k8sVersion, err := versionutil.ParseGeneric(c.KubeConf.Cluster.Kubernetes.Version); err == nil {
		if k8sVersion.AtLeast(versionutil.MustParseSemantic("v1.24.0")) && c.KubeConf.Cluster.Kubernetes.ContainerManager == common.Docker {
			fmt.Println("[Notice]")
			fmt.Println("Incorrect runtime. Please specify a container runtime other than Docker to install Kubernetes v1.24 or later.")
			fmt.Println("You can set \"spec.kubernetes.containerManager\" in the configuration file to \"containerd\" or add \"--container-manager containerd\" to the \"./kk create cluster\" command.")
			fmt.Println("For more information, see:")
			fmt.Println("https://github.com/kubesphere/kubekey/blob/master/docs/commands/kk-create-cluster.md")
			fmt.Println("https://kubernetes.io/docs/setup/production-environment/container-runtimes/#container-runtimes")
			fmt.Println("https://kubernetes.io/blog/2022/02/17/dockershim-faq/")
			fmt.Println("")
			stopFlag = true
		}
	}

	if stopFlag {
		os.Exit(1)
	}

	confirmOK := false
	for !confirmOK {
		fmt.Printf("Continue this init the cluster? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			logger.Log.Fatal(err)
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch strings.ToLower(input) {
		case "yes", "y":
			confirmOK = true
		case "no", "n":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}

type CreateKsConfirm struct {
	common.KubeAction
}

func (u *CreateKsConfirm) Execute(runtime connector.Runtime) error {
	pre := make([]map[string]string, len(runtime.GetAllHosts()), len(runtime.GetAllHosts()))
	for i, host := range runtime.GetAllHosts() {
		if v, ok := host.GetCache().Get(common.NodePreCheck); ok {
			pre[i] = v.(map[string]string)
		} else {
			return errors.New("get node check result failed by host cache")
		}
	}

	if u.KubeConf.Cluster.KubeSphere.Enabled {
		fmt.Printf("desired kubesphere version: %s\n", u.KubeConf.Cluster.KubeSphere.Version)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	confirmOK := false
	for !confirmOK {
		fmt.Printf("Continue install KubeSphere? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "yes", "y":
			confirmOK = true
		case "no", "n":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}
