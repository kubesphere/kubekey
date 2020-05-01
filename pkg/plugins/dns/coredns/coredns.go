package coredns

import (
	"encoding/base64"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
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

func OverrideCorednsService(mgr *manager.Manager) error {
	corednsSvc, err := GenerateCorednsService(mgr)
	if err != nil {
		return err
	}
	corednsSvcgBase64 := base64.StdEncoding.EncodeToString([]byte(corednsSvc))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/coredns-svc.yaml\"", corednsSvcgBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "failed to generate kubeadm config")
	}
	deleteKubednsSvcCmd := "/usr/local/bin/kubectl delete -n kube-system svc kube-dns"
	_, err2 := mgr.Runner.RunCmd(deleteKubednsSvcCmd)
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "failed to delete kubeadm Kube-DNS service")
	}
	_, err3 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/coredns-svc.yaml")
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "failed to create coredns service")
	}
	return nil
}
