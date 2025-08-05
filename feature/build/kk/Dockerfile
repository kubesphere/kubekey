ARG builder_image
# Build the manager binary
FROM ${builder_image} as builder

ARG goproxy=https://goproxy.cn,direct
ENV GOPROXY ${goproxy}

WORKDIR /workspace

COPY api/go.mod api/go.mod
COPY api/go.sum api/go.sum
COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy the go source
COPY ./ ./

ARG ldflags
ARG build_tags

# Cache the go build into the the Go’s compiler cache folder so we take benefits of compiler caching across docker build calls
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg/mod \
  CGO_ENABLED=0 go build -trimpath -tags "${build_tags}" -ldflags "${ldflags}" -o kk cmd/kk/kubekey.go

FROM alpine:3.19.0

WORKDIR /kubekey

# install tool
RUN apk update && apk add bash && apk add curl && apk add openssl && apk add sudo
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 && \
     chmod 700 get_helm.sh && \
     ./get_helm.sh

COPY --from=ghcr.io/oras-project/oras:v1.1.0 /bin/oras /usr/local/bin/oras
COPY --from=builder /workspace/kk /usr/local/bin/kk
# add builtin capkk project 
COPY --from=builder /workspace/builtin/capkk /capkk/project

ENTRYPOINT ["sh"]
