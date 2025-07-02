package web

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/executor"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
)

// NewCoreService creates and configures a new RESTful web service for managing inventories and playbooks.
// It sets up routes for CRUD operations on inventories and playbooks, including pagination, sorting, and filtering.
func NewCoreService(workdir string, client ctrlclient.Client, config *rest.Config) *restful.WebService {
	ws := new(restful.WebService)
	// the GroupVersion might be empty, we need to remove the final /
	ws.Path(strings.TrimRight(_const.APIPath+kkcorev1.SchemeGroupVersion.String(), "/")).
		Produces(restful.MIME_JSON).Consumes(
		string(types.JSONPatchType),
		string(types.MergePatchType),
		string(types.StrategicMergePatchType),
		string(types.ApplyPatchType),
		restful.MIME_JSON)

	h := newCoreHandler(workdir, client, config)

	// Inventory management routes
	ws.Route(ws.POST("/inventories").To(h.createInventory).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("create a inventory.").
		Reads(kkcorev1.Inventory{}).
		Returns(http.StatusOK, _const.StatusOK, kkcorev1.Inventory{}))

	ws.Route(ws.PATCH("/namespaces/{namespace}/inventories/{inventory}").To(h.patchInventory).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("patch a inventory.").
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.PathParameter("inventory", "the name of the inventory")).
		Reads(kkcorev1.Inventory{}).
		Returns(http.StatusOK, _const.StatusOK, kkcorev1.Inventory{}))

	ws.Route(ws.GET("/inventories").To(h.listInventories).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("list all inventories.").
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[kkcorev1.Inventory]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/inventories").To(h.listInventories).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("list all inventories in a namespace.").
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[kkcorev1.Inventory]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/inventories/{inventory}").To(h.inventoryInfo).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("get a inventory in a namespace.").
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.PathParameter("inventory", "the name of the inventory")).
		Returns(http.StatusOK, _const.StatusOK, kkcorev1.Inventory{}))

	ws.Route(ws.GET("/namespaces/{namespace}/inventories/{inventory}/hosts").To(h.listInventoryHosts).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("list all hosts in a inventory.").
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.PathParameter("inventory", "the name of the inventory")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[api.InventoryHostTable]{}))

	// Playbook management routes
	ws.Route(ws.POST("/playbooks").To(h.createPlaybook).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("create a playbook.").
		Reads(kkcorev1.Playbook{}).
		Returns(http.StatusOK, _const.StatusOK, kkcorev1.Playbook{}))

	ws.Route(ws.GET("/playbooks").To(h.listPlaybooks).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("list all playbooks.").
		Reads(kkcorev1.Playbook{}).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[kkcorev1.Playbook]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/playbooks").To(h.listPlaybooks).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("list all playbooks in a namespace.").
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[kkcorev1.Playbook]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/playbooks/{playbook}").To(h.playbookInfo).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("get or watch a playbook in a namespace.").
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.PathParameter("playbook", "the name of the playbook")).
		Param(ws.QueryParameter("watch", "set to true to watch this playbook")).
		Returns(http.StatusOK, _const.StatusOK, kkcorev1.Playbook{}))

	ws.Route(ws.GET("/namespaces/{namespace}/playbooks/{playbook}/log").To(h.logPlaybook).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("get a playbook execute log.").
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.PathParameter("playbook", "the name of the playbook")).
		Returns(http.StatusOK, _const.StatusOK, "text/plain"))

	ws.Route(ws.DELETE("/namespaces/{namespace}/playbooks/{playbook}").To(h.deletePlaybook).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("delete a playbook.").
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.PathParameter("playbook", "the name of the playbook")).
		Returns(http.StatusOK, _const.StatusOK, api.Result{}))

	return ws
}

// newInventoryHandler creates a new handler instance with the given workdir, client and config
// workdir: Base directory for storing work files
// client: Kubernetes client for API operations
// config: Kubernetes REST client configuration
func newCoreHandler(workdir string, client ctrlclient.Client, config *rest.Config) *coreHandler {
	// Create a new coreHandler with initialized playbookManager
	return &coreHandler{workdir: workdir, client: client, restconfig: config, playbookManager: playbookManager{manager: make(map[string]context.CancelFunc)}}
}

// playbookManager is responsible for managing playbook execution contexts and their cancellation.
// It uses a mutex to ensure thread-safe access to the manager map.
type playbookManager struct {
	sync.Mutex
	manager map[string]context.CancelFunc // Map of playbook key to its cancel function
}

