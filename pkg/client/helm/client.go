package helm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	helmclientlib "github.com/mittwald/go-helm-client"
	valueslib "github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	DefaultTimeout = 5 * time.Minute
)

type ChartSpec struct {
	ReleaseName string
	ChartName   string
	Namespace   string
	Version     string

	CreateNamespace bool
	Atomic          bool
	Wait            bool
	WaitForJobs     bool
	Timeout         time.Duration
	Silent          bool
	UpgradeCRDs     bool

	ValuesYaml  string
	ValueFiles  []string
	SetValues   map[string]string
	SetFileVals map[string]string
	SetJSONVals map[string]string

	RepoURL               string
	Username              string
	Password              string
	CertFile              string
	KeyFile               string
	CaFile                string
	InsecureSkipTLSverify bool
}

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

//go:generate mockery --name=HelmClient --output=. --filename=mocks.go
type HelmClient interface {
	InstallChart(context.Context, *ChartSpec) (*ReleaseInfo, error)
	InstallOrUpgradeChart(context.Context, *ChartSpec) (*ReleaseInfo, error)
	UninstallRelease(context.Context, string, string) error
	AddRepository(context.Context, *RepositoryEntry) error
}

type Client struct {
	inner helmclientlib.Client
}

var _ HelmClient = (*Client)(nil)

func NewClient(kubeConfig, kubeContext string) (*Client, error) {
	return newClient(kubeConfig, kubeContext, nil)
}

func NewClientWithDebug(
	kubeConfig, kubeContext string,
	debugFunc func(string, ...interface{}),
) (*Client, error) {
	return newClient(kubeConfig, kubeContext, debugFunc)
}

func newClient(
	kubeConfig, kubeContext string,
	debug func(string, ...interface{}),
) (*Client, error) {
	inner, err := createHelmClient(kubeConfig, kubeContext, debug)
	if err != nil {
		return nil, err
	}

	return &Client{inner: inner}, nil
}

func createHelmClient(
	kubeConfig, kubeContext string,
	debug func(string, ...interface{}),
) (helmclientlib.Client, error) {
	debugLog := debug
	if debugLog == nil {
		debugLog = func(string, ...interface{}) {}
	}

	options := &helmclientlib.Options{
		Linting:  false,
		Debug:    debug != nil,
		DebugLog: debugLog,
	}

	if kubeConfig != "" {
		if data, err := os.ReadFile(kubeConfig); err == nil {
			client, err := helmclientlib.NewClientFromKubeConf(&helmclientlib.KubeConfClientOptions{
				Options:     options,
				KubeConfig:  data,
				KubeContext: kubeContext,
			})
			if err == nil {
				return client, nil
			}
		}
	}

	client, err := helmclientlib.New(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create helm client: %w", err)
	}

	settings := client.GetSettings()
	if kubeConfig != "" {
		settings.KubeConfig = kubeConfig
	}
	if kubeContext != "" {
		settings.KubeContext = kubeContext
	}

	if hc, ok := client.(*helmclientlib.HelmClient); ok {
		if err := reinitActionConfig(hc); err != nil {
			return nil, fmt.Errorf("failed to initialize helm action config: %w", err)
		}
	}

	return client, nil
}

func reinitActionConfig(hc *helmclientlib.HelmClient) error {
	return hc.ActionConfig.Init(
		hc.Settings.RESTClientGetter(),
		hc.Settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		hc.DebugLog,
	)
}

func (c *Client) concreteClient() (*helmclientlib.HelmClient, error) {
	hc, ok := c.inner.(*helmclientlib.HelmClient)
	if !ok {
		return nil, errors.New("unsupported helm client implementation")
	}
	return hc, nil
}

func (c *Client) switchNamespace(namespace string) (func(), error) {
	if namespace == "" {
		return func() {}, nil
	}

	hc, err := c.concreteClient()
	if err != nil {
		return nil, err
	}

	settings := hc.Settings
	previous := settings.Namespace()
	if previous == namespace {
		return func() {}, nil
	}

	settings.SetNamespace(namespace)
	if err := reinitActionConfig(hc); err != nil {
		settings.SetNamespace(previous)
		_ = reinitActionConfig(hc)
		return nil, fmt.Errorf("failed to set helm namespace %q: %w", namespace, err)
	}

	return func() {
		settings.SetNamespace(previous)
		if err := reinitActionConfig(hc); err != nil {
			hc.DebugLog("failed to restore helm namespace: %v", err)
		}
	}, nil
}

func (c *Client) InstallChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error) {
	return c.installRelease(ctx, spec, false)
}

func (c *Client) InstallOrUpgradeChart(ctx context.Context, spec *ChartSpec) (*ReleaseInfo, error) {
	return c.installRelease(ctx, spec, true)
}

func (c *Client) installRelease(
	ctx context.Context,
	spec *ChartSpec,
	upgrade bool,
) (*ReleaseInfo, error) {
	return c.executeReleaseOp(
		ctx,
		spec,
		true,
		func(ctx context.Context, chartSpec *helmclientlib.ChartSpec) (*release.Release, error) {
			if upgrade {
				return c.inner.InstallOrUpgradeChart(ctx, chartSpec, nil)
			}

			return c.inner.InstallChart(ctx, chartSpec, nil)
		},
	)
}

func (c *Client) UninstallRelease(ctx context.Context, releaseName, namespace string) error {
	if releaseName == "" {
		return errors.New("release name is required")
	}

	cleanup, err := c.switchNamespace(namespace)
	if err != nil {
		return err
	}
	defer cleanup()

	chartSpec := &helmclientlib.ChartSpec{
		ReleaseName: releaseName,
		Namespace:   namespace,
	}

	return c.inner.UninstallRelease(chartSpec)
}

