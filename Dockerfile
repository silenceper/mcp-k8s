ARG BASEIMAGE=alpine:3.18.4
ARG GOVERSION=1.24.1
ARG LDFLAGS=""

# Build the manager binary
FROM golang:${GOVERSION} as builder
# Copy in the go src
WORKDIR /go/src/github.com/silenceper/mcp-k8s
COPY internal internal/
COPY cmd cmd/
COPY go.mod go.mod
COPY go.sum go.sum
ARG LDFLAGS
ARG TARGETOS
ARG TARGETARCH

# Build
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="${LDFLAGS}" -a -o mcp-k8s /go/src/github.com/silenceper/mcp-k8s/cmd/server

# Copy the cmd into a thin image
FROM ${BASEIMAGE}
WORKDIR /root
RUN apk add gcompat
COPY --from=builder /go/src/github.com/silenceper/mcp-k8s/mcp-k8s /usr/local/bin/mcp-k8s
ENTRYPOINT ["/usr/local/bin/mcp-k8s"]
