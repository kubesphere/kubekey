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
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/audit"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	apiendpoints "k8s.io/apiserver/pkg/endpoints"
	genericapifilters "k8s.io/apiserver/pkg/endpoints/filters"
	apihandlers "k8s.io/apiserver/pkg/endpoints/handlers"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	apirest "k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/internal"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/resources/inventory"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/resources/playbook"
	"github.com/kubesphere/kubekey/v4/pkg/proxy/resources/task"
)

// RestConfig replace the restconfig transport to proxy transport
func RestConfig(runtimedir string, restconfig *rest.Config) error {
	transport, err := newProxyTransport(runtimedir, restconfig)
	if err != nil {
		return err
	}
	restconfig.QPS = 500
	restconfig.Burst = 200
	restconfig.TLSClientConfig = rest.TLSClientConfig{}

	restconfig.Transport = transport

	return nil
}

// NewProxyTransport return a new http.RoundTripper use in ctrl.client.
// When restConfig is not empty: should connect a kubernetes cluster and store some resources in there.
// Such as: playbook.kubekey.kubesphere.io/v1, inventory.kubekey.kubesphere.io/v1, config.kubekey.kubesphere.io/v1
// when restConfig is empty: store all resource in local.
//
// SPECIFICALLY: since tasks is running data, which is reentrant and large in quantity,
// they should always store in local.
func newProxyTransport(runtimedir string, restConfig *rest.Config) (http.RoundTripper, error) {
	lt := &transport{
		authz: authorizerfactory.NewAlwaysAllowAuthorizer(),
		handlerChainFunc: func(handler http.Handler) http.Handler {
			return genericapifilters.WithRequestInfo(handler, &apirequest.RequestInfoFactory{
				APIPrefixes: sets.NewString("apis"),
			})
		},
	}
	if restConfig.Host != "" {
		clientFor, err := rest.HTTPClientFor(restConfig)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create http client")
		}
		lt.restClient = clientFor
	}

	var apiGroups []metav1.APIGroup
	// register kkcorev1alpha1 resources
	kkv1alpha1 := newAPIIResources(kkcorev1alpha1.SchemeGroupVersion)
	storage, err := task.NewStorage(internal.NewFileRESTOptionsGetter(runtimedir, kkcorev1alpha1.SchemeGroupVersion))
	if err != nil {
		return nil, err
	}
	if err := kkv1alpha1.AddResource(resourceOptions{
		path:    "tasks",
		storage: storage.Task,
	}); err != nil {
		return nil, err
	}
	if err := kkv1alpha1.AddResource(resourceOptions{
		path:    "tasks/status",
		storage: storage.TaskStatus,
	}); err != nil {
		return nil, err
	}
	if err := lt.registerResources(kkv1alpha1); err != nil {
		return nil, err
	}
	apiGroups = append(apiGroups, metav1.APIGroup{
		Name: kkv1alpha1.gv.Group,
		Versions: []metav1.GroupVersionForDiscovery{
			{
				GroupVersion: kkv1alpha1.gv.String(),
				Version:      kkv1alpha1.gv.Version,
			},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: kkv1alpha1.gv.String(),
			Version:      kkv1alpha1.gv.Version,
		},
	})

	// when restConfig is null. should store all resource local
	if restConfig.Host == "" {
		// register kkcorev1 resources
		kkv1 := newAPIIResources(kkcorev1.SchemeGroupVersion)
		// add inventory
		inventoryStorage, err := inventory.NewStorage(internal.NewFileRESTOptionsGetter(runtimedir, kkcorev1.SchemeGroupVersion))
		if err != nil {
			return nil, err
		}
		if err := kkv1.AddResource(resourceOptions{
			path:    "inventories",
			storage: inventoryStorage.Inventory,
		}); err != nil {
			return nil, err
		}
		// add playbook
		playbookStorage, err := playbook.NewStorage(internal.NewFileRESTOptionsGetter(runtimedir, kkcorev1.SchemeGroupVersion))
		if err != nil {
			return nil, err
		}
		if err := kkv1.AddResource(resourceOptions{
			path:    "playbooks",
			storage: playbookStorage.Playbook,
		}); err != nil {
			return nil, err
		}
		if err := kkv1.AddResource(resourceOptions{
			path:    "playbooks/status",
			storage: playbookStorage.PlaybookStatus,
		}); err != nil {
			return nil, err
		}

		if err := lt.registerResources(kkv1); err != nil {
			return nil, err
		}
		apiGroups = append(apiGroups, metav1.APIGroup{
			Name: kkv1.gv.Group,
			Versions: []metav1.GroupVersionForDiscovery{
				{
					GroupVersion: kkv1.gv.String(),
					Version:      kkv1.gv.Version,
				},
			},
			PreferredVersion: metav1.GroupVersionForDiscovery{
				GroupVersion: kkv1.gv.String(),
				Version:      kkv1.gv.Version,
			},
		})
	}

	lt.registerRouter(http.MethodGet, "/api", func(w http.ResponseWriter, r *http.Request) {
		obj := &metav1.APIVersions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIVersions",
				APIVersion: "v1",
			},
			Versions: []string{"v1"},
		}
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			klog.ErrorS(err, "failed to encode /api")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}, true)
	lt.registerRouter(http.MethodGet, "/apis", func(w http.ResponseWriter, r *http.Request) {
		obj := &metav1.APIGroupList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIGroupList",
				APIVersion: "v1",
			},
			Groups: apiGroups,
		}
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			klog.ErrorS(err, "failed to encode /api")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}, true)

	return lt, nil
}

