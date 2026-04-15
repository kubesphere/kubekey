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

LATEST_VERSION=
LATEST_WEB_INSTALLER_VERSION=

# Fetch latest version if VERSION not set
if [ "x${VERSION}" = "x" ]; then
  VERSION="${LATEST_VERSION}"
fi

is_valid_version() {
  echo "$1" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'
}

case "$(uname -m)" in
x86_64|amd64) ARCH=amd64 ;;
armv8*|aarch64*|arm64) ARCH=arm64 ;;
*)
  echo "$(uname -m) isn't supported"
  exit 1
  ;;
esac

case "$(uname -s)" in
Linux) OSTYPE=linux ;;
Darwin) OSTYPE=darwin ;;
*)
  echo "Unsupported OS. Use Linux or macOS"
  exit 1
  ;;
esac

DOWNLOAD_PREFIX="https://github.com/kubesphere/kubekey"
if [ "x${KKZONE}" = "xcn" ] && is_valid_version "${VERSION}"; then
  DOWNLOAD_PREFIX="https://kubekey.pek3b.qingstor.com/github.com/kubesphere/kubekey"
  if echo "$VERSION" | grep -E '^v3\.[0-9]+\.[0-9]+$' >/dev/null; then
    DOWNLOAD_PREFIX="https://kubernetes.pek3b.qingstor.com/kubekey"
  fi
fi

FILENAME="kubekey-${VERSION}-${OSTYPE}-${ARCH}.tar.gz"
DOWNLOAD_URL="${DOWNLOAD_PREFIX}/releases/download/${VERSION}/${FILENAME}"

echo ""
echo "Downloading kubekey ${VERSION} for ${OSTYPE}/${ARCH} from ${DOWNLOAD_URL} ..."
echo ""

curl -fsLO "$DOWNLOAD_URL"
if [ $? -ne 0 ]; then
  echo ""
  echo "Failed to download Kubekey ${VERSION} !"
  echo ""
  echo "Please verify the version you are trying to download."
  echo ""
  exit 1
fi

ret=0
command -v tar >/dev/null 2>&1 || ret=1

if [ ${ret} -eq 0 ]; then
  tar -xzf "${FILENAME}" --no-same-owner
else
  echo "Kubekey ${VERSION} Download Complete!"
  echo ""
  echo "tar command not found. Please unpack ${FILENAME} manually."
  echo "Then run: sudo mv kk /usr/local/bin/"
  exit 1
fi

if [ "${SKIP_WEB_INSTALLER}" != "true" ] && [ -n "${LATEST_WEB_INSTALLER_VERSION}" ] && is_valid_version "${LATEST_WEB_INSTALLER_VERSION}"; then
  WEB_DOWNLOAD_URL=https://kubekey.pek3b.qingstor.com/github.com/kubesphere/web-installer/releases/download/${LATEST_WEB_INSTALLER_VERSION}/web-installer.tgz
  echo ""
  echo "Downloading kubekey web_installer ${LATEST_WEB_INSTALLER_VERSION} from ${WEB_DOWNLOAD_URL} ..."

  curl -fsLO "$WEB_DOWNLOAD_URL"
  if tar -xzf "web-installer.tgz" --no-same-owner 2>/dev/null; then
    :
  else
    echo "tar not found. Please unpack web-installer.tgz manually."
    exit 1
  fi
fi

if [ "${SKIP_PACKAGE}" != "true" ] && ! echo "$VERSION" | grep -E '^v3\.[0-9]+\.[0-9]+$' >/dev/null; then
  # generate package.sh
cat > package.sh << EOF
#!/bin/sh

set -e

# Get the configuration file path from the first argument, default to config.yaml if not provided
if [ -n "\$1" ]; then
  CONFIG_FILE="\$1"
else
  CONFIG_FILE="config.yaml"
fi
if [ ! -f "\$CONFIG_FILE" ]; then
  echo "Configuration file \$CONFIG_FILE does not exist. Please check the file path."
  exit 1
fi

echo "Exporting artifact with kk..."
./kk artifact export -c "\$CONFIG_FILE" --workdir prepare --set download.tools.kubekey=$DOWNLOAD_PREFIX/releases/download/$VERSION/kubekey-$VERSION-linux-"{{ \"{{ .arch }}\" }}".tar.gz
if [ $? -ne 0 ]; then
  echo "Failed to export artifact with kk. Please check the command output above."
  exit 1
fi

if [ -f "web-installer.tgz" ]; then
  echo "Extracting web-installer.tgz to artifacts/web-installer/ ..."
  tar -xzf web-installer.tgz -C prepare/artifact --no-same-owner
fi

cd prepare/ && tar -czf ../artifact.tgz artifact

echo "Offline package artifact.tgz has been created successfully."
EOF
  chmod +x package.sh
fi

echo ""
echo "Kubekey ${VERSION} Download Complete!"
echo ""
