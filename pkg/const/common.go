/*
Copyright 2023 The KubeSphere Authors.

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

package _const

// variable specific key in system
const ( // === From inventory ===
	// VariableLocalHost is the default local host name in inventory.
	VariableLocalHost = "localhost"
	// VariableIPv4 is the ipv4 in inventory.
	VariableIPv4 = "internal_ipv4"
	// VariableIPv6 is the ipv6 in inventory.
	VariableIPv6 = "internal_ipv6"
	// VariableGroups the value is a host_name slice
	VariableGroups = "groups"
	// VariableConnector is connector parameter in inventory.
	VariableConnector = "connector"
	// VariableConnectorType is connected type for VariableConnector.
	VariableConnectorType = "type"
	// VariableConnectorHost is connected address for VariableConnector.
	VariableConnectorHost = "host"
	// VariableConnectorPort is connected address for VariableConnector.
	VariableConnectorPort = "port"
	// VariableConnectorUser is connected user for VariableConnector.
	VariableConnectorUser = "user"
	// VariableConnectorPassword is connected type for VariableConnector.
	VariableConnectorPassword = "password"
	// VariableConnectorPrivateKey is connected auth key for VariableConnector.
	VariableConnectorPrivateKey = "private_key"
	// VariableConnectorKubeconfig is connected auth key for VariableConnector.
	VariableConnectorKubeconfig = "kubeconfig"
)

const ( // === From system generate ===
	// VariableInventoryName the value which defined in inventory.spec.host.
	VariableInventoryName = "inventory_name"
	// VariableHostName the value is node hostname, default VariableInventoryName.
	// If VariableInventoryName is "localhost". try to set the actual name.
	VariableHostName = "hostname"
	// VariableGlobalHosts the value is host_var which defined in inventory.
	VariableGlobalHosts = "inventory_hosts"
	// VariableGroupsAll the value is a all host_name slice of VariableGroups.
	VariableGroupsAll = "all"
	// VariableUnGrouped the value is a all host_name slice of VariableGroups.
	VariableUnGrouped = "ungrouped"
)

const ( // === From GatherFact ===
	// VariableOS the value is os information.
	VariableOS = "os"
	// VariableOSRelease the value is os-release of VariableOS.
	VariableOSRelease = "release"
	// VariableOSKernelVersion the value is kernel version of VariableOS.
	VariableOSKernelVersion = "kernel_version"
	// VariableOSHostName the value is hostname of VariableOS.
	VariableOSHostName = "hostname"
	// VariableOSArchitecture the value is architecture of VariableOS.
	VariableOSArchitecture = "architecture"
	// VariableProcess the value is process information.
	VariableProcess = "process"
	// VariableProcessCPU the value is cpu info of VariableProcess.
	VariableProcessCPU = "cpuInfo"
	// VariableProcessMemory the value is memory info of VariableProcess.
	VariableProcessMemory = "memInfo"
)

const ( // === From runtime ===
	// VariableItem for "loop" argument when run a task.
	VariableItem = "item"
)

const ( // === From env ===
	// ENV_SHELL which shell operator use in local connector.
	ENV_SHELL = "SHELL"
	// ENV_EXECUTOR_VERBOSE which verbose use in playbook pod.
	ENV_EXECUTOR_VERBOSE = "EXECUTOR_VERBOSE"
	// ENV_EXECUTOR_IMAGE which image use in playbook pod.
	ENV_EXECUTOR_IMAGE = "EXECUTOR_IMAGE"
	// ENV_EXECUTOR_IMAGE_PULLPOLICY which imagePolicy use in playbook pod.
	ENV_EXECUTOR_IMAGE_PULLPOLICY = "EXECUTOR_IMAGE_PULLPOLICY"
	// ENV_EXECUTOR_CLUSTERROLE which clusterrole use in playbook pod.
	ENV_EXECUTOR_CLUSTERROLE = "EXECUTOR_CLUSTERROLE"
	// ENV_CAPKK_GROUP_CONTROLPLANE the control_plane groups for capkk playbook
	ENV_CAPKK_GROUP_CONTROLPLANE = "CAPKK_GROUP_CONTROLPLANE"
	// ENV_CAPKK_GROUP_WORKER the worker groups for capkk playbook
	ENV_CAPKK_GROUP_WORKER = "CAPKK_GROUP_WORKER"
	// ENV_CAPKK_VOLUME_BINARY is the binary dir for capkk playbook. used in offline installer.
	// the value should be a pvc name.
	ENV_CAPKK_VOLUME_BINARY = "CAPKK_VOLUME_BINARY"
	// ENV_CAPKK_VOLUME_PROJECT is the project dir for capkk playbook. the default project has contained in IMAGE.
	// the value should be a pvc name.
	ENV_CAPKK_VOLUME_PROJECT = "CAPKK_VOLUME_PROJECT"
	// ENV_CAPKK_VOLUME_WORKDIR is the workdir for capkk playbook.
	ENV_CAPKK_VOLUME_WORKDIR = "CAPKK_VOLUME_WORKDIR"
)

const ( // === From CAPKK base on GetCAPKKProject() ===
	// CAPKKWorkdir is the work dir for capkk playbook.
	CAPKKWorkdir = "/kubekey/"
	// CAPKKProjectdir is the project dir for capkk playbook.
	CAPKKProjectdir = "/capkk/project/"
	// CAPKKBinarydir is the path of binary.
	CAPKKBinarydir = "/capkk/kubekey/"
	// CAPKKCloudConfigPath is the cloud-config path.
	CAPKKCloudConfigPath = "/capkk/cloud/cloud-config"
	// CAPKKCloudKubeConfigPath is the cloud-config path.
	CAPKKCloudKubeConfigPath = "/capkk/cloud/kubeconfig"
	// CAPKKPlaybookHostCheck is the playbook for host check.
	CAPKKPlaybookHostCheck = "playbooks/host_check.yaml"
	// CAPKKPlaybookAddNode is the playbook for add node.
	CAPKKPlaybookAddNode = "playbooks/add_node.yaml"
	// CAPKKPlaybookDeleteNode is the playbook for delete node.
	CAPKKPlaybookDeleteNode = "playbooks/delete_node.yaml"
)
