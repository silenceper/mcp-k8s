# mcp-k8s

A Kubernetes MCP (Model Control Protocol) server that enables interaction with Kubernetes clusters through MCP tools.

## Features

- Query supported Kubernetes resource types (built-in resources and CRDs)
- Perform CRUD operations on Kubernetes resources
- Configurable write operations (create/update/delete can be enabled/disabled independently)
- Connects to Kubernetes cluster using kubeconfig

## Architecture

### 1. Project Overview

An stdio-based MCP server that connects to Kubernetes clusters and provides the following capabilities:
- Query Kubernetes resource types (including built-in resources and CRDs)
- CRUD operations on Kubernetes resources (with configurable write operations)

### 2. Technical Stack

- Go
- [mcp-go](https://github.com/mark3labs/mcp-go) SDK
- Kubernetes client-go library
- Stdio for communication

### 3. Core Components

1. **MCP Server**: Uses mcp-go's `server` package to create an stdio-based MCP server
2. **K8s Client**: Uses client-go to connect to Kubernetes clusters
3. **Tool Implementations**: Implements various MCP tools for different Kubernetes operations

### 4. Available Tools

#### Resource Type Query Tools
- `get_api_resources`: Get all supported API resource types in the cluster

#### Resource Operation Tools
- `get_resource`: Get detailed information about a specific resource
- `list_resources`: List all instances of a resource type
- `create_resource`: Create new resources (can be disabled)
- `update_resource`: Update existing resources (can be disabled)
- `delete_resource`: Delete resources (can be disabled)

## Getting Started

### Build

```bash
git clone https://github.com/silenceper/mcp-k8s.git
cd mcp-k8s
go build -o bin/mcp-k8s cmd/server/main.go
```

### Run

Default mode (read-only operations):
```bash
./bin/mcp-k8s --kubeconfig=/path/to/kubeconfig
```

Enable write operations:
```bash
./bin/mcp-k8s --kubeconfig=/path/to/kubeconfig --enable-create --enable-update --enable-delete
```

### Command Line Arguments

- `--kubeconfig`: Path to Kubernetes configuration file (uses default config if not specified)
- `--enable-create`: Enable resource creation operations (default: false)
- `--enable-update`: Enable resource update operations (default: false)
- `--enable-delete`: Enable resource deletion operations (default: false)

### Integration with MCP Clients

mcp-k8s is an stdio-based MCP server that can be integrated with any MCP-compatible LLM client. Refer to your MCP client's documentation for integration instructions.

## Security Considerations

- Write operations are strictly controlled through independent configuration switches
- Uses RBAC to ensure K8s client has only necessary permissions
- Validates all user inputs to prevent injection attacks
