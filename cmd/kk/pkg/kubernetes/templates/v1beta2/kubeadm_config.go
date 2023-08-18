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
	"strings"
	"text/template"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils"
)

var (
	// KubeadmConfig defines the template of kubeadm configuration file.
	KubeadmConfig = template.Must(template.New("kubeadm-config.yaml").Funcs(utils.FuncMap).Parse(
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
    - "{{ . }}"
    {{- end }}
{{- if .EnableAudit }} 
  extraVolumes:
  - name: k8s-audit
    hostPath: /etc/kubernetes/audit
    mountPath: /etc/kubernetes/audit
    pathType: DirectoryOrCreate
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
	FeatureGatesSecurityDefaultConfiguration = map[string]bool{
		"RotateKubeletServerCertificate": true, //k8s 1.7+
		"TTLAfterFinished":               true, //k8s 1.12+
		"ExpandCSIVolumes":               true, //k8s 1.14+
		"CSIStorageCapacity":             true, //k8s 1.19+
		"SeccompDefault":                 true, //kubelet
	}

	ApiServerArgs = map[string]string{
		"bind-address": "0.0.0.0",
	}
	ApiServerSecurityArgs = map[string]string{
		"bind-address":       "0.0.0.0",
		"authorization-mode": "Node,RBAC",
		// --enable-admission-plugins=EventRateLimit must have a configuration file
		"enable-admission-plugins": "AlwaysPullImages,ServiceAccount,NamespaceLifecycle,NodeRestriction,LimitRanger,ResourceQuota,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,PodNodeSelector,PodSecurity",
		// "audit-log-path":      "/var/log/apiserver/audit.log", // need audit policy
		"profiling":              "false",
		"request-timeout":        "120s",
		"service-account-lookup": "true",
		"tls-min-version":        "VersionTLS12",
		"tls-cipher-suites":      "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	}
	auditArgs = map[string]string{
		"audit-log-format":          "json",
		"audit-log-maxbackup":       "2",
		"audit-log-maxsize":         "200",
		"audit-policy-file":         "/etc/kubernetes/audit/audit-policy.yaml",
		"audit-webhook-config-file": "/etc/kubernetes/audit/audit-webhook.yaml",
	}
	ControllermanagerArgs = map[string]string{
		"bind-address":             "0.0.0.0",
		"cluster-signing-duration": "87600h",
	}
	ControllermanagerSecurityArgs = map[string]string{
		"bind-address":                    "127.0.0.1",
		"cluster-signing-duration":        "87600h",
		"profiling":                       "false",
		"terminated-pod-gc-threshold":     "50",
		"use-service-account-credentials": "true",
	}
	SchedulerArgs = map[string]string{
		"bind-address": "0.0.0.0",
	}
	SchedulerSecurityArgs = map[string]string{
		"bind-address": "127.0.0.1",
		"profiling":    "false",
	}
)

func GetApiServerArgs(securityEnhancement bool, enableAudit bool) map[string]string {
	if securityEnhancement {
		if enableAudit {
			for k, v := range auditArgs {
				ApiServerSecurityArgs[k] = v
			}
		}
		return ApiServerSecurityArgs
	}

	if enableAudit {
		for k, v := range auditArgs {
			ApiServerArgs[k] = v
		}
	}

	return ApiServerArgs
}

func GetControllermanagerArgs(version string, securityEnhancement bool) map[string]string {
	var args map[string]string
	if securityEnhancement {
		args = copyStringMap(ControllermanagerSecurityArgs)
	} else {
		args = copyStringMap(ControllermanagerArgs)
	}

	if versionutil.MustParseSemantic(version).LessThan(versionutil.MustParseSemantic("1.19.0")) {
		delete(args, "cluster-signing-duration")
		args["experimental-cluster-signing-duration"] = "87600h"
	}
	return args
}

func GetSchedulerArgs(securityEnhancement bool) map[string]string {
	if securityEnhancement {
		return SchedulerSecurityArgs
	}
	return SchedulerArgs
}

func UpdateFeatureGatesConfiguration(args map[string]string, kubeConf *common.KubeConf) map[string]string {
	var featureGates []string

	for k, v := range kubeConf.Cluster.Kubernetes.FeatureGates {
		featureGates = append(featureGates, fmt.Sprintf("%s=%v", k, v))
	}

	for k, v := range FeatureGatesDefaultConfiguration {
		// When kubernetes version is less than 1.21,`CSIStorageCapacity` should not be set.
		if k == "CSIStorageCapacity" &&
			versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).LessThan(versionutil.MustParseSemantic("v1.21.0")) {
			continue
		}
		if k == "TTLAfterFinished" &&
			versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.24.0")) {
			continue
		}

		if _, ok := kubeConf.Cluster.Kubernetes.FeatureGates[k]; !ok {
			featureGates = append(featureGates, fmt.Sprintf("%s=%v", k, v))
		}
	}

	args["feature-gates"] = strings.Join(featureGates, ",")

	return args
}

func GetKubeletConfiguration(runtime connector.Runtime, kubeConf *common.KubeConf, criSock string, securityEnhancement bool) map[string]interface{} {
	// When kubernetes version is less than 1.21,`CSIStorageCapacity` should not be set.
	cmp, _ := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.21.0")
	if cmp == -1 {
		delete(FeatureGatesDefaultConfiguration, "CSIStorageCapacity")
	}

	defaultKubeletConfiguration := map[string]interface{}{
		"clusterDomain":      kubeConf.Cluster.Kubernetes.DNSDomain,
		"clusterDNS":         []string{kubeConf.Cluster.ClusterDNS()},
		"maxPods":            kubeConf.Cluster.Kubernetes.MaxPods,
		"podPidsLimit":       kubeConf.Cluster.Kubernetes.PodPidsLimit,
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

	if securityEnhancement {
		defaultKubeletConfiguration["readOnlyPort"] = 0
		defaultKubeletConfiguration["protectKernelDefaults"] = true
		defaultKubeletConfiguration["eventRecordQPS"] = 1
		defaultKubeletConfiguration["streamingConnectionIdleTimeout"] = "5m"
		defaultKubeletConfiguration["makeIPTablesUtilChains"] = true
		defaultKubeletConfiguration["tlsCipherSuites"] = []string{
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		}
		defaultKubeletConfiguration["featureGates"] = FeatureGatesSecurityDefaultConfiguration
	}

	cgroupDriver, err := GetKubeletCgroupDriver(runtime, kubeConf)
	if err != nil {
		logger.Log.Fatal(err)
	}
	if len(cgroupDriver) == 0 {
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
		if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).LessThan(versionutil.MustParseSemantic("v1.21.0")) {
			delete(featureGates, "CSIStorageCapacity")
		}

		if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.24.0")) {
			delete(featureGates, "TTLAfterFinished")
		}

		for k, v := range kubeConf.Cluster.Kubernetes.FeatureGates {
			if _, ok := featureGates[k]; !ok {
				featureGates[k] = v
			}
		}
	}

	if kubeConf.Arg.Debug {
		logger.Log.Debugf("Set kubeletConfiguration: %v", kubeletConfiguration)
	}

	return kubeletConfiguration
}

