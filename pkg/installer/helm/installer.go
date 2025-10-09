package helminstaller

import (
	"context"
	"errors"
	"fmt"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
)

// ErrUnexpectedClientType is returned when the helm client constructor returns an unexpected type.
var ErrUnexpectedClientType = errors.New(
	"unexpected client type returned from helm client constructor",
)

// HelmInstaller implements the installer.Installer interface for Helm charts.
type HelmInstaller struct {
	client      HelmClient
	releaseName string
	chartName   string
	namespace   string
	version     string
	valuesYaml  string
	timeout     time.Duration
}

// NewHelmInstaller creates a new Helm installer instance.
func NewHelmInstaller(
	client HelmClient,
	releaseName, chartName, namespace, version, valuesYaml string,
	timeout time.Duration,
) *HelmInstaller {
	return &HelmInstaller{
		client:      client,
		releaseName: releaseName,
		chartName:   chartName,
		namespace:   namespace,
		version:     version,
		valuesYaml:  valuesYaml,
		timeout:     timeout,
	}
}

// Install installs or upgrades a Helm chart.
func (h *HelmInstaller) Install(ctx context.Context) error {
	err := h.helmInstallChart(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Helm chart: %w", err)
	}

	return nil
}

// Uninstall removes the Helm release.
func (h *HelmInstaller) Uninstall(_ context.Context) error {
	err := h.client.UninstallReleaseByName(h.releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall release %s: %w", h.releaseName, err)
	}

	return nil
}

// --- internals ---

func (h *HelmInstaller) helmInstallChart(ctx context.Context) error {
	spec := &helmclient.ChartSpec{
		ReleaseName:     h.releaseName,
		ChartName:       h.chartName,
		Namespace:       h.namespace,
		CreateNamespace: true,
		Atomic:          true,
		UpgradeCRDs:     true,
		Timeout:         h.timeout,
		ValuesYaml:      h.valuesYaml,
		Version:         h.version,
		Wait:            true,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	_, err := h.client.InstallChart(timeoutCtx, spec, nil)
	if err != nil {
		return fmt.Errorf("failed to install helm chart: %w", err)
	}

	return nil
}
