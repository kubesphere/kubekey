package web

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// NewSchemaService creates a new WebService that serves schema files from the specified root path.
// It sets up a route that handles GET requests to /resources/schema/{subpath} and serves files from the rootPath directory.
// The {subpath:*} parameter allows for matching any path under /resources/schema/.
func NewSchemaService(rootPath string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/resources/schema")

	ws.Route(ws.GET("/{subpath:*}").To(func(req *restful.Request, resp *restful.Response) {
		http.StripPrefix("/resources/schema/", http.FileServer(http.Dir(rootPath))).ServeHTTP(resp.ResponseWriter, req.Request)
	}).Metadata(restfulspec.KeyOpenAPITags, []string{_const.ResourceTag}))

	return ws
}
