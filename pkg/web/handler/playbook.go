package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/executor"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
)

// PlaybookHandler handles HTTP requests for playbook resources.
type PlaybookHandler struct {
	workdir    string            // Base directory for storing work files
	restconfig *rest.Config      // Kubernetes REST client configuration
	client     ctrlclient.Client // Kubernetes client for API operations
}

// NewPlaybookHandler creates a new PlaybookHandler with the given workdir, restconfig, and client.
func NewPlaybookHandler(workdir string, restconfig *rest.Config, client ctrlclient.Client) *PlaybookHandler {
	return &PlaybookHandler{workdir: workdir, restconfig: restconfig, client: client}
}

// Post handles the creation of a new playbook resource.
// It reads the playbook from the request, checks schema label constraints, sets the workdir, creates the resource, and starts execution in a goroutine.
func (h *PlaybookHandler) Post(request *restful.Request, response *restful.Response) {
	playbook := &kkcorev1.Playbook{}
	// Read the playbook entity from the request body
	if err := request.ReadEntity(playbook); err != nil {
		api.HandleError(response, request, err)
		return
	}

	// Check for schema label: only one allowed, must not be empty, and must be unique among playbooks
	hasSchemaLabel := false
	for labelKey, labelValue := range playbook.Labels {
		// Only consider labels with the schema label suffix
		if !strings.HasSuffix(labelKey, api.SchemaLabelSubfix) {
			continue
		}
		// If a schema label was already found, this is a conflict
		if hasSchemaLabel {
			api.HandleConflict(response, request, errors.New("a playbook can only have one schema label. Please ensure only one schema label is set"))
			return
		}
		// The schema label value must not be empty
		if labelValue == "" {
			api.HandleConflict(response, request, errors.New("the schema label value must not be empty. Please provide a valid schema label value"))
			return
		}
		hasSchemaLabel = true
		// Check if there is already a playbook with the same schema label
		playbookList := &kkcorev1.PlaybookList{}
		if err := h.client.List(request.Request.Context(), playbookList, ctrlclient.MatchingLabels{
			labelKey: labelValue,
		}); err != nil {
			api.HandleError(response, request, err)
			return
		}
		// If any playbook with the same schema label exists, this is a conflict
		if len(playbookList.Items) > 0 {
			api.HandleConflict(response, request, errors.New("a playbook with the same schema label already exists. Please use a different schema label or remove the existing playbook"))
			return
		}
	}

	// Set the workdir in the playbook's spec config
	if err := unstructured.SetNestedField(playbook.Spec.Config.Value(), h.workdir, _const.Workdir); err != nil {
		api.HandleError(response, request, err)
		return
	}
	playbook.Status.Phase = kkcorev1.PlaybookPhasePending
	// Create the playbook resource in Kubernetes
	if err := h.client.Create(context.TODO(), playbook); err != nil {
		api.HandleError(response, request, err)
		return
	}
	// Start playbook execution in a separate goroutine
	if err := executor.PlaybookManager.Executor(playbook, h.client, query.DefaultString(request.QueryParameter("promise"), "true")); err != nil {
		api.HandleError(response, request, errors.Wrap(err, "failed to execute playbook"))
		return
	}
	// For web UI: it does not run in Kubernetes, so execute playbook immediately.
	_ = response.WriteEntity(playbook)
}

// List handles listing playbook resources with filtering and pagination.
// It supports field selectors and label selectors for filtering the results.
func (h *PlaybookHandler) List(request *restful.Request, response *restful.Response) {
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
	playbookList := &kkcorev1.PlaybookList{}
	// List playbooks from the Kubernetes API with the specified options.
	err := h.client.List(request.Request.Context(), playbookList, &ctrlclient.ListOptions{Namespace: request.PathParameter("namespace"), LabelSelector: queryParam.Selector(), FieldSelector: fieldselector})
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	// Sort and filter the playbook list using DefaultList.
	results := query.DefaultList(playbookList.Items, queryParam, func(left, right kkcorev1.Playbook, sortBy string) bool {
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

// Info handles retrieving a single playbook or watching for changes.
// If the "watch" query parameter is set to "true", it streams updates to the client.
func (h *PlaybookHandler) Info(request *restful.Request, response *restful.Response) {
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

// Log handles streaming the log file for a playbook.
// It opens the log file and streams its contents to the client, supporting live updates.
func (h *PlaybookHandler) Log(request *restful.Request, response *restful.Response) {
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

// Delete handles deletion of a playbook resource and its associated tasks.
// It stops the playbook execution if running, deletes the playbook, removes related files, and deletes all related tasks.
func (h *PlaybookHandler) Delete(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("playbook")

	playbook := &kkcorev1.Playbook{}
	// Retrieve the playbook resource to delete.
	err := h.client.Get(request.Request.Context(), ctrlclient.ObjectKey{Namespace: namespace, Name: name}, playbook)
	if err != nil {
		if apierrors.IsNotFound(err) {
			_ = response.WriteEntity(api.SUCCESS.SetResult("playbook has deleted"))
		} else {
			api.HandleError(response, request, err)
		}
		return
	}
	// Stop the playbook execution if it is running.
	executor.PlaybookManager.StopPlaybook(playbook)
	// Delete the playbook resource.
	if err := h.client.Delete(request.Request.Context(), playbook); err != nil {
		if apierrors.IsNotFound(err) {
			_ = response.WriteEntity(api.SUCCESS.SetResult("playbook has deleted"))
		} else {
			api.HandleError(response, request, err)
		}
		return
	}
	// Delete related log file and directory.
	_ = os.Remove(filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.Group, kkcorev1.SchemeGroupVersion.Version, "playbooks", playbook.Namespace, playbook.Name, playbook.Name+".log"))
	_ = os.RemoveAll(filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.Group, kkcorev1.SchemeGroupVersion.Version, "playbooks", playbook.Namespace, playbook.Name))
	// Delete all tasks owned by this playbook.
	if err := h.client.DeleteAllOf(request.Request.Context(), &kkcorev1alpha1.Task{}, ctrlclient.InNamespace(playbook.Namespace), ctrlclient.MatchingFields{
		"playbook.name": playbook.Name,
		"playbook.uid":  string(playbook.UID),
	}); err != nil {
		if apierrors.IsNotFound(err) {
			_ = response.WriteEntity(api.SUCCESS.SetResult("playbook has deleted"))
		} else {
			api.HandleError(response, request, err)
		}
		return
	}

	_ = response.WriteEntity(api.SUCCESS)
}
