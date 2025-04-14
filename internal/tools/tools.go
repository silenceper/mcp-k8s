package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/silenceper/mcp-k8s/internal/k8s"
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
		includeNamespaceScoped := true
		includeClusterScoped := true

		if val, ok := request.Params.Arguments["includeNamespaceScoped"].(bool); ok {
			includeNamespaceScoped = val
		}

		if val, ok := request.Params.Arguments["includeClusterScoped"].(bool); ok {
			includeClusterScoped = val
		}

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
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: name")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

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
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)
		labelSelector, _ := request.Params.Arguments["labelSelector"].(string)
		fieldSelector, _ := request.Params.Arguments["fieldSelector"].(string)

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
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		manifest, ok := request.Params.Arguments["manifest"].(string)
		if !ok || manifest == "" {
			return nil, fmt.Errorf("missing required parameter: manifest")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

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
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: name")
		}

		manifest, ok := request.Params.Arguments["manifest"].(string)
		if !ok || manifest == "" {
			return nil, fmt.Errorf("missing required parameter: manifest")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

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
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: name")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		err := client.DeleteResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted resource %s/%s", kind, name)), nil
	}
}