func GetKubeletCgroupDriver(runtime connector.Runtime, kubeConf *common.KubeConf) (string, error) {
	var cmd, kubeletCgroupDriver string
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Docker, "":
		cmd = "docker info | grep 'Cgroup Driver'"
	case common.Crio:
		cmd = "crio config | grep cgroup_manager"
	case common.Containerd:
		cmd = "containerd config dump | grep SystemdCgroup || echo 'SystemdCgroup = false'"
	case common.Isula:
		cmd = "isula info | grep 'Cgroup Driver'"
	default:
		kubeletCgroupDriver = ""
	}

	checkResult, err := runtime.GetRunner().SudoCmd(cmd, false)
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "Failed to get container runtime cgroup driver.")
	}
	if strings.Contains(checkResult, "systemd") || strings.Contains(checkResult, "SystemdCgroup = true") {
		kubeletCgroupDriver = "systemd"
	} else if strings.Contains(checkResult, "cgroupfs") || strings.Contains(checkResult, "SystemdCgroup = false") {
		kubeletCgroupDriver = "cgroupfs"
	} else {
		return "", errors.Errorf("Failed to get container runtime cgroup driver from %s by run %s", checkResult, cmd)
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

func copyStringMap(m map[string]string) map[string]string {
	cp := make(map[string]string)
	for k, v := range m {
		cp[k] = v
	}

	return cp
}
