package manager

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/web"
)

// webManager handles the web server functionality for the application
type webManager struct {
	port    int
	workdir string

	schemaPath string

	ctrlclient.Client
	*rest.Config
}

// Run starts the web server and handles incoming requests
func (m webManager) Run(ctx context.Context) error {
	container := restful.DefaultContainer
	container.Filter(logRequestAndResponse)
	container.RecoverHandler(func(panicReason any, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})
	container.Add(web.NewSchemaService(m.schemaPath, m.workdir, m.Client)).
		Add(web.NewCoreService(m.workdir, m.Client, m.Config)).
		// openapi
		Add(web.NewSwaggerUIService()).
		Add(web.NewAPIService(container.RegisteredWebServices()))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", m.port),
		Handler:           container,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attack by timing out slow headers
	}

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}

// logStackOnRecover handles panic recovery and logs the stack trace
func logStackOnRecover(panicReason any, w http.ResponseWriter) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("recover from panic: %v\n", panicReason))
	for i := 2; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buf.WriteString(fmt.Sprintf("    %s:%d\n", file, line))
	}
	klog.Errorln(buf.String())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("Internal Server Error"))
}

// logRequestAndResponse logs HTTP request and response details
func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		remoteIP(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

// remoteIP extracts the client IP address from the request, handling various proxy headers
func remoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("X-Client-Ip"); ip != "" {
		remoteAddr = ip
	} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}
