FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o ops-controller-manager main.go

FROM alpine:latest
RUN apk add --update curl && rm -rf /var/cache/apk/*
WORKDIR /
COPY --from=builder /workspace/ops-controller-manager .
ENTRYPOINT ["/ops-controller-manager"]
