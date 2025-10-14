// Package helm provides a native Go helm client implementation.
package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

const (
	// DefaultTimeout is the default timeout for helm operations.
	DefaultTimeout = 5 * time.Minute
)

// ChartSpec defines the configuration for chart operations.
type ChartSpec struct {
	// Chart identification
	ReleaseName string
	ChartName   string
	Namespace   string
	Version     string

	// Installation options
	CreateNamespace bool
	UpgradeCRDs     bool
	SkipCRDs        bool
	Atomic          bool
	Wait            bool
	WaitForJobs     bool
	Timeout         time.Duration

	// Values configuration
	ValuesYaml  string
	ValueFiles  []string
	SetValues   map[string]string
	SetFileVals map[string]string
	SetJSONVals map[string]string

	// Upgrade options
	Force                bool
	ResetValues          bool
	ReuseValues          bool
	ResetThenReuseValues bool
	MaxHistory           int
	CleanupOnFail        bool

	// Other options
	DisableHooks     bool
	Replace          bool
	DependencyUpdate bool
	DryRun           bool
	Description      string
	GenerateName     bool
	NameTemplate     string
	SubNotes         bool
	KeepHistory      bool
	IgnoreNotFound   bool

	// Repository configuration
	RepoURL               string
	Username              string
	Password              string
	CertFile              string
	KeyFile               string
	CaFile                string
	InsecureSkipTLSverify bool
	PlainHTTP             bool
}

// RepositoryEntry represents a Helm repository entry.
type RepositoryEntry struct {
	Name                  string
	URL                   string
	Username              string
	Password              string
	CertFile              string
	KeyFile               string
	CaFile                string
	InsecureSkipTLSverify bool
	PlainHTTP             bool
}

// ReleaseInfo represents information about a Helm release.
type ReleaseInfo struct {
	Name       string
	Namespace  string
	Revision   int
	Status     string
	Chart      string
	AppVersion string
	Updated    time.Time
	Notes      string
}

// HelmClient defines the interface for Helm operations.
//
//go:generate mockery --name=HelmClient --output=. --filename=mocks.go
type HelmClient interface {
	// Chart operations
	InstallChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error)
	UpgradeChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error)
	InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error)
	UninstallRelease(ctx context.Context, releaseName string, namespace string) error

	// Release management
	GetRelease(ctx context.Context, releaseName string, namespace string) (*ReleaseInfo, error)
	ListReleases(ctx context.Context, namespace string) ([]*ReleaseInfo, error)
	RollbackRelease(ctx context.Context, releaseName string, namespace string, revision int) error

	// Repository operations
	AddRepository(ctx context.Context, entry *RepositoryEntry) error
	UpdateRepositories(ctx context.Context) error
	RemoveRepository(ctx context.Context, name string) error
	ListRepositories(ctx context.Context) ([]*RepositoryEntry, error)

	// Utility operations
	TemplateChart(ctx context.Context, spec *ChartSpec) (string, error)
	GetValues(
		ctx context.Context,
		releaseName string,
		namespace string,
	) (map[string]interface{}, error)
}

// Client implements the HelmClient interface using native Helm v3 actions.
type Client struct {
	settings *cli.EnvSettings
	debug    func(format string, v ...interface{})
}

// NewClient creates a new Helm client instance.
func NewClient(kubeConfig, kubeContext string) (*Client, error) {
	settings := cli.New()
	if kubeConfig != "" {
		settings.KubeConfig = kubeConfig
	}
	if kubeContext != "" {
		settings.KubeContext = kubeContext
	}

	return &Client{
		settings: settings,
		debug:    func(format string, v ...interface{}) {}, // Silent by default
	}, nil
}

// NewClientWithDebug creates a new Helm client with debug logging enabled.
func NewClientWithDebug(
	kubeConfig, kubeContext string,
	debugFunc func(format string, v ...interface{}),
) (*Client, error) {
	client, err := NewClient(kubeConfig, kubeContext)
	if err != nil {
		return nil, err
	}
	if debugFunc != nil {
		client.debug = debugFunc
	}
	return client, nil
}

