#!/bin/bash
{% if (renew_certs.is_kubeadm_alpha=="true") %}
kubeadmCerts='/usr/local/bin/kubeadm alpha certs'
{% else %}
kubeadmCerts='/usr/local/bin/kubeadm certs'
{% endif %}
getCertValidDays() {
  local earliestExpireDate; earliestExpireDate=$(${kubeadmCerts} check-expiration | grep -o "[A-Za-z]\{3,4\}\s\w\w,\s[0-9]\{4,\}\s\w*:\w*\s\w*\s*" | xargs -I {} date -d {} +%s | sort | head -n 1)
  local today; today="$(date +%s)"
  echo -n $(( ($earliestExpireDate - $today) / (24 * 60 * 60) ))
}
echo "## Expiration before renewal ##"
${kubeadmCerts} check-expiration
if [ $(getCertValidDays) -lt 30 ]; then
  echo "## Renewing certificates managed by kubeadm ##"
  ${kubeadmCerts} renew all
  echo "## Restarting control plane pods managed by kubeadm ##"
{% if (renew_certs.is_docker=="true") %}
  $(which docker | grep docker) ps -af 'name=k8s_POD_(kube-apiserver|kube-controller-manager|kube-scheduler|etcd)-*' -q | /usr/bin/xargs $(which docker | grep docker) rm -f
{% else %}
  $(which crictl | grep crictl) pods --namespace kube-system --name 'kube-scheduler-*|kube-controller-manager-*|kube-apiserver-*|etcd-*' -q | /usr/bin/xargs $(which crictl | grep crictl) rmp -f
{% endif %}
  echo "## Updating /root/.kube/config ##"
  cp /etc/kubernetes/admin.conf /root/.kube/config
fi
echo "## Waiting for apiserver to be up again ##"
until printf "" 2>>/dev/null >>/dev/tcp/127.0.0.1/6443; do sleep 1; done
echo "## Expiration after renewal ##"
${kubeadmCerts} check-expiration
