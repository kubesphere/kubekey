package preinstall

import (
	"bufio"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/mitchellh/mapstructure"
	"github.com/modood/table"
	"os"
	"strings"
)

type PrecheckResults struct {
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

var (
	checkResults  = make(map[string]interface{})
	baseSoftwares = []string{"sudo", "curl", "openssl", "ebtables", "socat", "ipset", "conntrack", "docker", "showmount", "rbd", "glusterfs"}
)

func Precheck(mgr *manager.Manager) error {
	return mgr.RunTaskOnAllNodes(precheck, true)
}

func precheck(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	var results = make(map[string]interface{})
	results["name"] = node.Name
	for _, software := range baseSoftwares {
		_, err := mgr.Runner.RunCmd(fmt.Sprintf("which %s", software))
		if err != nil {
			results[software] = ""
		} else {
			results[software] = "y"
		}
	}
	output, err := mgr.Runner.RunCmd("date +\"%Z %H:%M:%S\"")
	if err != nil {
		results["time"] = ""
	} else {
		results["time"] = strings.TrimSpace(output)
	}

	checkResults[node.Name] = results
	return nil
}

func PrecheckConfirm(mgr *manager.Manager) {

	results := []PrecheckResults{}
	for node := range checkResults {
		var result PrecheckResults
		mapstructure.Decode(checkResults[node], &result)
		results = append(results, result)
	}
	table.OutputA(results)
	reader := bufio.NewReader(os.Stdin)
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
			mgr.Logger.Fatal(err)
		}
		input = strings.TrimSpace(input)

		switch input {
		case "yes":
			break Loop
		case "no":
			os.Exit(0)
		default:
			continue
		}
	}
}
