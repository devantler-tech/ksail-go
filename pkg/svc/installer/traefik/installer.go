package traefikinstaller

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
)

// TraefikInstaller implements the installer.Installer interface for Traefik.
type TraefikInstaller struct {
	timeout time.Duration
	client  helm.Interface
}

// NewTraefikInstaller creates a new Traefik installer instance.
func NewTraefikInstaller(
	client helm.Interface,
	timeout time.Duration,
) *TraefikInstaller {
	return &TraefikInstaller{
		client:  client,
		timeout: timeout,
	}
}

// Install installs or upgrades Traefik via its Helm chart.
func (t *TraefikInstaller) Install(ctx context.Context) error {
	err := t.helmInstallOrUpgradeTraefik(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Traefik: %w", err)
	}

	return nil
}

// Uninstall removes the Helm release for Traefik.
func (t *TraefikInstaller) Uninstall(ctx context.Context) error {
	err := t.client.UninstallRelease(ctx, "traefik", "traefik")
	if err != nil {
		return fmt.Errorf("failed to uninstall traefik release: %w", err)
	}

	return nil
}

// --- internals ---

func (t *TraefikInstaller) helmInstallOrUpgradeTraefik(ctx context.Context) error {
	repoEntry := &helm.RepositoryEntry{
		Name: "traefik",
		URL:  "https://traefik.github.io/charts",
	}

	addRepoErr := t.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add traefik repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     "traefik",
		ChartName:       "traefik/traefik",
		Namespace:       "traefik",
		CreateNamespace: true,
		Atomic:          true,
		Wait:            true,
		WaitForJobs:     true,
		Timeout:         t.timeout,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	_, err := t.client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install traefik chart: %w", err)
	}

	return nil
}
