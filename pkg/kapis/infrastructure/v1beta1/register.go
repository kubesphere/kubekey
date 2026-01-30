package v1beta1

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/kubesphere/kubekey/v4/pkg/apiserver/errors"
	"github.com/kubesphere/kubekey/v4/pkg/apiserver/runtime"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/handler"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"net/http"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	GroupName = "kubekey.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(c *restful.Container, client ctrlclient.Client, cfg *rest.Config, workDir string) error {
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}
	h := NewHandler(client, cfg, workDir)
	playbookHandler := handler.NewPlaybookHandler(workDir, cfg, client)
	webservice := runtime.NewWebService(GroupVersion)
	webservice.Produces(mimePatch...)

	webservice.Route(webservice.POST("/kkcluster").
		Reads(v1beta1.KKCluster{}).
		To(h.CreateKKCluster).
		Returns(http.StatusOK, api.StatusOK, v1beta1.KKCluster{}).
		Doc("Create kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webservice.Route(webservice.DELETE("namespaces/{namespace}/kkcluster/{kk-cluster-name}").
		To(h.DeleteKKCluster).
		Param(webservice.PathParameter("namespace", "kk cluster namespace")).
		Param(webservice.PathParameter("kk-cluster-name", "kk cluster name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Doc("delete kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webservice.Route(webservice.POST("/kkmachine").
		Reads(v1beta1.KKMachine{}).
		To(h.CreateKKMachine).
		Returns(http.StatusOK, api.StatusOK, v1beta1.KKMachine{}).
		Doc("Create kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webservice.Route(webservice.DELETE("namespaces/{namespace}/kkmachine/{kk-machine-name}").
		To(h.DeleteKKMachine).
		Param(webservice.PathParameter("namespace", "kk machine namespace")).
		Param(webservice.PathParameter("kk-machine-name", "kk machine name")).
		Param(webservice.QueryParameter("kk-cluster-name", "kk cluster name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Doc("delete kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webservice.Route(webservice.PATCH("namespaces/{namespace}/kkmachine/{kk-machine-name}").
		To(h.UpdateKKMachine).
		Param(webservice.PathParameter("namespace", "kk machine namespace")).
		Param(webservice.PathParameter("kk-machine-name", "kk machine name")).
		Param(webservice.QueryParameter("kk-cluster-name", "kk cluster name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("patch a kk machine.").Operation("patchKKMachine").
		Consumes(string(types.JSONPatchType), string(types.MergePatchType), string(types.ApplyPatchType)).Produces(restful.MIME_JSON).
		Reads(v1beta1.KKMachine{}).
		Returns(http.StatusOK, api.StatusOK, v1beta1.KKMachine{}))

	webservice.Route(webservice.PATCH("/namespaces/{namespace}/kkcluster/{kk-cluster-name}").
		To(h.CreateConfig).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("patch a kk machine template.").Operation("patchKKMachineTemplate").
		Consumes(string(types.JSONPatchType), string(types.MergePatchType), string(types.ApplyPatchType)).Produces(restful.MIME_JSON).
		Reads(v1beta1.KKCluster{}).
		Param(webservice.PathParameter("namespace", "the namespace of the inventory")).
		Param(webservice.PathParameter("kk-cluster-name", "the name of the inventory")).
		Returns(http.StatusOK, api.StatusOK, v1beta1.KKCluster{}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/kkcluster/{kk-cluster-name}").
		To(h.GetKKCluster).
		Param(webservice.PathParameter("namespace", "kk cluster namespace")).
		Param(webservice.PathParameter("kk-cluster-name", "kk cluster  name")).
		Returns(http.StatusOK, api.StatusOK, v1beta1.KKCluster{}).
		Doc("describe kk cluster with config").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/kkmachine/{kk-cluster-name}").
		To(h.GetKKMachineList).
		Param(webservice.PathParameter("namespace", "kk cluster namespace")).
		Param(webservice.PathParameter("kk-cluster-name", "kk cluster  name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[v1beta1.KKMachine]{}).
		Doc("describe kk machine list").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webservice.Route(webservice.GET("/kkclusters").
		To(h.GetKKClusterList).
		Returns(http.StatusOK, api.StatusOK, api.ListResult[v1beta1.KKCluster]{}).
		Doc("describe kk machine list").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.CapkkTag}))

	webservice.Route(webservice.POST("/playbooks").To(playbookHandler.Post).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("create a playbook.").Operation("createPlaybook").
		Param(webservice.QueryParameter("promise", "promise to execute playbook").Required(false).DefaultValue("true")).
		Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON).
		Reads(kkcorev1.Playbook{}).
		Returns(http.StatusOK, api.StatusOK, kkcorev1.Playbook{}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/playbooks/{playbook}/log").To(playbookHandler.Log).
		Metadata(restfulspec.KeyOpenAPITags, []string{api.KubeKeyTag}).
		Doc("get a playbook execute log.").Operation("getPlaybookLog").
		Produces("text/plain").
		Param(webservice.PathParameter("namespace", "the namespace of the playbook")).
		Param(webservice.PathParameter("playbook", "the name of the playbook")).
		Returns(http.StatusOK, api.StatusOK, ""))

	c.Add(webservice)
	return nil
}
