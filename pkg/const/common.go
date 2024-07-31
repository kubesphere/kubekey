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
	// VariableHostName the value is host_name
	VariableHostName = "inventory_name"
	// VariableGlobalHosts the value is host_var which defined in inventory.
	VariableGlobalHosts = "inventory_hosts"
	// VariableGroupsAll the value is a all host_name slice of VariableGroups.
	VariableGroupsAll = "all"
)

const ( // === From GatherFact ===
	// VariableOS the value is os information.
	VariableOS = "os"
	// VariableOSRelease the value is os-release of VariableOS.
	VariableOSRelease = "release"
	// VariableOSKernelVersion the value is kernel version of VariableOS.
	VariableOSKernelVersion = "kernel_version"
	// VariableOSKHostName the value is hostname of VariableOS.
	VariableOSKHostName = "hostname"
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
	VariableItem = "item"
)
