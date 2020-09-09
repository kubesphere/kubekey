package tmpl

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"strings"
	"text/template"
)

var EtcdBackupScriptTmpl = template.Must(template.New("etcdBackupScript").Parse(
	dedent.Dedent(`#!/bin/bash

ETCDCTL_PATH='/usr/local/bin/etcdctl'
ENDPOINTS='{{ .Etcdendpoint }}'
ETCD_DATA_DIR="/var/lib/etcd"
BACKUP_DIR="/var/backups/kube_etcd/etcd-$(date +%Y-%m-%d_%H:%M:%S)"

ETCDCTL_CERT="/etc/ssl/etcd/ssl/admin-{{ .Hostname }}.pem"
ETCDCTL_KEY="/etc/ssl/etcd/ssl/admin-{{ .Hostname }}-key.pem"
ETCDCTL_CA_FILE="/etc/ssl/etcd/ssl/ca.pem"


[ ! -d $BACKUP_DIR ] && mkdir -p $BACKUP_DIR


export ETCDCTL_API=2;$ETCDCTL_PATH backup --data-dir $ETCD_DATA_DIR --backup-dir $BACKUP_DIR

sleep 3

{
export ETCDCTL_API=3;$ETCDCTL_PATH --endpoints="$ENDPOINTS" snapshot save $BACKUP_DIR/snapshot.db \
                                   --cacert="$ETCDCTL_CA_FILE" \
                                   --cert="$ETCDCTL_CERT" \
                                   --key="$ETCDCTL_KEY"
} > /dev/null 

sleep 3

cd $BACKUP_DIR/../;ls -lt |awk '{if(NR>14){print "rm -rf "$9}}'|sh
`)))

func EtcdBackupScript(mgr *manager.Manager, node *kubekeyapi.HostCfg ) (string, error) {
	ips := []string{}
	for _, host := range mgr.EtcdNodes {
		ips = append(ips, fmt.Sprintf("https://%s:2379", host.InternalAddress))
	}
	return util.Render(EtcdBackupScriptTmpl, util.Data{
		"Hostname": node.Name,
		"Etcdendpoint": strings.Join(ips, ","),
	})
}
