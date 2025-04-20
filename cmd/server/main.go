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
	enableHelm := flag.Bool("enable-helm", false, "Enable Helm operations")
	transport := flag.String("transport", "stdio", "Transport type (stdio or sse)")
	host := flag.String("host", "localhost", "Host for SSE transport")
	port := flag.Int("port", 8080, "TCP port for SSE transport")
	flag.Parse()

	// Create configuration
	cfg := config.NewConfig(*kubeconfigPath, *enableCreate, *enableUpdate, *enableDelete, *enableList)
	cfg.EnableHelm = *enableHelm
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
	if cfg.EnableHelm {
		fmt.Println("Registering Helm tools...")
		
		// Helm Release 管理
		s.AddTool(tools.CreateListHelmReleasesTool(), tools.HandleListHelmReleases(client))
		s.AddTool(tools.CreateGetHelmReleaseTool(), tools.HandleGetHelmRelease(client))
		s.AddTool(tools.CreateInstallHelmChartTool(), tools.HandleInstallHelmChart(client))
		s.AddTool(tools.CreateUpgradeHelmChartTool(), tools.HandleUpgradeHelmChart(client))
		s.AddTool(tools.CreateUninstallHelmChartTool(), tools.HandleUninstallHelmChart(client))
		
		// Helm 仓库管理
		s.AddTool(tools.CreateListHelmRepositoriesTool(), tools.HandleListHelmRepositories(client))
		s.AddTool(tools.CreateAddHelmRepositoryTool(), tools.HandleAddHelmRepository(client))
		s.AddTool(tools.CreateRemoveHelmRepositoryTool(), tools.HandleRemoveHelmRepository(client))
	}

	// Output functionality status
	fmt.Printf("\nStarting Kubernetes MCP Server with %s transport on %s:%d\n", *transport, *host, *port)
	fmt.Printf("Create operations: %v\n", cfg.EnableCreate)
	fmt.Printf("Update operations: %v\n", cfg.EnableUpdate)
	fmt.Printf("Delete operations: %v\n", cfg.EnableDelete)
	fmt.Printf("List operations: %v\n", cfg.EnableList)
	fmt.Printf("Helm operations: %v\n", cfg.EnableHelm)

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