// getActionConfig initializes and returns an action configuration for the given namespace.
func (c *Client) getActionConfig(namespace string) (*action.Configuration, error) {
	cfg := new(action.Configuration)

	// Use the provided namespace, fall back to settings namespace if empty
	ns := namespace
	if ns == "" {
		ns = c.settings.Namespace()
	}

	err := cfg.Init(
		c.settings.RESTClientGetter(),
		ns,
		os.Getenv("HELM_DRIVER"),
		c.debug,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize helm configuration: %w", err)
	}

	return cfg, nil
}

// InstallChart installs a Helm chart.
func (c *Client) InstallChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error) {
	cfg, err := c.getActionConfig(spec.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action config: %w", err)
	}

	install := action.NewInstall(cfg)
	c.configureInstallAction(install, spec)

	chartObj, err := c.loadChart(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	values, err := c.mergeValues(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %w", err)
	}

	rel, err := install.RunWithContext(ctx, chartObj, values)
	if err != nil {
		return nil, fmt.Errorf("failed to install chart %s: %w", spec.ChartName, err)
	}

	return c.releaseToInfo(rel), nil
}

// UpgradeChart upgrades an existing Helm chart.
func (c *Client) UpgradeChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error) {
	cfg, err := c.getActionConfig(spec.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action config: %w", err)
	}

	upgrade := action.NewUpgrade(cfg)
	c.configureUpgradeAction(upgrade, spec)

	chartObj, err := c.loadChart(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	values, err := c.mergeValues(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %w", err)
	}

	rel, err := upgrade.RunWithContext(ctx, spec.ReleaseName, chartObj, values)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade chart %s: %w", spec.ChartName, err)
	}

	return c.releaseToInfo(rel), nil
}

// InstallOrUpgradeChart installs a chart if it doesn't exist, otherwise upgrades it.
func (c *Client) InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error) {
	// Check if release already exists
	_, err := c.GetRelease(ctx, spec.ReleaseName, spec.Namespace)
	if err != nil {
		// Release doesn't exist, install it
		return c.InstallChart(ctx, spec)
	}

	// Release exists, upgrade it
	return c.UpgradeChart(ctx, spec)
}

// UninstallRelease uninstalls a Helm release.
func (c *Client) UninstallRelease(ctx context.Context, releaseName, namespace string) error {
	cfg, err := c.getActionConfig(namespace)
	if err != nil {
		return fmt.Errorf("failed to get action config: %w", err)
	}

	uninstall := action.NewUninstall(cfg)
	uninstall.KeepHistory = false

	_, err = uninstall.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall release %s: %w", releaseName, err)
	}

	return nil
}

// GetRelease retrieves information about a specific release.
func (c *Client) GetRelease(
	ctx context.Context,
	releaseName, namespace string,
) (*ReleaseInfo, error) {
	cfg, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action config: %w", err)
	}

	get := action.NewGet(cfg)
	rel, err := get.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s: %w", releaseName, err)
	}

	return c.releaseToInfo(rel), nil
}

// ListReleases lists all releases in the specified namespace.
func (c *Client) ListReleases(ctx context.Context, namespace string) ([]*ReleaseInfo, error) {
	cfg, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action config: %w", err)
	}

	list := action.NewList(cfg)
	list.All = true

	releases, err := list.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	infos := make([]*ReleaseInfo, len(releases))
	for i, rel := range releases {
		infos[i] = c.releaseToInfo(rel)
	}

	return infos, nil
}

// RollbackRelease rolls back a release to a specific revision.
func (c *Client) RollbackRelease(
	ctx context.Context,
	releaseName, namespace string,
	revision int,
) error {
	cfg, err := c.getActionConfig(namespace)
	if err != nil {
		return fmt.Errorf("failed to get action config: %w", err)
	}

	rollback := action.NewRollback(cfg)
	rollback.Version = revision

	err = rollback.Run(releaseName)
	if err != nil {
		return fmt.Errorf(
			"failed to rollback release %s to revision %d: %w",
			releaseName,
			revision,
			err,
		)
	}

	return nil
}

