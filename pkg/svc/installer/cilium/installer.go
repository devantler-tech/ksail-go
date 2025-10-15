package ciliuminstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
)

// CiliumInstaller implements the installer.Installer interface for Cilium.
type CiliumInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
	client     HelmClient
}

// NewCiliumInstaller creates a new Cilium installer instance.
func NewCiliumInstaller(
	client HelmClient,
	kubeconfig, context string,
	timeout time.Duration,
) *CiliumInstaller {
	return &CiliumInstaller{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
	}
}

// Install installs or upgrades Cilium via its Helm chart.
func (c *CiliumInstaller) Install(ctx context.Context) error {
	err := c.helmInstallOrUpgradeCilium(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Cilium: %w", err)
	}

	return nil
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
	spec := &helm.ChartSpec{
		ReleaseName:          "cilium",
		ChartName:            "cilium/cilium",
		Namespace:            "kube-system",
		CreateNamespace:      false,
		Atomic:               true,
		UpgradeCRDs:          true,
		Timeout:              c.timeout,
		Wait:                 true,
		WaitForJobs:          true,
		DisableHooks:         false,
		Replace:              false,
		DependencyUpdate:     false,
		GenerateName:         false,
		NameTemplate:         "",
		SkipCRDs:             false,
		SubNotes:             false,
		Force:                false,
		ResetValues:          false,
		ReuseValues:          false,
		ResetThenReuseValues: false,
		MaxHistory:           0,
		CleanupOnFail:        false,
		DryRun:               false,
		Description:          "",
		KeepHistory:          false,
		IgnoreNotFound:       false,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install cilium chart: %w", err)
	}

	return nil
}
