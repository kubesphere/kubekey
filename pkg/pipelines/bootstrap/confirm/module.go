package confirm

import (
	"bufio"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/table"
	"github.com/pkg/errors"
	"os"
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

type InstallConfirmModule struct {
	common.KubeCustomModule
	Skip bool
}

func (c *InstallConfirmModule) IsSkip() bool {
	return c.Skip
}

func (c *InstallConfirmModule) Init() {
	c.Name = "ConfirmModule"
	c.Desc = "display confirmation form"
}

func (c *InstallConfirmModule) Run() error {
	var (
		results  []PreCheckResults
		stopFlag bool
	)

	pre := make([]map[string]string, 0, len(c.Runtime.GetAllHosts()))
	for _, host := range c.Runtime.GetAllHosts() {
		if v, ok := host.GetCache().Get(common.NodePreCheck); ok {
			pre = append(pre, v.(map[string]string))
		} else {
			return errors.New("get node check result failed")
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

type DeleteClusterConfirmModule struct {
	common.KubeModule
}

func (d *DeleteClusterConfirmModule) Init() {
	d.Name = "DeleteClusterConfirmModule"
	d.Desc = "display delete confirmation form"
}

func (d *DeleteClusterConfirmModule) Run() error {
	reader := bufio.NewReader(os.Stdin)

	var res string
	for {
		fmt.Printf("Are you sure to delete this cluster? [yes/no]: ")
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

type DeleteNodeConfirmModule struct {
	common.KubeModule
}

func (d *DeleteNodeConfirmModule) Init() {
	d.Name = "DeleteNodeConfirmModule"
	d.Desc = "display delete node confirmation form"
}

func (d *DeleteNodeConfirmModule) Run() error {
	reader := bufio.NewReader(os.Stdin)

	var res string
	for {
		fmt.Printf("Are you sure to delete this node? [yes/no]: ")
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
