package v1

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"github.com/kubesphere/kubekey/v4/pkg/apiserver/runtime"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/handler"
	"k8s.io/client-go/rest"
	"net/http"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func AddToContainer(c *restful.Container, client ctrlclient.Client, cfg *rest.Config, workDir string) error {
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}
	h := NewHandler(client, cfg, workDir)
	webService := new(restful.WebService)
	webService.Path(strings.TrimRight(api.ResourcesAPIPath, "/")).
		Produces(mimePatch...)

	// only used for pre check host ,root path not needed
	resourceHandler := handler.NewResourceHandler("", workDir, client)

	webService.Route(webService.POST("/ip").To(resourceHandler.PreCheckHost).
		Doc("pre check host ssh connect information").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.ResourceTag}).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[api.IPHostCheckResult]{}))

	webService.Route(webService.GET("/schema/summary").
		To(h.GetSchemaSummary).
		Returns(http.StatusOK, api.StatusOK, v1beta1.KKClusterList{}).
		Param(webService.QueryParameter("namespaces", "kk cluster name")).
		Param(webService.QueryParameter("kk-cluster-name", "kk cluster name")).
		Doc("describe kk machine list").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	c.Add(webService)
	return nil
}
