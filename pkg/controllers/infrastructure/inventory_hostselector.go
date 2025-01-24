package infrastructure

import (
	"context"
	"crypto/rand"
	"math/big"

	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func getHostSelectorFunc(policy capkkinfrav1beta1.HostSelectorPolicy) HostSelectorFunc {
	switch policy {
	case capkkinfrav1beta1.HostSelectorRandom:
		return RandomSelector()
	case capkkinfrav1beta1.HostSelectorSequence:
		return SequenceSelector()
	default:
		return RandomSelector()
	}
}

// SequenceSelector returns a HostSelectorFunc that adjusts the number of hosts in a specified group
// within the inventory. If the specified number of hosts (groupHostNum) is less than the current number
// of hosts in the group, it removes the excess hosts. If the specified number is greater, it adds hosts
// from the ungrouped hosts to the group.
//
// The returned function takes the following parameters:
// - ctx: The context for the operation.
// - groupName: The name of the group to adjust.
// - groupHostNum: The desired number of hosts in the group.
// - inventory: The inventory object containing the groups and hosts.
func SequenceSelector() HostSelectorFunc {
	return func(ctx context.Context, groupName string, groupHostNum int, inventory *kkcorev1.Inventory) {
		groups := inventory.Spec.Groups[groupName]
		currentHosts := groups.Hosts

		if groupHostNum < len(currentHosts) {
			// Remove excess hosts
			groups.Hosts = currentHosts[:groupHostNum]
			inventory.Spec.Groups[groupName] = groups
		} else if groupHostNum > len(currentHosts) {
			// Add hosts from ungrouped
			ungrouped, ok := variable.ConvertGroup(*inventory)[_const.VariableUnGrouped].([]string)
			if !ok {
				klog.ErrorS(nil, "Failed to get ungrouped hosts")

				return
			}
			groups.Hosts = append(groups.Hosts, ungrouped[:groupHostNum-len(currentHosts)]...)
			inventory.Spec.Groups[groupName] = groups
		}
	}
}

// RandomSelector returns a HostSelectorFunc that randomly selects hosts for a given group.
// If the number of requested hosts (groupHostNum) is less than the current number of hosts in the group,
// it shuffles the current hosts and trims the excess hosts.
// If the number of requested hosts is greater than the current number of hosts in the group,
// it adds hosts from the ungrouped hosts to meet the requested number.
// The function modifies the inventory to reflect the changes in the group hosts.
//
// Returns:
//
//	HostSelectorFunc: A function that selects hosts for a group based on the specified criteria.
func RandomSelector() HostSelectorFunc {
	return func(ctx context.Context, groupName string, groupHostNum int, inventory *kkcorev1.Inventory) {
		groups := inventory.Spec.Groups[groupName]
		currentHosts := groups.Hosts

		if groupHostNum < len(currentHosts) {
			// Shuffle and trim excess hosts
			shuffleHosts(currentHosts)
			groups.Hosts = currentHosts[:groupHostNum]
			inventory.Spec.Groups[groupName] = groups
		} else if groupHostNum > len(currentHosts) {
			// Add hosts from ungrouped
			ungrouped, ok := variable.ConvertGroup(*inventory)[_const.VariableUnGrouped].([]string)
			if !ok {
				klog.ErrorS(nil, "Failed to get ungrouped hosts")

				return
			}
			shuffleHosts(ungrouped)
			groups.Hosts = append(currentHosts, ungrouped[:groupHostNum-len(currentHosts)]...)
			inventory.Spec.Groups[groupName] = groups
		}
	}
}

// shuffleHosts securely shuffles a slice of hosts using crypto/rand.
func shuffleHosts(hosts []string) {
	for i := len(hosts) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue // Skip in case of error
		}
		hosts[i], hosts[j.Int64()] = hosts[j.Int64()], hosts[i]
	}
}
