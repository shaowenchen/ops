# Build the manager binary
FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the go source
COPY . .

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o ops-controller-manager main.go

FROM ubuntu:22.04
RUN apt-get update && \
    apt-get install -y curl python3 python3-pip && \
    rm -rf /var/cache/apt/* && \
    rm -rf /root/.cache && \
    rm -rf /tmp/*
RUN pip3 install --no-cache-dir requests
RUN mkdir /app && chown -R 65532:65532 /app
COPY --from=builder /workspace/ops-controller-manager /app/
USER 65532:65532
ENV HOME=/app
ENTRYPOINT ["/app/ops-controller-manager"]
