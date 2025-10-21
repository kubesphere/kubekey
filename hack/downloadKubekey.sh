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
    version="$1"
    if echo "$version" | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' >/dev/null; then
        return 0
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

if [ "x${KKZONE}" = "xcn" ] && check_version "${VERSION}"; then
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

if [ "${ISLINUX}" = "true" ]; then
  filename="kubekey-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
  ret='0'
  command -v tar >/dev/null 2>&1 || { ret='1'; }
  if [ "$ret" -eq 0 ]; then
    tar -xzf "${filename}" --no-same-owner
  else
    echo "Kubekey ${VERSION} Download Complete!"
    echo ""
    echo "Try to unpack the ${filename} failed."
    echo "tar: command not found, please unpack the ${filename} manually."
    exit
  fi

fi

if check_version "${WEB_INSTALLER_VERSION}"; then
  WEB_DOWNLOAD_URL=https://kubekey.pek3b.qingstor.com/github.com/kubesphere/web-installer/releases/download/${WEB_INSTALLER_VERSION}/web-installer.tgz
  echo ""
  echo "Downloading kubekey web_installer ${VERSION} from ${DOWNLOAD_URL} ..."
  echo ""

  curl -fsLO "$WEB_DOWNLOAD_URL"
  ret='0'
  command -v tar >/dev/null 2>&1 || { ret='1'; }
  if [ "$ret" -eq 0 ]; then
    tar -xzf "web-installer.tgz" --no-same-owner
  else
    echo "Web Installer Download Complete!"
    echo ""
    echo "Try to unpack the web-installer.tgz failed."
    echo "tar: command not found, please unpack the web-installer.tgz manually."
    exit
  fi
fi

# generate package.sh
cat > package.sh << 'EOF'
#!/bin/sh

set -e

# Get the configuration file path from the first argument, default to config.yaml if not provided
if [ -n "$1" ]; then
  CONFIG_FILE="$1"
else
  CONFIG_FILE="config.yaml"
fi
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Configuration file $CONFIG_FILE does not exist. Please check the file path."
  exit 1
fi

echo "Exporting artifact with kk..."
./kk artifact export -c "$CONFIG_FILE" --workdir prepare -a $(pwd)/artifact.tgz
if [ $? -ne 0 ]; then
  echo "Failed to export artifact with kk. Please check the command output above."
  exit 1
fi

echo "Preparing offline package directory..."
mkdir -p offline

echo "Extracting artifact.tgz to offline/ ..."
tar -xzf artifact.tgz -C offline/ --no-same-owner

echo "Extracting web-installer.tgz to offline/ ..."
tar -xzf web-installer.tgz -C offline/ --no-same-owner

echo "Copying config.yaml and kk to offline/ ..."
cp "$CONFIG_FILE" offline/
cp kk offline/

echo "Creating offline.tgz package ..."
tar -czf offline.tgz offline

echo "Offline package offline.tgz has been created successfully."
EOF

chmod +x package.sh


echo ""
echo "Kubekey ${VERSION} Download Complete!"
echo ""
