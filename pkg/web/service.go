package web

import (
	"context"
	"net/http"
	"strings"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
)

// NewWebService creates and configures a new RESTful web service for managing inventories and playbooks.
// It sets up routes for CRUD operations on inventories and playbooks, including pagination, sorting, and filtering.
// Parameters:
//   - ctx: Context for the web service
//   - workdir: Working directory for file operations
//   - client: Kubernetes controller client
//   - config: REST configuration
//
// Returns a configured WebService instance
func NewWebService(ctx context.Context, workdir string, client ctrlclient.Client, config *rest.Config) *restful.WebService {
	ws := new(restful.WebService)
	// the GroupVersion might be empty, we need to remove the final /
	ws.Path(strings.TrimRight(_const.APIPath+kkcorev1.SchemeGroupVersion.String(), "/")).
		Produces(restful.MIME_JSON).Consumes(
		string(types.JSONPatchType),
		string(types.MergePatchType),
		string(types.StrategicMergePatchType),
		string(types.ApplyPatchType),
		restful.MIME_JSON)

	h := newHandler(workdir, client, config)

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
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[kkcorev1.Inventory]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/inventories").To(h.listInventories).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("list all inventories in a namespace.").
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
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
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
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
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[kkcorev1.Playbook]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/playbooks").To(h.listInventories).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.KubeKeyTag}).
		Doc("list all playbooks in a namespace.").
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
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
		Doc("watch a playbook in a namespace.").
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.PathParameter("playbook", "the name of the playbook")).
		Returns(http.StatusOK, _const.StatusOK, "text/plain"))

	return ws
}
