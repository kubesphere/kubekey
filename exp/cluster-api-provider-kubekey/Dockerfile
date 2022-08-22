# Download dependencies
FROM alpine:3.11 as base_os_context


ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /tmp

RUN apk add --no-cache ca-certificates

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

FROM alpine:3.11

WORKDIR /

RUN mkdir -p /var/lib/kubekey/rootfs

COPY --from=base_os_context /out/ /
COPY --from=builder /workspace/manager .

ENTRYPOINT ["/manager"]
