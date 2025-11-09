package ciliuminstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CiliumInstaller implements the installer.Installer interface for Cilium.
type CiliumInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
	client     helm.Interface
	waitFn     func(context.Context) error
}

// NewCiliumInstaller creates a new Cilium installer instance.
func NewCiliumInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *CiliumInstaller {
	installer := &CiliumInstaller{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
	}

	installer.waitFn = installer.waitForReadiness

	return installer
}

// Install installs or upgrades Cilium via its Helm chart.
func (c *CiliumInstaller) Install(ctx context.Context) error {
	err := c.helmInstallOrUpgradeCilium(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Cilium: %w", err)
	}

	return nil
}

// WaitForReadiness waits for the Cilium components to become ready.
func (c *CiliumInstaller) WaitForReadiness(ctx context.Context) error {
	if c.waitFn == nil {
		return nil
	}

	return c.waitFn(ctx)
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (c *CiliumInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	if waitFunc == nil {
		c.waitFn = c.waitForReadiness

		return
	}

	c.waitFn = waitFunc
}

// Uninstall removes the Helm release for Cilium.
func (c *CiliumInstaller) Uninstall(ctx context.Context) error {
	err := c.client.UninstallRelease(ctx, "cilium", "kube-system")
	if err != nil {
		return fmt.Errorf("failed to uninstall cilium release: %w", err)
	}

	return nil
}

// --- internals ---

func (c *CiliumInstaller) helmInstallOrUpgradeCilium(ctx context.Context) error {
	repoEntry := &helm.RepositoryEntry{
		Name: "cilium",
		URL:  "https://helm.cilium.io",
	}

	addRepoErr := c.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add cilium repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName: "cilium",
		ChartName:   "cilium/cilium",
		Namespace:   "kube-system",
		RepoURL:     "https://helm.cilium.io",
		Atomic:      true,
		Silent:      true,
		UpgradeCRDs: true,
		Timeout:     c.timeout,
		Wait:        true,
		WaitForJobs: true,
	}

	applyDefaultValues(spec)

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install cilium chart: %w", err)
	}

	return nil
}

func applyDefaultValues(spec *helm.ChartSpec) {
	if spec.SetJSONVals == nil {
		spec.SetJSONVals = make(map[string]string, 1)
	}

	if _, ok := spec.SetJSONVals["operator.replicas"]; !ok {
		spec.SetJSONVals["operator.replicas"] = "1"
	}
}

func (c *CiliumInstaller) waitForReadiness(ctx context.Context) error {
	restConfig, err := k8sutil.BuildRESTConfig(c.kubeconfig, c.context)
	if err != nil {
		return fmt.Errorf("build kubernetes client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	checks := []k8sutil.ReadinessCheck{
		{Type: "daemonset", Namespace: "kube-system", Name: "cilium"},
		{Type: "deployment", Namespace: "kube-system", Name: "cilium-operator"},
	}

	err = k8sutil.WaitForMultipleResources(ctx, clientset, checks, c.timeout)
	if err != nil {
		return fmt.Errorf("wait for cilium components: %w", err)
	}

	return nil
}

func (c *CiliumInstaller) buildRESTConfig() (*rest.Config, error) {
	config, err := k8sutil.BuildRESTConfig(c.kubeconfig, c.context)
	if err != nil {
		return nil, fmt.Errorf("build REST config: %w", err)
	}

	return config, nil
}
