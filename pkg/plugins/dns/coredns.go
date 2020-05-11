package dns

import (
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"text/template"
)

var (
	CorednsServiceTempl = template.Must(template.New("CorednsService").Parse(
		dedent.Dedent(`---
apiVersion: v1
kind: Service
metadata:
  name: coredns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "coredns"
    addonmanager.kubernetes.io/mode: Reconcile
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: {{ .ClusterIP }}
  ports:
    - name: dns
      port: 53
      protocol: UDP
    - name: dns-tcp
      port: 53
      protocol: TCP
    - name: metrics
      port: 9153
      protocol: TCP
    `)))
)

func GenerateCorednsService(mgr *manager.Manager) (string, error) {
	return util.Render(CorednsServiceTempl, util.Data{
		"ClusterIP": mgr.Cluster.ClusterIP(),
	})
}