// addPlaybook adds a playbook and its cancel function to the manager map.
func (m *playbookManager) addPlaybook(playbook *kkcorev1.Playbook, cancel context.CancelFunc) {
	m.Lock()
	defer m.Unlock()

	m.manager[ctrlclient.ObjectKeyFromObject(playbook).String()] = cancel
}

// deletePlaybook removes a playbook from the manager map.
func (m *playbookManager) deletePlaybook(playbook *kkcorev1.Playbook) {
	m.Lock()
	defer m.Unlock()

	delete(m.manager, ctrlclient.ObjectKeyFromObject(playbook).String())
}

// stopPlaybook cancels the context for a running playbook, if it exists.
func (m *playbookManager) stopPlaybook(playbook *kkcorev1.Playbook) {
	// Attempt to cancel the playbook's context if it exists in the manager
	if cancel, ok := m.manager[ctrlclient.ObjectKeyFromObject(playbook).String()]; ok {
		cancel()
	}
}

// coreHandler implements HTTP handlers for managing inventories and playbooks
// It provides methods for CRUD operations on inventories and playbooks
type coreHandler struct {
	workdir    string            // Base directory for storing work files
	restconfig *rest.Config      // Kubernetes REST client configuration
	client     ctrlclient.Client // Kubernetes client for API operations

	// playbookManager control to cancel playbook
	playbookManager playbookManager
}

// createInventory creates a new inventory resource
// It reads the inventory from the request body and creates it in the Kubernetes cluster
func (h *coreHandler) createInventory(request *restful.Request, response *restful.Response) {
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

// patchInventory patches an existing inventory resource
// It reads the patch data from the request body and applies it to the specified inventory
func (h *coreHandler) patchInventory(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("inventory")
	data, err := io.ReadAll(request.Request.Body)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	patchType := request.HeaderParameter("Content-Type")

	// get old inventory
	inventory := &kkcorev1.Inventory{}
	if err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, inventory); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if err := h.client.Patch(request.Request.Context(), inventory, ctrlclient.RawPatch(types.PatchType(patchType), data)); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	_ = response.WriteEntity(inventory)
}

