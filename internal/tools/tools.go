package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/silenceper/mcp-k8s/internal/k8s"
)

const (
	// DefaultPodLogTailLinesStr is the default number of lines to retrieve from pod logs as string
	DefaultPodLogTailLinesStr = "50"
)

// CreateGetAPIResourcesTool creates a tool for getting API resources
func CreateGetAPIResourcesTool() mcp.Tool {
	return mcp.NewTool("get_api_resources",
		mcp.WithDescription("Get all supported API resource types in the cluster, including built-in resources and CRDs"),
		mcp.WithBoolean("includeNamespaceScoped",
			mcp.Description("Include namespace-scoped resources"),
			mcp.DefaultBool(true),
		),
		mcp.WithBoolean("includeClusterScoped",
			mcp.Description("Include cluster-scoped resources"),
			mcp.DefaultBool(true),
		),
	)
}

// CreateGetResourceTool creates a tool for getting a specific resource
func CreateGetResourceTool() mcp.Tool {
	return mcp.NewTool("get_resource",
		mcp.WithDescription("Get detailed information about a specific resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Resource name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// CreateListResourcesTool creates a tool for listing resources
func CreateListResourcesTool() mcp.Tool {
	return mcp.NewTool("list_resources",
		mcp.WithDescription("List all instances of a resource type"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (only list resources in this namespace)"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Label selector (format: key1=value1,key2=value2)"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Field selector (format: key1=value1,key2=value2)"),
		),
	)
}

// CreateCreateResourceTool creates a tool for creating resources
func CreateCreateResourceTool() mcp.Tool {
	return mcp.NewTool("create_resource",
		mcp.WithDescription("Create a new resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
		mcp.WithString("manifest",
			mcp.Required(),
			mcp.Description("Resource manifest (Only JSON is supported)"),
		),
	)
}

// CreateUpdateResourceTool creates a tool for updating resources
func CreateUpdateResourceTool() mcp.Tool {
	return mcp.NewTool("update_resource",
		mcp.WithDescription("Update an existing resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Resource name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
		mcp.WithString("manifest",
			mcp.Required(),
			mcp.Description("Resource manifest (Only JSON is supported)"),
		),
	)
}

// CreateDeleteResourceTool creates a tool for deleting resources
func CreateDeleteResourceTool() mcp.Tool {
	return mcp.NewTool("delete_resource",
		mcp.WithDescription("Delete a resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Resource name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// HandleGetAPIResources handles the get API resources tool
func HandleGetAPIResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		includeNamespaceScoped := request.GetBool("includeNamespaceScoped", true)
		includeClusterScoped := request.GetBool("includeClusterScoped", true)

		resources, err := client.GetAPIResources(ctx, includeNamespaceScoped, includeClusterScoped)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetResource handles the get resource tool
func HandleGetResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: kind: %w", err)
		}

		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		namespace := request.GetString("namespace", "")

		resource, err := client.GetResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleListResources handles the list resources tool
func HandleListResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: kind: %w", err)
		}

		namespace := request.GetString("namespace", "")
		labelSelector := request.GetString("labelSelector", "")
		fieldSelector := request.GetString("fieldSelector", "")

		resources, err := client.ListResources(ctx, kind, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleCreateResource handles the create resource tool
func HandleCreateResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: kind: %w", err)
		}

		manifest, err := request.RequireString("manifest")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: manifest: %w", err)
		}

		namespace := request.GetString("namespace", "")

		resource, err := client.CreateResource(ctx, kind, namespace, manifest)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleUpdateResource handles the update resource tool
func HandleUpdateResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: kind: %w", err)
		}

		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		manifest, err := request.RequireString("manifest")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: manifest: %w", err)
		}

		namespace := request.GetString("namespace", "")

		resource, err := client.UpdateResource(ctx, kind, name, namespace, manifest)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleDeleteResource handles the delete resource tool
func HandleDeleteResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, err := request.RequireString("kind")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: kind: %w", err)
		}

		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		namespace := request.GetString("namespace", "")

		err = client.DeleteResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted resource %s/%s", kind, name)), nil
	}
}

// CreateGetPodLogsTool creates a tool for getting pod logs
func CreateGetPodLogsTool() mcp.Tool {
	return mcp.NewTool("get_pod_logs",
		mcp.WithDescription("Retrieve logs from a specific pod"),
		mcp.WithString("pod_name",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
		mcp.WithString("namespace",
			mcp.Description(fmt.Sprintf("Namespace (default: %s)", k8s.DefaultNamespace)),
		),
		mcp.WithString("container",
			mcp.Description("Container name (optional, defaults to first container in multi-container pods)"),
		),
		mcp.WithString("tail_lines",
			mcp.Description(fmt.Sprintf("Number of lines to retrieve from the end of logs (optional, default: %d)", k8s.DefaultPodLogTailLines)),
		),
	)
}

// HandleGetPodLogs handles the get pod logs tool
func HandleGetPodLogs(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		podName, err := request.RequireString("pod_name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: pod_name: %w", err)
		}

		namespace := request.GetString("namespace", k8s.DefaultNamespace)
		container := request.GetString("container", "")

		// Parse tail_lines as string and convert to int
		tailLinesStr := request.GetString("tail_lines", DefaultPodLogTailLinesStr)
		tailLines, err := strconv.Atoi(tailLinesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tail_lines value: %s, must be a number", tailLinesStr)
		}
		logs, err := client.GetPodLogs(ctx, namespace, podName, container, tailLines)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(logs), nil
	}
}

// CreateListEventsTool creates a tool for listing events
func CreateListEventsTool() mcp.Tool {
	return mcp.NewTool("list_events",
		mcp.WithDescription("List events within a namespace or for a specific resource"),
		mcp.WithString("namespace",
			mcp.Description("Namespace to list events from (if not specified, lists from all namespaces)"),
		),
		mcp.WithString("kind",
			mcp.Description("Resource kind to filter events (e.g., Pod, Deployment)"),
		),
		mcp.WithString("name",
			mcp.Description("Resource name to filter events"),
		),
		mcp.WithString("field_selector",
			mcp.Description("Field selector to filter events (format: key1=value1,key2=value2)"),
		),
	)
}

// HandleListEvents handles the list events tool
func HandleListEvents(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.GetString("namespace", "")
		kind := request.GetString("kind", "")
		name := request.GetString("name", "")
		fieldSelector := request.GetString("field_selector", "")

		events, err := client.ListEvents(ctx, namespace, kind, name, fieldSelector)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(events)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
