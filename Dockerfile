ARG BASEIMAGES=3.20.2
FROM golang:1.24.1 as builder
LABEL maintainer="wang.t <wang.t.nice@gmail.com>"

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn

WORKDIR /opt/
RUN git clone https://github.com/silenceper/mcp-k8s.git && \
    cd mcp-k8s && \
    go build -o bin/mcp-k8s cmd/server/main.go

FROM  --platform=$TARGETPLATFORM alpine:${BASEIMAGES}
ARG TARGETPLATFORM

WORKDIR /opt/mcp-k8s/bin/
LABEL maintainer="wang.t <wang.t.nice@gmail.com>"
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk update  \
    && apk add --no-cache ca-certificates bash tree tzdata libc6-compat dumb-init \
    && cp -rf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /opt/mcp-k8s/bin/mcp-k8s /opt/mcp-k8s/bin/mcp-k8s
EXPOSE 8080
ENTRYPOINT ["./mcp-k8s"]
