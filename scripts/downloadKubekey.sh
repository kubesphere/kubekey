#!/bin/sh

# Copyright 2020 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [ "x$(uname)" = "xDarwin" ]; then
  OSTYPE="osx"
else
  OSTYPE="linux"
fi

if [ "x${KUBEKEY_VERSION}" = "x" ]; then
  KUBEKEY_VERSION="$(curl -sL https://api.github.com/repos/kubesphere/kubekey/releases |
    grep -o 'download/v[0-9]*.[0-9]*.[0-9]*/' |
    sort --version-sort |
    tail -1 | awk -F'/' '{ print $2}')"
  KUBEKEY_VERSION="${KUBEKEY_VERSION##*/}"
fi

if [ -z "${ARCH}" ]; then
  case "$(uname -m)" in
  x86_64)
    ARCH=amd64
    ;;
  armv8*)
    ARCH=arm64
    ;;
  aarch64*)
    ARCH=arm64
    ;;
  armv*)
    ARCH=armv7
    ;;
  *)
    echo "${ARCH}, isn't supported"
    exit 1
    ;;
  esac
fi

if [ "x${KUBEKEY_VERSION}" = "x" ]; then
  echo "Unable to get latest Kubekey version. Set KUBEKEY_VERSION env var and re-run. For example: export KUBEKEY_VERSION=v1.0.0"
  echo ""
  exit
fi

DOWNLOAD_URL="https://github.com/kubesphere/kubekey/releases/download/${KUBEKEY_VERSION}/kubekey-${KUBEKEY_VERSION}-${OSTYPE}-${ARCH}.tar.gz"
if [ "x${KKZONE}" = "xcn" ]; then
  export KKZONE=cn
  DOWNLOAD_URL="https://kubernetes.pek3b.qingstor.com/kubekey/releases/download/${KUBEKEY_VERSION}/kubekey-${KUBEKEY_VERSION}-${OSTYPE}-${ARCH}.tar.gz"
fi

echo "\nDownloading kubekey ${KUBEKEY_VERSION} from ${DOWNLOAD_URL} ...\n"

curl -fsLO "$DOWNLOAD_URL"
if [ $? -ne 0 ]; then
  echo "\n\n"
  echo "Failed to download Kubekey ${KUBEKEY_VERSION} !\n"
  echo "Please verify the version you are trying to download.\n\n"
  exit
fi

filename="kubekey-${KUBEKEY_VERSION}-${OSTYPE}-${ARCH}.tar.gz"
tar -xzf "${filename}"
# rm "${filename}"

echo "\nKubekey ${KUBEKEY_VERSION} Download Complete!\n"
