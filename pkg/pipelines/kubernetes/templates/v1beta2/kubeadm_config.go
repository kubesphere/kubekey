package v1beta2

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
	"text/template"
)

var (
	funcMap = template.FuncMap{"toYaml": toYAML, "indent": Indent}
	// KubeadmConfig defines the template of kubeadm configuration file.
	KubeadmConfig = template.Must(template.New("kubeadm-config.yaml").Funcs(funcMap).Parse(
		dedent.Dedent(`---
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
etcd:
  external:
    endpoints:
    {{- range .ExternalEtcd.Endpoints }}
    - {{ . }}
    {{- end }}
    caFile: {{ .ExternalEtcd.CaFile }}
    certFile: {{ .ExternalEtcd.CertFile }}
    keyFile: {{ .ExternalEtcd.KeyFile }}
dns:
  type: CoreDNS
  imageRepository: {{ .CorednsRepo }}
  imageTag: {{ .CorednsTag }}
imageRepository: {{ .ImageRepo }}
kubernetesVersion: {{ .Version }}
certificatesDir: /etc/kubernetes/pki
clusterName: {{ .ClusterName }}
controlPlaneEndpoint: {{ .ControlPlaneEndpoint }}
networking:
  dnsDomain: {{ .ClusterName }}
  podSubnet: {{ .PodSubnet }}
  serviceSubnet: {{ .ServiceSubnet }}
apiServer:
  extraArgs:
{{ toYaml .ApiServerArgs | indent 4}}
  certSANs:
    {{- range .CertSANs }}
    - {{ . }}
    {{- end }}
controllerManager:
  extraArgs:
    node-cidr-mask-size: "{{ .NodeCidrMaskSize }}"
{{ toYaml .ControllerManagerArgs | indent 4 }}
  extraVolumes:
  - name: host-time
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
scheduler:
  extraArgs:
{{ toYaml .SchedulerArgs | indent 4 }}

---
apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: {{ .ControlPlanAddr }}
  bindPort: {{ .ControlPlanPort }}
{{- if .CriSock }}
nodeRegistration:
  criSocket: {{ .CriSock }}
{{- end }}
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
{{ toYaml .KubeProxyConfiguration }}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
{{ toYaml .KubeletConfiguration }}

    `)))
)

var (
	ApiServerArgs = map[string]string{
		"bind-address":        "0.0.0.0",
		"audit-log-maxage":    "30",
		"audit-log-maxbackup": "10",
		"audit-log-maxsize":   "100",
		"feature-gates":       "ExpandCSIVolumes=true,RotateKubeletServerCertificate=true",
	}
	ControllermanagerArgs = map[string]string{
		"bind-address":                          "0.0.0.0",
		"experimental-cluster-signing-duration": "87600h",
		"feature-gates":                         "ExpandCSIVolumes=true,RotateKubeletServerCertificate=true",
	}
	SchedulerArgs = map[string]string{
		"bind-address":  "0.0.0.0",
		"feature-gates": "ExpandCSIVolumes=true,RotateKubeletServerCertificate=true",
	}
)

func GetKubeletConfiguration(runtime connector.Runtime, kubeConf *common.KubeConf, criSock string) map[string]interface{} {
	defaultKubeletConfiguration := map[string]interface{}{
		"clusterDomain":      kubeConf.Cluster.Kubernetes.ClusterName,
		"clusterDNS":         []string{kubeConf.Cluster.ClusterDNS()},
		"maxPods":            kubeConf.Cluster.Kubernetes.MaxPods,
		"rotateCertificates": true,
		"kubeReserved": map[string]string{
			"cpu":    "200m",
			"memory": "250Mi",
		},
		"systemReserved": map[string]string{
			"cpu":    "200m",
			"memory": "250Mi",
		},
		"evictionHard": map[string]string{
			"memory.available": "5%",
		},
		"evictionSoft": map[string]string{
			"memory.available": "10%",
		},
		"evictionSoftGracePeriod": map[string]string{
			"memory.available": "2m",
		},
		"evictionMaxPodGracePeriod":        120,
		"evictionPressureTransitionPeriod": "30s",
		"featureGates": map[string]bool{
			"ExpandCSIVolumes":               true,
			"RotateKubeletServerCertificate": true,
		},
	}

	cgroupDriver, err := getKubeletCgroupDriver(runtime, kubeConf)
	if err != nil {
		logger.Log.Fatal(err)
	}
	if len(cgroupDriver) != 0 {
		defaultKubeletConfiguration["cgroupDriver"] = "systemd"
	}

	if len(criSock) != 0 {
		defaultKubeletConfiguration["containerLogMaxSize"] = "5Mi"
		defaultKubeletConfiguration["containerLogMaxFiles"] = 3
	}

	customKubeletConfiguration := make(map[string]interface{})
	if len(kubeConf.Cluster.Kubernetes.KubeletConfiguration.Raw) != 0 {
		err := yaml.Unmarshal(kubeConf.Cluster.Kubernetes.KubeletConfiguration.Raw, &customKubeletConfiguration)
		if err != nil {
			logger.Log.Fatal("failed to parse kubelet configuration")
		}
	}

	kubeletConfiguration := make(map[string]interface{})
	if len(customKubeletConfiguration) != 0 {
		for customArg := range customKubeletConfiguration {
			if _, ok := defaultKubeletConfiguration[customArg]; ok {
				kubeletConfiguration[customArg] = customKubeletConfiguration[customArg]
				delete(defaultKubeletConfiguration, customArg)
				delete(customKubeletConfiguration, customArg)
			} else {
				kubeletConfiguration[customArg] = customKubeletConfiguration[customArg]
			}
		}
	}

	if len(defaultKubeletConfiguration) != 0 {
		for defaultArg := range defaultKubeletConfiguration {
			kubeletConfiguration[defaultArg] = defaultKubeletConfiguration[defaultArg]
		}
	}
	return kubeletConfiguration
}

