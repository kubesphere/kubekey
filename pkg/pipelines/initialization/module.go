package initialization

import (
	"bufio"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/initialization/templates"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/table"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

type NodeInitializationModule struct {
	common.KubeModule
	Skip bool
}

func (n *NodeInitializationModule) IsSkip() bool {
	return n.Skip
}

func (n *NodeInitializationModule) Init() {
	n.Name = "NodeInitializationModule"

	preCheck := modules.Task{
		Name:  "NodePreCheck",
		Desc:  "a pre-check on nodes",
		Hosts: n.Runtime.GetAllHosts(),
		Prepare: &prepare.FastPrepare{
			Inject: func() (bool, error) {
				if len(n.Runtime.GetHostsByRole(common.Etcd))%2 == 0 {
					logger.Log.Error("The number of etcd is even. Please configure it to be odd.")
					return false, errors.New("the number of etcd is even")
				}
				return true, nil
			}},
		Action:   new(NodePreCheck),
		Parallel: true,
	}

	n.Tasks = []modules.Task{
		preCheck,
	}
}

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

type ConfirmModule struct {
	common.KubeCustomModule
	Skip bool
}

func (c *ConfirmModule) IsSkip() bool {
	return c.Skip
}

func (c *ConfirmModule) Init() {
	c.Name = "ConfirmModule"
	c.Desc = "display confirmation form"
}

func (c *ConfirmModule) Run() error {
	var (
		results  []PreCheckResults
		stopFlag bool
	)

	nodePreChecks, ok := c.RootCache.Get("nodePreCheck")
	if !ok {
		return errors.New("get node check result failed")
	}

	pre := nodePreChecks.(map[string]map[string]string)
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
Loop:
	for {
		fmt.Printf("Continue this installation? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			logger.Log.Fatal(err)
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "yes":
			break Loop
		case "no":
			os.Exit(0)
		default:
			continue
		}
	}
	return nil
}

type ConfigureOSModule struct {
	common.KubeModule
}

func (c *ConfigureOSModule) Init() {
	c.Name = "ConfigureOSModule"

	initOS := modules.Task{
		Name:     "InitOS",
		Desc:     "prepare to init OS",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeConfigureOS),
		Parallel: true,
	}

	GenerateScript := modules.Task{
		Name:  "GenerateScript",
		Desc:  "generate init os script",
		Hosts: c.Runtime.GetAllHosts(),
		Action: &action.Template{
			Template: templates.InitOsScriptTmpl,
			Dst:      filepath.Join(common.KubeScriptDir, "initOS.sh"),
			Data: util.Data{
				"Hosts": c.KubeConf.ClusterHosts,
			},
		},
		Parallel: true,
	}

	ExecScript := modules.Task{
		Name:     "ExecScript",
		Desc:     "exec init os script",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeExecScript),
		Parallel: true,
	}

	c.Tasks = []modules.Task{
		initOS,
		GenerateScript,
		ExecScript,
	}
}
