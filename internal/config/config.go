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
	// 是否启用资源列表操作
	EnableList bool
	// 是否启用 Helm 安装操作
	EnableHelmInstall bool
	// 是否启用 Helm 升级操作
	EnableHelmUpgrade bool
	// 是否启用 Helm 卸载操作
	EnableHelmUninstall bool
	// 是否启用 Helm 仓库添加操作
	EnableHelmRepoAdd bool
	// 是否启用 Helm 仓库删除操作
	EnableHelmRepoRemove bool
	// 是否启用 Helm 发布版列表操作
	EnableHelmReleaseList bool
	// 是否启用 Helm 发布版查询操作
	EnableHelmReleaseGet bool
	// 是否启用 Helm 仓库列表操作
	EnableHelmRepoList bool
}

// NewConfig 从命令行参数创建配置
func NewConfig(kubeconfigPath string, enableCreate, enableUpdate, enableDelete, enableList bool) *Config {
	return &Config{
		KubeconfigPath: kubeconfigPath,
		EnableCreate:   enableCreate,
		EnableUpdate:   enableUpdate,
		EnableDelete:   enableDelete,
		EnableList:     enableList,
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

// InitHelmDefaults 初始化Helm相关的默认配置
func (c *Config) InitHelmDefaults() {
	// 读操作默认开启
	c.EnableHelmReleaseList = true
	c.EnableHelmReleaseGet = true
	c.EnableHelmRepoList = true
	
	// 写操作默认关闭
	c.EnableHelmInstall = false
	c.EnableHelmUpgrade = false
	c.EnableHelmUninstall = false
	c.EnableHelmRepoAdd = false
	c.EnableHelmRepoRemove = false
}
