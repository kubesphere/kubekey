package handler

import (
	"io"
	"reflect"
	"slices"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/emicklei/go-restful/v3"
	jsonpatch "github.com/evanphx/json-patch"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/utils"
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
		api.HandleError(response, request, err)
		return
	}
	if err := h.client.Create(request.Request.Context(), inventory); err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(inventory)
}

// Patch updates an existing inventory resource with clear variable definitions and English comments.
func (h *InventoryHandler) Patch(request *restful.Request, response *restful.Response) {
	// Get namespace and inventory name from path parameters
	namespace := request.PathParameter("namespace")
	inventoryName := request.PathParameter("inventory")

	// Read the patch body from the request
	patchBody, err := io.ReadAll(request.Request.Body)
	if err != nil {
		api.HandleError(response, request, errors.Wrap(err, "failed to read patch body from request"))
		return
	}
	// Get the Content-Type header to determine patch type
	contentType := request.HeaderParameter("Content-Type")
	patchType := types.PatchType(contentType)

	// Get the codec for encoding/decoding Inventory objects
	codec := _const.CodecFactory.LegacyCodec(kkcorev1.SchemeGroupVersion)

	// Retrieve the old inventory object from the cluster
	oldInventory := &kkcorev1.Inventory{}
	if err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: inventoryName}, oldInventory); err != nil {
		api.HandleError(response, request, errors.Wrapf(err, "failed to get Inventory %s/%s from cluster", namespace, inventoryName))
		return
	}
	// Pre-process: ensure all groups have hosts as arrays instead of null
	// This is necessary because JSON patch operations like "add" with "-" path
	// require the target to be an array, not null
	if oldInventory.Spec.Groups != nil {
		for groupName, group := range oldInventory.Spec.Groups {
			if group.Hosts == nil {
				group.Hosts = []string{}
				oldInventory.Spec.Groups[groupName] = group
			}
		}
	}
	// Encode the old inventory object to JSON
	oldInventoryJSON, err := runtime.Encode(codec, oldInventory)
	if err != nil {
		api.HandleError(response, request, errors.Wrap(err, "failed to encode old inventory object to JSON"))
		return
	}

	// Apply the patch to the old inventory and decode the result
	applyPatchAndDecode := func(objectJSON []byte) (*kkcorev1.Inventory, error) {
		var patchedJSON []byte
		switch patchType {
		case types.JSONPatchType:
			patchObj, err := jsonpatch.DecodePatch(patchBody)
			if err != nil {
				return nil, errors.Wrap(err, "failed to decode JSON patch")
			}
			patchedJSON, err = patchObj.Apply(objectJSON)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply JSON patch to old inventory JSON")
			}
		case types.MergePatchType:
			var err error
			patchedJSON, err = jsonpatch.MergePatch(objectJSON, patchBody)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply merge patch to old inventory JSON")
			}
		default:
			return nil, errors.Errorf("unknown Content-Type header for patch: %v", patchType)
		}
		newInventory := &kkcorev1.Inventory{}
		err := runtime.DecodeInto(codec, patchedJSON, newInventory)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode patched inventory JSON")
		}
		return newInventory, nil
	}

	updatedInventory, err := applyPatchAndDecode(oldInventoryJSON)
	if err != nil {
		api.HandleError(response, request, errors.Wrap(err, "failed to apply patch and decode inventory"))
		return
	}

	// completeInventory normalizes the inventory groups:
	// - Synchronizes the "kube_control_plane" group to the "etcd" group.
	// - Removes duplicate hosts and groups within each group.
	completeInventory := func(inventory *kkcorev1.Inventory) {
		// sync kube_control_plane group to etcd group
		inventory.Spec.Groups["etcd"] = inventory.Spec.Groups["kube_control_plane"]
		for k, gv := range inventory.Spec.Groups {
			gv.Hosts = utils.RemoveDuplicatesInOrder(gv.Hosts)
			gv.Groups = utils.RemoveDuplicatesInOrder(gv.Groups)
			inventory.Spec.Groups[k] = gv
		}
	}
	completeInventory(updatedInventory)

	// Patch the inventory resource in the cluster
	if err := h.client.Patch(request.Request.Context(), updatedInventory, ctrlclient.MergeFrom(oldInventory)); err != nil {
		api.HandleError(response, request, errors.Wrapf(err, "failed to patch Inventory %s/%s in cluster", namespace, inventoryName))
		return
	}

	// Create a host-check playbook and set the workdir
	hostCheckPlaybook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "host-check-",
			Namespace:    namespace,
		},
		Spec: kkcorev1.PlaybookSpec{
			InventoryRef: &corev1.ObjectReference{
				Kind:      "Inventory",
				Namespace: namespace,
				Name:      inventoryName,
			},
			Playbook: "host_check.yaml",
		},
		Status: kkcorev1.PlaybookStatus{
			Phase: kkcorev1.PlaybookPhasePending,
		},
	}
	// Set the workdir in the playbook's config
	if err := unstructured.SetNestedField(hostCheckPlaybook.Spec.Config.Value(), h.workdir, _const.Workdir); err != nil {
		api.HandleError(response, request, errors.Wrap(err, "failed to set workdir in playbook config"))
		return
	}
	// Create the playbook resource in the cluster
	if err := h.client.Create(request.Request.Context(), hostCheckPlaybook); err != nil {
		api.HandleError(response, request, errors.Wrap(err, "failed to create host-check playbook in cluster"))
		return
	}

	// Execute the playbook asynchronously if "promise" is true (default)
	if err := playbookManager.executor(hostCheckPlaybook, h.client, query.DefaultString(request.QueryParameter("promise"), "true")); err != nil {
		api.HandleError(response, request, errors.Wrap(err, "failed to execute host-check playbook"))
		return
	}

	// Patch the inventory annotation with the host-check playbook name
	if updatedInventory.Annotations == nil {
		updatedInventory.Annotations = make(map[string]string)
	}
	updatedInventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation] = hostCheckPlaybook.Name

	patchObj := &kkcorev1.Inventory{
		ObjectMeta: metav1.ObjectMeta{
			Name:        updatedInventory.Name,
			Namespace:   updatedInventory.Namespace,
			Annotations: updatedInventory.Annotations,
		},
	}
	baseObj := &kkcorev1.Inventory{
		ObjectMeta: metav1.ObjectMeta{
			Name:      updatedInventory.Name,
			Namespace: updatedInventory.Namespace,
		},
	}
	if err := h.client.Patch(request.Request.Context(), patchObj, ctrlclient.MergeFrom(baseObj)); err != nil {
		api.HandleError(response, request, errors.Wrapf(err, "failed to patch inventory annotation for %s/%s", updatedInventory.Namespace, updatedInventory.Name))
		return
	}

	_ = response.WriteEntity(updatedInventory)
}

