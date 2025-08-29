package web

import (
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/config"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/handler"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
	"github.com/kubesphere/kubekey/v4/version"
)

// NewCoreService creates and configures a new RESTful web service for managing inventories and playbooks.
// It sets up routes for CRUD operations on inventories and playbooks, including pagination, sorting, and filtering.
func NewCoreService(workdir string, client ctrlclient.Client, restconfig *rest.Config) *restful.WebService {
	ws := new(restful.WebService).
		// the GroupVersion might be empty, we need to remove the final /
		Path(strings.TrimRight(api.CoreAPIPath+kkcorev1.SchemeGroupVersion.String(), "/"))

	inventoryHandler := handler.NewInventoryHandler(workdir, restconfig, client)
	// Inventory management routes
	ws.Route(ws.POST("/inventories").To(inventoryHandler.Post).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("create a inventory.").Operation("createInventory").
		Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON).
		Reads(kkcorev1.Inventory{}).
		Returns(http.StatusOK, api.StatusOK, kkcorev1.Inventory{}))

	ws.Route(ws.PATCH("/namespaces/{namespace}/inventories/{inventory}").To(inventoryHandler.Patch).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("patch a inventory.").Operation("patchInventory").
		Consumes(string(types.JSONPatchType), string(types.MergePatchType), string(types.ApplyPatchType)).Produces(restful.MIME_JSON).
		Reads(kkcorev1.Inventory{}).
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.PathParameter("inventory", "the name of the inventory")).
		Param(ws.QueryParameter("promise", "promise to execute playbook").Required(false).DefaultValue("true")).
		Returns(http.StatusOK, api.StatusOK, kkcorev1.Inventory{}))

	ws.Route(ws.GET("/inventories").To(inventoryHandler.List).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("list all inventories.").Operation("listInventory").
		Produces(restful.MIME_JSON).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[kkcorev1.Inventory]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/inventories").To(inventoryHandler.List).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("list all inventories in a namespace.").
		Produces(restful.MIME_JSON).Operation("listInventoryInNamespace").
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[kkcorev1.Inventory]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/inventories/{inventory}").To(inventoryHandler.Info).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("get a inventory in a namespace.").Operation("getInventory").
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.PathParameter("inventory", "the name of the inventory")).
		Returns(http.StatusOK, api.StatusOK, kkcorev1.Inventory{}))

	ws.Route(ws.GET("/namespaces/{namespace}/inventories/{inventory}/hosts").To(inventoryHandler.ListHosts).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("list all hosts in a inventory.").Operation("listInventoryHosts").
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("namespace", "the namespace of the inventory")).
		Param(ws.PathParameter("inventory", "the name of the inventory")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[api.InventoryHostTable]{}))

	playbookHandler := handler.NewPlaybookHandler(workdir, restconfig, client)
	// Playbook management routes
	ws.Route(ws.POST("/playbooks").To(playbookHandler.Post).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("create a playbook.").Operation("createPlaybook").
		Param(ws.QueryParameter("promise", "promise to execute playbook").Required(false).DefaultValue("true")).
		Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON).
		Reads(kkcorev1.Playbook{}).
		Returns(http.StatusOK, api.StatusOK, kkcorev1.Playbook{}))

	ws.Route(ws.GET("/playbooks").To(playbookHandler.List).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("list all playbooks.").Operation("listPlaybook").
		Produces(restful.MIME_JSON).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[kkcorev1.Playbook]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/playbooks").To(playbookHandler.List).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("list all playbooks in a namespace.").Operation("listPlaybookInNamespace").
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[kkcorev1.Playbook]{}))

	ws.Route(ws.GET("/namespaces/{namespace}/playbooks/{playbook}").To(playbookHandler.Info).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("get or watch a playbook in a namespace.").Operation("getPlaybook").
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.PathParameter("playbook", "the name of the playbook")).
		Param(ws.QueryParameter("watch", "set to true to watch this playbook")).
		Returns(http.StatusOK, api.StatusOK, kkcorev1.Playbook{}))

	ws.Route(ws.GET("/namespaces/{namespace}/playbooks/{playbook}/log").To(playbookHandler.Log).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("get a playbook execute log.").Operation("getPlaybookLog").
		Produces("text/plain").
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.PathParameter("playbook", "the name of the playbook")).
		Returns(http.StatusOK, api.StatusOK, ""))

	ws.Route(ws.DELETE("/namespaces/{namespace}/playbooks/{playbook}").To(playbookHandler.Delete).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("delete a playbook.").Operation("deletePlaybook").
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("namespace", "the namespace of the playbook")).
		Param(ws.PathParameter("playbook", "the name of the playbook")).
		Returns(http.StatusOK, api.StatusOK, api.Result{}))

	return ws
}

