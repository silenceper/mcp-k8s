FROM alpine:latest
LABEL maintainer="wang.t <wang.t.nice@gmail.com>,silenceper"

ARG VERSION=v1.0.2
ARG PLATFORM=linux_amd64
ARG MCP_K8S_URL=https://github.com/silenceper/mcp-k8s/releases/download/${VERSION}/mcp-k8s_${PLATFORM}

RUN wget -cO mcp-k8s $MCP_K8S_URL && \
    chmod  +x mcp-k8s
    
ENTRYPOINT ["./mcp-k8s"]
