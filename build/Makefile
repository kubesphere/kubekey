SHELL      = /usr/bin/env bash

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

VERSION_METADATA = unreleased
VERSION = latest
# Clear the "unreleased" string in BuildMetadata
ifneq ($(GIT_TAG),)
	VERSION_METADATA =
	VERSION = ${GIT_TAG}
endif

LDFLAGS += -X github.com/kubesphere/kubekey/version.version=${VERSION}
LDFLAGS += -X github.com/kubesphere/kubekey/version.metadata=${VERSION_METADATA}
LDFLAGS += -X github.com/kubesphere/kubekey/version.gitCommit=${GIT_COMMIT}
LDFLAGS += -X github.com/kubesphere/kubekey/version.gitTreeState=${GIT_DIRTY}

.PHONY: build
build: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	docker run --rm \
		-v $(shell pwd)/../:/usr/src/myapp \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-e GO111MODULE=on \
		-w /usr/src/myapp golang:1.14 \
		go build -ldflags '$(LDFLAGS)' -v -o output/linux/amd64/kk ./kubekey.go  # linux
	sha256sum ../output/linux/amd64/kk || shasum -a 256 ../output/linux/amd64/kk

build-linux-arm64:
	docker run --rm \
		-v $(shell pwd)../:/usr/src/myapp \
		-e GOOS=linux \
		-e GOARCH=arm64 \
		-e CGO_ENABLED=0 \
		-e GO111MODULE=on \
		-w /usr/src/myapp golang:1.14 \
		go build -ldflags '$(LDFLAGS)' -v -o output/linux/arm64/kk ./kubekey.go  # linux
	sha256sum ../output/linux/arm64/kk || shasum -a 256 ../output/linux/arm64/kk
