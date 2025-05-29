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
	"time"

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

// newHandler creates a new handler instance with the given workdir, client and config
// workdir: Base directory for storing work files
// client: Kubernetes client for API operations
// config: Kubernetes REST client configuration
func newHandler(workdir string, client ctrlclient.Client, config *rest.Config) *handler {
	return &handler{workdir: workdir, client: client, restconfig: config}
}

// handler implements HTTP handlers for managing inventories and playbooks
// It provides methods for CRUD operations on inventories and playbooks
type handler struct {
	workdir    string            // Base directory for storing work files
	restconfig *rest.Config      // Kubernetes REST client configuration
	client     ctrlclient.Client // Kubernetes client for API operations
}

// createInventory creates a new inventory resource
// It reads the inventory from the request body and creates it in the Kubernetes cluster
func (h handler) createInventory(request *restful.Request, response *restful.Response) {
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
func (h handler) patchInventory(request *restful.Request, response *restful.Response) {
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
func (h handler) listInventories(request *restful.Request, response *restful.Response) {
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

// getInventory retrieves a specific inventory resource
// It returns the inventory with the specified name in the given namespace
func (h handler) inventoryInfo(request *restful.Request, response *restful.Response) {
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
func (h handler) listInventoryHosts(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("inventory")

	inventory := &kkcorev1.Inventory{}
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, inventory)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	taskList := &kkcorev1alpha1.TaskList{}
	_ = h.client.List(request.Request.Context(), taskList, ctrlclient.InNamespace(namespace), ctrlclient.MatchingFields{
		"playbook.name": inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation],
	})

	buildHostItem := func(hostname string, raw runtime.RawExtension) api.InventoryHostTable {
		vars := variable.Extension2Variables(raw)
		internalIPV4, _ := variable.StringVar(nil, vars, _const.VariableIPv4)
		internalIPV6, _ := variable.StringVar(nil, vars, _const.VariableIPv6)
		sshHost, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorHost)
		sshPort, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPort)
		sshUser, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorUser)
		sshPassword, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPassword)
		sshPrivateKey, _ := variable.StringVar(nil, vars, _const.VariableConnector, _const.VariableConnectorPrivateKey)

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

	groups := variable.ConvertGroup(*inventory)
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

	less := func(left, right api.InventoryHostTable, sortBy query.Field) bool {
		leftVal := left.Name
		if val := reflect.ValueOf(left).FieldByName(string(sortBy)); val.Kind() == reflect.String {
			leftVal = val.String()
		}
		rightVal := right.Name
		if val := reflect.ValueOf(right).FieldByName(string(sortBy)); val.Kind() == reflect.String {
			rightVal = val.String()
		}
		return leftVal > rightVal
	}

	filter := func(o api.InventoryHostTable, f query.Filter) bool {
		if f.Field == query.ParameterFieldSelector {
			return true
		}
		objectMeta, err := meta.Accessor(o)
		if err != nil {
			return false
		}
		return query.DefaultObjectMetaFilter(objectMeta, f)
	}

	hostTable := make([]api.InventoryHostTable, 0)
	for hostname, raw := range inventory.Spec.Hosts {
		item := buildHostItem(hostname, raw)
		fillGroups(&item)
		fillTaskInfo(&item)
		hostTable = append(hostTable, item)
	}

	results := query.DefaultList(hostTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}

// createPlaybook handles the creation of a new playbook resource
func (h handler) createPlaybook(request *restful.Request, response *restful.Response) {
	playbook := &kkcorev1.Playbook{}
	if err := request.ReadEntity(playbook); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	// set workdir to playbook
	if err := unstructured.SetNestedField(playbook.Spec.Config.Value(), h.workdir, _const.Workdir); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if err := h.client.Create(request.Request.Context(), playbook); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	go func() {
		// create playbook log
		filename := filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.Group, kkcorev1.SchemeGroupVersion.Version, "playbooks", playbook.Namespace, playbook.Name, playbook.Name+".log")
		if _, err := os.Stat(filepath.Dir(filename)); err != nil {
			if !os.IsNotExist(err) {
				api.HandleBadRequest(response, request, err)
				return
			}
			// if dir is not exist, create it.
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				api.HandleBadRequest(response, request, err)
				return
			}
		}
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			klog.ErrorS(err, "failed to open file", "file", filename)
			return
		}
		defer file.Close()

		ctx := context.TODO()
		if err := executor.NewPlaybookExecutor(ctx, h.client, playbook, file).Exec(ctx); err != nil {
			klog.ErrorS(err, "failed to exec playbook", "playbook", playbook.Name)
		}
	}()

	// for web ui. it not run in kubernetes. executor playbook right now
	_ = response.WriteEntity(playbook)
}

// listPlaybooks handles listing playbook resources with filtering and pagination
func (h handler) listPlaybooks(request *restful.Request, response *restful.Response) {
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

	playbookList := &kkcorev1.PlaybookList{}
	err := h.client.List(request.Request.Context(), playbookList, &ctrlclient.ListOptions{Namespace: namespace, LabelSelector: queryParam.Selector(), FieldSelector: fieldselector})
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

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

// playbookInfo handles retrieving a single playbook or watching for changes
func (h handler) playbookInfo(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("playbook")
	watch := request.QueryParameter("watch")

	playbook := &kkcorev1.Playbook{}

	if watch == "true" {
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
		for event := range watchInterface.ResultChan() {
			if err := encoder.Encode(event.Object); err != nil {
				break
			}
			flusher.Flush()
		}
		return
	}

	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, playbook)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(playbook)
}

// logPlaybook handles streaming the log file for a playbook
func (h handler) logPlaybook(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("playbook")

	playbook := &kkcorev1.Playbook{}
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, playbook)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

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

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
		fmt.Fprint(writer, line)
		flusher.Flush()
	}
}
