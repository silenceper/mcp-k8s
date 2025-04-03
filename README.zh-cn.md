# mcp-k8s

一个基于 MCP（Model Control Protocol）的 Kubernetes 服务器，通过 MCP 工具实现与 Kubernetes 集群的交互。

## 功能特性

- 查询支持的 Kubernetes 资源类型（内置资源和 CRD）
- 对 Kubernetes 资源执行 CRUD 操作
- 可配置的写操作（创建/更新/删除可独立启用/禁用）
- 使用 kubeconfig 连接 Kubernetes 集群

## 使用场景

### 1. 基于 LLM 的 Kubernetes 资源管理

- **交互式资源管理**：通过自然语言与 LLM 交互来管理 Kubernetes 资源，无需记忆复杂的 kubectl 命令
- **批量操作**：使用自然语言描述复杂的批量操作需求，由 LLM 转换为具体的资源操作
- **资源状态查询**：使用自然语言查询集群中的资源状态，获得更易理解的响应

### 2. 自动化运维场景

- **智能运维助手**：作为运维人员的智能助手，协助进行日常的集群管理工作
- **问题诊断**：通过自然语言描述问题，由 LLM 协助进行集群问题诊断
- **配置审查**：利用 LLM 的理解能力，帮助审查和优化 Kubernetes 资源配置

### 3. 开发与测试支持

- **快速原型验证**：开发人员可以通过自然语言快速创建和验证资源配置
- **环境管理**：简化测试环境的资源管理，快速创建、修改和清理测试资源
- **配置生成**：根据需求描述，自动生成符合最佳实践的资源配置

### 4. 教育培训场景

- **交互式学习**：新手可以通过自然语言交互来学习 Kubernetes 概念和操作
- **最佳实践指导**：LLM 可以在资源操作时提供最佳实践建议
- **错误解释**：当操作出错时，提供更易理解的错误解释和修正建议

## 架构

### 1. 项目概述

一个基于标准输入输出的 MCP 服务器，连接到 Kubernetes 集群并提供以下功能：
- 查询 Kubernetes 资源类型（包括内置资源和 CRD）
- 对 Kubernetes 资源进行 CRUD 操作（可配置写操作）

### 2. 技术栈

- Go
- [mcp-go](https://github.com/mark3labs/mcp-go) SDK
- Kubernetes client-go 库
- 标准输入输出通信

### 3. 核心组件

1. **MCP 服务器**：使用 mcp-go 的 `server` 包创建基于标准输入输出的 MCP 服务器
2. **K8s 客户端**：使用 client-go 连接 Kubernetes 集群
3. **工具实现**：实现各种用于不同 Kubernetes 操作的 MCP 工具

### 4. 可用工具

#### 资源类型查询工具
- `get_api_resources`：获取集群中所有支持的 API 资源类型

#### 资源操作工具
- `get_resource`：获取特定资源的详细信息
- `list_resources`：列出某类资源的所有实例
- `create_resource`：创建新资源（可禁用）
- `update_resource`：更新现有资源（可禁用）
- `delete_resource`：删除资源（可禁用）

## 快速开始

### 构建

```bash
git clone https://github.com/silenceper/mcp-k8s.git
cd mcp-k8s
go build -o bin/mcp-k8s cmd/server/main.go
```

### 运行

默认模式（只读操作）：
```bash
./bin/mcp-k8s --kubeconfig=/path/to/kubeconfig
```

启用写操作：
```bash
./bin/mcp-k8s --kubeconfig=/path/to/kubeconfig --enable-create --enable-update --enable-delete
```

### 命令行参数

- `--kubeconfig`：Kubernetes 配置文件路径（如果未指定则使用默认配置）
- `--enable-create`：启用资源创建操作（默认：false）
- `--enable-update`：启用资源更新操作（默认：false）
- `--enable-delete`：启用资源删除操作（默认：false）

### 与 MCP 客户端集成

mcp-k8s 是一个基于标准输入输出的 MCP 服务器，可以与任何兼容 MCP 的 LLM 客户端集成。请参考您的 MCP 客户端文档获取集成说明。

## 安全考虑

- 通过独立的配置开关严格控制写操作
- 使用 RBAC 确保 K8s 客户端只具有必要的权限
- 验证所有用户输入以防止注入攻击 