// AddRepository adds a Helm repository.
func (c *Client) AddRepository(ctx context.Context, entry *RepositoryEntry) error {
	settings := c.settings

	repoFile := settings.RepositoryConfig
	repoCache := settings.RepositoryCache

	// Ensure the file directory exists
	err := os.MkdirAll(settings.RepositoryCache, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create repository cache directory: %w", err)
	}

	// Load existing repositories
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		f = repo.NewFile()
	}

	// Create repository entry
	repoEntry := &repo.Entry{
		Name:                  entry.Name,
		URL:                   entry.URL,
		Username:              entry.Username,
		Password:              entry.Password,
		CertFile:              entry.CertFile,
		KeyFile:               entry.KeyFile,
		CAFile:                entry.CaFile,
		InsecureSkipTLSverify: entry.InsecureSkipTLSverify,
	}

	// Add repository to index
	chartRepo, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	chartRepo.CachePath = repoCache
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("failed to download repository index: %w", err)
	}

	// Update repo file
	f.Update(repoEntry)
	err = f.WriteFile(repoFile, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	return nil
}

// UpdateRepositories updates all configured repositories.
func (c *Client) UpdateRepositories(ctx context.Context) error {
	settings := c.settings
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository file: %w", err)
	}

	for _, cfg := range f.Repositories {
		chartRepo, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return fmt.Errorf("failed to create chart repository for %s: %w", cfg.Name, err)
		}

		chartRepo.CachePath = settings.RepositoryCache
		_, err = chartRepo.DownloadIndexFile()
		if err != nil {
			return fmt.Errorf("failed to update repository %s: %w", cfg.Name, err)
		}
	}

	return nil
}

// RemoveRepository removes a Helm repository.
func (c *Client) RemoveRepository(ctx context.Context, name string) error {
	settings := c.settings
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository file: %w", err)
	}

	if !f.Remove(name) {
		return fmt.Errorf("repository %s not found", name)
	}

	err = f.WriteFile(repoFile, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	return nil
}

// ListRepositories lists all configured repositories.
func (c *Client) ListRepositories(ctx context.Context) ([]*RepositoryEntry, error) {
	settings := c.settings
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load repository file: %w", err)
	}

	entries := make([]*RepositoryEntry, len(f.Repositories))
	for i, repo := range f.Repositories {
		entries[i] = &RepositoryEntry{
			Name:                  repo.Name,
			URL:                   repo.URL,
			Username:              repo.Username,
			Password:              repo.Password,
			CertFile:              repo.CertFile,
			KeyFile:               repo.KeyFile,
			CaFile:                repo.CAFile,
			InsecureSkipTLSverify: repo.InsecureSkipTLSverify,
		}
	}

	return entries, nil
}

// TemplateChart renders a chart template without installing it.
func (c *Client) TemplateChart(ctx context.Context, spec *ChartSpec) (string, error) {
	cfg, err := c.getActionConfig(spec.Namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get action config: %w", err)
	}

	install := action.NewInstall(cfg)
	install.DryRun = true
	install.ReleaseName = spec.ReleaseName
	install.Namespace = spec.Namespace
	install.Replace = true
	install.ClientOnly = true
	install.APIVersions = []string{}
	install.IncludeCRDs = !spec.SkipCRDs

	chartObj, err := c.loadChart(spec)
	if err != nil {
		return "", fmt.Errorf("failed to load chart: %w", err)
	}

	values, err := c.mergeValues(spec)
	if err != nil {
		return "", fmt.Errorf("failed to merge values: %w", err)
	}

	rel, err := install.RunWithContext(ctx, chartObj, values)
	if err != nil {
		return "", fmt.Errorf("failed to template chart: %w", err)
	}

	return rel.Manifest, nil
}

// GetValues retrieves the values for a specific release.
func (c *Client) GetValues(
	ctx context.Context,
	releaseName, namespace string,
) (map[string]interface{}, error) {
	cfg, err := c.getActionConfig(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action config: %w", err)
	}

	getValues := action.NewGetValues(cfg)
	values, err := getValues.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get values for release %s: %w", releaseName, err)
	}

	return values, nil
}

// Helper methods for chart operations

