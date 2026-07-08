package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func newFakeClient(t *testing.T) client.Client {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(_const.Scheme).Build()
}

func collectRoutePaths(ws *restful.WebService) []string {
	paths := make([]string, 0, len(ws.Routes()))
	for _, r := range ws.Routes() {
		paths = append(paths, r.Path)
	}
	return paths
}

func TestNewHealthzService(t *testing.T) {
	container := restful.NewContainer()
	container.Add(NewHealthzService())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestNewReadyzService(t *testing.T) {
	container := restful.NewContainer()
	container.Add(NewReadyzService())

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestNewCoreService_RoutesExist(t *testing.T) {
	ws := NewCoreService(t.TempDir(), newFakeClient(t), &rest.Config{})
	paths := collectRoutePaths(ws)
	base := "/kapis/" + kkcorev1.SchemeGroupVersion.String()

	expected := []string{
		base + "/inventories",
		base + "/namespaces/{namespace}/inventories/{inventory}",
		base + "/namespaces/{namespace}/inventories/{inventory}/hosts",
		base + "/playbooks",
		base + "/namespaces/{namespace}/playbooks",
		base + "/namespaces/{namespace}/playbooks/{playbook}",
		base + "/namespaces/{namespace}/playbooks/{playbook}/log",
	}
	for _, p := range expected {
		assert.Contains(t, paths, p, "expected route %s to be registered", p)
	}
}

func TestNewSchemaService_RoutesExist(t *testing.T) {
	ws := NewSchemaService(t.TempDir(), t.TempDir(), newFakeClient(t))
	paths := collectRoutePaths(ws)

	expected := []string{
		"/resources/ip",
		"/resources/schema/{subpath:*}",
		"/resources/schema",
		"/resources/schema/config",
	}
	for _, p := range expected {
		assert.Contains(t, paths, p, "expected route %s to be registered", p)
	}
}

func TestNewUIService(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('ok')"), 0o644))

	container := restful.NewContainer()
	container.Add(NewUIService(dir))

	t.Run("root path serves index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		container.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "<html></html>", rec.Body.String())
		assert.Equal(t, "text/html", rec.Header().Get("Content-Type"))
	})

	t.Run("static asset is served", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		rec := httptest.NewRecorder()
		container.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "console.log('ok')", rec.Body.String())
	})

	t.Run("api path falls through to 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/kapis/kubekey.kubesphere.io/v1/inventories", nil)
		rec := httptest.NewRecorder()
		container.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("spa route serves index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/clusters/foo", nil)
		rec := httptest.NewRecorder()
		container.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "<html></html>", rec.Body.String())
	})
}

func TestNewSwaggerUIService(t *testing.T) {
	container := restful.NewContainer()
	container.Add(NewSwaggerUIService())

	req := httptest.NewRequest(http.MethodGet, "/swagger-ui/", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rec.Body.String(), "Swagger")
}

func TestNewAPIService(t *testing.T) {
	container := restful.NewContainer()
	container.Add(NewAPIService([]*restful.WebService{NewHealthzService()}))

	req := httptest.NewRequest(http.MethodGet, "/apidocs.json", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	assert.Contains(t, rec.Body.String(), "KubeKey Web API")
}

// Compile-time check that client.Client is satisfied by the fake client.
var _ client.Client = fake.NewClientBuilder().Build()

// Keep corev1 imported for potential future test data usage.
var _ = corev1.Namespace{}
