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

// HelmClient provides Helm operation functionality
type HelmClient struct {
	settings  *cli.EnvSettings
	config    *action.Configuration
	namespace string
}

// HelmRelease represents Helm deployment information
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

// HelmRepository represents a Helm repository
type HelmRepository struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// NewHelmClient creates a Helm client
func NewHelmClient(namespace string, kubeconfigPath string) (*HelmClient, error) {
	settings := cli.New()

	// If kubeconfig path is provided, set it
	if kubeconfigPath != "" {
		settings.KubeConfig = kubeconfigPath
	}

	// Set default namespace
	if namespace != "" {
		settings.SetNamespace(namespace)
	} else {
		settings.SetNamespace(DefaultNamespace)
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		fmt.Printf(format, v...)
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm config: %w", err)
	}

	return &HelmClient{
		settings:  settings,
		config:    actionConfig,
		namespace: settings.Namespace(),
	}, nil
}

// SetNamespace sets the namespace for Helm client operations
func (c *HelmClient) SetNamespace(namespace string) error {
	c.settings.SetNamespace(namespace)
	c.namespace = namespace

	// Re-initialize config
	if err := c.config.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		fmt.Printf(format, v...)
	}); err != nil {
		return fmt.Errorf("failed to update Helm config namespace: %w", err)
	}

	return nil
}

// ListReleases lists all deployed Helm charts
func (c *HelmClient) ListReleases(allNamespaces bool) ([]HelmRelease, error) {
	client := action.NewList(c.config)

	// Whether to list releases from all namespaces
	if allNamespaces {
		client.AllNamespaces = true
	}

	results, err := client.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list Helm releases: %w", err)
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

// GetRelease gets detailed information about a specific Helm release
func (c *HelmClient) GetRelease(name string) (*HelmRelease, error) {
	client := action.NewGet(c.config)

	release, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get Helm release: %w", err)
	}

	// Get values
	var values map[string]interface{}
	// Merge user-provided values and computed values
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

// InstallChart installs a Helm chart
func (c *HelmClient) InstallChart(name, chartName string, values map[string]interface{}, version string, repo string) (*HelmRelease, error) {
	client := action.NewInstall(c.config)
	client.ReleaseName = name
	client.Namespace = c.namespace

	// If version is specified, set it
	if version != "" {
		client.Version = version
	}

	// Parse chart name and repository
	var chartPath string
	if repo != "" {
		// Use specified repository
		chartPath = fmt.Sprintf("%s/%s", repo, chartName)
	} else {
		// Default to local or remote chart
		chartPath = chartName
	}

	// Locate chart
	chartPathOptions := client.ChartPathOptions
	cp, err := chartPathOptions.LocateChart(chartPath, c.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart '%s': %w", chartPath, err)
	}

	// Load chart
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// Install chart
	release, err := client.Run(chartRequested, values)
	if err != nil {
		return nil, fmt.Errorf("failed to install chart: %w", err)
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

// UpgradeChart upgrades a Helm chart
func (c *HelmClient) UpgradeChart(name, chartName string, values map[string]interface{}, version string, repo string) (*HelmRelease, error) {
	client := action.NewUpgrade(c.config)

	// If version is specified, set it
	if version != "" {
		client.Version = version
	}

	// Parse chart name and repository
	var chartPath string
	if repo != "" {
		// Use specified repository
		chartPath = fmt.Sprintf("%s/%s", repo, chartName)
	} else {
		// Default to local or remote chart
		chartPath = chartName
	}

	// Locate chart
	chartPathOptions := client.ChartPathOptions
	cp, err := chartPathOptions.LocateChart(chartPath, c.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart '%s': %w", chartPath, err)
	}

	// Load chart
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// Upgrade chart
	release, err := client.Run(name, chartRequested, values)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade chart: %w", err)
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

// UninstallChart uninstalls a Helm chart
func (c *HelmClient) UninstallChart(name string) error {
	client := action.NewUninstall(c.config)

	_, err := client.Run(name)
	if err != nil {
		return fmt.Errorf("failed to uninstall chart: %w", err)
	}

	return nil
}

// RollbackRelease rolls back a Helm release to a specified revision
func (c *HelmClient) RollbackRelease(name string, revision int) error {
	client := action.NewRollback(c.config)
	client.Version = revision

	return client.Run(name)
}

// GetReleaseHistory gets the history of a Helm release
func (c *HelmClient) GetReleaseHistory(name string) ([]HelmRelease, error) {
	client := action.NewHistory(c.config)

	hist, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release history: %w", err)
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

// AddRepository adds a Helm repository
func (c *HelmClient) AddRepository(repository *HelmRepository) error {
	// Create repository entry
	entry := &repo.Entry{
		Name:     repository.Name,
		URL:      repository.URL,
		Username: repository.Username,
		Password: repository.Password,
	}

	// Get repository file path
	repoFile := c.settings.RepositoryConfig

	// Ensure repository directory exists
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create repository config directory: %w", err)
	}

	// Load repository file
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		// If file doesn't exist, create a new file
		if os.IsNotExist(err) {
			f = repo.NewFile()
		} else {
			return err
		}
	}

	// Check if repository with same name already exists
	if f.Has(entry.Name) {
		// Update existing entry
		f.Update(entry)
	} else {
		// Add new entry
		f.Add(entry)
	}

	// Save repository file
	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("failed to save repository config: %w", err)
	}

	return nil
}

// RemoveRepository removes a Helm repository
func (c *HelmClient) RemoveRepository(name string) error {
	// Get repository file path
	repoFile := c.settings.RepositoryConfig

	// Load repository file
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository config: %w", err)
	}

	// Check if repository exists
	if !f.Has(name) {
		return fmt.Errorf("repository %s does not exist", name)
	}

	// Remove repository
	f.Remove(name)

	// Save repository file
	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("failed to save repository config: %w", err)
	}

	return nil
}

// ListRepositories lists all Helm repositories
func (c *HelmClient) ListRepositories() ([]*HelmRepository, error) {
	// Get repository file path
	repoFile := c.settings.RepositoryConfig

	// Load repository file
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*HelmRepository{}, nil
		}
		return nil, fmt.Errorf("failed to load repository config: %w", err)
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

// ParseYamlValues parses YAML format values into a map
func ParseYamlValues(valuesYaml string) (map[string]interface{}, error) {
	if valuesYaml == "" {
		return map[string]interface{}{}, nil
	}

	values := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(valuesYaml), &values)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML values: %w", err)
	}

	return values, nil
}

// GetChartInfo gets detailed information about a chart
func (c *HelmClient) GetChartInfo(name string, version string, repo string) (*chart.Metadata, error) {
	client := action.NewShowWithConfig(action.ShowChart, c.config)
	chartPathOptions := client.ChartPathOptions
	chartPathOptions.Version = version

	// Parse chart name and repository
	var chartPath string
	if repo != "" {
		// Use specified repository
		chartPath = fmt.Sprintf("%s/%s", repo, name)
	} else {
		// Default to local or remote chart
		chartPath = name
	}

	// Locate chart
	cp, err := chartPathOptions.LocateChart(chartPath, c.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart '%s': %w", chartPath, err)
	}

	// Load chart
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	return chartRequested.Metadata, nil
}

// SearchCharts searches for charts in Helm repositories
func (c *HelmClient) SearchCharts(keyword string, repoName string) ([]*HelmRepository, error) {
	// Search logic needs to be implemented here
	// Since search in Helm v3 API is complex, temporarily return all repository list
	return c.ListRepositories()
}
