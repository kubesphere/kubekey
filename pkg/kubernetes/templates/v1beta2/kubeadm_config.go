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

package v1beta2

import (
	"fmt"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"regexp"
	"strings"
	"text/template"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	funcMap = template.FuncMap{"toYaml": toYAML, "indent": Indent}
	// KubeadmConfig defines the template of kubeadm configuration file.
	KubeadmConfig = template.Must(template.New("kubeadm-config.yaml").Funcs(funcMap).Parse(
		dedent.Dedent(`
{{- if .IsInitCluster -}}
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
etcd:
{{- if .EtcdTypeIsKubeadm }}
  local:
    imageRepository: {{ .EtcdRepo }}
    imageTag: {{ .EtcdTag }}
    serverCertSANs:
    {{- range .ExternalEtcd.Endpoints }}
    - {{ . }}
    {{- end }}
{{- else }}
  external:
    endpoints:
    {{- range .ExternalEtcd.Endpoints }}
    - {{ . }}
    {{- end }}
{{- if .ExternalEtcd.CAFile }}
    caFile: {{ .ExternalEtcd.CAFile }}
{{- end }}
{{- if .ExternalEtcd.CertFile }}
    certFile: {{ .ExternalEtcd.CertFile }}
{{- end }}
{{- if .ExternalEtcd.KeyFile }}
    keyFile: {{ .ExternalEtcd.KeyFile }}
{{- end }}
{{- end }}
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
  dnsDomain: {{ .DNSDomain }}
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
  advertiseAddress: {{ .AdvertiseAddress }}
  bindPort: {{ .BindPort }}
nodeRegistration:
{{- if .CriSock }}
  criSocket: {{ .CriSock }}
{{- end }}
  kubeletExtraArgs:
    cgroup-driver: {{ .CgroupDriver }}
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
{{ toYaml .KubeProxyConfiguration }}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
{{ toYaml .KubeletConfiguration }}

{{- else -}}
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: {{ .ControlPlaneEndpoint }}
    token: "{{ .BootstrapToken }}"
    unsafeSkipCAVerification: true
  tlsBootstrapToken: "{{ .BootstrapToken }}"
{{- if .IsControlPlane }}
controlPlane:
  localAPIEndpoint:
    advertiseAddress: {{ .AdvertiseAddress }}
    bindPort: {{ .BindPort }}
  certificateKey: {{ .CertificateKey }}
{{- end }}
nodeRegistration:
{{- if .CriSock }}
  criSocket: {{ .CriSock }}
{{- end }}
  kubeletExtraArgs:
    cgroup-driver: {{ .CgroupDriver }}

{{- end }}
    `)))
)

var (
	// ref: https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/
	FeatureGatesDefaultConfiguration = map[string]bool{
		"RotateKubeletServerCertificate": true, //k8s 1.7+
		"TTLAfterFinished":               true, //k8s 1.12+
		"ExpandCSIVolumes":               true, //k8s 1.14+
		"CSIStorageCapacity":             true, //k8s 1.19+
	}

	ApiServerArgs = map[string]string{
		"bind-address":        "0.0.0.0",
		"audit-log-maxage":    "30",
		"audit-log-maxbackup": "10",
		"audit-log-maxsize":   "100",
	}
	ControllermanagerArgs = map[string]string{
		"bind-address":                          "0.0.0.0",
		"experimental-cluster-signing-duration": "87600h",
	}
	SchedulerArgs = map[string]string{
		"bind-address": "0.0.0.0",
	}
)

func UpdateFeatureGatesConfiguration(args map[string]string, kubeConf *common.KubeConf) map[string]string {
	// When kubernetes version is less than 1.21,`CSIStorageCapacity` should not be set.
	cmp, _ := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.21.0")
	if cmp == -1 {
		delete(FeatureGatesDefaultConfiguration, "CSIStorageCapacity")
	}

	var featureGates []string

	for k, v := range kubeConf.Cluster.Kubernetes.FeatureGates {
		featureGates = append(featureGates, fmt.Sprintf("%s=%v", k, v))
	}

	for k, v := range FeatureGatesDefaultConfiguration {
		if _, ok := kubeConf.Cluster.Kubernetes.FeatureGates[k]; !ok {
			featureGates = append(featureGates, fmt.Sprintf("%s=%v", k, v))
		}
	}

	args["feature-gates"] = strings.Join(featureGates, ",")

	return args
}

func GetKubeletConfiguration(runtime connector.Runtime, kubeConf *common.KubeConf, criSock string) map[string]interface{} {
	// When kubernetes version is less than 1.21,`CSIStorageCapacity` should not be set.
	cmp, _ := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.21.0")
	if cmp == -1 {
		delete(FeatureGatesDefaultConfiguration, "CSIStorageCapacity")
	}

	defaultKubeletConfiguration := map[string]interface{}{
		"clusterDomain":      kubeConf.Cluster.Kubernetes.DNSDomain,
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
		"podPidsLimit": 1000,
		"evictionHard": map[string]string{
			"memory.available": "5%",
			"pid.available":    "10%",
		},
		"evictionSoft": map[string]string{
			"memory.available": "10%",
		},
		"evictionSoftGracePeriod": map[string]string{
			"memory.available": "2m",
		},
		"evictionMaxPodGracePeriod":        120,
		"evictionPressureTransitionPeriod": "30s",
		"featureGates":                     FeatureGatesDefaultConfiguration,
	}

	cgroupDriver, err := GetKubeletCgroupDriver(runtime, kubeConf)
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
		for k, v := range defaultKubeletConfiguration {
			kubeletConfiguration[k] = v
		}
	}

	if featureGates, ok := kubeletConfiguration["featureGates"].(map[string]bool); ok {
		for k, v := range kubeConf.Cluster.Kubernetes.FeatureGates {
			if _, ok := featureGates[k]; !ok {
				featureGates[k] = v
			}
		}

		for k, v := range FeatureGatesDefaultConfiguration {
			if _, ok := featureGates[k]; !ok {
				featureGates[k] = v
			}
		}
	}

	if kubeConf.Arg.Debug {
		logger.Log.Debug("Set kubeletConfiguration: %v", kubeletConfiguration)
	}

	return kubeletConfiguration
}

func GetKubeletCgroupDriver(runtime connector.Runtime, kubeConf *common.KubeConf) (string, error) {
	var cmd, kubeletCgroupDriver string
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Docker, "":
		cmd = "docker info | grep 'Cgroup Driver' | awk -F': ' '{ print $2; }'"
	case common.Crio:
		cmd = "crio config | grep cgroup_manager | awk -F'= ' '{ print $2; }'"
	case common.Conatinerd:
		cmd = "containerd config dump | grep systemd_cgroup | awk -F'= ' '{ print $2; }'"
	case common.Isula:
		cmd = "isula info | grep 'Cgroup Driver' | awk -F': ' '{ print $2; }'"
	default:
		kubeletCgroupDriver = ""
	}

	checkResult, err := runtime.GetRunner().SudoCmd(cmd, false)
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "Failed to get container runtime cgroup driver.")
	}
	if strings.Contains(checkResult, "systemd") && !strings.Contains(checkResult, "false") {
		kubeletCgroupDriver = "systemd"
	} else {
		kubeletCgroupDriver = "cgroupfs"
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
