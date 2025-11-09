package calicoinstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil"
	"k8s.io/client-go/rest"
)

// CalicoInstaller implements the installer.Installer interface for Calico.
type CalicoInstaller struct {
	*installer.CNIInstallerBase
}

// NewCalicoInstaller creates a new Calico installer instance.
func NewCalicoInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *CalicoInstaller {
	calicoInstaller := &CalicoInstaller{}
	calicoInstaller.CNIInstallerBase = installer.NewCNIInstallerBase(
		client,
		kubeconfig,
		context,
		timeout,
		calicoInstaller.waitForReadiness,
	)

	return calicoInstaller
}

// Install installs or upgrades Calico via its Helm chart.
func (c *CalicoInstaller) Install(ctx context.Context) error {
	err := c.helmInstallOrUpgradeCalico(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Calico: %w", err)
	}

	return nil
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (c *CalicoInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	c.CNIInstallerBase.SetWaitForReadinessFunc(waitFunc, c.waitForReadiness)
}

// Uninstall removes the Helm release for Calico.
func (c *CalicoInstaller) Uninstall(ctx context.Context) error {
	client, err := c.GetClient()
	if err != nil {
		return fmt.Errorf("get helm client: %w", err)
	}

	err = client.UninstallRelease(ctx, "calico", "tigera-operator")
	if err != nil {
		return fmt.Errorf("failed to uninstall calico release: %w", err)
	}

	return nil
}

// --- internals ---

func (c *CalicoInstaller) helmInstallOrUpgradeCalico(ctx context.Context) error {
	client, err := c.GetClient()
	if err != nil {
		return fmt.Errorf("get helm client: %w", err)
	}

	repoConfig := installer.HelmRepoConfig{
		Name:     "projectcalico",
		URL:      "https://docs.tigera.io/calico/charts",
		RepoName: "calico",
	}

	chartConfig := installer.HelmChartConfig{
		ReleaseName:     "calico",
		ChartName:       "projectcalico/tigera-operator",
		Namespace:       "tigera-operator",
		RepoURL:         "https://docs.tigera.io/calico/charts",
		CreateNamespace: true,
	}

	return installer.InstallOrUpgradeHelmChart(ctx, client, repoConfig, chartConfig, c.GetTimeout())
}

func (c *CalicoInstaller) waitForReadiness(ctx context.Context) error {
	checks := []k8sutil.ReadinessCheck{
		{Type: "deployment", Namespace: "tigera-operator", Name: "tigera-operator"},
		{Type: "daemonset", Namespace: "calico-system", Name: "calico-node"},
		{Type: "deployment", Namespace: "calico-system", Name: "calico-kube-controllers"},
	}

	return installer.WaitForResourceReadiness(
		ctx,
		c.GetKubeconfig(),
		c.GetContext(),
		checks,
		c.GetTimeout(),
		"calico",
	)
}

func (c *CalicoInstaller) buildRESTConfig() (*rest.Config, error) {
	return c.BuildRESTConfig()
}
