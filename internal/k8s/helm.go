package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

// HelmClient 提供 Helm 操作功能
type HelmClient struct {
	settings  *cli.EnvSettings
	config    *action.Configuration
	namespace string
}

// HelmRelease 表示 Helm 部署的信息
type HelmRelease struct {
	Name         string                 `json:"name"`
	Namespace    string                 `json:"namespace"`
	Revision     int                    `json:"revision"`
	Status       string                 `json:"status"`
	Chart        string                 `json:"chart"`
	ChartVersion string                 `json:"chartVersion"`
	AppVersion   string                 `json:"appVersion,omitempty"`
	Updated      time.Time              `json:"updated"`
	Values       map[string]interface{} `json:"values,omitempty"`
}

// HelmRepository 表示一个 Helm 仓库
type HelmRepository struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// NewHelmClient 创建一个 Helm 客户端
func NewHelmClient(namespace string, kubeconfigPath string) (*HelmClient, error) {
	settings := cli.New()
	
	// 如果提供了 kubeconfig 路径，设置它
	if kubeconfigPath != "" {
		settings.KubeConfig = kubeconfigPath
	}
	
	// 设置默认命名空间
	if namespace != "" {
		settings.SetNamespace(namespace)
	} else {
		settings.SetNamespace("default")
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		fmt.Printf(format, v...)
	}); err != nil {
		return nil, fmt.Errorf("初始化 Helm 配置失败: %w", err)
	}

	return &HelmClient{
		settings:  settings,
		config:    actionConfig,
		namespace: settings.Namespace(),
	}, nil
}

// SetNamespace 设置 Helm 客户端操作的命名空间
func (c *HelmClient) SetNamespace(namespace string) error {
	c.settings.SetNamespace(namespace)
	c.namespace = namespace
	
	// 重新初始化配置
	if err := c.config.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		fmt.Printf(format, v...)
	}); err != nil {
		return fmt.Errorf("更新 Helm 配置命名空间失败: %w", err)
	}
	
	return nil
}

// ListReleases 列出所有部署的 Helm charts
func (c *HelmClient) ListReleases(allNamespaces bool) ([]HelmRelease, error) {
	client := action.NewList(c.config)
	
	// 是否列出所有命名空间的 releases
	if allNamespaces {
		client.AllNamespaces = true
	}
	
	results, err := client.Run()
	if err != nil {
		return nil, fmt.Errorf("列出 Helm releases 失败: %w", err)
	}
	
	releases := []HelmRelease{}
	for _, r := range results {
		releases = append(releases, HelmRelease{
			Name:         r.Name,
			Namespace:    r.Namespace,
			Revision:     r.Version,
			Status:       r.Info.Status.String(),
			Chart:        r.Chart.Metadata.Name,
			ChartVersion: r.Chart.Metadata.Version,
			AppVersion:   r.Chart.Metadata.AppVersion,
			Updated:      r.Info.LastDeployed.Time,
		})
	}
	
	return releases, nil
}

// GetRelease 获取特定 Helm release 的详细信息
func (c *HelmClient) GetRelease(name string) (*HelmRelease, error) {
	client := action.NewGet(c.config)
	
	release, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("获取 Helm release 失败: %w", err)
	}
	
	// 获取值
	var values map[string]interface{}
	// 合并用户提供的值和计算的值
	if release.Config != nil && release.Chart.Values != nil {
		values = release.Config
	}
	
	return &HelmRelease{
		Name:         release.Name,
		Namespace:    release.Namespace,
		Revision:     release.Version,
		Status:       release.Info.Status.String(),
		Chart:        release.Chart.Metadata.Name,
		ChartVersion: release.Chart.Metadata.Version,
		AppVersion:   release.Chart.Metadata.AppVersion,
		Updated:      release.Info.LastDeployed.Time,
		Values:       values,
	}, nil
}

