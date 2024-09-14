#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status.

### >>>>>>>> >>>>>>>> >>>>>>>> >>>>>>>> >>>>>>>> >>>>>>>> >>>>>>>> >>>>>>>> >>>>>>>> >>>>>>>> ###
### Used for capkk-template.yaml
export CLUSTER_NAME="capkk-test-cluster"
export CLUSTER_NAMESPACE="default"
export K8S_VERSION="1.23.15"
export CONTROL_PLANE_NODE_COUNT=1
export WORKER_NODE_COUNT=2
### Used for sh
export LOCAL_KK_DIR="/Users/dyl/GoProjects/kubekey/"
export REMOTE_KK_DIR="/root/kubekey/"
export BINARY="${REMOTE_KK_DIR}binary/"
export CRDS="${REMOTE_KK_DIR}crds/"
export CAPKK="${REMOTE_KK_DIR}capkk/"
export CAPKK_PLAYBOOK_PVC="/var/openebs/local/pvc-81d1bbad-863e-49eb-b9d4-076fa8b8fad6/"
export CAPKK_ARTIFACTS_PVC="/var/openebs/local/pvc-b71118d8-7992-4e14-9861-2ad111b7727a/"
### <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< ###

# Function for uploading files
upload_files() {
    echo ">>> Upload kubekey-1.0.0.tgz to server"
    scp -P 30001 "${LOCAL_KK_DIR}kubekey-1.0.0.tgz" root@139.198.121.174:"${BINARY}kubekey-1.0.0.tgz"
    echo ""

    echo ">>> Upload CRDs to server"
    scp -P 30001 -r "${LOCAL_KK_DIR}config/capkk/" root@139.198.121.174:"${CRDS}"
    scp -P 30001 -r "${LOCAL_KK_DIR}config/kubekey/crds/" root@139.198.121.174:"${CRDS}kubekey/"
    echo ""

    echo ">>> Sync playbooks to OpenEBS PVC"
    rsync -avz --delete -e "ssh -p 30002" "${LOCAL_KK_DIR}_playbooks/" root@139.198.121.174:"${CAPKK_PLAYBOOK_PVC}playbooks/"
    rsync -avz --delete -e "ssh -p 30002" "${LOCAL_KK_DIR}_roles/" root@139.198.121.174:"${CAPKK_PLAYBOOK_PVC}roles/"
    echo ""
}

# Function to apply CRDs
apply_crds() {
    echo ">>> Apply CRD resources on the server"
    ssh -p 30001 root@139.198.121.174 "\
    find ${CRDS}kubekey/crds ${CRDS}capkk/crds -type f -name '*.yaml' -exec kubectl apply -f {} \;"
    echo ""
}

# Function to update capkk-template.yaml
update_template() {
    echo ">>> Delete old capkk-template.yaml and apply the new one"
    # Must remove finalizers before delete (if exist).
    ssh -p 30001 root@139.198.121.174 "kubectl patch kkc/${CLUSTER_NAME} -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge || true"

    ssh -p 30001 root@139.198.121.174 "\
    for machine in \$(kubectl get kkmachine -o jsonpath='{.items[*].metadata.name}'); \
    do kubectl patch kkmachine \$machine -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge || true; \
    done"

    ssh -p 30001 root@139.198.121.174 "kubectl delete -f ${REMOTE_KK_DIR}capkk-template.yaml || true"

    sleep 10

    envsubst < ${LOCAL_KK_DIR}capkk-template.yaml > ${LOCAL_KK_DIR}capkk-template-resolved.yaml
    scp -P 30001 "${LOCAL_KK_DIR}capkk-template-resolved.yaml" root@139.198.121.174:"${REMOTE_KK_DIR}capkk-template.yaml"

    ssh -p 30001 root@139.198.121.174 "kubectl apply -f ${REMOTE_KK_DIR}capkk-template.yaml"
}

# Main steps
echo ">>> Generate CRD resources: make generate"
make generate
printf "\n### <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< ###\n"

upload_files
printf "\n### <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< ###\n"

apply_crds
printf "\n### <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< ###\n"

echo ">>> Check CAPKK related CRD"
ssh -p 30001 root@139.198.121.174 "kubectl get crd | grep -E 'kkclusters|kkmachines|kubekey.kubesphere.io'"
printf "\n### <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< ###\n"

update_template
printf "### <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< <<<<<<<< ###\n"

echo ">>> Finished."

# 如果出现 webhook 问题，查看以下 mu 和 va 资源
#kubectl get mutatingwebhookconfigurations.admissionregistration.k8s.io
#kubectl get validatingwebhookconfigurations.admissionregistration.k8s.io

# 如果 kkcluster 删不掉，去掉finalizers
#kubectl patch kkc/${CLUSTER_NAME} -p '{"metadata":{"finalizers":[]}}' --type=merge