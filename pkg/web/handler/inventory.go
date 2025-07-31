package handler

import (
	"io"
	"reflect"
	"slices"
	"strings"

	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
)

// InventoryHandler handles HTTP requests for inventory resources.
type InventoryHandler struct {
	workdir    string            // Base directory for storing work files
	restconfig *rest.Config      // Kubernetes REST client configuration
	client     ctrlclient.Client // Kubernetes client for API operations
}

// NewInventoryHandler creates a new InventoryHandler instance.
func NewInventoryHandler(workdir string, restconfig *rest.Config, client ctrlclient.Client) *InventoryHandler {
	return &InventoryHandler{workdir: workdir, restconfig: restconfig, client: client}
}

// Post creates a new inventory resource.
// It reads the inventory from the request body and creates it in the Kubernetes cluster.
func (h *InventoryHandler) Post(request *restful.Request, response *restful.Response) {
	inventory := &kkcorev1.Inventory{}
	if err := request.ReadEntity(inventory); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	if err := h.client.Create(request.Request.Context(), inventory); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	_ = response.WriteEntity(inventory)
}

// Patch updates an existing inventory resource.
// It reads the patch data from the request body and applies it to the specified inventory.
func (h *InventoryHandler) Patch(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("inventory")
	data, err := io.ReadAll(request.Request.Body)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	patchType := request.HeaderParameter("Content-Type")

	// Get the existing inventory object.
	inventory := &kkcorev1.Inventory{}
	if err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, inventory); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	// Apply the patch.
	if err := h.client.Patch(request.Request.Context(), inventory, ctrlclient.RawPatch(types.PatchType(patchType), data)); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	_ = response.WriteEntity(inventory)
}

// List returns all inventory resources with optional filtering and sorting.
// It supports field selectors and label selectors for filtering the results.
func (h *InventoryHandler) List(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	var fieldselector fields.Selector
	// Parse field selector from query parameters if present.
	if v, ok := queryParam.Filters[query.ParameterFieldSelector]; ok {
		fs, err := fields.ParseSelector(string(v))
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		fieldselector = fs
	}
	namespace := request.PathParameter("namespace")

	inventoryList := &kkcorev1.InventoryList{}
	// List inventory resources from the Kubernetes API.
	err := h.client.List(request.Request.Context(), inventoryList, &ctrlclient.ListOptions{Namespace: namespace, LabelSelector: queryParam.Selector(), FieldSelector: fieldselector})
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	// Sort and filter the inventory list using DefaultList.
	results := query.DefaultList(inventoryList.Items, queryParam, func(left, right kkcorev1.Inventory, sortBy string) bool {
		leftMeta, err := meta.Accessor(left)
		if err != nil {
			return false
		}
		rightMeta, err := meta.Accessor(right)
		if err != nil {
			return false
		}

		return query.DefaultObjectMetaCompare(leftMeta, rightMeta, sortBy)
	}, func(o kkcorev1.Inventory, filter query.Filter) bool {
		// Skip fieldselector filter.
		if filter.Field == query.ParameterFieldSelector {
			return true
		}
		objectMeta, err := meta.Accessor(o)
		if err != nil {
			return false
		}

		return query.DefaultObjectMetaFilter(objectMeta, filter)
	})

	_ = response.WriteEntity(results)
}

// Info retrieves a specific inventory resource.
// It returns the inventory with the specified name in the given namespace.
func (h *InventoryHandler) Info(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("inventory")

	inventory := &kkcorev1.Inventory{}

	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, inventory)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(inventory)
}

