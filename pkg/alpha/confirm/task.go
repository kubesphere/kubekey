package confirm

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

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
