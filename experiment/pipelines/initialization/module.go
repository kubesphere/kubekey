package initialization

import (
	"bufio"
	"fmt"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"github.com/kubesphere/kubekey/experiment/core/modules"
	"github.com/kubesphere/kubekey/experiment/core/prepare"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/table"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type InitializationModule struct {
	modules.BaseTaskModule
}

func (i *InitializationModule) Init() {
	i.Name = "NodeInitializationModule"

	preCheck := modules.Task{
		Name:  "NodePreCheck",
		Hosts: i.Runtime.AllNodes,
		Prepare: &prepare.FastPrepare{
			Inject: func() (bool, error) {
				if len(i.Runtime.EtcdNodes)%2 == 0 {
					logger.Log.Warnln("The number of etcd is even. Please configure it to be odd.")
					return false, errors.New("the number of etcd is even")
				}
				return true, nil
			}},
		Action:   new(NodePreCheck),
		Parallel: true,
	}

	i.Tasks = []modules.Task{
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
	modules.CustomModule
}

func (c *ConfirmModule) Init() {
	c.Name = "ConfirmModule"
	logger.Log.SetModule(c.Name)
}

func (c *ConfirmModule) Run() error {
	logger.Log.Info("Begin Run")
	var (
		results  []PreCheckResults
		stopFlag bool
	)

	nodePreChecks, ok := c.RootCache.Get("nodePreCheck")
	if !ok {
		return errors.New("get node check result failed")
	}

	pre := nodePreChecks.(map[string]interface{})
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