// ListHosts lists all hosts in an inventory with their details.
// It includes information about SSH configuration, IP addresses, and group membership.
func (h *InventoryHandler) ListHosts(request *restful.Request, response *restful.Response) {
	// Parse query parameters from the request.
	queryParam := query.ParseQueryParameter(request)
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("inventory")

	// Retrieve the inventory object from the cluster.
	inventory := &kkcorev1.Inventory{}
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, inventory)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	// Get host-check playbook if annotation exists.
	playbook := &kkcorev1.Playbook{}
	if playbookName, ok := inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation]; ok {
		if err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Name: playbookName, Namespace: inventory.Namespace}, playbook); err != nil {
			klog.Warningf("cannot found host-check playbook for inventory %q", ctrlclient.ObjectKeyFromObject(inventory))
		}
	}

	// buildHostItem constructs an InventoryHostTable from the hostname and raw extension.
	buildHostItem := func(hostname string, raw runtime.RawExtension) api.InventoryHostTable {
		// Convert the raw extension to a map of variables.
		vars := variable.Extension2Variables(raw)
		// Extract relevant fields from the variables.
		internalIPV4, _ := variable.StringVar(nil, vars, _const.VariableIPv4)
		internalIPV6, _ := variable.StringVar(nil, vars, _const.VariableIPv6)
		sshHost, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorHost)
		sshPort, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPort)
		sshUser, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorUser)
		sshPassword, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPassword)
		sshPrivateKeyContent, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPrivateKeyContent)

		// Remove sensitive or redundant variables from the vars map.
		delete(vars, _const.VariableIPv4)
		delete(vars, _const.VariableIPv6)
		delete(vars, _const.VariableConnector)

		// Return the constructed InventoryHostTable.
		return api.InventoryHostTable{
			Name:                 hostname,
			InternalIPV4:         internalIPV4,
			InternalIPV6:         internalIPV6,
			SSHHost:              sshHost,
			SSHPort:              sshPort,
			SSHUser:              sshUser,
			SSHPassword:          sshPassword,
			SSHPrivateKeyContent: sshPrivateKeyContent,
			Vars:                 vars,
			Groups:               []api.InventoryHostGroups{},
		}
	}

	// Convert inventory groups for host membership lookup.
	groups := variable.ConvertGroup(*inventory)

	// getGroupIndex checks if a host is in a group and returns its index.
	getGroupIndex := func(groupName, hostName string) int {
		for i, h := range inventory.Spec.Groups[groupName].Hosts {
			if h == hostName {
				return i
			}
		}
		return -1
	}

	// fillGroups adds group names to the InventoryHostTable item if the host is a member.
	fillGroups := func(item *api.InventoryHostTable) {
		for groupName, hosts := range groups {
			// Skip special groups.
			if groupName == _const.VariableGroupsAll || groupName == _const.VariableUnGrouped || groupName == "k8s_cluster" {
				continue
			}
			// If the host is in the group, add the group info to the item.
			if slices.Contains(hosts, item.Name) {
				g := api.InventoryHostGroups{
					Role:  groupName,
					Index: getGroupIndex(groupName, item.Name),
				}
				item.Groups = append(item.Groups, g)
			}
		}
	}

	// fillByPlaybook populates status and architecture info for the host from task results.
	fillByPlaybook := func(playbook kkcorev1.Playbook, item *api.InventoryHostTable) {
		// Set status and architecture based on playbook phase and result.
		switch playbook.Status.Phase {
		case kkcorev1.PlaybookPhaseFailed:
			item.Status = api.ResultFailed
		case kkcorev1.PlaybookPhaseSucceeded:
			item.Status = api.ResultFailed
			// Extract architecture info from playbook result.
			results := variable.Extension2Variables(playbook.Status.Result)
			if arch, ok := results[item.Name].(string); ok && arch != "" {
				item.Arch = arch
				item.Status = api.ResultSucceed
			}
		}
	}

	// less is a comparison function for sorting InventoryHostTable items by a given field.
	less := func(left, right api.InventoryHostTable, sortBy string) bool {
		// Compare fields for sorting.
		leftVal := query.GetFieldByJSONTag(reflect.ValueOf(left), sortBy)
		rightVal := query.GetFieldByJSONTag(reflect.ValueOf(right), sortBy)
		switch leftVal.Kind() {
		case reflect.String:
			return leftVal.String() > rightVal.String()
		case reflect.Int, reflect.Int64:
			return leftVal.Int() > rightVal.Int()
		default:
			return left.Name > right.Name
		}
	}

	// filter is a function to filter InventoryHostTable items based on query filters.
	filter := func(o api.InventoryHostTable, f query.Filter) bool {
		// Filter by string fields, otherwise always true.
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return strings.Contains(val.String(), string(f.Value))
		default:
			return true
		}
	}

	// Build the host table for the inventory.
	hostTable := make([]api.InventoryHostTable, 0, len(inventory.Spec.Hosts))
	for hostname, raw := range inventory.Spec.Hosts {
		// Build the host item from raw data.
		item := buildHostItem(hostname, raw)
		// Fill in group membership.
		fillGroups(&item)
		// Fill in playbook status and architecture.
		fillByPlaybook(*playbook, &item)
		// Add the item to the host table.
		hostTable = append(hostTable, item)
	}

	// Sort and filter the host table, then write the result.
	results := query.DefaultList(hostTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}
