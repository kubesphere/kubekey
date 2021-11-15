/*
 Copyright 2021 The KubeSphere Authors.

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

package templates

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/lithammer/dedent"
	"strconv"
	"text/template"
)

// EtcdBackupScriptTmpl defines the template of etcd backup script.
var EtcdBackupScript = template.Must(template.New("etcd-backup.sh").Parse(
	dedent.Dedent(`#!/bin/bash

ETCDCTL_PATH='/usr/local/bin/etcdctl'
ENDPOINTS='{{ .Etcdendpoint }}'
ETCD_DATA_DIR="/var/lib/etcd"
BACKUP_DIR="{{ .Backupdir }}/etcd-$(date +%Y-%m-%d-%H-%M-%S)"
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

func BackupTimeInterval(runtime connector.Runtime, kubeConf *common.KubeConf) string {
	var etcdBackupHour string
	if kubeConf.Cluster.Kubernetes.EtcdBackupPeriod != 0 {
		period := kubeConf.Cluster.Kubernetes.EtcdBackupPeriod
		if period > 60 && period < 1440 {
			kubeConf.Cluster.Kubernetes.EtcdBackupPeriod = period % 60
			etcdBackupHour = strconv.Itoa(period / 60)
		}
		if period > 1440 {
			logger.Log.Message(runtime.RemoteHost().GetName(), "etcd backup cannot last more than one day, Please change it to within one day or KubeKey will set it to the default value '24'.")
			return "24"
		}
	}
	return etcdBackupHour
}
