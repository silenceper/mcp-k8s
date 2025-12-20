package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/silenceper/mcp-k8s/internal/k8s"
)

// CreateListHelmReleasesTool creates a tool for listing all Helm Releases
func CreateListHelmReleasesTool() mcp.Tool {
	return mcp.NewTool("list_helm_releases",
		mcp.WithDescription("List all installed Helm charts"),
		mcp.WithBoolean("all_namespaces",
			mcp.Description("Whether to list releases from all namespaces"),
			mcp.DefaultBool(false),
		),
	)
}

// CreateGetHelmReleaseTool creates a tool for getting a single Helm Release
func CreateGetHelmReleaseTool() mcp.Tool {
	return mcp.NewTool("get_helm_release",
		mcp.WithDescription("Get detailed information about a specific Helm release"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the release"),
		),
	)
}

// CreateInstallHelmChartTool creates a tool for installing a Helm Chart
func CreateInstallHelmChartTool() mcp.Tool {
	return mcp.NewTool("install_helm_chart",
		mcp.WithDescription("Install a Helm chart"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the release to install"),
		),
		mcp.WithString("chart",
			mcp.Required(),
			mcp.Description("Chart name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Target namespace (if not specified, use current namespace)"),
		),
		mcp.WithString("version",
			mcp.Description("Chart version"),
		),
		mcp.WithString("repo",
			mcp.Description("Repository name (if not specified, use local or remote chart)"),
		),
		mcp.WithString("values",
			mcp.Description("YAML format values that will override default values"),
		),
	)
}

// CreateUpgradeHelmChartTool creates a tool for upgrading a Helm Chart
func CreateUpgradeHelmChartTool() mcp.Tool {
	return mcp.NewTool("upgrade_helm_chart",
		mcp.WithDescription("Upgrade a Helm chart"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the release to upgrade"),
		),
		mcp.WithString("chart",
			mcp.Required(),
			mcp.Description("Chart name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace where the release is located (if not specified, use current namespace)"),
		),
		mcp.WithString("version",
			mcp.Description("Chart version"),
		),
		mcp.WithString("repo",
			mcp.Description("Repository name (if not specified, use local or remote chart)"),
		),
		mcp.WithString("values",
			mcp.Description("YAML format values that will override default values"),
		),
	)
}

// CreateUninstallHelmChartTool creates a tool for uninstalling a Helm Chart
func CreateUninstallHelmChartTool() mcp.Tool {
	return mcp.NewTool("uninstall_helm_chart",
		mcp.WithDescription("Uninstall a Helm chart"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the release to uninstall"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace where the release is located (if not specified, use current namespace)"),
		),
	)
}

// CreateListHelmRepositoriesTool creates a tool for listing Helm repositories
func CreateListHelmRepositoriesTool() mcp.Tool {
	return mcp.NewTool("list_helm_repos",
		mcp.WithDescription("List all configured Helm repositories"),
	)
}

// CreateAddHelmRepositoryTool creates a tool for adding a Helm repository
func CreateAddHelmRepositoryTool() mcp.Tool {
	return mcp.NewTool("add_helm_repo",
		mcp.WithDescription("Add a Helm repository"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("Repository URL"),
		),
		mcp.WithString("username",
			mcp.Description("Username to access the repository (if needed)"),
		),
		mcp.WithString("password",
			mcp.Description("Password to access the repository (if needed)"),
		),
	)
}

// CreateRemoveHelmRepositoryTool creates a tool for removing a Helm repository
func CreateRemoveHelmRepositoryTool() mcp.Tool {
	return mcp.NewTool("remove_helm_repo",
		mcp.WithDescription("Remove a Helm repository"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
	)
}

// GetHelmClient gets a shared Helm client
func GetHelmClient(client *k8s.Client, namespace string) (*k8s.HelmClient, error) {
	return k8s.NewHelmClient(namespace, client.GetKubeconfigPath())
}

// HandleListHelmReleases handles the request to list Helm Releases
func HandleListHelmReleases(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		allNamespaces := request.GetBool("all_namespaces", false)

		// Get Helm client
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// List all releases
		releases, err := helmClient.ListReleases(allNamespaces)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(releases)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetHelmRelease handles the request to get a single Helm Release
func HandleGetHelmRelease(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		// Get Helm client
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// Get release details
		release, err := helmClient.GetRelease(name)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleInstallHelmChart handles the request to install a Helm Chart
func HandleInstallHelmChart(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		chart, err := request.RequireString("chart")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: chart: %w", err)
		}

		namespace := request.GetString("namespace", "")
		version := request.GetString("version", "")
		repo := request.GetString("repo", "")
		valuesYaml := request.GetString("values", "")

		// Parse YAML format values
		values, err := k8s.ParseYamlValues(valuesYaml)
		if err != nil {
			return nil, err
		}

		// Get Helm client
		helmClient, err := GetHelmClient(client, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// Install chart
		release, err := helmClient.InstallChart(name, chart, values, version, repo)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleUpgradeHelmChart handles the request to upgrade a Helm Chart
func HandleUpgradeHelmChart(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		chart, err := request.RequireString("chart")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: chart: %w", err)
		}

		namespace := request.GetString("namespace", "")
		version := request.GetString("version", "")
		repo := request.GetString("repo", "")
		valuesYaml := request.GetString("values", "")

		// Parse YAML format values
		values, err := k8s.ParseYamlValues(valuesYaml)
		if err != nil {
			return nil, err
		}

		// Get Helm client
		helmClient, err := GetHelmClient(client, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// Upgrade chart
		release, err := helmClient.UpgradeChart(name, chart, values, version, repo)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleUninstallHelmChart handles the request to uninstall a Helm Chart
func HandleUninstallHelmChart(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		namespace := request.GetString("namespace", "")

		// Get Helm client
		helmClient, err := GetHelmClient(client, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// Uninstall chart
		err = helmClient.UninstallChart(name)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully uninstalled Helm release: %s", name)), nil
	}
}

// HandleListHelmRepositories handles the request to list Helm repositories
func HandleListHelmRepositories(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get Helm client
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// List all repositories
		repos, err := helmClient.ListRepositories()
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(repos)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleAddHelmRepository handles the request to add a Helm repository
func HandleAddHelmRepository(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		url, err := request.RequireString("url")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: url: %w", err)
		}

		username := request.GetString("username", "")
		password := request.GetString("password", "")

		// Get Helm client
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// Add repository
		repo := &k8s.HelmRepository{
			Name:     name,
			URL:      url,
			Username: username,
			Password: password,
		}

		err = helmClient.AddRepository(repo)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully added Helm repository: %s", name)), nil
	}
}

// HandleRemoveHelmRepository handles the request to remove a Helm repository
func HandleRemoveHelmRepository(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return nil, fmt.Errorf("missing required parameter: name: %w", err)
		}

		// Get Helm client
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create Helm client: %w", err)
		}

		// Remove repository
		err = helmClient.RemoveRepository(name)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully removed Helm repository: %s", name)), nil
	}
}
