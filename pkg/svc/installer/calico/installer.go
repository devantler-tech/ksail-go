package calicoinstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CalicoInstaller implements the installer.Installer interface for Calico.
type CalicoInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
	client     helm.Interface
	waitFn     func(context.Context) error
}

// NewCalicoInstaller creates a new Calico installer instance.
func NewCalicoInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *CalicoInstaller {
	installer := &CalicoInstaller{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
	}

	installer.waitFn = installer.waitForReadiness

	return installer
}

// Install installs or upgrades Calico via its Helm chart.
func (c *CalicoInstaller) Install(ctx context.Context) error {
	err := c.helmInstallOrUpgradeCalico(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Calico: %w", err)
	}

	return nil
}

// WaitForReadiness waits for the Calico components to become ready.
func (c *CalicoInstaller) WaitForReadiness(ctx context.Context) error {
	if c.waitFn == nil {
		return nil
	}

	return c.waitFn(ctx)
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (c *CalicoInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	if waitFunc == nil {
		c.waitFn = c.waitForReadiness

		return
	}

	c.waitFn = waitFunc
}

// Uninstall removes the Helm release for Calico.
func (c *CalicoInstaller) Uninstall(ctx context.Context) error {
	err := c.client.UninstallRelease(ctx, "calico", "tigera-operator")
	if err != nil {
		return fmt.Errorf("failed to uninstall calico release: %w", err)
	}

	return nil
}

// --- internals ---

func (c *CalicoInstaller) helmInstallOrUpgradeCalico(ctx context.Context) error {
	repoEntry := &helm.RepositoryEntry{
		Name: "projectcalico",
		URL:  "https://docs.tigera.io/calico/charts",
	}

	addRepoErr := c.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add calico repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     "calico",
		ChartName:       "projectcalico/tigera-operator",
		Namespace:       "tigera-operator",
		RepoURL:         "https://docs.tigera.io/calico/charts",
		CreateNamespace: true,
		Atomic:          true,
		Silent:          true,
		UpgradeCRDs:     true,
		Timeout:         c.timeout,
		Wait:            true,
		WaitForJobs:     true,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install calico chart: %w", err)
	}

	return nil
}

func (c *CalicoInstaller) waitForReadiness(ctx context.Context) error {
	restConfig, err := k8sutil.BuildRESTConfig(c.kubeconfig, c.context)
	if err != nil {
		return fmt.Errorf("build kubernetes client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Wait for tigera-operator deployment
	deploymentErr := k8sutil.WaitForDeploymentReady(
		waitCtx,
		clientset,
		"tigera-operator",
		"tigera-operator",
		c.timeout,
	)
	if deploymentErr != nil {
		return fmt.Errorf("tigera-operator not ready: %w", deploymentErr)
	}

	// Wait for calico-node daemonset in calico-system namespace
	daemonSetErr := k8sutil.WaitForDaemonSetReady(
		waitCtx,
		clientset,
		"calico-system",
		"calico-node",
		c.timeout,
	)
	if daemonSetErr != nil {
		return fmt.Errorf("calico-node daemonset not ready: %w", daemonSetErr)
	}

	// Wait for calico-kube-controllers deployment
	kubeControllersErr := k8sutil.WaitForDeploymentReady(
		waitCtx,
		clientset,
		"calico-system",
		"calico-kube-controllers",
		c.timeout,
	)
	if kubeControllersErr != nil {
		return fmt.Errorf("calico-kube-controllers not ready: %w", kubeControllersErr)
	}

	return nil
}

func (c *CalicoInstaller) buildRESTConfig() (*rest.Config, error) {
	config, err := k8sutil.BuildRESTConfig(c.kubeconfig, c.context)
	if err != nil {
		return nil, fmt.Errorf("build REST config: %w", err)
	}

	return config, nil
}
