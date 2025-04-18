# mcp-k8s

一个基于 MCP（Model Control Protocol）的 Kubernetes 服务器，支持通过 MCP 工具与 Kubernetes 集群进行交互。

## 特性

- 查询支持的 Kubernetes 资源类型（内置资源和 CRD）
- 对 Kubernetes 资源执行 CRUD 操作
- 可配置的写操作（创建/更新/删除可以独立启用/禁用）
- 使用 kubeconfig 连接到 Kubernetes 集群

## 预览
> 通过 cursor 进行交互

![](./docs/create-deployment.png)

## 使用场景

### 1. 通过 LLM 管理 Kubernetes 资源

- **交互式资源管理**：通过自然语言与 LLM 交互来管理 Kubernetes 资源，无需记忆复杂的 kubectl 命令
- **批量操作**：用自然语言描述复杂的批量操作需求，让 LLM 将其转换为具体的资源操作
- **资源状态查询**：使用自然语言查询集群资源状态，获得易于理解的响应

### 2. 自动化运维场景

- **智能运维助手**：作为运维人员在日常集群管理任务中的智能助手
- **问题诊断**：通过自然语言问题描述协助集群问题诊断
- **配置审查**：利用 LLM 的理解能力帮助审查和优化 Kubernetes 资源配置

### 3. 开发和测试支持

- **快速原型验证**：开发者可以通过自然语言快速创建和验证资源配置
- **环境管理**：简化测试环境资源管理，快速创建、修改和清理测试资源
- **配置生成**：根据需求描述自动生成遵循最佳实践的资源配置

### 4. 教育和培训场景

- **交互式学习**：新手可以通过自然语言交互学习 Kubernetes 概念和操作
- **最佳实践指导**：在资源操作过程中，LLM 提供最佳实践建议
- **错误解释**：在操作失败时提供易于理解的错误解释和修正建议

## 架构

### 1. 项目概述

一个基于 stdio 的 MCP 服务器，连接到 Kubernetes 集群并提供以下功能：
- 查询 Kubernetes 资源类型（包括内置资源和 CRD）
- 对 Kubernetes 资源进行 CRUD 操作（可配置写操作）

### 2. 技术栈

- Go
- [mcp-go](https://github.com/mark3labs/mcp-go) SDK
- Kubernetes client-go 库
- Stdio 用于通信

### 3. 核心组件

1. **MCP 服务器**：使用 mcp-go 的 `server` 包创建基于 stdio 的 MCP 服务器
2. **K8s 客户端**：使用 client-go 连接到 Kubernetes 集群
3. **工具实现**：实现各种 MCP 工具用于不同的 Kubernetes 操作

### 4. 可用工具

#### 资源类型查询工具
- `get_api_resources`：获取集群中所有支持的 API 资源类型

#### 资源操作工具
- `get_resource`：获取特定资源的详细信息
- `list_resources`：列出资源类型的所有实例
- `create_resource`：创建新资源（可禁用）
- `update_resource`：更新现有资源（可禁用）
- `delete_resource`：删除资源（可禁用）

## 使用方式

mcp-k8s 支持两种通信模式：

### 1. Stdio 模式（默认）

在 stdio 模式下，mcp-k8s 通过标准输入/输出流与客户端通信。这是默认模式，适合大多数使用场景。

```bash
# 以 stdio 模式运行（默认）
{
    "mcpServers":
    {
        "mcp-k8s":
        {
            "command": "/path/to/mcp-k8s",
            "args":
            [
                "-kubeconfig",
                "/path/to/kubeconfig",
                "-enable-create",
                "-enable-delete",
                "-enable-update",
                "-enable-list"
            ]
        }
    }
}
```

### 2. SSE 模式

在 SSE（Server-Sent Events）模式下，mcp-k8s 向 mcp 客户端暴露 HTTP 端点。
您可以将服务部署在远程服务器上（但需要注意安全性）

```bash
# 以 SSE 模式运行
./bin/mcp-k8s -kubeconfig=/path/to/kubeconfig -transport=sse -port=8080 -host=localhost -enable-create -enable-delete -enable-list -enable-update
# 此命令将开启所有操作
```

mcp 配置
```json
{
  "mcpServers": {
    "mcp-k8s": {
      "url": "http://localhost:8080/sse",
      "args": []
    }
  }
}
```

SSE 模式配置：
- `-transport`：设置为 "sse" 以启用 SSE 模式
- `-port`：HTTP 服务器端口（默认：8080）
- `-host`：HTTP 服务器主机（默认："localhost"）

### 3. 使用 Docker

您可以像这样使用docker环境运行

```bash
docker build -t mcp-k8s:latest .  
docker run --rm  -i -v   ~/.kube/config:/root/.kube/config  mcp-k8s:latest  args
```
对于 docker 的 mcp 配置
```json
{
  "mcpServers": {
    "mcp-k8s": {
      "url": "http://you ip:8080/sse",
      "args": []
    }
  }
}
```
对于 docker 在 stdio 模式下
```json
{
  "mcpServers": {
    "mmcp-k8s": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "-v",
        "~/.kube/config:/root/.kube/config",
        "--rm",
        "mcp-k8s:latest"
      ]
    }
  }
}
```

## 快速开始

### 直接使用
您可以直接从 [releases 页面](https://github.com/silenceper/mcp-k8s/releases) 下载适合您平台的二进制文件并立即使用。

### 构建

```bash
git clone https://github.com/silenceper/mcp-k8s.git
cd mcp-k8s
go build -o bin/mcp-k8s cmd/server/main.go
```

### 命令行参数

- `-kubeconfig`：Kubernetes 配置文件路径（如果未指定则使用默认配置）
- `-enable-create`：启用资源创建操作（默认：false）
- `-enable-update`：启用资源更新操作（默认：false）
- `-enable-delete`：启用资源删除操作（默认：false）
- `-enable-list`：启用资源列表操作（默认：true）
- `-transport`：传输类型（stdio 或 sse）（默认："stdio"）
- `-host`：SSE 传输的主机（默认 "localhost"）
- `-port`：SSE 传输的 TCP 端口（默认 8080）

### 与 MCP 客户端集成

mcp-k8s 是一个基于 stdio 的 MCP 服务器，可以与任何兼容 MCP 的 LLM 客户端集成。请参考您的 MCP 客户端的文档了解集成说明。

## 安全考虑

- 通过独立的配置开关严格控制写操作
- 使用 RBAC 确保 K8s 客户端仅具有必要的权限
- 验证所有用户输入以防止注入攻击

## 关注微信公众号
![AI技术小林](./docs/qrcode.png)
