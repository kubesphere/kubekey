package templates

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/lithammer/dedent"
	"strconv"
	"text/template"
)

var HaproxyConfig = template.Must(template.New("haproxy.cfg").Parse(
	dedent.Dedent(`
global
    maxconn                 4000
    log                     127.0.0.1 local0

defaults
    mode                    http
    log                     global
    option                  httplog
    option                  dontlognull
    option                  http-server-close
    option                  redispatch
    retries                 5
    timeout http-request    5m
    timeout queue           5m
    timeout connect         30s
    timeout client          30s
    timeout server          15m
    timeout http-keep-alive 30s
    timeout check           30s
    maxconn                 4000

frontend healthz
  bind *:{{ .LoadbalancerApiserverHealthcheckPort }}
  mode http
  monitor-uri /healthz

frontend kube_api_frontend
  bind 127.0.0.1:{{ .LoadbalancerApiserverPort }}
  mode tcp
  option tcplog
  default_backend kube_api_backend

backend kube_api_backend
  mode tcp
  balance leastconn
  default-server inter 15s downinter 15s rise 2 fall 2 slowstart 60s maxconn 1000 maxqueue 256 weight 100
  {{- if ne .KubernetesType "k3s"}}
  option httpchk GET /healthz
  {{- end }}
  http-check expect status 200
  {{- range .MasterNodes }}
  server {{ . }} check check-ssl verify none
  {{- end }}
`)))

func MasterNodeStr(runtime connector.ModuleRuntime, conf *common.KubeConf) []string {
	masterNodes := make([]string, len(runtime.GetHostsByRole(common.Master)))
	for i, node := range runtime.GetHostsByRole(common.Master) {
		masterNodes[i] = node.GetName() + " " + node.GetAddress() + ":" + strconv.Itoa(conf.Cluster.ControlPlaneEndpoint.Port)
	}
	return masterNodes
}
