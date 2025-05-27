package web

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

// NewSchemaService creates a new WebService that serves schema files from the specified root path.
// It sets up a route that handles GET requests to /schema/{subpath} and serves files from the rootPath directory.
// The {subpath:*} parameter allows for matching any path under /schema/.
func NewSchemaService(rootPath string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/schema")

	ws.Route(ws.GET("/{subpath:*}").To(func(req *restful.Request, resp *restful.Response) {
		http.StripPrefix("/schema/", http.FileServer(http.Dir(rootPath))).ServeHTTP(resp.ResponseWriter, req.Request)
	}))

	return ws
}
