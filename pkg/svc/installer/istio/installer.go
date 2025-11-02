package istioinstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
)

const istioRepoURL = "https://istio-release.storage.googleapis.com/charts"

// IstioInstaller implements the installer.Installer interface for Istio.
type IstioInstaller struct {
	timeout time.Duration
	client  helm.Interface
}

// NewIstioInstaller creates a new Istio installer instance.
func NewIstioInstaller(
	client helm.Interface,
	timeout time.Duration,
) *IstioInstaller {
	return &IstioInstaller{
		client:  client,
		timeout: timeout,
	}
}

// Install installs or upgrades Istio via its Helm charts.
func (i *IstioInstaller) Install(ctx context.Context) error {
	err := i.helmInstallOrUpgradeIstioBase(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Istio base: %w", err)
	}

	err = i.helmInstallOrUpgradeIstiod(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Istiod: %w", err)
	}

	return nil
}

// Uninstall removes the Helm releases for Istio.
func (i *IstioInstaller) Uninstall(ctx context.Context) error {
	// Uninstall istiod first, then base (reverse order of installation)
	err := i.client.UninstallRelease(ctx, "istiod", "istio-system")
	if err != nil {
		return fmt.Errorf("failed to uninstall istiod release: %w", err)
	}

	err = i.client.UninstallRelease(ctx, "istio-base", "istio-system")
	if err != nil {
		return fmt.Errorf("failed to uninstall istio-base release: %w", err)
	}

	return nil
}

// --- internals ---

func (i *IstioInstaller) helmInstallOrUpgradeIstioBase(ctx context.Context) error {
	return i.installChart(ctx, "istio-base", "istio/base")
}

func (i *IstioInstaller) helmInstallOrUpgradeIstiod(ctx context.Context) error {
	return i.installChart(ctx, "istiod", "istio/istiod")
}

// installChart is a helper method to install or upgrade an Istio chart.
func (i *IstioInstaller) installChart(ctx context.Context, releaseName, chartName string) error {
	repoEntry := &helm.RepositoryEntry{
		Name: "istio",
		URL:  istioRepoURL,
	}

	addRepoErr := i.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add istio repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     releaseName,
		ChartName:       chartName,
		Namespace:       "istio-system",
		RepoURL:         istioRepoURL,
		CreateNamespace: true,
		Atomic:          true,
		UpgradeCRDs:     true,
		Timeout:         i.timeout,
		Wait:            true,
		WaitForJobs:     true,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	_, err := i.client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install %s chart: %w", releaseName, err)
	}

	return nil
}