// List returns all inventory resources with optional filtering and sorting.
// It supports field selectors and label selectors for filtering the results.
func (h *InventoryHandler) List(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	var fieldselector fields.Selector
	// Parse field selector from query parameters if present.
	if v, ok := queryParam.Filters[query.ParameterFieldSelector]; ok {
		fs, err := fields.ParseSelector(v)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		fieldselector = fs
	}

	inventoryList := &kkcorev1.InventoryList{}
	// List inventory resources from the Kubernetes API.
	err := h.client.List(request.Request.Context(), inventoryList, &ctrlclient.ListOptions{Namespace: request.PathParameter("namespace"), LabelSelector: queryParam.Selector(), FieldSelector: fieldselector})
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
		if apierrors.IsNotFound(err) {
			_ = response.WriteEntity(api.SUCCESS.SetResult("waiting for inventory to be created"))
		} else {
			api.HandleError(response, request, err)
		}
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
		if apierrors.IsNotFound(err) {
			_ = response.WriteEntity(api.SUCCESS.SetResult("waiting for inventory to be created"))
		} else {
			api.HandleError(response, request, err)
		}
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
		case kkcorev1.PlaybookPhasePending, kkcorev1.PlaybookPhaseRunning:
			item.Status = api.ResultPending
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
			return strings.Contains(val.String(), f.Value)
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
