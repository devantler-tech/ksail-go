package cni

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CNIInstallerBase provides common fields and methods for CNI installers.
type CNIInstallerBase struct {
	kubeconfig string
	context    string
	timeout    time.Duration
	client     helm.Interface
	waitFn     func(context.Context) error
}

// NewCNIInstallerBase creates a new base installer instance.
func NewCNIInstallerBase(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
	waitFn func(context.Context) error,
) *CNIInstallerBase {
	return &CNIInstallerBase{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
		waitFn:     waitFn,
	}
}

// WaitForReadiness waits for the CNI components to become ready.
func (b *CNIInstallerBase) WaitForReadiness(ctx context.Context) error {
	if b.waitFn == nil {
		return nil
	}

	err := b.waitFn(ctx)
	if err != nil {
		return fmt.Errorf("wait for readiness: %w", err)
	}

	return nil
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (b *CNIInstallerBase) SetWaitForReadinessFunc(
	waitFunc func(context.Context) error,
	defaultWaitFn func(context.Context) error,
) {
	if waitFunc == nil {
		b.waitFn = defaultWaitFn

		return
	}

	b.waitFn = waitFunc
}

// BuildRESTConfig builds a Kubernetes REST configuration.
func (b *CNIInstallerBase) BuildRESTConfig() (*rest.Config, error) {
	config, err := k8sutil.BuildRESTConfig(b.kubeconfig, b.context)
	if err != nil {
		return nil, fmt.Errorf("build REST config: %w", err)
	}

	return config, nil
}

var errHelmClientNil = errors.New("helm client is nil")

// GetClient returns the Helm client.
//
//nolint:ireturn // Method returns interface by design for flexibility.
func (b *CNIInstallerBase) GetClient() (helm.Interface, error) {
	if b.client == nil {
		return nil, errHelmClientNil
	}

	return b.client, nil
}

// GetTimeout returns the timeout duration.
func (b *CNIInstallerBase) GetTimeout() time.Duration {
	return b.timeout
}

// GetKubeconfig returns the kubeconfig path.
func (b *CNIInstallerBase) GetKubeconfig() string {
	return b.kubeconfig
}

// GetContext returns the kubeconfig context.
func (b *CNIInstallerBase) GetContext() string {
	return b.context
}

// GetWaitFn returns the wait function for testing purposes.
// This method is primarily used in tests to verify wait function behavior.
func (b *CNIInstallerBase) GetWaitFn() func(context.Context) error {
	return b.waitFn
}

// SetWaitFn sets the wait function directly for testing purposes.
// This is a low-level method used primarily in tests. Prefer using SetWaitForReadinessFunc for production code.
func (b *CNIInstallerBase) SetWaitFn(fn func(context.Context) error) {
	b.waitFn = fn
}

// HelmRepoConfig holds repository configuration for a Helm chart.
type HelmRepoConfig struct {
	// Name is the repository identifier used in Helm commands.
	Name string
	// URL is the Helm repository URL.
	URL string
	// RepoName is the human-readable name used in error messages.
	RepoName string
}

// HelmChartConfig holds chart installation configuration.
type HelmChartConfig struct {
	// ReleaseName is the Helm release name.
	ReleaseName string
	// ChartName is the chart identifier (e.g., "repo/chart").
	ChartName string
	// Namespace is the Kubernetes namespace for installation.
	Namespace string
	// RepoURL is the Helm repository URL.
	RepoURL string
	// CreateNamespace determines if the namespace should be created.
	CreateNamespace bool
	// SetJSONVals contains JSON values to set during installation.
	SetJSONVals map[string]string
}

// InstallOrUpgradeHelmChart performs a Helm install or upgrade operation.
func InstallOrUpgradeHelmChart(
	ctx context.Context,
	client helm.Interface,
	repoConfig HelmRepoConfig,
	chartConfig HelmChartConfig,
	timeout time.Duration,
) error {
	repoEntry := &helm.RepositoryEntry{
		Name: repoConfig.Name,
		URL:  repoConfig.URL,
	}

	addRepoErr := client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add %s repository: %w", repoConfig.RepoName, addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     chartConfig.ReleaseName,
		ChartName:       chartConfig.ChartName,
		Namespace:       chartConfig.Namespace,
		RepoURL:         chartConfig.RepoURL,
		CreateNamespace: chartConfig.CreateNamespace,
		Atomic:          true,
		Silent:          true,
		UpgradeCRDs:     true,
		Timeout:         timeout,
		Wait:            true,
		WaitForJobs:     true,
		SetJSONVals:     chartConfig.SetJSONVals,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install %s chart: %w", repoConfig.RepoName, err)
	}

	return nil
}

// WaitForResourceReadiness waits for multiple Kubernetes resources to become ready.
func WaitForResourceReadiness(
	ctx context.Context,
	kubeconfig, context string,
	checks []k8sutil.ReadinessCheck,
	timeout time.Duration,
	componentName string,
) error {
	restConfig, err := k8sutil.BuildRESTConfig(kubeconfig, context)
	if err != nil {
		return fmt.Errorf("build kubernetes client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	err = k8sutil.WaitForMultipleResources(ctx, clientset, checks, timeout)
	if err != nil {
		return fmt.Errorf("wait for %s components: %w", componentName, err)
	}

	return nil
}
