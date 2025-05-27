package web

import (
	"io/fs"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/config"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/version"
)

// NewSwaggerUIService creates a new WebService that serves the Swagger UI interface
// It mounts the embedded swagger-ui files and handles requests to display the API documentation
func NewSwaggerUIService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/swagger-ui")

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
	}).Metadata(restfulspec.KeyOpenAPITags, []string{_const.OpenAPITag}))
	ws.Route(ws.GET("/{subpath:*}").To(func(req *restful.Request, resp *restful.Response) {
		fileServer.ServeHTTP(resp.ResponseWriter, req.Request)
	}).Metadata(restfulspec.KeyOpenAPITags, []string{_const.OpenAPITag}))

	return ws
}

// NewAPIService creates a new WebService that serves the OpenAPI/Swagger specification
// It takes a list of WebServices and generates the API documentation
func NewAPIService(webservice []*restful.WebService) *restful.WebService {
	restconfig := restfulspec.Config{
		WebServices:                   webservice, // you control what services are visible
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	for _, ws := range webservice {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	return restfulspec.NewOpenAPIService(restconfig)
}

// enrichSwaggerObject customizes the Swagger documentation with KubeKey-specific information
// It sets the API title, description, version and contact information
func enrichSwaggerObject(swo *spec.Swagger) {
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
}
