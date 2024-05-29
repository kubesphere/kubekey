/*
Copyright 2024 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	apiendpoints "k8s.io/apiserver/pkg/endpoints"
	genericapifilters "k8s.io/apiserver/pkg/endpoints/filters"
	apihandlers "k8s.io/apiserver/pkg/endpoints/handlers"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	apirest "k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/internal"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/resources/config"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/resources/inventory"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/resources/pipeline"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/resources/task"
)

func NewConfig() (*rest.Config, error) {
	restconfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Infof("kubeconfig in empty, store resources local")
		restconfig = &rest.Config{}
	}
	restconfig.Transport, err = newProxyTransport(restconfig)
	if err != nil {
		return nil, fmt.Errorf("create proxy transport error: %w", err)
	}
	restconfig.TLSClientConfig = rest.TLSClientConfig{}
	return restconfig, nil
}

// NewProxyTransport return a new http.RoundTripper use in ctrl.client.
// when restConfig is not empty: should connect a kubernetes cluster and store some resources in there.
// such as: pipeline.kubekey.kubesphere.io/v1, inventory.kubekey.kubesphere.io/v1, config.kubekey.kubesphere.io/v1
// when restConfig is empty: store all resource in local.
//
// SPECIFICALLY: since tasks is running data, which is reentrant and large in quantity,
// they should always store in local.
func newProxyTransport(restConfig *rest.Config) (http.RoundTripper, error) {
	lt := &transport{
		authz:            authorizerfactory.NewAlwaysAllowAuthorizer(),
		handlerChainFunc: defaultHandlerChain,
	}
	if restConfig.Host != "" {
		clientFor, err := rest.HTTPClientFor(restConfig)
		if err != nil {
			return nil, err
		}
		lt.restClient = clientFor
	}

	// register kubekeyv1alpha1 resources
	kkv1alpha1 := newApiIResources(kubekeyv1alpha1.SchemeGroupVersion)
	storage, err := task.NewStorage(internal.NewFileRESTOptionsGetter(kubekeyv1alpha1.SchemeGroupVersion))
	if err != nil {
		klog.V(4).ErrorS(err, "failed to create storage")
		return nil, err
	}
	if err := kkv1alpha1.AddResource(resourceOptions{
		path:    "tasks",
		storage: storage.Task,
	}); err != nil {
		klog.V(4).ErrorS(err, "failed to add resource")
		return nil, err
	}
	if err := kkv1alpha1.AddResource(resourceOptions{
		path:    "tasks/status",
		storage: storage.TaskStatus,
	}); err != nil {
		klog.V(4).ErrorS(err, "failed to add resource")
		return nil, err
	}
	if err := lt.registerResources(kkv1alpha1); err != nil {
		klog.V(4).ErrorS(err, "failed to register resources")
	}

	// when restConfig is null. should store all resource local
	if restConfig.Host == "" {
		// register kubekeyv1 resources
		kkv1 := newApiIResources(kubekeyv1.SchemeGroupVersion)
		// add config
		configStorage, err := config.NewStorage(internal.NewFileRESTOptionsGetter(kubekeyv1.SchemeGroupVersion))
		if err != nil {
			klog.V(4).ErrorS(err, "failed to create storage")
			return nil, err
		}
		if err := kkv1.AddResource(resourceOptions{
			path:    "configs",
			storage: configStorage.Config,
		}); err != nil {
			klog.V(4).ErrorS(err, "failed to add resource")
			return nil, err
		}
		// add inventory
		inventoryStorage, err := inventory.NewStorage(internal.NewFileRESTOptionsGetter(kubekeyv1.SchemeGroupVersion))
		if err != nil {
			klog.V(4).ErrorS(err, "failed to create storage")
			return nil, err
		}
		if err := kkv1.AddResource(resourceOptions{
			path:    "inventories",
			storage: inventoryStorage.Inventory,
		}); err != nil {
			klog.V(4).ErrorS(err, "failed to add resource")
			return nil, err
		}
		// add pipeline
		pipelineStorage, err := pipeline.NewStorage(internal.NewFileRESTOptionsGetter(kubekeyv1.SchemeGroupVersion))
		if err != nil {
			klog.V(4).ErrorS(err, "failed to create storage")
			return nil, err
		}
		if err := kkv1.AddResource(resourceOptions{
			path:    "pipelines",
			storage: pipelineStorage.Pipeline,
		}); err != nil {
			klog.V(4).ErrorS(err, "failed to add resource")
			return nil, err
		}
		if err := kkv1.AddResource(resourceOptions{
			path:    "pipelines/status",
			storage: pipelineStorage.PipelineStatus,
		}); err != nil {
			klog.V(4).ErrorS(err, "failed to add resource")
			return nil, err
		}

		if err := lt.registerResources(kkv1); err != nil {
			klog.V(4).ErrorS(err, "failed to register resources")
			return nil, err
		}
	}

	return lt, nil
}

type responseWriter struct {
	*http.Response
}

func (r *responseWriter) Header() http.Header {
	return r.Response.Header
}

func (r *responseWriter) Write(bs []byte) (int, error) {
	r.Response.Body = io.NopCloser(bytes.NewBuffer(bs))
	return 0, nil
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.Response.StatusCode = statusCode
}

type transport struct {
	// use to connect remote
	restClient *http.Client

	authz authorizer.Authorizer
	// routers is a list of routers
	routers []router

	// handlerChain will be called after each request.
	handlerChainFunc func(handler http.Handler) http.Handler
}

func (l *transport) RoundTrip(request *http.Request) (*http.Response, error) {
	if l.restClient != nil && !strings.HasPrefix(request.URL.Path, "/apis/"+kubekeyv1alpha1.SchemeGroupVersion.String()) {
		return l.restClient.Transport.RoundTrip(request)
	}

	response := &http.Response{
		Proto:  "local",
		Header: make(http.Header),
	}
	// dispatch request
	handler, err := l.detectDispatcher(request)
	if err != nil {
		return response, fmt.Errorf("no router for request. url: %s, method: %s", request.URL.Path, request.Method)
	}
	// call handler
	l.handlerChainFunc(handler).ServeHTTP(&responseWriter{response}, request)
	return response, nil
}

// http://jsr311.java.net/nonav/releases/1.1/spec/spec3.html#x3-360003.7.2 (step 1)
func (l transport) detectDispatcher(request *http.Request) (http.HandlerFunc, error) {
	filtered := &sortableDispatcherCandidates{}
	for _, each := range l.routers {
		matches := each.pathExpr.Matcher.FindStringSubmatch(request.URL.Path)
		if matches != nil {
			filtered.candidates = append(filtered.candidates,
				dispatcherCandidate{each, matches[len(matches)-1], len(matches), each.pathExpr.LiteralCount, each.pathExpr.VarCount})
		}
	}
	if len(filtered.candidates) == 0 {
		return nil, fmt.Errorf("not found")
	}
	sort.Sort(sort.Reverse(filtered))

	handler, ok := filtered.candidates[0].router.handlers[request.Method]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return handler, nil
}

func (l *transport) registerResources(resources *apiResources) error {
	// register apiResources router
	l.registerRouter(http.MethodGet, resources.prefix, resources.handlerApiResources(), true)
	// register resources router
	for _, o := range resources.resourceOptions {
		// what verbs are supported by the storage, used to know what verbs we support per path
		creater, isCreater := o.storage.(apirest.Creater)
		namedCreater, isNamedCreater := o.storage.(apirest.NamedCreater)
		lister, isLister := o.storage.(apirest.Lister)
		getter, isGetter := o.storage.(apirest.Getter)
		getterWithOptions, isGetterWithOptions := o.storage.(apirest.GetterWithOptions)
		gracefulDeleter, isGracefulDeleter := o.storage.(apirest.GracefulDeleter)
		collectionDeleter, isCollectionDeleter := o.storage.(apirest.CollectionDeleter)
		updater, isUpdater := o.storage.(apirest.Updater)
		patcher, isPatcher := o.storage.(apirest.Patcher)
		watcher, isWatcher := o.storage.(apirest.Watcher)
		connecter, isConnecter := o.storage.(apirest.Connecter)
		tableProvider, isTableProvider := o.storage.(apirest.TableConvertor)
		if isLister && !isTableProvider {
			// All listers must implement TableProvider
			return fmt.Errorf("%q must implement TableConvertor", o.path)
		}
		gvAcceptor, _ := o.storage.(apirest.GroupVersionAcceptor)

		if isNamedCreater {
			isCreater = true
		}

		allowWatchList := isWatcher && isLister
		var (
			connectSubpath bool
			getSubpath     bool
		)
		if isConnecter {
			_, connectSubpath, _ = connecter.NewConnectOptions()
		}
		if isGetterWithOptions {
			_, getSubpath, _ = getterWithOptions.NewGetOptions()
		}
		resource, subresource, err := splitSubresource(o.path)
		if err != nil {
			return err
		}
		isSubresource := len(subresource) > 0
		scoper, ok := o.storage.(apirest.Scoper)
		if !ok {
			return fmt.Errorf("%q must implement scoper", o.path)
		}

		// Get the list of actions for the given scope.
		switch {
		case !scoper.NamespaceScoped():
			// do nothing. The current managed resources are all  namespace scope.
		default:
			resourcePath := "/namespaces/{namespace}/" + resource
			itemPath := resourcePath + "/{name}"
			if isSubresource {
				itemPath = itemPath + "/" + subresource
				resourcePath = itemPath
			}
			// request scope
			fqKindToRegister, err := apiendpoints.GetResourceKind(resources.gv, o.storage, _const.Scheme)
			if err != nil {
				return err
			}
			reqScope := apihandlers.RequestScope{
				Namer: apihandlers.ContextBasedNaming{
					Namer:         meta.NewAccessor(),
					ClusterScoped: false,
				},
				Serializer:      _const.Codecs,
				ParameterCodec:  _const.ParameterCodec,
				Creater:         _const.Scheme,
				Convertor:       _const.Scheme,
				Defaulter:       _const.Scheme,
				Typer:           _const.Scheme,
				UnsafeConvertor: _const.Scheme,
				Authorizer:      l.authz,

				EquivalentResourceMapper: runtime.NewEquivalentResourceRegistry(),

				// TODO: Check for the interface on storage
				TableConvertor: tableProvider,

				// TODO: This seems wrong for cross-group subresources. It makes an assumption that a subresource and its parent are in the same group version. Revisit this.
				Resource:    resources.gv.WithResource(resource),
				Subresource: subresource,
				Kind:        fqKindToRegister,

				AcceptsGroupVersionDelegate: gvAcceptor,

				HubGroupVersion: schema.GroupVersion{Group: fqKindToRegister.Group, Version: runtime.APIVersionInternal},

				MetaGroupVersion: metav1.SchemeGroupVersion,

				MaxRequestBodyBytes: 0,
			}
			var resetFields map[fieldpath.APIVersion]*fieldpath.Set
			if resetFieldsStrategy, isResetFieldsStrategy := o.storage.(apirest.ResetFieldsStrategy); isResetFieldsStrategy {
				resetFields = resetFieldsStrategy.GetResetFields()
			}
			reqScope.FieldManager, err = managedfields.NewDefaultFieldManager(
				managedfields.NewDeducedTypeConverter(),
				_const.Scheme,
				_const.Scheme,
				_const.Scheme,
				fqKindToRegister,
				reqScope.HubGroupVersion,
				subresource,
				resetFields,
			)
			if err != nil {
				return err
			}

			// LIST
			l.registerRouter(http.MethodGet, resources.prefix+resourcePath, apihandlers.ListResource(lister, watcher, &reqScope, false, resources.minRequestTimeout), isLister)
			// POST
			if isNamedCreater {
				l.registerRouter(http.MethodPost, resources.prefix+resourcePath, apihandlers.CreateNamedResource(namedCreater, &reqScope, o.admit), isCreater)
			} else {
				l.registerRouter(http.MethodPost, resources.prefix+resourcePath, apihandlers.CreateResource(creater, &reqScope, o.admit), isCreater)
			}
			// DELETECOLLECTION
			l.registerRouter(http.MethodDelete, resources.prefix+resourcePath, apihandlers.DeleteCollection(collectionDeleter, isCollectionDeleter, &reqScope, o.admit), isCollectionDeleter)
			// DEPRECATED in 1.11 WATCHLIST
			l.registerRouter(http.MethodGet, resources.prefix+"/watch"+resourcePath, apihandlers.ListResource(lister, watcher, &reqScope, true, resources.minRequestTimeout), allowWatchList)
			// GET
			if isGetterWithOptions {
				l.registerRouter(http.MethodGet, resources.prefix+itemPath, apihandlers.GetResourceWithOptions(getterWithOptions, &reqScope, isSubresource), isGetter)
				l.registerRouter(http.MethodGet, resources.prefix+itemPath+"/{path:*}", apihandlers.GetResourceWithOptions(getterWithOptions, &reqScope, isSubresource), isGetter && getSubpath)
			} else {
				l.registerRouter(http.MethodGet, resources.prefix+itemPath, apihandlers.GetResource(getter, &reqScope), isGetter)
				l.registerRouter(http.MethodGet, resources.prefix+itemPath+"/{path:*}", apihandlers.GetResource(getter, &reqScope), isGetter && getSubpath)
			}
			// PUT
			l.registerRouter(http.MethodPut, resources.prefix+itemPath, apihandlers.UpdateResource(updater, &reqScope, o.admit), isUpdater)
			// PATCH
			supportedTypes := []string{
				string(types.JSONPatchType),
				string(types.MergePatchType),
				string(types.StrategicMergePatchType),
				string(types.ApplyPatchType),
			}
			l.registerRouter(http.MethodPatch, resources.prefix+itemPath, apihandlers.PatchResource(patcher, &reqScope, o.admit, supportedTypes), isPatcher)
			// DELETE
			l.registerRouter(http.MethodDelete, resources.prefix+itemPath, apihandlers.DeleteResource(gracefulDeleter, isGracefulDeleter, &reqScope, o.admit), isGracefulDeleter)
			// DEPRECATED in 1.11 WATCH
			l.registerRouter(http.MethodGet, resources.prefix+"/watch"+itemPath, apihandlers.ListResource(lister, watcher, &reqScope, true, resources.minRequestTimeout), isWatcher)
			// CONNECT
			l.registerRouter(http.MethodConnect, resources.prefix+itemPath, apihandlers.ConnectResource(connecter, &reqScope, o.admit, o.path, isSubresource), isConnecter)
			l.registerRouter(http.MethodConnect, resources.prefix+itemPath+"/{path:*}", apihandlers.ConnectResource(connecter, &reqScope, o.admit, o.path, isSubresource), isConnecter && connectSubpath)
			// list or post across namespace.
			// For ex: LIST all pods in all namespaces by sending a LIST request at /api/apiVersion/pods.
			// LIST
			l.registerRouter(http.MethodGet, resources.prefix+"/"+resource, apihandlers.ListResource(lister, watcher, &reqScope, false, resources.minRequestTimeout), !isSubresource && isLister)
			// WATCHLIST
			l.registerRouter(http.MethodGet, resources.prefix+"/watch/"+resource, apihandlers.ListResource(lister, watcher, &reqScope, true, resources.minRequestTimeout), !isSubresource && allowWatchList)
		}
	}
	return nil
}

func (l *transport) registerRouter(verb string, path string, handler http.HandlerFunc, shouldAdd bool) {
	if !shouldAdd {
		// if the router should not be added. return
		return
	}
	for i, r := range l.routers {
		if r.path != path {
			continue
		}
		// add handler to router
		if _, ok := r.handlers[verb]; ok {
			// if handler is exists. throw error
			klog.V(4).ErrorS(fmt.Errorf("handler has already register"), "failed to register router", "path", path, "verb", verb)
			return
		}
		l.routers[i].handlers[verb] = handler
		return
	}

	// add new router
	expression, err := newPathExpression(path)
	if err != nil {
		klog.V(4).ErrorS(err, "failed to register router", "path", path, "verb", verb)
		return
	}
	l.routers = append(l.routers, router{
		path:     path,
		pathExpr: expression,
		handlers: map[string]http.HandlerFunc{
			verb: handler,
		},
	})
}

// splitSubresource checks if the given storage path is the path of a subresource and returns
// the resource and subresource components.
func splitSubresource(path string) (string, string, error) {
	var resource, subresource string
	switch parts := strings.Split(path, "/"); len(parts) {
	case 2:
		resource, subresource = parts[0], parts[1]
	case 1:
		resource = parts[0]
	default:
		return "", "", fmt.Errorf("api_installer allows only one or two segment paths (resource or resource/subresource)")
	}
	return resource, subresource, nil
}

var defaultRequestInfoResolver = &apirequest.RequestInfoFactory{
	APIPrefixes: sets.NewString("apis"),
}

func defaultHandlerChain(handler http.Handler) http.Handler {
	return genericapifilters.WithRequestInfo(handler, defaultRequestInfoResolver)
}
