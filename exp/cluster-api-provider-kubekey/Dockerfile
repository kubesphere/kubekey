# Build the manager binary
FROM golang:1.18 as builder

# Run this with docker build --build_arg $(go env GOPROXY) to override the goproxy
ARG goproxy=https://goproxy.cn,direct
ENV GOPROXY=$goproxy

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY ./ ./

# Build
ARG package=.
ARG ARCH
ARG LDFLAGS
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags "${LDFLAGS}" -o manager ${package}
ENTRYPOINT ["/workspace/manager" ]

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ubuntu:20.04

WORKDIR /

RUN mkdir -p /var/lib/kubekey/rootfs

COPY --from=builder /workspace/manager .

ENTRYPOINT ["/manager"]