// InstallChart 安装 Helm chart
func (c *HelmClient) InstallChart(name, chartName string, values map[string]interface{}, version string, repo string) (*HelmRelease, error) {
	client := action.NewInstall(c.config)
	client.ReleaseName = name
	client.Namespace = c.namespace
	
	// 如果指定了版本，设置版本
	if version != "" {
		client.Version = version
	}
	
	// 解析 chart 名称和仓库
	var chartPath string
	if repo != "" {
		// 使用指定的仓库
		chartPath = fmt.Sprintf("%s/%s", repo, chartName)
	} else {
		// 默认使用本地或远程 chart
		chartPath = chartName
	}
	
	// 定位 chart
	cp, err := client.ChartPathOptions.LocateChart(chartPath, c.settings)
	if err != nil {
		return nil, fmt.Errorf("查找 chart '%s' 失败: %w", chartPath, err)
	}
	
	// 加载 chart
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("加载 chart 失败: %w", err)
	}
	
	// 安装 chart
	release, err := client.Run(chartRequested, values)
	if err != nil {
		return nil, fmt.Errorf("安装 chart 失败: %w", err)
	}
	
	return &HelmRelease{
		Name:         release.Name,
		Namespace:    release.Namespace,
		Revision:     release.Version,
		Status:       release.Info.Status.String(),
		Chart:        release.Chart.Metadata.Name,
		ChartVersion: release.Chart.Metadata.Version,
		AppVersion:   release.Chart.Metadata.AppVersion,
		Updated:      release.Info.LastDeployed.Time,
		Values:       values,
	}, nil
}

// UpgradeChart 升级 Helm chart
func (c *HelmClient) UpgradeChart(name, chartName string, values map[string]interface{}, version string, repo string) (*HelmRelease, error) {
	client := action.NewUpgrade(c.config)
	
	// 如果指定了版本，设置版本
	if version != "" {
		client.Version = version
	}
	
	// 解析 chart 名称和仓库
	var chartPath string
	if repo != "" {
		// 使用指定的仓库
		chartPath = fmt.Sprintf("%s/%s", repo, chartName)
	} else {
		// 默认使用本地或远程 chart
		chartPath = chartName
	}
	
	// 定位 chart
	cp, err := client.ChartPathOptions.LocateChart(chartPath, c.settings)
	if err != nil {
		return nil, fmt.Errorf("查找 chart '%s' 失败: %w", chartPath, err)
	}
	
	// 加载 chart
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("加载 chart 失败: %w", err)
	}
	
	// 升级 chart
	release, err := client.Run(name, chartRequested, values)
	if err != nil {
		return nil, fmt.Errorf("升级 chart 失败: %w", err)
	}
	
	return &HelmRelease{
		Name:         release.Name,
		Namespace:    release.Namespace,
		Revision:     release.Version,
		Status:       release.Info.Status.String(),
		Chart:        release.Chart.Metadata.Name,
		ChartVersion: release.Chart.Metadata.Version,
		AppVersion:   release.Chart.Metadata.AppVersion,
		Updated:      release.Info.LastDeployed.Time,
		Values:       values,
	}, nil
}

// UninstallChart 卸载 Helm chart
func (c *HelmClient) UninstallChart(name string) error {
	client := action.NewUninstall(c.config)
	
	_, err := client.Run(name)
	if err != nil {
		return fmt.Errorf("卸载 chart 失败: %w", err)
	}
	
	return nil
}

// RollbackRelease 回滚 Helm release 到指定版本
func (c *HelmClient) RollbackRelease(name string, revision int) error {
	client := action.NewRollback(c.config)
	client.Version = revision
	
	return client.Run(name)
}

// GetReleaseHistory 获取 Helm release 的历史记录
func (c *HelmClient) GetReleaseHistory(name string) ([]HelmRelease, error) {
	client := action.NewHistory(c.config)
	
	hist, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("获取 release 历史记录失败: %w", err)
	}
	
	releases := []HelmRelease{}
	for _, r := range hist {
		releases = append(releases, HelmRelease{
			Name:         r.Name,
			Namespace:    r.Namespace,
			Revision:     r.Version,
			Status:       r.Info.Status.String(),
			Chart:        r.Chart.Metadata.Name,
			ChartVersion: r.Chart.Metadata.Version,
			AppVersion:   r.Chart.Metadata.AppVersion,
			Updated:      r.Info.LastDeployed.Time,
		})
	}
	
	return releases, nil
}

