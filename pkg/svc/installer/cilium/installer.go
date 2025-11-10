package ciliuminstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil"
)

// CiliumInstaller implements the installer.Installer interface for Cilium.
type CiliumInstaller struct {
	*installer.CNIInstallerBase
}

// NewCiliumInstaller creates a new Cilium installer instance.
func NewCiliumInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *CiliumInstaller {
	ciliumInstaller := &CiliumInstaller{}
	ciliumInstaller.CNIInstallerBase = installer.NewCNIInstallerBase(
		client,
		kubeconfig,
		context,
		timeout,
		ciliumInstaller.waitForReadiness,
	)

	return ciliumInstaller
}

// Install installs or upgrades Cilium via its Helm chart.
func (c *CiliumInstaller) Install(ctx context.Context) error {
	err := c.helmInstallOrUpgradeCilium(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Cilium: %w", err)
	}

	return nil
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (c *CiliumInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	c.CNIInstallerBase.SetWaitForReadinessFunc(waitFunc, c.waitForReadiness)
}

// Uninstall removes the Helm release for Cilium.
func (c *CiliumInstaller) Uninstall(ctx context.Context) error {
	client, err := c.GetClient()
	if err != nil {
		return fmt.Errorf("get helm client: %w", err)
	}

	err = client.UninstallRelease(ctx, "cilium", "kube-system")
	if err != nil {
		return fmt.Errorf("failed to uninstall cilium release: %w", err)
	}

	return nil
}

// --- internals ---

func (c *CiliumInstaller) helmInstallOrUpgradeCilium(ctx context.Context) error {
	client, err := c.GetClient()
	if err != nil {
		return fmt.Errorf("get helm client: %w", err)
	}

	repoConfig := installer.HelmRepoConfig{
		Name:     "cilium",
		URL:      "https://helm.cilium.io",
		RepoName: "cilium",
	}

	chartConfig := installer.HelmChartConfig{
		ReleaseName:     "cilium",
		ChartName:       "cilium/cilium",
		Namespace:       "kube-system",
		RepoURL:         "https://helm.cilium.io",
		CreateNamespace: false,
		SetJSONVals:     applyDefaultValues(),
	}

	return installer.InstallOrUpgradeHelmChart(ctx, client, repoConfig, chartConfig, c.GetTimeout())
}

func applyDefaultValues() map[string]string {
	return map[string]string{
		"operator.replicas": "1",
	}
}

func (c *CiliumInstaller) waitForReadiness(ctx context.Context) error {
	checks := []k8sutil.ReadinessCheck{
		{Type: "daemonset", Namespace: "kube-system", Name: "cilium"},
		{Type: "deployment", Namespace: "kube-system", Name: "cilium-operator"},
	}

	return installer.WaitForResourceReadiness(
		ctx,
		c.GetKubeconfig(),
		c.GetContext(),
		checks,
		c.GetTimeout(),
		"cilium",
	)
}


