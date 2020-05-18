package create

import (
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/spf13/cobra"
)

func NewCmdCreateCfg() *cobra.Command {
	var addons, name, clusterCfgPath string
	var clusterCfgCmd = &cobra.Command{
		Use:   "config",
		Short: "Create cluster info configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.GenerateClusterObj(addons, name, clusterCfgPath)
			if err != nil {
				return err
			}
			return err
			//clusterConfig()
		},
	}

	clusterCfgCmd.Flags().StringVarP(&addons, "add", "", "", "Add plugins")
	clusterCfgCmd.Flags().StringVarP(&name, "name", "", "", "Name of config file")
	clusterCfgCmd.Flags().StringVarP(&clusterCfgPath, "config", "f", "", "cluster info config")
	return clusterCfgCmd
}

//func getConfig(reader *bufio.Reader, text, def string) (string, error) {
//	for {
//		if def == "" {
//			fmt.Printf("[+] %s [%s]: ", text, "none")
//		} else {
//			fmt.Printf("[+] %s [%s]: ", text, def)
//		}
//		input, err := reader.ReadString('\n')
//		if err != nil {
//			return "", err
//		}
//		input = strings.TrimSpace(input)
//
//		if input != "" {
//			return input, nil
//		}
//		return def, nil
//	}
//}
//
//func writeConfig(cluster *kubekeyapi.ClusterSpec, configFile string, print bool) error {
//	yamlConfig, err := yaml.Marshal(*cluster)
//	if err != nil {
//		return err
//	}
//	log.Infof("Deploying cluster info file: %s", configFile)
//
//	configString := fmt.Sprintf("%s", string(yamlConfig))
//	if print {
//		fmt.Printf("%s", configString)
//		//return nil
//	}
//	return ioutil.WriteFile(configFile, []byte(configString), 0640)
//}
//
//func clusterConfig() error {
//	clusterCfg := kubekeyapi.ClusterSpec{}
//	reader := bufio.NewReader(os.Stdin)
//
//	// Get number of hosts
//	numberOfHostsString, err := getConfig(reader, "Number of Hosts", "1")
//	if err != nil {
//		return err
//	}
//	numberOfHostsInt, err := strconv.Atoi(numberOfHostsString)
//	if err != nil {
//		return err
//	}
//
//	//sshKeyPath, err := getConfig(reader, "Cluster Level SSH Private Key Path", "~/.ssh-bak/id_rsa")
//	//if err != nil {
//	//	return err
//	//}
//	//clusterCfg.SSHKeyPath = sshKeyPath
//	// Get Hosts config
//	masterNumber := 0
//	clusterCfg.Hosts = make([]kubekeyapi.HostCfg, 0)
//	for i := 0; i < numberOfHostsInt; i++ {
//		hostCfg, isMaster, err := getHostConfig(reader, i)
//		if err != nil {
//			return err
//		}
//		clusterCfg.Hosts = append(clusterCfg.Hosts, *hostCfg)
//		if isMaster == true {
//			masterNumber = masterNumber + 1
//		}
//	}
//	if masterNumber > 1 {
//		lbCfg := kubekeyapi.LBKubeApiserverCfg{}
//		address, err := getConfig(reader, fmt.Sprintf("Address of LoadBalancer for KubeApiserver"), "")
//		if err != nil {
//			return err
//		}
//		lbCfg.Address = address
//
//		port, err := getConfig(reader, fmt.Sprintf("Port of LoadBalancer for KubeApiserver"), kubekeyapi.DefaultLBPort)
//		if err != nil {
//			return err
//		}
//		lbCfg.Port = port
//
//		domain, err := getConfig(reader, fmt.Sprintf("Address of LoadBalancer for KubeApiserver"), kubekeyapi.DefaultLBDomain)
//		if err != nil {
//			return err
//		}
//		lbCfg.Domain = domain
//
//		clusterCfg.LBKubeApiserver = lbCfg
//	}
//	// Get Kubernetes version
//	version, err := getConfig(reader, fmt.Sprintf("Kubernetes Version"), kubekeyapi.DefaultKubeVersion)
//	if err != nil {
//		return err
//	}
//	clusterCfg.KubeCluster.Version = version
//
//	// Get Kubernetes images repo
//	repo, err := getConfig(reader, fmt.Sprintf("Kubernetes Images Repo"), kubekeyapi.DefaultKubeImageRepo)
//	if err != nil {
//		return err
//	}
//	clusterCfg.KubeCluster.ImageRepo = repo
//
//	// Get Kubernetes cluster name
//	clusterName, err := getConfig(reader, fmt.Sprintf("Kubernetes Cluster Name"), kubekeyapi.DefaultClusterName)
//	if err != nil {
//		return err
//	}
//	clusterCfg.KubeCluster.ClusterName = clusterName
//
//	// Get Network config
//	networkConfig, err := getNetworkConfig(reader)
//	if err != nil {
//		return err
//	}
//	clusterCfg.Network = *networkConfig
//
//	dir, _ := os.Executable()
//	exPath := filepath.Dir(dir)
//	configFile := fmt.Sprintf("%s/%s", exPath, "cluster-info.yaml")
//	return writeConfig(&clusterCfg, configFile, true)
//}
//
//func getHostConfig(reader *bufio.Reader, index int) (*kubekeyapi.HostCfg, bool, error) {
//	isMaster := false
//	host := kubekeyapi.HostCfg{}
//	address, err := getConfig(reader, fmt.Sprintf("SSH Address of host (%d)", index+1), "")
//	if err != nil {
//		return nil, false, err
//	}
//	host.SSHAddress = address
//
//	port, err := getConfig(reader, fmt.Sprintf("SSH Port of host (%s)", address), kubekeyapi.DefaultSSHPort)
//	if err != nil {
//		return nil, false, err
//	}
//	host.Port = port
//
//	sshUser, err := getConfig(reader, fmt.Sprintf("SSH User of host (%s)", address), "root")
//	if err != nil {
//		return nil, false, err
//	}
//	host.User = sshUser
//
//	password, err := getConfig(reader, fmt.Sprintf("SSH Password of host (%s)", address), "")
//	if err != nil {
//		return nil, false, err
//	}
//	host.Password = password
//
//	hostRole, err := getConfig(reader, fmt.Sprintf("What's host (%s) role? (0: etcd, 1: master, 2: worker)", address), "012")
//	if err != nil {
//		return nil, false, err
//	}
//
//	if strings.Contains(hostRole, "0") {
//		host.Role = append(host.Role, kubekeyapi.ETCDRole)
//	}
//	if strings.Contains(hostRole, "1") {
//		host.Role = append(host.Role, kubekeyapi.MasterRole)
//		isMaster = true
//	}
//	if strings.Contains(hostRole, "2") {
//		host.Role = append(host.Role, kubekeyapi.WorkerRole)
//	}
//
//	hostnameOverride, err := getConfig(reader, fmt.Sprintf("Override Hostname of host (%s)", address), "")
//	if err != nil {
//		return nil, false, err
//	}
//	host.HostName = hostnameOverride
//
//	internalAddress, err := getConfig(reader, fmt.Sprintf("Internal IP of host (%s)", address), "")
//	if err != nil {
//		return nil, false, err
//	}
//	host.InternalAddress = internalAddress
//
//	return &host, isMaster, nil
//}
//
//func getNetworkConfig(reader *bufio.Reader) (*kubekeyapi.NetworkConfig, error) {
//	networkConfig := kubekeyapi.NetworkConfig{}
//
//	networkPlugin, err := getConfig(reader, "Network Plugin Type (0: calico, 1: flannel)", "0")
//	if err != nil {
//		return nil, err
//	}
//	networkPluginInt, _ := strconv.Atoi(networkPlugin)
//	if networkPluginInt == 0 {
//		networkConfig.Plugin = "calico"
//	}
//	if networkPluginInt == 1 {
//		networkConfig.Plugin = "flannel"
//	}
//
//	podsCIDR, err := getConfig(reader, "Specify range of IP addresses for the pod network.", kubekeyapi.DefaultPodsCIDR)
//	if err != nil {
//		return nil, err
//	}
//	networkConfig.KubePodsCIDR = podsCIDR
//
//	serviceCIDR, err := getConfig(reader, "Use alternative range of IP address for service VIPs.", kubekeyapi.DefaultServiceCIDR)
//	if err != nil {
//		return nil, err
//	}
//	networkConfig.KubeServiceCIDR = serviceCIDR
//	return &networkConfig, nil
//}