func (c *Client) AddRepository(ctx context.Context, entry *RepositoryEntry) error {
	if entry == nil {
		return errors.New("repository entry is required")
	}
	if entry.Name == "" {
		return errors.New("repository name is required")
	}

	settings := c.inner.GetSettings()
	repoFile := settings.RepositoryConfig
	if repoFile == "" {
		return errors.New("helm repository config path is not set")
	}

	if err := os.MkdirAll(filepath.Dir(repoFile), 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		f = repo.NewFile()
	}

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

	f.Update(repoEntry)

	if err := f.WriteFile(repoFile, 0o644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	return nil
}

func (c *Client) executeReleaseOp(
	ctx context.Context,
	spec *ChartSpec,
	applyDefaultTimeout bool,
	operation func(context.Context, *helmclientlib.ChartSpec) (*release.Release, error),
) (*ReleaseInfo, error) {
	chartSpec, cleanup, err := c.prepareChartSpec(spec, applyDefaultTimeout)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	run := func() (*release.Release, error) {
		return operation(ctx, chartSpec)
	}

	var rel *release.Release
	if spec != nil && spec.Silent {
		rel, err = runWithSilencedStderr(run)
	} else {
		rel, err = run()
	}
	if err != nil {
		return nil, err
	}

	return releaseToInfo(rel), nil
}

func (c *Client) prepareChartSpec(
	spec *ChartSpec,
	applyDefaultTimeout bool,
) (*helmclientlib.ChartSpec, func(), error) {
	if spec == nil {
		return nil, nil, errors.New("chart spec is required")
	}

	chartSpec := convertChartSpec(spec)
	if applyDefaultTimeout && chartSpec.Timeout == 0 {
		chartSpec.Timeout = DefaultTimeout
	}

	if err := c.ensureRepository(spec, chartSpec); err != nil {
		return nil, nil, err
	}

	cleanup, err := c.switchNamespace(chartSpec.Namespace)
	if err != nil {
		return nil, nil, err
	}

	return chartSpec, cleanup, nil
}

func (c *Client) ensureRepository(spec *ChartSpec, chartSpec *helmclientlib.ChartSpec) error {
	if spec.RepoURL == "" {
		return nil
	}

	_, chartName, err := parseChartRef(spec.ChartName)
	if err != nil {
		return fmt.Errorf("failed to parse chart reference: %w", err)
	}
	if chartName == "" {
		chartName = spec.ChartName
	}

	settings := c.inner.GetSettings()
	chartURL, err := repo.FindChartInAuthAndTLSAndPassRepoURL(
		spec.RepoURL,
		spec.Username,
		spec.Password,
		chartName,
		spec.Version,
		spec.CertFile,
		spec.KeyFile,
		spec.CaFile,
		spec.InsecureSkipTLSverify,
		false,
		getter.All(settings),
	)
	if err != nil {
		return fmt.Errorf(
			"failed to locate chart %q in repository %s: %w",
			chartName,
			spec.RepoURL,
			err,
		)
	}

	chartSpec.ChartName = chartURL
	return nil
}

func convertChartSpec(spec *ChartSpec) *helmclientlib.ChartSpec {
	return &helmclientlib.ChartSpec{
		ReleaseName: spec.ReleaseName,
		ChartName:   spec.ChartName,
		Namespace:   spec.Namespace,
		ValuesYaml:  spec.ValuesYaml,
		ValuesOptions: valueslib.Options{
			ValueFiles:   copyStringSlice(spec.ValueFiles),
			StringValues: convertMapToSlice(spec.SetValues),
			FileValues:   convertMapToSlice(spec.SetFileVals),
			JSONValues:   convertMapToSlice(spec.SetJSONVals),
		},
		Version:         spec.Version,
		CreateNamespace: spec.CreateNamespace,
		Wait:            spec.Wait,
		WaitForJobs:     spec.WaitForJobs,
		Timeout:         spec.Timeout,
		Atomic:          spec.Atomic,
		UpgradeCRDs:     spec.UpgradeCRDs,
	}
}

func convertMapToSlice(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]string, 0, len(keys))
	for _, k := range keys {
		result = append(result, fmt.Sprintf("%s=%s", k, m[k]))
	}

	return result
}

func copyStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dup := make([]string, len(src))
	copy(dup, src)
	return dup
}

func parseChartRef(chartRef string) (string, string, error) {
	parts := strings.SplitN(chartRef, "/", 2)
	if len(parts) == 1 {
		return "", parts[0], nil
	}
	return parts[0], parts[1], nil
}

func releaseToInfo(rel *release.Release) *ReleaseInfo {
	if rel == nil {
		return nil
	}

	return &ReleaseInfo{
		Name:       rel.Name,
		Namespace:  rel.Namespace,
		Revision:   rel.Version,
		Status:     rel.Info.Status.String(),
		Chart:      rel.Chart.Metadata.Name,
		AppVersion: rel.Chart.Metadata.AppVersion,
		Updated:    rel.Info.LastDeployed.Time,
		Notes:      rel.Info.Notes,
	}
}

func runWithSilencedStderr[T any](fn func() (T, error)) (result T, err error) {
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return fn()
	}

	originalStderr := os.Stderr
	var buffer bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(&buffer, r)
	}()

	os.Stderr = w

	defer func() {
		_ = w.Close()
		wg.Wait()
		_ = r.Close()
		os.Stderr = originalStderr

		if err != nil {
			logs := strings.TrimSpace(buffer.String())
			if logs != "" {
				err = fmt.Errorf("%w: %s", err, logs)
			}
		}
	}()

	result, err = fn()

	return result, err
}