// listInventories lists all inventory resources with optional filtering and sorting
// It supports field selectors and label selectors for filtering the results
func (h *coreHandler) listInventories(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	var fieldselector fields.Selector
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
	err := h.client.List(request.Request.Context(), inventoryList, &ctrlclient.ListOptions{Namespace: namespace, LabelSelector: queryParam.Selector(), FieldSelector: fieldselector})
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	// Sort and filter the inventory list using DefaultList
	results := query.DefaultList(inventoryList.Items, queryParam, func(left, right kkcorev1.Inventory, sortBy query.Field) bool {
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
		// skip fieldselector
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

// inventoryInfo retrieves a specific inventory resource
// It returns the inventory with the specified name in the given namespace
func (h *coreHandler) inventoryInfo(request *restful.Request, response *restful.Response) {
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

// listInventoryHosts lists all hosts in an inventory with their details
// It includes information about SSH configuration, IP addresses, and group membership
func (h *coreHandler) listInventoryHosts(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("inventory")

	inventory := &kkcorev1.Inventory{}
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, inventory)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	// Retrieve the list of tasks associated with the inventory
	taskList := &kkcorev1alpha1.TaskList{}
	_ = h.client.List(request.Request.Context(), taskList, ctrlclient.InNamespace(namespace), ctrlclient.MatchingFields{
		"playbook.name": inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation],
	})

	// buildHostItem constructs an InventoryHostTable from the hostname and raw extension
	buildHostItem := func(hostname string, raw runtime.RawExtension) api.InventoryHostTable {
		vars := variable.Extension2Variables(raw)
		internalIPV4, _ := variable.StringVar(nil, vars, _const.VariableIPv4)
		internalIPV6, _ := variable.StringVar(nil, vars, _const.VariableIPv6)
		sshHost, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorHost)
		sshPort, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPort)
		sshUser, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorUser)
		sshPassword, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPassword)
		sshPrivateKey, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPrivateKey)

		// Remove sensitive or redundant variables from the vars map
		delete(vars, _const.VariableIPv4)
		delete(vars, _const.VariableIPv6)
		delete(vars, _const.VariableConnector)

		return api.InventoryHostTable{
			Name:          hostname,
			InternalIPV4:  internalIPV4,
			InternalIPV6:  internalIPV6,
			SSHHost:       sshHost,
			SSHPort:       sshPort,
			SSHUser:       sshUser,
			SSHPassword:   sshPassword,
			SSHPrivateKey: sshPrivateKey,
			Vars:          vars,
			Groups:        []string{},
		}
	}

	// Convert inventory groups for host membership lookup
	groups := variable.ConvertGroup(*inventory)

	// fillGroups adds group names to the InventoryHostTable item if the host is a member
	fillGroups := func(item *api.InventoryHostTable) {
		for groupName, hosts := range groups {
			if groupName == _const.VariableGroupsAll || groupName == _const.VariableUnGrouped {
				continue
			}
			if slices.Contains(hosts, item.Name) {
				item.Groups = append(item.Groups, groupName)
			}
		}
	}

	// fillTaskInfo populates status and architecture info for the host from task results
	fillTaskInfo := func(item *api.InventoryHostTable) {
		for _, task := range taskList.Items {
			switch task.Name {
			case _const.Getenv(_const.TaskNameGatherFacts):
				for _, result := range task.Status.HostResults {
					if result.Host == item.Name {
						item.Status = result.Stdout
						break
					}
				}
			case _const.Getenv(_const.TaskNameGetArch):
				for _, result := range task.Status.HostResults {
					if result.Host == item.Name {
						item.Arch = result.Stdout
						break
					}
				}
			}
		}
	}

	// less is a comparison function for sorting InventoryHostTable items by a given field
	less := func(left, right api.InventoryHostTable, sortBy query.Field) bool {
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

	// filter is a function to filter InventoryHostTable items based on query filters
	filter := func(o api.InventoryHostTable, f query.Filter) bool {
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return val.String() == string(f.Value)
		default:
			return true
		}
	}

	// Build the host table for the inventory
	hostTable := make([]api.InventoryHostTable, 0)
	for hostname, raw := range inventory.Spec.Hosts {
		item := buildHostItem(hostname, raw)
		fillGroups(&item)
		fillTaskInfo(&item)
		hostTable = append(hostTable, item)
	}

	// Sort and filter the host table, then write the result
	results := query.DefaultList(hostTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}

// createPlaybook handles the creation of a new playbook resource.
// It reads the playbook from the request, sets the workdir, creates the resource, and starts execution in a goroutine.
func (h *coreHandler) createPlaybook(request *restful.Request, response *restful.Response) {
	playbook := &kkcorev1.Playbook{}
	if err := request.ReadEntity(playbook); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	// Set workdir to playbook spec config.
	if err := unstructured.SetNestedField(playbook.Spec.Config.Value(), h.workdir, _const.Workdir); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if err := h.client.Create(request.Request.Context(), playbook); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	go func() {
		// Create playbook log file and execute the playbook, writing output to the log.
		filename := filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.Group, kkcorev1.SchemeGroupVersion.Version, "playbooks", playbook.Namespace, playbook.Name, playbook.Name+".log")
		// Check if the directory for the log file exists, and create it if it does not.
		if _, err := os.Stat(filepath.Dir(filename)); err != nil {
			if !os.IsNotExist(err) {
				api.HandleBadRequest(response, request, err)
				return
			}
			// If directory does not exist, create it.
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				api.HandleBadRequest(response, request, err)
				return
			}
		}
		// Open the log file for writing.
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			klog.ErrorS(err, "failed to open file", "file", filename)
			return
		}
		defer file.Close()

		// Create a cancellable context for the playbook execution.
		ctx, cancel := context.WithCancel(context.Background())
		// Add the playbook and its cancel function to the playbookManager.
		h.playbookManager.addPlaybook(playbook, cancel)
		// Execute the playbook and write output to the log file.
		if err := executor.NewPlaybookExecutor(ctx, h.client, playbook, file).Exec(ctx); err != nil {
			klog.ErrorS(err, "failed to exec playbook", "playbook", playbook.Name)
		}
		// Remove the playbook from the playbookManager after execution.
		h.playbookManager.deletePlaybook(playbook)
	}()

	// For web UI: it does not run in Kubernetes, so execute playbook immediately.
	_ = response.WriteEntity(playbook)
}

