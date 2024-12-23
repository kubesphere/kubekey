package infrastructure

import (
	"context"
	"crypto/rand"
	"math/big"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// SequenceSelector is a HostSelectorFunc that adjusts the number of hosts in a specified group
// within the inventory. If the specified number of hosts (groupHostNum) is less than the current number
// of hosts in the group, it removes the excess hosts. If the specified number is greater, it adds hosts
// from the ungrouped hosts to the group.
//
// The returned function takes the following parameters:
// - ctx: The context for the operation.
// - groupName: The name of the group to adjust.
// - groupHostNum: The desired number of hosts in the group.
// - inventory: The inventory object containing the groups and hosts.
var SequenceSelector = func(ctx context.Context, groupName string, remain int, inventory *kkcorev1.Inventory) []string {
	var availableHosts []string
	groups := inventory.Spec.Groups[groupName]
	if remain > 0 {
		// Add hosts from ungrouped
		ungrouped, ok := variable.ConvertGroup(*inventory)[_const.VariableUnGrouped].([]string)
		if !ok {
			klog.ErrorS(nil, "Failed to get ungrouped hosts")

			return availableHosts
		}
		availableHosts = ungrouped[:remain]
		groups.Hosts = append(groups.Hosts, availableHosts...)
		if inventory.Spec.Groups == nil {
			inventory.Spec.Groups = make(map[string]kkcorev1.InventoryGroup, 0)
		}
		inventory.Spec.Groups[groupName] = groups
	}

	return availableHosts
}

// RandomSelector is a HostSelectorFunc that randomly selects hosts for a given group.
// If the number of requested hosts (groupHostNum) is less than the current number of hosts in the group,
// it shuffles the current hosts and trims the excess hosts.
// If the number of requested hosts is greater than the current number of hosts in the group,
// it adds hosts from the ungrouped hosts to meet the requested number.
// The function modifies the inventory to reflect the changes in the group hosts.
//
// Returns:
//
//	HostSelectorFunc: A function that selects hosts for a group based on the specified criteria.
var RandomSelector = func(ctx context.Context, groupName string, remain int, inventory *kkcorev1.Inventory) []string {
	var availableHosts []string
	groups := inventory.Spec.Groups[groupName]
	if remain > 0 {
		// Add hosts from ungrouped
		ungrouped, ok := variable.ConvertGroup(*inventory)[_const.VariableUnGrouped].([]string)
		if !ok {
			klog.ErrorS(nil, "Failed to get ungrouped hosts")

			return nil
		}
		shuffleHosts(ungrouped)
		availableHosts = ungrouped[:remain]
		groups.Hosts = append(groups.Hosts, availableHosts...)
		if inventory.Spec.Groups == nil {
			inventory.Spec.Groups = make(map[string]kkcorev1.InventoryGroup, 0)
		}
		inventory.Spec.Groups[groupName] = groups
	}

	return availableHosts
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
