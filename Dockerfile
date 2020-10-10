# Build the manager binary
FROM golang:1.14 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
ADD ./ /workspace

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go


FROM debian:stable

RUN useradd -m kubekey && apt-get update && apt-get install bash curl -y; apt-get autoclean; rm -rf /var/lib/apt/lists/*
WORKDIR /home/kubekey
#ADD addons /home/kubekey/addons

#ADD kubekey /home/kubekey/kubekey
#RUN chown kubekey:kubekey -R /home/kubekey/kubekey

COPY --from=builder /workspace/manager /home/kubekey
USER kubekey:kubekey

ENTRYPOINT ["/home/kubekey/manager"]