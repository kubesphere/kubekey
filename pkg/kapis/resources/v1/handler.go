package v1

import (
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"reflect"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

// Handler handle web-installer resource apis
type Handler struct {
	workDir  string
	rootPath string
	client   ctrlclient.Client
	config   *rest.Config
}

// NewHandler create a new Handler
func NewHandler(client ctrlclient.Client, config *rest.Config, workDir, rootPath string) *Handler {
	return &Handler{
		workDir:  workDir,
		client:   client,
		config:   config,
		rootPath: rootPath,
	}
}

// GetSchemaSummary get resources summary data
// should return cluster summary and node summary
func (h *Handler) GetSchemaSummary(req *restful.Request, resp *restful.Response) {

	namespace := req.QueryParameter("namespaces")
	kkClusterName := req.QueryParameter("kk-cluster-name")

	var err error
	var kkClusterList v1beta1.KKClusterList
	var kkMachineList v1beta1.KKMachineList

	ctx := req.Request.Context()

	if namespace != "" && kkClusterName != "" {
		// if two arg not empty then query summary of the cluster
		var kkcluster v1beta1.KKCluster
		err = h.client.Get(ctx, ctrlclient.ObjectKey{
			Name:      kkClusterName,
			Namespace: namespace,
		}, &kkcluster)
		if err != nil {
			klog.Error(err)
			api.HandleBadRequest(resp, req, err)
			return
		}
		kkClusterList.Items = append(kkClusterList.Items, kkcluster)
		err = h.client.List(ctx, &kkMachineList, ctrlclient.InNamespace(namespace), ctrlclient.MatchingLabels{
			_const.ClusterApiClusterNameLabelKey: kkClusterName,
		})
	} else {
		// else query all summary
		err = h.client.List(ctx, &kkClusterList)
		if err != nil {
			klog.Error(err)
			api.HandleBadRequest(resp, req, err)
			return
		}
		err = h.client.List(ctx, &kkMachineList)
	}

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	var result = api.ResourcesSummaryResult{
		Clusters: api.ResourcesSummaryClusterResult{},
		Nodes:    api.ResourcesSummaryNodeResult{},
	}

	for _, item := range kkClusterList.Items {
		switch item.Status.Status {
		case v1beta1.StatusNotInstall:
			result.Clusters.NotInstall++
		case v1beta1.StatusInitializing:
			result.Clusters.Initializing++
		case v1beta1.StatusRunning:
			result.Clusters.Running++
		case v1beta1.StatusUpgrading:
			result.Clusters.Upgrading++
		case v1beta1.StatusScaling:
			result.Clusters.Scaling++
		case v1beta1.StatusNodeError:
			result.Clusters.NodeError++
		case v1beta1.StatusUpgradeError:
			result.Clusters.UpgradeError++
		case v1beta1.StatusFailed:
			result.Clusters.Failed++
		}
	}

	for _, item := range kkMachineList.Items {
		switch item.Status.Status {
		case v1beta1.KKMachineStatusCreating:
			result.Nodes.Creating++
		case v1beta1.KKMachineStatusWarning:
			result.Nodes.Warning++
		case v1beta1.KKMachineStatusReady:
			result.Nodes.Ready++
		case v1beta1.KKMachineStatusRunning:
			result.Nodes.Running++
		case v1beta1.KKMachineStatusFault:
			result.Nodes.Fault++
		case v1beta1.KKMachineStatusUnschedulable:
			result.Nodes.Unschedulable++
		}
	}

	resp.WriteEntity(result)
}

// ListSchema lists all schema JSON files in the rootPath directory as a table.
// It supports filtering, sorting, and pagination via query parameters.
func (h *Handler) ListSchema(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	// Read all entries in the rootPath directory.
	entries, err := os.ReadDir(h.rootPath)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	schemaTable := make([]api.SchemaTable, 0)
	for _, entry := range entries {
		// Skip directories, non-JSON files, and special schema files.
		if entry.IsDir() ||
			!strings.HasSuffix(entry.Name(), ".json") ||
			entry.Name() == api.SchemaProductFile || entry.Name() == api.SchemaConfigFile {
			continue
		}
		// Read the JSON file.
		data, err := os.ReadFile(filepath.Join(h.rootPath, entry.Name()))
		if err != nil {
			api.HandleError(response, request, errors.Wrapf(err, "failed to read file for schema %q", entry.Name()))
			return
		}
		var schemaFile api.SchemaFile
		// Unmarshal the JSON data into a SchemaTable struct.
		if err := json.Unmarshal(data, &schemaFile); err != nil {
			api.HandleError(response, request, errors.Wrapf(err, "failed to unmarshal file for schema %q", entry.Name()))
			return
		}
		schema := api.SchemaFile2Table(schemaFile, filepath.Join(h.rootPath, api.SchemaConfigFile), entry.Name())
		schemaTable = append(schemaTable, schema)
	}
	// less is a comparison function for sorting SchemaTable items by a given field.
	less := func(left, right api.SchemaTable, sortBy string) bool {
		leftVal := query.GetFieldByJSONTag(reflect.ValueOf(left), sortBy)
		rightVal := query.GetFieldByJSONTag(reflect.ValueOf(right), sortBy)
		switch leftVal.Kind() {
		case reflect.String:
			return leftVal.String() > rightVal.String()
		case reflect.Int, reflect.Int64:
			return leftVal.Int() > rightVal.Int()
		default:
			// If the field is not a string or int, sort by Priority as a fallback.
			return left.Priority > right.Priority
		}
	}
	// filter is a function for filtering SchemaTable items by a given field and value.
	filter := func(o api.SchemaTable, f query.Filter) bool {
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return strings.Contains(val.String(), f.Value)
		case reflect.Int:
			v, err := strconv.Atoi(f.Value)
			if err != nil {
				return false
			}
			return v == int(val.Int())
		default:
			return true
		}
	}

	// Use the DefaultList function to apply filtering, sorting, and pagination.
	// The results variable contains the filtered, sorted, and paginated schemaTable.
	results := query.DefaultList(schemaTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}