func getKubeletCgroupDriver(runtime connector.Runtime, kubeConf *common.KubeConf) (string, error) {
	var cmd, kubeletCgroupDriver string
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case "docker":
		cmd = "docker info | grep 'Cgroup Driver' | awk -F': ' '{ print $2; }'"
	case "crio":
		cmd = "crio config | grep cgroup_manager | awk -F'= ' '{ print $2; }'"
	case "containerd":
		cmd = "containerd config dump | grep systemd_cgroup | awk -F'= ' '{ print $2; }'"
	case "isula":
		cmd = "isula info | grep 'Cgroup Driver' | awk -F': ' '{ print $2; }'"
	default:
		kubeletCgroupDriver = ""
	}

	checkResult, err := runtime.GetRunner().Cmd(fmt.Sprintf("sudo env PATH=$PATH /bin/sh -c \"%s\"", cmd), false)
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get container runtime cgroup driver failed")
	}
	if strings.Contains(checkResult, "systemd") && !strings.Contains(checkResult, "false") {
		kubeletCgroupDriver = "systemd"
	} else {
		kubeletCgroupDriver = ""
	}

	return kubeletCgroupDriver, nil
}

func GetKubeProxyConfiguration(kubeConf *common.KubeConf) map[string]interface{} {
	defaultKubeProxyConfiguration := map[string]interface{}{
		"clusterCIDR": kubeConf.Cluster.Network.KubePodsCIDR,
		"mode":        kubeConf.Cluster.Kubernetes.ProxyMode,
		"iptables": map[string]interface{}{
			"masqueradeAll": kubeConf.Cluster.Kubernetes.MasqueradeAll,
			"masqueradeBit": 14,
			"minSyncPeriod": "0s",
			"syncPeriod":    "30s",
		},
	}

	customKubeProxyConfiguration := make(map[string]interface{})
	if len(kubeConf.Cluster.Kubernetes.KubeProxyConfiguration.Raw) != 0 {
		err := yaml.Unmarshal(kubeConf.Cluster.Kubernetes.KubeProxyConfiguration.Raw, &customKubeProxyConfiguration)
		if err != nil {
			logger.Log.Fatal("failed to parse kube-proxy's configuration")
		}
	}

	kubeProxyConfiguration := make(map[string]interface{})
	if len(customKubeProxyConfiguration) != 0 {
		for customArg := range customKubeProxyConfiguration {
			if _, ok := defaultKubeProxyConfiguration[customArg]; ok {
				kubeProxyConfiguration[customArg] = customKubeProxyConfiguration[customArg]
				delete(defaultKubeProxyConfiguration, customArg)
				delete(customKubeProxyConfiguration, customArg)
			} else {
				kubeProxyConfiguration[customArg] = customKubeProxyConfiguration[customArg]
			}
		}
	}

	if len(defaultKubeProxyConfiguration) != 0 {
		for defaultArg := range defaultKubeProxyConfiguration {
			kubeProxyConfiguration[defaultArg] = defaultKubeProxyConfiguration[defaultArg]
		}
	}

	return kubeProxyConfiguration
}

func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func Indent(n int, text string) string {
	startOfLine := regexp.MustCompile(`(?m)^`)
	indentation := strings.Repeat(" ", n)
	return startOfLine.ReplaceAllLiteralString(text, indentation)
}
