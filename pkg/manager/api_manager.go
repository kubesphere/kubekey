package manager

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/controllers/apiserver"
	"github.com/kubesphere/kubekey/v4/pkg/kapis/infrastructure/v1beta1"
	resourcesv1 "github.com/kubesphere/kubekey/v4/pkg/kapis/resources/v1"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
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

	Controllers []options.Controller
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
		Controllers: []options.Controller{
			&apiserver.KKMachineController{
				Workdir: opts.Workdir,
			},
			&apiserver.KKClusterController{
				Workdir: opts.Workdir,
			},
		},
	}
}

func (a *ApiManager) Run(ctx context.Context) error {
	ctrl.SetLogger(klog.NewKlogr())

	mgr, err := ctrl.NewManager(a.config, ctrl.Options{
		NewClient: func(_ *rest.Config, _ ctrlclient.Options) (ctrlclient.Client, error) {
			return a.client, nil
		},
		Scheme:                 _const.Scheme,
		HealthProbeBindAddress: ":9440",
		Metrics: server.Options{
			BindAddress: ":9441",
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create controller manager")
	}
	if err := mgr.AddHealthzCheck("default", healthz.Ping); err != nil {
		return errors.Wrap(err, "failed to add default healthcheck")
	}
	if err := mgr.AddReadyzCheck("default", healthz.Ping); err != nil {
		return errors.Wrap(err, "failed to add default readycheck")
	}

	if err := a.register(mgr); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		_ = a.server.Shutdown(ctx)
	}()

	go func() {
		if err = mgr.Start(ctx); err != nil {
			fmt.Println(err)
			ctx.Done()
		}
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

func (a *ApiManager) register(mgr ctrl.Manager) error {
	for _, c := range a.Controllers {
		if err := c.SetupWithManager(mgr, options.ControllerManagerServerOptions{}); err != nil {
			return errors.Wrapf(err, "failed to register controller %q", c.Name())
		}
	}

	return nil
}