// listPlaybooks handles listing playbook resources with filtering and pagination.
// It supports field selectors and label selectors for filtering the results.
func (h *coreHandler) listPlaybooks(request *restful.Request, response *restful.Response) {
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

	playbookList := &kkcorev1.PlaybookList{}
	// List playbooks from the Kubernetes API with the specified options.
	err := h.client.List(request.Request.Context(), playbookList, &ctrlclient.ListOptions{Namespace: namespace, LabelSelector: queryParam.Selector(), FieldSelector: fieldselector})
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	// Sort and filter the playbook list using DefaultList.
	results := query.DefaultList(playbookList.Items, queryParam, func(left, right kkcorev1.Playbook, sortBy query.Field) bool {
		leftMeta, err := meta.Accessor(left)
		if err != nil {
			return false
		}
		rightMeta, err := meta.Accessor(right)
		if err != nil {
			return false
		}

		return query.DefaultObjectMetaCompare(leftMeta, rightMeta, sortBy)
	}, func(o kkcorev1.Playbook, filter query.Filter) bool {
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

// playbookInfo handles retrieving a single playbook or watching for changes.
// If the "watch" query parameter is set to "true", it streams updates to the client.
func (h *coreHandler) playbookInfo(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("playbook")
	watch := request.QueryParameter("watch")

	playbook := &kkcorev1.Playbook{}

	if watch == "true" {
		// Watch for changes to the playbook resource and stream events as JSON.
		h.restconfig.GroupVersion = &kkcorev1.SchemeGroupVersion
		client, err := rest.RESTClientFor(h.restconfig)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		watchInterface, err := client.Get().Namespace(namespace).Resource("playbooks").Name(name).Param("watch", "true").Watch(request.Request.Context())
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		defer watchInterface.Stop()

		response.AddHeader("Content-Type", "application/json")
		flusher, ok := response.ResponseWriter.(http.Flusher)
		if !ok {
			http.Error(response.ResponseWriter, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		encoder := json.NewEncoder(response.ResponseWriter)
		// Stream each event object to the client as JSON.
		for event := range watchInterface.ResultChan() {
			if err := encoder.Encode(event.Object); err != nil {
				break
			}
			flusher.Flush()
		}
		return
	}

	// Retrieve the playbook resource by namespace and name.
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, playbook)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(playbook)
}

// logPlaybook handles streaming the log file for a playbook.
// It opens the log file and streams its contents to the client, supporting live updates.
func (h *coreHandler) logPlaybook(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("playbook")

	playbook := &kkcorev1.Playbook{}
	// Retrieve the playbook resource to get its config for log file path.
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, playbook)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	// Build the log file path for the playbook.
	filename := filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.Group, kkcorev1.SchemeGroupVersion.Version, "playbooks", playbook.Namespace, playbook.Name, playbook.Name+".log")
	file, err := os.Open(filename)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	defer file.Close()

	response.AddHeader("Content-Type", "text/plain; charset=utf-8")
	writer := response.ResponseWriter
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Stream the log file line by line, waiting for new lines if at EOF.
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// If EOF, wait for new log lines to be written.
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
		fmt.Fprint(writer, line)
		flusher.Flush()
	}
}

// deletePlaybook handles deletion of a playbook resource and its associated tasks.
// It stops the playbook execution if running, deletes the playbook, and deletes all related tasks.
func (h *coreHandler) deletePlaybook(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("playbook")

	playbook := &kkcorev1.Playbook{}
	// Retrieve the playbook resource to delete.
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, playbook)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	// Stop the playbook execution if it is running.
	h.playbookManager.stopPlaybook(playbook)
	// Delete the playbook resource.
	if err := h.client.Delete(request.Request.Context(), playbook); err != nil {
		api.HandleError(response, request, err)
		return
	}
	// delete relative filepath: variable and log
	_ = os.Remove(filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.Group, kkcorev1.SchemeGroupVersion.Version, "playbooks", playbook.Namespace, playbook.Name, playbook.Name+".log"))
	_ = os.RemoveAll(filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.Group, kkcorev1.SchemeGroupVersion.Version, "playbooks", playbook.Namespace, playbook.Name, playbook.Name))
	// Delete all tasks owned by this playbook.
	if err := h.client.DeleteAllOf(request.Request.Context(), &kkcorev1alpha1.Task{}, ctrlclient.MatchingFields{
		"playbook.name": playbook.Name, "playbook.namespace": playbook.Namespace,
	}); err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(api.SUCCESS)
}
