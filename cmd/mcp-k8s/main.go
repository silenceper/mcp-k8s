package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/silenceper/mcp-k8s/internal/config"
	"github.com/silenceper/mcp-k8s/internal/k8s"
	"github.com/silenceper/mcp-k8s/internal/tools"
	"github.com/spf13/cobra"
)

var (
	kubeconfigPath        string
	enableCreate          bool
	enableUpdate          bool
	enableDelete          bool
	enableList            bool
	enableHelmInstall     bool
	enableHelmUpgrade     bool
	enableHelmUninstall   bool
	enableHelmRepoAdd     bool
	enableHelmRepoRemove  bool
	enableHelmReleaseList bool
	enableHelmReleaseGet  bool
	enableHelmRepoList    bool
	transport             string
	host                  string
	port                  int
	endpointPath          string
)

var (
	// version is injected at build time via ldflags
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-k8s",
	Short: "A Kubernetes MCP (Model Control Protocol) server",
	Long: `mcp-k8s is a Kubernetes MCP server that enables interaction with Kubernetes clusters 
through MCP tools. It supports stdio, SSE, and Streamable HTTP transport modes.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	Run:     runServer,
}

func init() {
	// Kubernetes resource operations
	rootCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "Path to Kubernetes configuration file (uses default config if not specified)")
	rootCmd.Flags().BoolVar(&enableCreate, "enable-create", false, "Enable resource creation operations")
	rootCmd.Flags().BoolVar(&enableUpdate, "enable-update", false, "Enable resource update operations")
	rootCmd.Flags().BoolVar(&enableDelete, "enable-delete", false, "Enable resource deletion operations")
	rootCmd.Flags().BoolVar(&enableList, "enable-list", true, "Enable resource list operations")

	// Helm operations
	rootCmd.Flags().BoolVar(&enableHelmInstall, "enable-helm-install", false, "Enable Helm install operations")
	rootCmd.Flags().BoolVar(&enableHelmUpgrade, "enable-helm-upgrade", false, "Enable Helm upgrade operations")
	rootCmd.Flags().BoolVar(&enableHelmUninstall, "enable-helm-uninstall", false, "Enable Helm uninstall operations")
	rootCmd.Flags().BoolVar(&enableHelmRepoAdd, "enable-helm-repo-add", false, "Enable Helm repository add operations")
	rootCmd.Flags().BoolVar(&enableHelmRepoRemove, "enable-helm-repo-remove", false, "Enable Helm repository remove operations")
	rootCmd.Flags().BoolVar(&enableHelmReleaseList, "enable-helm-release-list", true, "Enable Helm release list operations")
	rootCmd.Flags().BoolVar(&enableHelmReleaseGet, "enable-helm-release-get", true, "Enable Helm release get operations")
	rootCmd.Flags().BoolVar(&enableHelmRepoList, "enable-helm-repo-list", true, "Enable Helm repository list operations")

	// Transport configuration
	rootCmd.Flags().StringVar(&transport, "transport", "stdio", "Transport type (stdio, sse, or streamable-http)")
	rootCmd.Flags().StringVar(&host, "host", "localhost", "Host for HTTP transport (SSE or Streamable HTTP)")
	rootCmd.Flags().IntVar(&port, "port", 8080, "TCP port for HTTP transport (SSE or Streamable HTTP)")
	rootCmd.Flags().StringVar(&endpointPath, "endpoint-path", "/mcp", "Endpoint path for Streamable HTTP transport")
}

func runServer(cmd *cobra.Command, args []string) {
	// Create configuration
	cfg := config.NewConfig(kubeconfigPath, enableCreate, enableUpdate, enableDelete, enableList)

	// Set Helm-related configuration
	// Initialize Helm default configuration
	cfg.InitHelmDefaults()

	// Override default configuration with command line arguments
	cfg.EnableHelmInstall = enableHelmInstall
	cfg.EnableHelmUpgrade = enableHelmUpgrade
	cfg.EnableHelmUninstall = enableHelmUninstall
	cfg.EnableHelmRepoAdd = enableHelmRepoAdd
	cfg.EnableHelmRepoRemove = enableHelmRepoRemove
	cfg.EnableHelmReleaseList = enableHelmReleaseList
	cfg.EnableHelmReleaseGet = enableHelmReleaseGet
	cfg.EnableHelmRepoList = enableHelmRepoList

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
		version,
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

	// Helm Release management - read operations
	if cfg.EnableHelmReleaseList {
		fmt.Println("Registering Helm release list tool...")
		s.AddTool(tools.CreateListHelmReleasesTool(), tools.HandleListHelmReleases(client))
	}

	if cfg.EnableHelmReleaseGet {
		fmt.Println("Registering Helm release get tool...")
		s.AddTool(tools.CreateGetHelmReleaseTool(), tools.HandleGetHelmRelease(client))
	}

	// Helm Release management - write operations
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

	// Helm repository management - read operations
	if cfg.EnableHelmRepoList {
		fmt.Println("Registering Helm repository list tool...")
		s.AddTool(tools.CreateListHelmRepositoriesTool(), tools.HandleListHelmRepositories(client))
	}

	// Helm repository management - write operations
	if cfg.EnableHelmRepoAdd {
		fmt.Println("Registering Helm repository add tool...")
		s.AddTool(tools.CreateAddHelmRepositoryTool(), tools.HandleAddHelmRepository(client))
	}

	if cfg.EnableHelmRepoRemove {
		fmt.Println("Registering Helm repository remove tool...")
		s.AddTool(tools.CreateRemoveHelmRepositoryTool(), tools.HandleRemoveHelmRepository(client))
	}

	// Output functionality status
	fmt.Printf("\nStarting Kubernetes MCP Server with %s transport on %s:%d\n", transport, host, port)
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

	// Start server based on transport type
	fmt.Println("\nServer started, waiting for MCP client connections...")
	switch transport {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	case "sse":
		sseUrl := fmt.Sprintf("http://%s:%d", host, port)
		sseServer := server.NewSSEServer(s, server.WithBaseURL(sseUrl))
		if err := sseServer.Start(fmt.Sprintf(":%d", port)); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case "streamable-http":
		streamableUrl := fmt.Sprintf("http://%s:%d%s", host, port, endpointPath)
		fmt.Printf("Streamable HTTP endpoint: %s\n", streamableUrl)
		streamableServer := server.NewStreamableHTTPServer(s, server.WithEndpointPath(endpointPath))
		if err := streamableServer.Start(fmt.Sprintf(":%d", port)); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown transport type: %s. Supported types: stdio, sse, streamable-http\n", transport)
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