// responseWriter implements http.ResponseWriter for capturing HTTP responses locally.
// It writes response body to an internal buffer but directly sets headers/status on the *http.Response.
type responseWriter struct {
	resp *http.Response // The response object to write headers and status to.
	buf  bytes.Buffer   // Buffer to capture the response body.
}

// Header returns the header map that will be sent by WriteHeader.
func (r *responseWriter) Header() http.Header {
	return r.resp.Header
}

// Write writes the data to the buffer as part of the HTTP response body.
func (r *responseWriter) Write(bs []byte) (int, error) {
	return r.buf.Write(bs)
}

// WriteHeader sets the HTTP status code in the response.
func (r *responseWriter) WriteHeader(statusCode int) {
	r.resp.StatusCode = statusCode
}

// finalize prepares the http.Response by setting its Body to the contents of the buffer.
// If the status code has not been set, it defaults to http.StatusOK (200).
func (r *responseWriter) finalize() {
	// Set the HTTP response body to the buffered data.
	r.resp.Body = io.NopCloser(bytes.NewReader(r.buf.Bytes()))
	// Default status code to 200 OK if it was not set by the handler.
	if r.resp.StatusCode == 0 {
		r.resp.StatusCode = http.StatusOK
	}
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

// RoundTrip deal proxy transport http.Request.
func (l *transport) RoundTrip(request *http.Request) (*http.Response, error) {
	if l.restClient != nil && !strings.HasPrefix(request.URL.Path, "/apis/"+kkcorev1alpha1.SchemeGroupVersion.String()) {
		return l.restClient.Transport.RoundTrip(request)
	}

	response := &http.Response{
		Proto:  "local",
		Header: make(http.Header),
	}
	request = request.WithContext(audit.WithAuditContext(request.Context()))
	handler, err := l.detectDispatcher(request)
	if err != nil {
		return response, err
	}
	// Use a buffered responseWriter to collect the complete response
	rw := &responseWriter{resp: response}
	l.handlerChainFunc(handler).ServeHTTP(rw, request)
	rw.finalize()

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
		return nil, errors.New("not found")
	}
	sort.Sort(sort.Reverse(filtered))

	handler, ok := filtered.candidates[0].router.handlers[request.Method]
	if !ok {
		return nil, errors.New("not found")
	}

	return handler, nil
}

func (l *transport) registerResources(resources *apiResources) error {
	// register apiResources router
	l.registerRouter(http.MethodGet, resources.prefix, resources.handlerAPIResources(), true)
	// register resources router
	for _, o := range resources.resourceOptions {
		// what verbs are supported by the storage, used to know what verbs we support per path

		_, isLister := o.storage.(apirest.Lister)
		_, isTableProvider := o.storage.(apirest.TableConvertor)
		if isLister && !isTableProvider {
			// All listers must implement TableProvider
			return errors.Errorf("%q must implement TableConvertor", o.path)
		}

		// Get the list of actions for the given scope.
		// namespace
		reqScope, err := newReqScope(resources, o, l.authz)
		if err != nil {
			return err
		}
		// LIST
		l.registerList(resources, reqScope, o)
		// POST
		l.registerPost(resources, reqScope, o)
		// DELETECOLLECTION
		l.registerDeleteCollection(resources, reqScope, o)
		// DEPRECATED in 1.11 WATCHLIST
		l.registerWatchList(resources, reqScope, o)
		// GET
		l.registerGet(resources, reqScope, o)
		// PUT
		l.registerPut(resources, reqScope, o)
		// PATCH
		l.registerPatch(resources, reqScope, o)
		// DELETE
		l.registerDelete(resources, reqScope, o)
		// DEPRECATED in 1.11 WATCH
		l.registerWatch(resources, reqScope, o)
		// CONNECT
		l.registerConnect(resources, reqScope, o)
	}

	return nil
}