// configureInstallAction configures the install action with spec parameters.
func (c *Client) configureInstallAction(install *action.Install, spec *ChartSpec) {
	install.ReleaseName = spec.ReleaseName
	install.Namespace = spec.Namespace
	install.CreateNamespace = spec.CreateNamespace
	install.SkipCRDs = spec.SkipCRDs
	install.Atomic = spec.Atomic
	install.Wait = spec.Wait
	install.WaitForJobs = spec.WaitForJobs
	install.Timeout = spec.Timeout
	install.DisableHooks = spec.DisableHooks
	install.Replace = spec.Replace
	install.DependencyUpdate = spec.DependencyUpdate
	install.GenerateName = spec.GenerateName
	install.NameTemplate = spec.NameTemplate
	install.Description = spec.Description
	install.SubNotes = spec.SubNotes
	install.Force = spec.Force
	install.Version = spec.Version
	install.RepoURL = spec.RepoURL
	install.Username = spec.Username
	install.Password = spec.Password
	install.CertFile = spec.CertFile
	install.KeyFile = spec.KeyFile
	install.InsecureSkipTLSverify = spec.InsecureSkipTLSverify
	install.PlainHTTP = spec.PlainHTTP

	if spec.Timeout == 0 {
		install.Timeout = DefaultTimeout
	}

	if spec.DryRun {
		install.DryRun = true
		install.ClientOnly = true
	}
}

// configureUpgradeAction configures the upgrade action with spec parameters.
func (c *Client) configureUpgradeAction(upgrade *action.Upgrade, spec *ChartSpec) {
	upgrade.Namespace = spec.Namespace
	upgrade.SkipCRDs = spec.SkipCRDs
	upgrade.Atomic = spec.Atomic
	upgrade.Wait = spec.Wait
	upgrade.WaitForJobs = spec.WaitForJobs
	upgrade.Timeout = spec.Timeout
	upgrade.DisableHooks = spec.DisableHooks
	upgrade.DependencyUpdate = spec.DependencyUpdate
	upgrade.Description = spec.Description
	upgrade.SubNotes = spec.SubNotes
	upgrade.Force = spec.Force
	upgrade.ResetValues = spec.ResetValues
	upgrade.ReuseValues = spec.ReuseValues
	upgrade.ResetThenReuseValues = spec.ResetThenReuseValues
	upgrade.MaxHistory = spec.MaxHistory
	upgrade.CleanupOnFail = spec.CleanupOnFail
	upgrade.Version = spec.Version
	upgrade.RepoURL = spec.RepoURL
	upgrade.Username = spec.Username
	upgrade.Password = spec.Password
	upgrade.CertFile = spec.CertFile
	upgrade.KeyFile = spec.KeyFile
	upgrade.InsecureSkipTLSverify = spec.InsecureSkipTLSverify
	upgrade.PlainHTTP = spec.PlainHTTP

	if spec.Timeout == 0 {
		upgrade.Timeout = DefaultTimeout
	}

	if spec.DryRun {
		upgrade.DryRun = true
	}
}

// loadChart loads a chart from the specified chart name.
func (c *Client) loadChart(spec *ChartSpec) (*chart.Chart, error) {
	chartName := spec.ChartName

	// Check if it's an OCI reference
	if strings.HasPrefix(chartName, "oci://") {
		return c.loadOCIChart(spec)
	}

	// Check if it's a local path
	if isLocalPath(chartName) {
		return loader.Load(chartName)
	}

	// It's a repository chart
	return c.loadRepoChart(spec)
}

// loadOCIChart loads a chart from an OCI registry.
func (c *Client) loadOCIChart(spec *ChartSpec) (*chart.Chart, error) {
	cfg, err := c.getActionConfig(spec.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get action config: %w", err)
	}

	pull := action.NewPullWithOpts(action.WithConfig(cfg))
	pull.Settings = c.settings
	pull.Version = spec.Version
	pull.DestDir = os.TempDir()
	pull.Untar = true

	output, err := pull.Run(spec.ChartName)
	if err != nil {
		return nil, fmt.Errorf("failed to pull OCI chart: %w", err)
	}

	return loader.Load(output)
}

