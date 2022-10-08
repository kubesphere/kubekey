/*
 Copyright 2022 The KubeSphere Authors.

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

package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
)

// Test suite constants for e2e config variables.
const (
	KubernetesVersionManagement     = "KUBERNETES_VERSION_MANAGEMENT"
	KubernetesVersion               = "KUBERNETES_VERSION"
	CNIPath                         = "CNI"
	CNIResources                    = "CNI_RESOURCES"
	KubernetesVersionUpgradeFrom    = "KUBERNETES_VERSION_UPGRADE_FROM"
	KubernetesVersionUpgradeTo      = "KUBERNETES_VERSION_UPGRADE_TO"
	CPMachineTemplateUpgradeTo      = "CONTROL_PLANE_MACHINE_TEMPLATE_UPGRADE_TO"
	WorkersMachineTemplateUpgradeTo = "WORKERS_MACHINE_TEMPLATE_UPGRADE_TO"
	EtcdVersionUpgradeTo            = "ETCD_VERSION_UPGRADE_TO"
	CoreDNSVersionUpgradeTo         = "COREDNS_VERSION_UPGRADE_TO"
	IPFamily                        = "IP_FAMILY"
)

func Byf(format string, a ...interface{}) {
	By(fmt.Sprintf(format, a...))
}
