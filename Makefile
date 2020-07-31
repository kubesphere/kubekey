.PHONY: build
build: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	docker run --rm \
		-v $(shell pwd):/usr/src/myapp \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-e GO111MODULE=on \
		-w /usr/src/myapp golang:1.14 \
		go build -v -o output/linux/amd64/kk ./kubekey.go  # linux
	sha256sum output/linux/amd64/kk || shasum -a 256 output/linux/amd64/kk

build-linux-arm64:
	docker run --rm \
		-v $(shell pwd):/usr/src/myapp \
		-e GOOS=linux \
		-e GOARCH=arm64 \
		-e CGO_ENABLED=0 \
		-e GO111MODULE=on \
		-w /usr/src/myapp golang:1.14 \
		go build -v -o output/linux/arm64/kk ./kubekey.go  # linux
	sha256sum output/linux/arm64/kk || shasum -a 256 output/linux/arm64/kk