// newReqScope for resource.
func newReqScope(resources *apiResources, o resourceOptions, authz authorizer.Authorizer) (apihandlers.RequestScope, error) {
	tableProvider, _ := o.storage.(apirest.TableConvertor)
	gvAcceptor, _ := o.storage.(apirest.GroupVersionAcceptor)
	// request scope
	fqKindToRegister, err := apiendpoints.GetResourceKind(resources.gv, o.storage, _const.Scheme)
	if err != nil {
		return apihandlers.RequestScope{}, errors.Wrap(err, "failed to get resourcekind")
	}
	reqScope := apihandlers.RequestScope{
		Namer: apihandlers.ContextBasedNaming{
			Namer:         meta.NewAccessor(),
			ClusterScoped: false,
		},
		Serializer:                  _const.CodecFactory,
		ParameterCodec:              _const.ParameterCodec,
		Creater:                     _const.Scheme,
		Convertor:                   _const.Scheme,
		Defaulter:                   _const.Scheme,
		Typer:                       _const.Scheme,
		UnsafeConvertor:             _const.Scheme,
		Authorizer:                  authz,
		EquivalentResourceMapper:    runtime.NewEquivalentResourceRegistry(),
		TableConvertor:              tableProvider,
		Resource:                    resources.gv.WithResource(o.resource),
		Subresource:                 o.subresource,
		Kind:                        fqKindToRegister,
		AcceptsGroupVersionDelegate: gvAcceptor,
		HubGroupVersion:             schema.GroupVersion{Group: fqKindToRegister.Group, Version: runtime.APIVersionInternal},
		MetaGroupVersion:            metav1.SchemeGroupVersion,
		MaxRequestBodyBytes:         0,
	}
	reqScope.FieldManager, err = managedfields.NewDefaultFieldManager(
		managedfields.NewDeducedTypeConverter(),
		_const.Scheme,
		_const.Scheme,
		_const.Scheme,
		fqKindToRegister,
		reqScope.HubGroupVersion,
		o.subresource,
		nil,
	)
	if err != nil {
		return apihandlers.RequestScope{}, errors.Wrap(err, "failed to create default fieldManager")
	}

	return reqScope, nil
}

