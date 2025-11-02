package argocdinstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
)

// ArgoCDInstaller implements the installer.Installer interface for ArgoCD.
type ArgoCDInstaller struct {
	timeout time.Duration
	client  helm.Interface
}

// NewArgoCDInstaller creates a new ArgoCD installer instance.
func NewArgoCDInstaller(
	client helm.Interface,
	timeout time.Duration,
) *ArgoCDInstaller {
	return &ArgoCDInstaller{
		client:  client,
		timeout: timeout,
	}
}

// Install installs or upgrades ArgoCD via its Helm chart.
func (a *ArgoCDInstaller) Install(ctx context.Context) error {
	err := a.helmInstallOrUpgradeArgoCD(ctx)
	if err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	return nil
}

// Uninstall removes the Helm release for ArgoCD.
func (a *ArgoCDInstaller) Uninstall(ctx context.Context) error {
	err := a.client.UninstallRelease(ctx, "argocd", "argocd")
	if err != nil {
		return fmt.Errorf("failed to uninstall argocd release: %w", err)
	}

	return nil
}

// --- internals ---

func (a *ArgoCDInstaller) helmInstallOrUpgradeArgoCD(ctx context.Context) error {
	repoEntry := &helm.RepositoryEntry{
		Name: "argo",
		URL:  "https://argoproj.github.io/argo-helm",
	}

	addRepoErr := a.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add argo repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     "argocd",
		ChartName:       "argo/argo-cd",
		Namespace:       "argocd",
		RepoURL:         "https://argoproj.github.io/argo-helm",
		CreateNamespace: true,
		Atomic:          true,
		UpgradeCRDs:     true,
		Timeout:         a.timeout,
	}

	_, err := a.client.InstallOrUpgradeChart(ctx, spec)
	if err != nil {
		return fmt.Errorf("failed to install argocd chart: %w", err)
	}

	return nil
}
