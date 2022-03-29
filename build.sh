#! /bin/bash

GIT_COMMIT=$(git rev-parse HEAD)
GIT_SHA=$(git rev-parse --short HEAD)
GIT_TAG=$(git describe --tags --abbrev=0 --exact-match 2>/dev/null )
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "dirty" || echo "clean")



VERSION_METADATA=unreleased
VERSION=latest
# Clear the "unreleased" string in BuildMetadata
if [[ -n $GIT_TAG ]]
then
  VERSION_METADATA=
  VERSION=${GIT_TAG}
fi

LDFLAGS="-X github.com/kubesphere/kubekey/version.version=${VERSION}
         -X github.com/kubesphere/kubekey/version.metadata=${VERSION_METADATA}
         -X github.com/kubesphere/kubekey/version.gitCommit=${GIT_COMMIT}
         -X github.com/kubesphere/kubekey/version.gitTreeState=${GIT_DIRTY}"

if [ -n "$1" ]; then 
    if [ "$1" == "-p" ] || [ "$1" == "--proxy" ]; then
        # Using the most trusted Go module proxy in China
        docker run --rm -e GO111MODULE=on -e GOPROXY=https://goproxy.cn -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17 go build -tags='containers_image_openpgp' -ldflags "$LDFLAGS" -v -o output/kk ./cmd/main.go
    else
        echo "The option should be '-p' or '--proxy'"
    fi
else
    docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17 go build -tags='containers_image_openpgp' -ldflags "$LDFLAGS" -v -o output/kk ./cmd/main.go
fi