// NewSchemaService creates a new WebService that serves schema files from the specified root path.
// It sets up a route that handles GET requests to /resources/schema/{subpath} and serves files from the rootPath directory.
// The {subpath:*} parameter allows for matching any path under /resources/schema/.
func NewSchemaService(rootPath string, workdir string, client ctrlclient.Client) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(strings.TrimRight(api.ResourcesAPIPath, "/")).
		Produces(restful.MIME_JSON, "text/plain")

	resourceHandler := handler.NewResourceHandler(rootPath, workdir, client)
	ws.Route(ws.GET("/ip").To(resourceHandler.ListIP).
		Doc("list available ip from ip cidr").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}).
		Param(ws.QueryParameter("cidr", "the cidr for ip").Required(true)).
		Param(ws.QueryParameter("sshPort", "the ssh port for ip").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=ip").Required(false).DefaultValue("ip")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[api.IPTable]{}))

	ws.Route(ws.GET("/schema/{subpath:*}").To(resourceHandler.SchemaInfo).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}))

	ws.Route(ws.GET("/schema").To(resourceHandler.ListSchema).
		Doc("list all schema as table").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}).
		Param(ws.QueryParameter("cluster", "The namespace where the cluster resides").Required(false).DefaultValue("default")).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=priority")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[api.SchemaTable]{}))

	ws.Route(ws.POST("/schema/config").To(resourceHandler.PostConfig).
		Doc("storing user-defined configuration information").
		Reads(struct{}{}).
		Param(ws.QueryParameter("cluster", "The namespace where the cluster resides").Required(false).DefaultValue("default")).
		Param(ws.QueryParameter("inventory", "the inventory of the playbook").Required(false).DefaultValue("default")).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}))

	ws.Route(ws.GET("/schema/config").To(resourceHandler.ConfigInfo).
		Doc("get user-defined configuration information").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}))

	return ws
}

// NewUIService creates a new WebService that serves the static web UI and handles SPA routing.
// - Serves "/" with index.html
// - Serves static assets (e.g., .js, .css, .png) from the embedded web directory
// - Forwards all other non-API paths to index.html for SPA client-side routing
func NewUIService(path string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/")

	// Create a sub-filesystem for the embedded web UI assets
	fileServer := http.FileServer(http.FS(os.DirFS(path)))

	// Serve the root path "/" with index.html
	ws.Route(ws.GET("").To(func(req *restful.Request, resp *restful.Response) {
		data, err := fs.ReadFile(os.DirFS(path), "index.html")
		if err != nil {
			_ = resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.AddHeader("Content-Type", "text/html")
		_, _ = resp.Write(data)
	}))

	// Serve all subpaths:
	// - If the path matches an API prefix, return 404 to let other WebServices handle it
	// - If the path looks like a static asset (contains a dot), serve the file
	// - Otherwise, serve index.html for SPA routing
	ws.Route(ws.GET("/{subpath:*}").To(func(req *restful.Request, resp *restful.Response) {
		requestedPath := req.PathParameter("subpath")

		// If the path matches any API route, return 404 so other WebServices can handle it
		if strings.HasPrefix(requestedPath, strings.TrimLeft(api.CoreAPIPath, "/")) ||
			strings.HasPrefix(requestedPath, strings.TrimLeft(api.SwaggerAPIPath, "/")) ||
			strings.HasPrefix(requestedPath, strings.TrimLeft(api.ResourcesAPIPath, "/")) {
			_ = resp.WriteError(http.StatusNotFound, errors.New("not found"))
			return
		}

		// If the path looks like a static asset (e.g., .js, .css, .ico, .png, etc.), serve it
		if strings.Contains(requestedPath, ".") {
			fileServer.ServeHTTP(resp.ResponseWriter, req.Request)
			return
		}

		// For all other paths, serve index.html (SPA client-side routing)
		data, err := fs.ReadFile(os.DirFS(path), "index.html")
		if err != nil {
			_ = resp.WriteError(http.StatusInternalServerError, err)
			return
		}
		resp.AddHeader("Content-Type", "text/html")
		_, _ = resp.Write(data)
	}))

	return ws
}

// NewSwaggerUIService creates a new WebService that serves the Swagger UI interface
// It mounts the embedded swagger-ui files and handles requests to display the API documentation
func NewSwaggerUIService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(strings.TrimRight(api.SwaggerAPIPath, "/"))

	subFS, err := fs.Sub(config.Swagger, "swagger-ui")
	if err != nil {
		panic(err)
	}

	fileServer := http.StripPrefix("/swagger-ui/", http.FileServer(http.FS(subFS)))

	ws.Route(ws.GET("").To(func(req *restful.Request, resp *restful.Response) {
		data, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			_ = resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.AddHeader("Content-Type", "text/html")
		_, _ = resp.Write(data)
	}).Metadata(restfulspec.KeyOpenAPITags, []string{api.OpenAPITag}))
	ws.Route(ws.GET("/{subpath:*}").To(func(req *restful.Request, resp *restful.Response) {
		fileServer.ServeHTTP(resp.ResponseWriter, req.Request)
	}).Metadata(restfulspec.KeyOpenAPITags, []string{api.OpenAPITag}))

	return ws
}

// NewAPIService creates a new WebService that serves the OpenAPI/Swagger specification
// It takes a list of WebServices and generates the API documentation
func NewAPIService(webservice []*restful.WebService) *restful.WebService {
	restconfig := restfulspec.Config{
		WebServices: webservice, // you control what services are visible
		APIPath:     "/apidocs.json",
		PostBuildSwaggerObjectHandler: func(swo *spec.Swagger) {
			swo.Info = &spec.Info{
				InfoProps: spec.InfoProps{
					Title:       "KubeKey Web API",
					Description: "KubeKey Web OpenAPI",
					Version:     version.Get().String(),
					Contact: &spec.ContactInfo{
						ContactInfoProps: spec.ContactInfoProps{
							Name: "KubeKey",
							URL:  "https://github.com/kubesphere/kubekey",
						},
					},
				},
			}
		}}
	for _, ws := range webservice {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	return restfulspec.NewOpenAPIService(restconfig)
}
