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
	kubeconfig string
	context    string
	timeout    time.Duration
	client     helm.Interface
}

// NewIstioInstaller creates a new Istio installer instance.
func NewIstioInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *IstioInstaller {
	return &IstioInstaller{
		client:     client,
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
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
	repoEntry := &helm.RepositoryEntry{
		Name: "istio",
		URL:  istioRepoURL,
	}

	addRepoErr := i.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add istio repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     "istio-base",
		ChartName:       "istio/base",
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
		return fmt.Errorf("failed to install istio-base chart: %w", err)
	}

	return nil
}

func (i *IstioInstaller) helmInstallOrUpgradeIstiod(ctx context.Context) error {
	// Ensure repository is available for istiod installation
	repoEntry := &helm.RepositoryEntry{
		Name: "istio",
		URL:  istioRepoURL,
	}

	addRepoErr := i.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add istio repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     "istiod",
		ChartName:       "istio/istiod",
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
		return fmt.Errorf("failed to install istiod chart: %w", err)
	}

	return nil
}
