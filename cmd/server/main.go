package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/silenceper/mcp-k8s/internal/config"
	"github.com/silenceper/mcp-k8s/internal/k8s"
	"github.com/silenceper/mcp-k8s/internal/tools"
)

func main() {
	// Parse command line arguments
	kubeconfigPath := flag.String("kubeconfig", "", "Path to Kubernetes configuration file (uses default config if not specified)")
	enableCreate := flag.Bool("enable-create", false, "Enable resource creation operations")
	enableUpdate := flag.Bool("enable-update", false, "Enable resource update operations")
	enableDelete := flag.Bool("enable-delete", false, "Enable resource deletion operations")
	enableList := flag.Bool("enable-list", true, "Enable resource list operations")
	
	// Helm 操作的细粒度控制
	enableHelmInstall := flag.Bool("enable-helm-install", false, "Enable Helm install operations")
	enableHelmUpgrade := flag.Bool("enable-helm-upgrade", false, "Enable Helm upgrade operations")
	enableHelmUninstall := flag.Bool("enable-helm-uninstall", false, "Enable Helm uninstall operations")
	enableHelmRepoAdd := flag.Bool("enable-helm-repo-add", false, "Enable Helm repository add operations")
	enableHelmRepoRemove := flag.Bool("enable-helm-repo-remove", false, "Enable Helm repository remove operations")
	enableHelmReleaseList := flag.Bool("enable-helm-release-list", true, "Enable Helm release list operations")
	enableHelmReleaseGet := flag.Bool("enable-helm-release-get", true, "Enable Helm release get operations")
	enableHelmRepoList := flag.Bool("enable-helm-repo-list", true, "Enable Helm repository list operations")
	
	transport := flag.String("transport", "stdio", "Transport type (stdio or sse)")
	host := flag.String("host", "localhost", "Host for SSE transport")
	port := flag.Int("port", 8080, "TCP port for SSE transport")
	flag.Parse()

	// Create configuration
	cfg := config.NewConfig(*kubeconfigPath, *enableCreate, *enableUpdate, *enableDelete, *enableList)
	
	// 设置Helm相关配置
	// 初始化Helm默认配置
	cfg.InitHelmDefaults()
	
	// 使用命令行参数覆盖默认配置
	cfg.EnableHelmInstall = *enableHelmInstall
	cfg.EnableHelmUpgrade = *enableHelmUpgrade
	cfg.EnableHelmUninstall = *enableHelmUninstall
	cfg.EnableHelmRepoAdd = *enableHelmRepoAdd
	cfg.EnableHelmRepoRemove = *enableHelmRepoRemove
	cfg.EnableHelmReleaseList = *enableHelmReleaseList
	cfg.EnableHelmReleaseGet = *enableHelmReleaseGet
	cfg.EnableHelmRepoList = *enableHelmRepoList
	
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(cfg.KubeconfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"Kubernetes MCP Server",
		"1.0.0",
	)

	// Add basic tools
	fmt.Println("Registering basic tools...")
	s.AddTool(tools.CreateGetAPIResourcesTool(), tools.HandleGetAPIResources(client))
	s.AddTool(tools.CreateGetResourceTool(), tools.HandleGetResource(client))
	if cfg.EnableList {
		s.AddTool(tools.CreateListResourcesTool(), tools.HandleListResources(client))
	}

	// Add write operation tools (if enabled)
	if cfg.EnableCreate {
		fmt.Println("Registering resource creation tool...")
		s.AddTool(tools.CreateCreateResourceTool(), tools.HandleCreateResource(client))
	}

	if cfg.EnableUpdate {
		fmt.Println("Registering resource update tool...")
		s.AddTool(tools.CreateUpdateResourceTool(), tools.HandleUpdateResource(client))
	}

	if cfg.EnableDelete {
		fmt.Println("Registering resource deletion tool...")
		s.AddTool(tools.CreateDeleteResourceTool(), tools.HandleDeleteResource(client))
	}
	
	// Add Helm tools (if enabled)
	fmt.Println("Registering Helm tools...")
	
	// Helm Release 管理 - 读操作
	if cfg.EnableHelmReleaseList {
		fmt.Println("Registering Helm release list tool...")
		s.AddTool(tools.CreateListHelmReleasesTool(), tools.HandleListHelmReleases(client))
	}
	
	if cfg.EnableHelmReleaseGet {
		fmt.Println("Registering Helm release get tool...")
		s.AddTool(tools.CreateGetHelmReleaseTool(), tools.HandleGetHelmRelease(client))
	}
	
	// Helm Release 管理 - 写操作
	if cfg.EnableHelmInstall {
		fmt.Println("Registering Helm chart install tool...")
		s.AddTool(tools.CreateInstallHelmChartTool(), tools.HandleInstallHelmChart(client))
	}
	
	if cfg.EnableHelmUpgrade {
		fmt.Println("Registering Helm chart upgrade tool...")
		s.AddTool(tools.CreateUpgradeHelmChartTool(), tools.HandleUpgradeHelmChart(client))
	}
	
	if cfg.EnableHelmUninstall {
		fmt.Println("Registering Helm chart uninstall tool...")
		s.AddTool(tools.CreateUninstallHelmChartTool(), tools.HandleUninstallHelmChart(client))
	}
	
	// Helm 仓库管理 - 读操作
	if cfg.EnableHelmRepoList {
		fmt.Println("Registering Helm repository list tool...")
		s.AddTool(tools.CreateListHelmRepositoriesTool(), tools.HandleListHelmRepositories(client))
	}
	
	// Helm 仓库管理 - 写操作
	if cfg.EnableHelmRepoAdd {
		fmt.Println("Registering Helm repository add tool...")
		s.AddTool(tools.CreateAddHelmRepositoryTool(), tools.HandleAddHelmRepository(client))
	}
	
	if cfg.EnableHelmRepoRemove {
		fmt.Println("Registering Helm repository remove tool...")
		s.AddTool(tools.CreateRemoveHelmRepositoryTool(), tools.HandleRemoveHelmRepository(client))
	}

	// Output functionality status
	fmt.Printf("\nStarting Kubernetes MCP Server with %s transport on %s:%d\n", *transport, *host, *port)
	fmt.Printf("Create operations: %v\n", cfg.EnableCreate)
	fmt.Printf("Update operations: %v\n", cfg.EnableUpdate)
	fmt.Printf("Delete operations: %v\n", cfg.EnableDelete)
	fmt.Printf("List operations: %v\n", cfg.EnableList)
	
	fmt.Println("\nHelm operations details:")
	fmt.Printf("  Helm release list: %v\n", cfg.EnableHelmReleaseList)
	fmt.Printf("  Helm release get: %v\n", cfg.EnableHelmReleaseGet)
	fmt.Printf("  Helm install: %v\n", cfg.EnableHelmInstall)
	fmt.Printf("  Helm upgrade: %v\n", cfg.EnableHelmUpgrade)
	fmt.Printf("  Helm uninstall: %v\n", cfg.EnableHelmUninstall)
	fmt.Printf("  Helm repository list: %v\n", cfg.EnableHelmRepoList)
	fmt.Printf("  Helm repository add: %v\n", cfg.EnableHelmRepoAdd)
	fmt.Printf("  Helm repository remove: %v\n", cfg.EnableHelmRepoRemove)

	// Start stdio server
	fmt.Println("\nServer started, waiting for MCP client connections...\n")
	if *transport == "stdio" {
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	} else if *transport == "sse" {
		sseUrl := fmt.Sprintf("http://%s:%d", *host, *port)
		sseServer := server.NewSSEServer(s, server.WithBaseURL(sseUrl))
		if err := sseServer.Start(fmt.Sprintf(":%d", *port)); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
