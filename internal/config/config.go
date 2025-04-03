package config

import (
	"fmt"
	"os"
)

// Config 表示应用程序配置
type Config struct {
	// Kubeconfig 文件路径
	KubeconfigPath string
	// 是否启用资源创建操作
	EnableCreate bool
	// 是否启用资源更新操作
	EnableUpdate bool
	// 是否启用资源删除操作
	EnableDelete bool
}

// NewConfig 从命令行参数创建配置
func NewConfig(kubeconfigPath string, enableCreate, enableUpdate, enableDelete bool) *Config {
	return &Config{
		KubeconfigPath: kubeconfigPath,
		EnableCreate:   enableCreate,
		EnableUpdate:   enableUpdate,
		EnableDelete:   enableDelete,
	}
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	// 检查 kubeconfig 是否可访问
	if c.KubeconfigPath != "" {
		_, err := os.Stat(c.KubeconfigPath)
		if err != nil {
			return fmt.Errorf("无法访问 kubeconfig 文件: %w", err)
		}
	}
	return nil
} 