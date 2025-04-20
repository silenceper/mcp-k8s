package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/silenceper/mcp-k8s/internal/k8s"
)

// CreateListHelmReleasesTool 创建列出所有 Helm Releases 的工具
func CreateListHelmReleasesTool() mcp.Tool {
	return mcp.NewTool("list_helm_releases",
		mcp.WithDescription("列出所有已安装的 Helm charts"),
		mcp.WithBoolean("all_namespaces",
			mcp.Description("是否列出所有命名空间的 releases"),
			mcp.DefaultBool(false),
		),
	)
}

// CreateGetHelmReleaseTool 创建获取单个 Helm Release 的工具
func CreateGetHelmReleaseTool() mcp.Tool {
	return mcp.NewTool("get_helm_release",
		mcp.WithDescription("获取指定 Helm release 的详细信息"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Release 的名称"),
		),
	)
}

// CreateInstallHelmChartTool 创建安装 Helm Chart 的工具
func CreateInstallHelmChartTool() mcp.Tool {
	return mcp.NewTool("install_helm_chart",
		mcp.WithDescription("安装 Helm chart"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("要安装的 Release 名称"),
		),
		mcp.WithString("chart",
			mcp.Required(),
			mcp.Description("Chart 名称"),
		),
		mcp.WithString("namespace",
			mcp.Description("目标命名空间（如果不指定，使用当前命名空间）"),
		),
		mcp.WithString("version",
			mcp.Description("Chart 版本"),
		),
		mcp.WithString("repo",
			mcp.Description("仓库名称（如果不指定，使用本地或远程 chart）"),
		),
		mcp.WithString("values",
			mcp.Description("YAML 格式的值，将覆盖默认值"),
		),
	)
}

// CreateUpgradeHelmChartTool 创建升级 Helm Chart 的工具
func CreateUpgradeHelmChartTool() mcp.Tool {
	return mcp.NewTool("upgrade_helm_chart",
		mcp.WithDescription("升级 Helm chart"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("要升级的 Release 名称"),
		),
		mcp.WithString("chart",
			mcp.Required(),
			mcp.Description("Chart 名称"),
		),
		mcp.WithString("namespace",
			mcp.Description("Release 所在的命名空间（如果不指定，使用当前命名空间）"),
		),
		mcp.WithString("version",
			mcp.Description("Chart 版本"),
		),
		mcp.WithString("repo",
			mcp.Description("仓库名称（如果不指定，使用本地或远程 chart）"),
		),
		mcp.WithString("values",
			mcp.Description("YAML 格式的值，将覆盖默认值"),
		),
	)
}

// CreateUninstallHelmChartTool 创建卸载 Helm Chart 的工具
func CreateUninstallHelmChartTool() mcp.Tool {
	return mcp.NewTool("uninstall_helm_chart",
		mcp.WithDescription("卸载 Helm chart"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("要卸载的 Release 名称"),
		),
		mcp.WithString("namespace",
			mcp.Description("Release 所在的命名空间（如果不指定，使用当前命名空间）"),
		),
	)
}

// CreateListHelmRepositoriesTool 创建列出 Helm 仓库的工具
func CreateListHelmRepositoriesTool() mcp.Tool {
	return mcp.NewTool("list_helm_repos",
		mcp.WithDescription("列出所有配置的 Helm 仓库"),
	)
}

// CreateAddHelmRepositoryTool 创建添加 Helm 仓库的工具
func CreateAddHelmRepositoryTool() mcp.Tool {
	return mcp.NewTool("add_helm_repo",
		mcp.WithDescription("添加 Helm 仓库"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("仓库名称"),
		),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("仓库 URL"),
		),
		mcp.WithString("username",
			mcp.Description("访问仓库的用户名（如果需要）"),
		),
		mcp.WithString("password",
			mcp.Description("访问仓库的密码（如果需要）"),
		),
	)
}

// CreateRemoveHelmRepositoryTool 创建移除 Helm 仓库的工具
func CreateRemoveHelmRepositoryTool() mcp.Tool {
	return mcp.NewTool("remove_helm_repo",
		mcp.WithDescription("移除 Helm 仓库"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("仓库名称"),
		),
	)
}

