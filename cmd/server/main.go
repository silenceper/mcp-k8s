package main

import (
	"flag"
	"fmt"
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
	flag.Parse()

	// Create configuration
	cfg := config.NewConfig(*kubeconfigPath, *enableCreate, *enableUpdate, *enableDelete)
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
	s.AddTool(tools.CreateListResourcesTool(), tools.HandleListResources(client))
	
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

	// Output functionality status
	fmt.Println("\nStarting Kubernetes MCP Server...")
	fmt.Printf("Read operations: Enabled\n")
	fmt.Printf("Create operations: %v\n", cfg.EnableCreate)
	fmt.Printf("Update operations: %v\n", cfg.EnableUpdate)
	fmt.Printf("Delete operations: %v\n", cfg.EnableDelete)

	// Start stdio server
	fmt.Println("\nServer started, waiting for MCP client connections...\n")
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
} 