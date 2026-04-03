#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

ETCDCTL_PATH='/usr/local/bin/etcdctl'
{{- $endpoints := list }}
{{- range .groups.etcd | default list }}
  {{- if $.need_uninstall_etcd | default list | has . | not }}
    {{- $internalIPv4 := index $.hostvars . "internal_ipv4" | default "" }}
    {{- $internalIPv6 := index $.hostvars . "internal_ipv6" | default "" }}
    {{- if $internalIPv4 | empty | not }}
      {{- $endpoints = append $endpoints (printf "https://%s:%d" $internalIPv4 $.etcd.port) }}
    {{- end }}
    {{- if $internalIPv6 | empty | not }}
      {{- $endpoints = append $endpoints (printf "https://[%s]:%d" $internalIPv6 $.etcd.port) }}
    {{- end }}
  {{- end }}
{{- end }}
ETCD_ENDPOINTS="{{ join "," $endpoints }}"
ETCD_DATA_DIR="{{ .etcd.env.data_dir }}"
BACKUP_DIR="${BACKUP_DIR:-{{ .etcd.backup.backup_dir }}/timer/etcd-$(date +%Y-%m-%d-%H-%M-%S)}"
KEEPBACKUPNUMBER='{{ .etcd.backup.keep_backup_number }}'
((KEEPBACKUPNUMBER++))

ETCDCTL_CERT="/etc/ssl/etcd/ssl/server.crt"
ETCDCTL_KEY="/etc/ssl/etcd/ssl/server.key"
ETCDCTL_CA_FILE="/etc/ssl/etcd/ssl/ca.crt"

[ ! -d $BACKUP_DIR ] && mkdir -p $BACKUP_DIR

# export ETCDCTL_API=3;$ETCDCTL_PATH backup --data-dir $ETCD_DATA_DIR --backup-dir $BACKUP_DIR

# sleep 3

{
export ETCDCTL_API=3;$ETCDCTL_PATH --endpoints="$ENDPOINTS" snapshot save $BACKUP_DIR/snapshot.db \
                                   --cacert="$ETCDCTL_CA_FILE" \
                                   --cert="$ETCDCTL_CERT" \
                                   --key="$ETCDCTL_KEY"
} > /dev/null

sleep 3

cd $BACKUP_DIR/../ && ls -1r |awk '{if(NR > '$KEEPBACKUPNUMBER'){print "rm -rf "$1}}'|sh