func (l *transport) registerRouter(verb, path string, handler http.HandlerFunc, shouldAdd bool) {
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
			klog.V(6).ErrorS(errors.New("handler has already register"), "failed to register router", "path", path, "verb", verb)

			return
		}
		l.routers[i].handlers[verb] = handler

		return
	}

	// add new router
	expression, err := newPathExpression(path)
	if err != nil {
		klog.V(6).ErrorS(err, "failed to register router", "path", path, "verb", verb)

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

func (l *transport) registerList(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	lister, isLister := o.storage.(apirest.Lister)
	watcher, isWatcher := o.storage.(apirest.Watcher)
	l.registerRouter(http.MethodGet, resources.prefix+o.resourcePath, apihandlers.ListResource(lister, watcher, &reqScope, false, resources.minRequestTimeout), isLister)
	// list or post across namespace.
	// For ex: LIST all pods in all namespaces by sending a LIST request at /api/apiVersion/pods.
	// LIST
	l.registerRouter(http.MethodGet, resources.prefix+"/"+o.resource, apihandlers.ListResource(lister, watcher, &reqScope, false, resources.minRequestTimeout), o.subresource == "" && isLister)
	// WATCHLIST
	l.registerRouter(http.MethodGet, resources.prefix+"/watch/"+o.resource, apihandlers.ListResource(lister, watcher, &reqScope, true, resources.minRequestTimeout), o.subresource == "" && isWatcher && isLister)
}

func (l *transport) registerPost(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	creater, isCreater := o.storage.(apirest.Creater)
	namedCreater, isNamedCreater := o.storage.(apirest.NamedCreater)
	if isNamedCreater {
		l.registerRouter(http.MethodPost, resources.prefix+o.resourcePath, apihandlers.CreateNamedResource(namedCreater, &reqScope, o.admit), isCreater)
	} else {
		l.registerRouter(http.MethodPost, resources.prefix+o.resourcePath, apihandlers.CreateResource(creater, &reqScope, o.admit), isCreater)
	}
}

func (l *transport) registerDeleteCollection(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	collectionDeleter, isCollectionDeleter := o.storage.(apirest.CollectionDeleter)
	l.registerRouter(http.MethodDelete, resources.prefix+o.resourcePath, apihandlers.DeleteCollection(collectionDeleter, isCollectionDeleter, &reqScope, o.admit), isCollectionDeleter)
}

func (l *transport) registerWatchList(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	lister, isLister := o.storage.(apirest.Lister)
	watcher, isWatcher := o.storage.(apirest.Watcher)
	l.registerRouter(http.MethodGet, resources.prefix+"/watch"+o.resourcePath, apihandlers.ListResource(lister, watcher, &reqScope, true, resources.minRequestTimeout), isWatcher && isLister)
}

func (l *transport) registerGet(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	getterWithOptions, isGetterWithOptions := o.storage.(apirest.GetterWithOptions)
	getter, isGetter := o.storage.(apirest.Getter)
	if isGetterWithOptions {
		_, getSubpath, _ := getterWithOptions.NewGetOptions()
		l.registerRouter(http.MethodGet, resources.prefix+o.itemPath, apihandlers.GetResourceWithOptions(getterWithOptions, &reqScope, o.subresource != ""), isGetter)
		l.registerRouter(http.MethodGet, resources.prefix+o.itemPath+"/{path:*}", apihandlers.GetResourceWithOptions(getterWithOptions, &reqScope, o.subresource != ""), isGetter && getSubpath)
	} else {
		l.registerRouter(http.MethodGet, resources.prefix+o.itemPath, apihandlers.GetResource(getter, &reqScope), isGetter)
		l.registerRouter(http.MethodGet, resources.prefix+o.itemPath+"/{path:*}", apihandlers.GetResource(getter, &reqScope), false)
	}
}

func (l *transport) registerPut(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	updater, isUpdater := o.storage.(apirest.Updater)
	l.registerRouter(http.MethodPut, resources.prefix+o.itemPath, apihandlers.UpdateResource(updater, &reqScope, o.admit), isUpdater)
}

func (l *transport) registerPatch(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	patcher, isPatcher := o.storage.(apirest.Patcher)
	l.registerRouter(http.MethodPatch, resources.prefix+o.itemPath, apihandlers.PatchResource(patcher, &reqScope, o.admit, []string{
		string(types.JSONPatchType),
		string(types.MergePatchType),
		string(types.StrategicMergePatchType),
		string(types.ApplyPatchType),
	}), isPatcher)
}

func (l *transport) registerDelete(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	gracefulDeleter, isGracefulDeleter := o.storage.(apirest.GracefulDeleter)
	l.registerRouter(http.MethodDelete, resources.prefix+o.itemPath, apihandlers.DeleteResource(gracefulDeleter, isGracefulDeleter, &reqScope, o.admit), isGracefulDeleter)
}

func (l *transport) registerWatch(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	lister, _ := o.storage.(apirest.Lister)
	watcher, isWatcher := o.storage.(apirest.Watcher)
	l.registerRouter(http.MethodGet, resources.prefix+"/watch"+o.itemPath, apihandlers.ListResource(lister, watcher, &reqScope, true, resources.minRequestTimeout), isWatcher)
}

func (l *transport) registerConnect(resources *apiResources, reqScope apihandlers.RequestScope, o resourceOptions) {
	var connectSubpath bool
	connecter, isConnecter := o.storage.(apirest.Connecter)
	if isConnecter {
		_, connectSubpath, _ = connecter.NewConnectOptions()
	}
	l.registerRouter(http.MethodConnect, resources.prefix+o.itemPath, apihandlers.ConnectResource(connecter, &reqScope, o.admit, o.path, o.subresource != ""), isConnecter)
	l.registerRouter(http.MethodConnect, resources.prefix+o.itemPath+"/{path:*}", apihandlers.ConnectResource(connecter, &reqScope, o.admit, o.path, o.subresource != ""), isConnecter && connectSubpath)
}
