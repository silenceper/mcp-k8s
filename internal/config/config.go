package config

import (
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	// Kubeconfig file path
	KubeconfigPath string
	// Whether to enable resource creation operations
	EnableCreate bool
	// Whether to enable resource update operations
	EnableUpdate bool
	// Whether to enable resource deletion operations
	EnableDelete bool
	// Whether to enable resource list operations
	EnableList bool
	// Whether to enable Helm install operations
	EnableHelmInstall bool
	// Whether to enable Helm upgrade operations
	EnableHelmUpgrade bool
	// Whether to enable Helm uninstall operations
	EnableHelmUninstall bool
	// Whether to enable Helm repository add operations
	EnableHelmRepoAdd bool
	// Whether to enable Helm repository remove operations
	EnableHelmRepoRemove bool
	// Whether to enable Helm release list operations
	EnableHelmReleaseList bool
	// Whether to enable Helm release get operations
	EnableHelmReleaseGet bool
	// Whether to enable Helm repository list operations
	EnableHelmRepoList bool
}

// NewConfig creates a configuration from command line arguments
func NewConfig(kubeconfigPath string, enableCreate, enableUpdate, enableDelete, enableList bool) *Config {
	return &Config{
		KubeconfigPath: kubeconfigPath,
		EnableCreate:   enableCreate,
		EnableUpdate:   enableUpdate,
		EnableDelete:   enableDelete,
		EnableList:     enableList,
	}
}

// Validate validates whether the configuration is valid
func (c *Config) Validate() error {
	// Check if kubeconfig is accessible
	if c.KubeconfigPath != "" {
		_, err := os.Stat(c.KubeconfigPath)
		if err != nil {
			return fmt.Errorf("cannot access kubeconfig file: %w", err)
		}
	}
	return nil
}

// InitHelmDefaults initializes Helm-related default configuration
func (c *Config) InitHelmDefaults() {
	// Read operations are enabled by default
	c.EnableHelmReleaseList = true
	c.EnableHelmReleaseGet = true
	c.EnableHelmRepoList = true

	// Write operations are disabled by default
	c.EnableHelmInstall = false
	c.EnableHelmUpgrade = false
	c.EnableHelmUninstall = false
	c.EnableHelmRepoAdd = false
	c.EnableHelmRepoRemove = false
}