// loadRepoChart loads a chart from a repository.
func (c *Client) loadRepoChart(spec *ChartSpec) (*chart.Chart, error) {
	settings := c.settings

	// Parse chart reference
	chartRef := spec.ChartName
	_, chartName, err := parseChartRef(chartRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chart reference: %w", err)
	}

	// Set up downloader
	dl := downloader.ChartDownloader{
		Out:     os.Stderr,
		Keyring: "",
		Getters: getter.All(settings),
		Options: []getter.Option{
			getter.WithBasicAuth(spec.Username, spec.Password),
			getter.WithTLSClientConfig(spec.CertFile, spec.KeyFile, ""),
		},
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
	}

	if spec.InsecureSkipTLSverify {
		dl.Options = append(dl.Options, getter.WithInsecureSkipVerifyTLS(true))
	}

	if spec.RepoURL != "" {
		chartURL, err := repo.FindChartInRepoURL(
			spec.RepoURL,
			chartName,
			spec.Version,
			"",
			"",
			"",
			getter.All(settings),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to find chart in repo URL: %w", err)
		}
		chartRef = chartURL
	}

	// Download chart
	chartPath, _, err := dl.DownloadTo(chartRef, spec.Version, os.TempDir())
	if err != nil {
		return nil, fmt.Errorf("failed to download chart: %w", err)
	}

	return loader.Load(chartPath)
}

// mergeValues merges all value sources into a single values map.
func (c *Client) mergeValues(spec *ChartSpec) (map[string]interface{}, error) {
	base := map[string]interface{}{}

	// Load value files
	for _, filePath := range spec.ValueFiles {
		currentMap := map[string]interface{}{}

		bytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s: %w", filePath, err)
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal values file %s: %w", filePath, err)
		}

		base = mergeMaps(base, currentMap)
	}

	// Merge YAML values
	if spec.ValuesYaml != "" {
		currentMap := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.ValuesYaml), &currentMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal values YAML: %w", err)
		}
		base = mergeMaps(base, currentMap)
	}

	// Process set values using Helm's values package
	valueOpts := &values.Options{
		ValueFiles:   spec.ValueFiles,
		StringValues: convertMapToSlice(spec.SetValues),
		FileValues:   convertMapToSlice(spec.SetFileVals),
		JSONValues:   convertMapToSlice(spec.SetJSONVals),
	}

	// Use Helm's value merging logic
	vals, err := valueOpts.MergeValues(getter.All(c.settings))
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %w", err)
	}

	return vals, nil
}

// releaseToInfo converts a Helm release to ReleaseInfo.
func (c *Client) releaseToInfo(rel *release.Release) *ReleaseInfo {
	info := &ReleaseInfo{
		Name:      rel.Name,
		Namespace: rel.Namespace,
		Revision:  rel.Version,
		Status:    rel.Info.Status.String(),
	}

	if rel.Chart != nil && rel.Chart.Metadata != nil {
		info.Chart = rel.Chart.Metadata.Name + "-" + rel.Chart.Metadata.Version
		info.AppVersion = rel.Chart.Metadata.AppVersion
	}

	if rel.Info != nil {
		if !rel.Info.LastDeployed.IsZero() {
			info.Updated = rel.Info.LastDeployed.Time
		}
		info.Notes = rel.Info.Notes
	}

	return info
}

// Helper utility functions

// isLocalPath checks if a path is a local file system path.
func isLocalPath(path string) bool {
	return strings.HasPrefix(path, "/") ||
		strings.HasPrefix(path, "./") ||
		strings.HasPrefix(path, "../") ||
		filepath.IsAbs(path)
}

// parseChartRef parses a chart reference into repository name and chart name.
func parseChartRef(chartRef string) (string, string, error) {
	parts := strings.SplitN(chartRef, "/", 2)
	if len(parts) == 1 {
		return "", parts[0], nil
	}
	return parts[0], parts[1], nil
}

// convertMapToSlice converts a map[string]string to []string in key=value format.
func convertMapToSlice(m map[string]string) []string {
	var result []string
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// mergeMaps merges two maps, with the second map taking precedence.
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
