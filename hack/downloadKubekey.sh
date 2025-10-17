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

ISLINUX=true
OSTYPE="linux"

check_version() {
    input_version="$1"
	
    if [ -z "$input_version" ]; then
        return 1
    fi
    version_number=$(echo "$input_version" | sed 's/^v//')
    
    major=$(echo "$version_number" | cut -d. -f1)
    minor=$(echo "$version_number" | cut -d. -f2)
    patch=$(echo "$version_number" | cut -d. -f3)
    
    : ${minor:=0}
    : ${patch:=0}
    
    if [ "$major" -gt 4 ]; then
        return 0
    elif [ "$major" -eq 4 ]; then
        if [ "$minor" -gt 0 ]; then
            return 0
        elif [ "$minor" -eq 0 ] && [ "$patch" -ge 0 ]; then
            return 0
        else
            return 1
        fi
    else
        return 1
    fi
}

if [ "x$(uname)" != "xLinux" ]; then
  echo ""
  echo 'Warning: Non-Linux operating systems are not supported! After downloading, please copy the tar.gz file to linux.'  
  ISLINUX=false
fi

# Fetch latest version of 3.x
if [ "x${VERSION}" = "x" ]; then
  VERSION="$(curl -sL https://api.github.com/repos/kubesphere/kubekey/releases |
    grep -o 'download/v3*.[0-9]*.[0-9]*/' |
    sort --version-sort |
    tail -1 | awk -F'/' '{ print $2}')"
  VERSION="${VERSION##*/}"
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
  *)
    echo "${ARCH}, isn't supported"
    exit 1
    ;;
  esac
fi

if [ "x${VERSION}" = "x" ]; then
  echo "Unable to get latest Kubekey version. Set VERSION env var and re-run. For example: export VERSION=v1.0.0"
  echo ""
  exit
fi

DOWNLOAD_URL="https://github.com/kubesphere/kubekey/releases/download/${VERSION}/kubekey-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"

if [ "x${KKZONE}" = "xcn" ] && check_version; then
  if echo "$VERSION" | grep -E '^v3\.[0-9]+\.[0-9]+$' >/dev/null; then
    DOWNLOAD_URL="https://kubernetes.pek3b.qingstor.com/kubekey/releases/download/${VERSION}/kubekey-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
  else
    DOWNLOAD_URL="https://kubekey.pek3b.qingstor.com/github.com/kubesphere/kubekey/releases/download/${VERSION}/kubekey-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
  fi
fi

echo ""
echo "Downloading kubekey ${VERSION} from ${DOWNLOAD_URL} ..."
echo ""

curl -fsLO "$DOWNLOAD_URL"
if [ $? -ne 0 ]; then
  echo ""
  echo "Failed to download Kubekey ${VERSION} !"
  echo ""
  echo "Please verify the version you are trying to download."
  echo ""
  exit
fi

if [ ${ISLINUX} = true ]; then
  filename="kubekey-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
  ret='0'
  command -v tar >/dev/null 2>&1 || { ret='1'; }
  if [ "$ret" -eq 0 ]; then
    tar -xzf "${filename}"
  else
    echo "Kubekey ${VERSION} Download Complete!"
    echo ""
    echo "Try to unpack the ${filename} failed."
    echo "tar: command not found, please unpack the ${filename} manually."
    exit
  fi
fi

if [ x${WITH_WEB_INSTALLER} = true ]; then
  curl -fsLO "https://kubekey.pek3b.qingstor.com/github.com/kubesphere/web-installer/releases/download/v1.0.0/web-installer.tgz"
  ret='0'
  command -v tar >/dev/null 2>&1 || { ret='1'; }
  if [ "$ret" -eq 0 ]; then
    tar -xzf "web-installer.tgz"
  else
    echo "Web Installer Download Complete!"
    echo ""
    echo "Try to unpack the web-installer.tgz failed."
    echo "tar: command not found, please unpack the web-installer.tgz manually."
    exit
  fi
fi

echo ""
echo "Kubekey ${VERSION} Download Complete!"
echo ""
