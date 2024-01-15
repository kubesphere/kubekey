# Build architecture
ARG ARCH
ARG builder_image

# Download dependencies
FROM alpine:3.19.0 as base_os_context


ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /tmp

RUN apk add --no-cache ca-certificates


# Build the manager binary
FROM ${builder_image} as builder

# Run this with docker build --build_arg $(go env GOPROXY) to override the goproxy
ARG goproxy=https://goproxy.cn,direct
ENV GOPROXY=$goproxy

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the go source
COPY ./ ./

# Cache the go build into the the Go’s compiler cache folder so we take benefits of compiler caching across docker build calls
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o controller-manager cmd/controller-manager/controller_manager.go

# Build
ARG ARCH
ARG LDFLAGS

# Do not force rebuild of up-to-date packages (do not use -a) and use the compiler cache folder
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} \
    go build -o controller-manager cmd/controller-manager/controller_manager.go

FROM --platform=${ARCH} alpine:3.19.0

WORKDIR /

RUN mkdir -p /var/lib/kubekey/rootfs

COPY --from=base_os_context /out/ /
COPY --from=builder /workspace/controller-manager /usr/local/bin

ENTRYPOINT ["sh"]
