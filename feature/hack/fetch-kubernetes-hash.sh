#!/bin/bash

v22_patch_max=15
v23_patch_max=13
v24_patch_max=7
v25_patch_max=3

versions=()

append_k8s_version() {
  prefix=$1
  max=$2
  for i in $(seq 0 "$max");
  do
    versions+=("${prefix}${i}")
  done
}

append_k8s_version "v1.22." $v22_patch_max
append_k8s_version "v1.23." $v23_patch_max
append_k8s_version "v1.24." $v24_patch_max
append_k8s_version "v1.25." $v25_patch_max

#versions=("v1.22.12" "v1.23.9" "v1.24.3")

arches=("amd64" "arm64")
apps=("kubeadm" "kubelet" "kubectl")
json="{}"
for app in "${apps[@]}";
do
  for arch in "${arches[@]}"
  do
    echo "${app}@${arch}"
    for ver in "${versions[@]}"
    do
      url="https://dl.k8s.io/release/${ver}/bin/linux/${arch}/${app}.sha256"
      hash=$(wget --quiet -O - "$url")
      echo "\"${ver}\": \"${hash}\","
      json=$(echo "$json" | jq ".${app}.${arch} += {\"${ver}\":\"${hash}\"}")
    done
    done
done

file="kubernetes-hashes.json"
echo "$json" | jq --indent 4 > "${file}" && echo -e "\n\nThe hash info have saved to file ${file}.\n\n"
