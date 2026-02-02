package v1

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"github.com/kubesphere/kubekey/v4/pkg/apiserver/runtime"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/handler"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
	"k8s.io/client-go/rest"
	"net/http"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func AddToContainer(c *restful.Container, client ctrlclient.Client, cfg *rest.Config, workDir, schemaPath string) error {
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}
	h := NewHandler(client, cfg, workDir, schemaPath)
	webService := new(restful.WebService)
	webService.Path(strings.TrimRight(api.ResourcesAPIPath, "/")).
		Produces(mimePatch...)

	// only used for pre check host ,root path not needed
	resourceHandler := handler.NewResourceHandler(schemaPath, workDir, client)

	webService.Route(webService.POST("/ip").To(resourceHandler.PreCheckHost).
		Doc("pre check host ssh connect information").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[api.IPHostCheckResult]{}))

	webService.Route(webService.GET("/ip").To(resourceHandler.ListIP).
		Doc("list available ip from ip cidr").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}).
		Param(webService.QueryParameter("cidr", "the cidr for ip").Required(true)).
		Param(webService.QueryParameter("sshPort", "the ssh port for ip").Required(false)).
		Param(webService.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(webService.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webService.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(webService.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=ip").Required(false).DefaultValue("ip")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[api.IPTable]{}))

	webService.Route(webService.GET("/schema/{subpath:*}").To(resourceHandler.SchemaInfo).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}))

	webService.Route(webService.GET("/schema/summary").
		To(h.GetSchemaSummary).
		Returns(http.StatusOK, api.StatusOK, v1beta1.KKClusterList{}).
		Param(webService.QueryParameter("namespaces", "kk cluster name")).
		Param(webService.QueryParameter("kk-cluster-name", "kk cluster name")).
		Doc("describe kk machine list").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webService.Route(webService.GET("/schema").To(h.ListSchema).
		Doc("list all schema as table").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}).
		Param(webService.QueryParameter("cluster", "The namespace where the cluster resides").Required(false).DefaultValue("default")).
		Param(webService.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(webService.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webService.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(webService.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=priority")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[api.SchemaTable]{}))

	webService.Route(webService.GET("/schema/config").To(resourceHandler.ConfigInfo).
		Doc("get user-defined configuration information").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}))

	c.Add(webService)
	return nil
}