// AddRepository 添加 Helm 仓库
func (c *HelmClient) AddRepository(repository *HelmRepository) error {
	// 创建存储库条目
	entry := &repo.Entry{
		Name:     repository.Name,
		URL:      repository.URL,
		Username: repository.Username,
		Password: repository.Password,
	}
	
	// 获取存储库文件路径
	repoFile := c.settings.RepositoryConfig
	
	// 确保存储库目录存在
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("创建 repository 配置目录失败: %w", err)
	}
	
	// 加载存储库文件
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		// 如果文件不存在，创建一个新的文件
		if os.IsNotExist(err) {
			f = repo.NewFile()
		} else {
			return err
		}
	}
	
	// 检查是否已存在同名存储库
	if f.Has(entry.Name) {
		// 更新存在的条目
		f.Update(entry)
	} else {
		// 添加新条目
		f.Add(entry)
	}
	
	// 保存存储库文件
	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("保存 repository 配置失败: %w", err)
	}
	
	return nil
}

// RemoveRepository 移除 Helm 仓库
func (c *HelmClient) RemoveRepository(name string) error {
	// 获取存储库文件路径
	repoFile := c.settings.RepositoryConfig
	
	// 加载存储库文件
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("加载 repository 配置失败: %w", err)
	}
	
	// 检查是否存在该存储库
	if !f.Has(name) {
		return fmt.Errorf("repository %s 不存在", name)
	}
	
	// 移除存储库
	f.Remove(name)
	
	// 保存存储库文件
	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("保存 repository 配置失败: %w", err)
	}
	
	return nil
}

// ListRepositories 列出所有 Helm 仓库
func (c *HelmClient) ListRepositories() ([]*HelmRepository, error) {
	// 获取存储库文件路径
	repoFile := c.settings.RepositoryConfig
	
	// 加载存储库文件
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*HelmRepository{}, nil
		}
		return nil, fmt.Errorf("加载 repository 配置失败: %w", err)
	}
	
	repos := []*HelmRepository{}
	for _, entry := range f.Repositories {
		repos = append(repos, &HelmRepository{
			Name: entry.Name,
			URL:  entry.URL,
		})
	}
	
	return repos, nil
}

// ParseYamlValues 解析 YAML 格式的值为 map
func ParseYamlValues(valuesYaml string) (map[string]interface{}, error) {
	if valuesYaml == "" {
		return map[string]interface{}{}, nil
	}
	
	values := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(valuesYaml), &values)
	if err != nil {
		return nil, fmt.Errorf("解析 YAML 值失败: %w", err)
	}
	
	return values, nil
}

// GetChartInfo 获取 chart 的详细信息
func (c *HelmClient) GetChartInfo(name string, version string, repo string) (*chart.Metadata, error) {
	client := action.NewShowWithConfig(action.ShowChart, c.config)
	client.ChartPathOptions.Version = version
	
	// 解析 chart 名称和仓库
	var chartPath string
	if repo != "" {
		// 使用指定的仓库
		chartPath = fmt.Sprintf("%s/%s", repo, name)
	} else {
		// 默认使用本地或远程 chart
		chartPath = name
	}
	
	// 定位 chart
	cp, err := client.ChartPathOptions.LocateChart(chartPath, c.settings)
	if err != nil {
		return nil, fmt.Errorf("查找 chart '%s' 失败: %w", chartPath, err)
	}
	
	// 加载 chart
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("加载 chart 失败: %w", err)
	}
	
	return chartRequested.Metadata, nil
}

// SearchCharts 搜索 Helm 仓库中的 charts
func (c *HelmClient) SearchCharts(keyword string, repoName string) ([]*HelmRepository, error) {
	// 这里需要实现搜索逻辑
	// 由于 Helm v3 API 中搜索有些复杂，暂时只返回所有仓库列表
	return c.ListRepositories()
} 