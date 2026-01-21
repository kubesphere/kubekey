package manager

import (
	"context"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/v4/pkg/kapis/infrastructure/v1beta1"
	resourcesv1 "github.com/kubesphere/kubekey/v4/pkg/kapis/resources/v1"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"net/http"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ApiManager struct {
	port       int
	workdir    string
	schemaPath string

	// webservice container, where all webservice defines
	container *restful.Container
	server    *http.Server

	client ctrlclient.Client
	config *rest.Config
}

type ApiManagerOptions struct {
	Port       int
	Workdir    string
	SchemaPath string
	Client     ctrlclient.Client
	Config     *rest.Config
}

func NewApiManager(opts ApiManagerOptions) *ApiManager {

	return &ApiManager{
		port:       opts.Port,
		workdir:    opts.Workdir,
		schemaPath: opts.SchemaPath,
		client:     opts.Client,
		config:     opts.Config,
		server: &http.Server{
			Addr: fmt.Sprintf(":%d", opts.Port),
		},
		container: restful.NewContainer(),
	}
}

func (a *ApiManager) Run(ctx context.Context) error {

	go func() {
		<-ctx.Done()
		_ = a.server.Shutdown(ctx)
	}()

	return a.server.ListenAndServe()
}

func (a *ApiManager) PrepareRun(stopCh <-chan struct{}) error {

	a.installWebInstallerApi(stopCh)

	a.server.Handler = a.container
	return nil
}

func (a *ApiManager) installWebInstallerApi(stopCh <-chan struct{}) {
	urlruntime.Must(v1beta1.AddToContainer(a.container, a.client, a.config, a.workdir))
	urlruntime.Must(resourcesv1.AddToContainer(a.container, a.client, a.config, a.workdir, a.schemaPath))
}
