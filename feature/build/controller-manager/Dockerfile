ARG builder_image
# Build the manager binary
FROM ${builder_image} as builder

ARG goproxy=https://goproxy.cn,direct
ENV GOPROXY ${goproxy}

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy the go source
COPY ./ ./

ARG ldflags
ARG build_tags

ENV LDFLAGS ${ldflags}
ENV BUILDTAGS ${build_tags}

# Cache the go build into the the Go’s compiler cache folder so we take benefits of compiler caching across docker build calls
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg/mod \
  CGO_ENABLED=0 go build -trimpath -tags "$BUILDTAGS" -ldflags "$LDFLAGS" -o controller-manager cmd/controller-manager/controller_manager.go

FROM alpine:3.19.0

WORKDIR /kubekey

COPY --from=builder /workspace/controller-manager /usr/local/bin/controller-manager

ENTRYPOINT ["sh"]
