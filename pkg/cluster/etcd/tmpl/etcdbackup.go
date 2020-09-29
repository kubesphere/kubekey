package tmpl

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/api/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"os"
	"strconv"
	"strings"
	"text/template"
)

var EtcdBackupScriptTmpl = template.Must(template.New("etcdBackupScript").Parse(
	dedent.Dedent(`#!/bin/bash

ETCDCTL_PATH='/usr/local/bin/etcdctl'
ENDPOINTS='{{ .Etcdendpoint }}'
ETCD_DATA_DIR="/var/lib/etcd"
BACKUP_DIR="{{ .Backupdir }}/etcd-$(date +%Y-%m-%d_%H:%M:%S)"
KEEPBACKUPNUMBER='{{ .KeepbackupNumber }}'
ETCDBACKUPPERIOD='{{ .EtcdBackupPeriod }}'
ETCDBACKUPSCIPT='{{ .EtcdBackupScriptDir }}'
ETCDBACKUPHOUR='{{ .EtcdBackupHour }}'

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

cd $BACKUP_DIR/../;ls -lt |awk '{if(NR > '$KEEPBACKUPNUMBER'){print "rm -rf "$9}}'|sh

if [[ ! $ETCDBACKUPHOUR ]]; then
  time="*/$ETCDBACKUPPERIOD * * * *"
else
  if [[ 0 == $ETCDBACKUPPERIOD ]];then
    time="* */$ETCDBACKUPHOUR * * *"
  else
    time="*/$ETCDBACKUPPERIOD */$ETCDBACKUPHOUR * * *"
  fi
fi

crontab -l | grep -v '#' > /tmp/file
echo "$time sh $ETCDBACKUPSCIPT/etcd-backup.sh" >> /tmp/file && awk ' !x[$0]++{print > "/tmp/file"}' /tmp/file
crontab /tmp/file
rm -rf /tmp/file

`)))

func EtcdBackupScript(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) (string, error) {
	ips := []string{}
	var etcdBackupHour string
	for _, host := range mgr.EtcdNodes {
		ips = append(ips, fmt.Sprintf("https://%s:2379", host.InternalAddress))
	}
	if mgr.Cluster.Kubernetes.EtcdBackupPeriod != "" {
		period, _ := strconv.Atoi(mgr.Cluster.Kubernetes.EtcdBackupPeriod)
		if period > 60 && period < 1440 {
			mgr.Cluster.Kubernetes.EtcdBackupPeriod = strconv.Itoa(period % 60)
			etcdBackupHour = strconv.Itoa(period / 60)
		}
		if period > 1440 {
			fmt.Println("Etcd backup cannot last more than one day, Please change it to within one day.")
			os.Exit(0)
		}
	}

	return util.Render(EtcdBackupScriptTmpl, util.Data{
		"Hostname":            node.Name,
		"Etcdendpoint":        strings.Join(ips, ","),
		"Backupdir":           mgr.Cluster.Kubernetes.EtcdBackupDir,
		"KeepbackupNumber":    mgr.Cluster.Kubernetes.KeepBackupNumber,
		"EtcdBackupPeriod":    mgr.Cluster.Kubernetes.EtcdBackupPeriod,
		"EtcdBackupScriptDir": mgr.Cluster.Kubernetes.EtcdBackupScriptDir,
		"EtcdBackupHour":      etcdBackupHour,
	})
}