// GetHelmClient 获取共享的 Helm 客户端
func GetHelmClient(client *k8s.Client, namespace string) (*k8s.HelmClient, error) {
	return k8s.NewHelmClient(namespace, client.GetKubeconfigPath())
}

// HandleListHelmReleases 处理列出 Helm Releases 的请求
func HandleListHelmReleases(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		allNamespaces := false
		if val, ok := request.Params.Arguments["all_namespaces"].(bool); ok {
			allNamespaces = val
		}

		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 列出所有 releases
		releases, err := helmClient.ListReleases(allNamespaces)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(releases)
		if err != nil {
			return nil, fmt.Errorf("序列化响应失败: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetHelmRelease 处理获取单个 Helm Release 的请求
func HandleGetHelmRelease(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("缺少必需的参数: name")
		}

		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 获取 release 详情
		release, err := helmClient.GetRelease(name)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("序列化响应失败: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleInstallHelmChart 处理安装 Helm Chart 的请求
func HandleInstallHelmChart(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("缺少必需的参数: name")
		}

		chart, ok := request.Params.Arguments["chart"].(string)
		if !ok || chart == "" {
			return nil, fmt.Errorf("缺少必需的参数: chart")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)
		version, _ := request.Params.Arguments["version"].(string)
		repo, _ := request.Params.Arguments["repo"].(string)
		valuesYaml, _ := request.Params.Arguments["values"].(string)

		// 解析 YAML 格式的值
		values, err := k8s.ParseYamlValues(valuesYaml)
		if err != nil {
			return nil, err
		}

		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, namespace)
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 安装 chart
		release, err := helmClient.InstallChart(name, chart, values, version, repo)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("序列化响应失败: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleUpgradeHelmChart 处理升级 Helm Chart 的请求
func HandleUpgradeHelmChart(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("缺少必需的参数: name")
		}

		chart, ok := request.Params.Arguments["chart"].(string)
		if !ok || chart == "" {
			return nil, fmt.Errorf("缺少必需的参数: chart")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)
		version, _ := request.Params.Arguments["version"].(string)
		repo, _ := request.Params.Arguments["repo"].(string)
		valuesYaml, _ := request.Params.Arguments["values"].(string)

		// 解析 YAML 格式的值
		values, err := k8s.ParseYamlValues(valuesYaml)
		if err != nil {
			return nil, err
		}

		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, namespace)
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 升级 chart
		release, err := helmClient.UpgradeChart(name, chart, values, version, repo)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("序列化响应失败: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleUninstallHelmChart 处理卸载 Helm Chart 的请求
func HandleUninstallHelmChart(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("缺少必需的参数: name")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, namespace)
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 卸载 chart
		err = helmClient.UninstallChart(name)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("成功卸载 Helm release: %s", name)), nil
	}
}

// HandleListHelmRepositories 处理列出 Helm 仓库的请求
func HandleListHelmRepositories(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 列出所有仓库
		repos, err := helmClient.ListRepositories()
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(repos)
		if err != nil {
			return nil, fmt.Errorf("序列化响应失败: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleAddHelmRepository 处理添加 Helm 仓库的请求
func HandleAddHelmRepository(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("缺少必需的参数: name")
		}

		url, ok := request.Params.Arguments["url"].(string)
		if !ok || url == "" {
			return nil, fmt.Errorf("缺少必需的参数: url")
		}

		username, _ := request.Params.Arguments["username"].(string)
		password, _ := request.Params.Arguments["password"].(string)

		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 添加仓库
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

		return mcp.NewToolResultText(fmt.Sprintf("成功添加 Helm 仓库: %s", name)), nil
	}
}

// HandleRemoveHelmRepository 处理移除 Helm 仓库的请求
func HandleRemoveHelmRepository(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("缺少必需的参数: name")
		}

		// 获取 Helm 客户端
		helmClient, err := GetHelmClient(client, "")
		if err != nil {
			return nil, fmt.Errorf("创建 Helm 客户端失败: %w", err)
		}

		// 移除仓库
		err = helmClient.RemoveRepository(name)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("成功移除 Helm 仓库: %s", name)), nil
	}
} 