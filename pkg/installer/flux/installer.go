// Package fluxinstaller provides a Flux installer implementation.
package fluxinstaller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	pathutils "github.com/devantler-tech/ksail-go/internal/utils/path"
	helmclient "github.com/mittwald/go-helm-client"
)

// ErrUnexpectedClientType is returned when the helm client constructor returns an unexpected type.
var ErrUnexpectedClientType = errors.New("unexpected client type returned from helm client constructor")

// FluxInstaller implements the installer.Installer interface for Flux.
type FluxInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
}

// NewFluxInstaller creates a new Flux installer instance.
func NewFluxInstaller(kubeconfig, context string, timeout time.Duration) *FluxInstaller {
	return &FluxInstaller{
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
	}
}

// Install installs or upgrades the Flux Operator via its OCI Helm chart.
func (b *FluxInstaller) Install() error {
	err := b.helmInstallOrUpgradeFluxOperator()
	if err != nil {
		return fmt.Errorf("failed to install Flux operator: %w", err)
	}

	// FluxInstance sync configuration will be added in future iterations

	return nil
}

// Uninstall removes the Helm release for the Flux Operator.
func (b *FluxInstaller) Uninstall() error {
	client, err := b.newHelmClient()
	if err != nil {
		return fmt.Errorf("failed to create Helm client: %w", err)
	}

	err = client.UninstallReleaseByName("flux-operator")
	if err != nil {
		return fmt.Errorf("failed to uninstall flux-operator release: %w", err)
	}

	return nil
}

// --- internals ---

func (b *FluxInstaller) helmInstallOrUpgradeFluxOperator() error {
	client, err := b.newHelmClient()
	if err != nil {
		return fmt.Errorf("failed to create Helm client: %w", err)
	}

	spec := &helmclient.ChartSpec{
		ReleaseName:        "flux-operator",
		ChartName:          "oci://ghcr.io/controlplaneio-fluxcd/charts/flux-operator",
		Namespace:          "flux-system",
		CreateNamespace:    true,
		Atomic:             true,
		UpgradeCRDs:        true,
		Timeout:            b.timeout,
		// Only set fields that have valid zero values for their types
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	_, err = client.InstallOrUpgradeChart(ctx, spec, nil)
	if err != nil {
		return fmt.Errorf("failed to install or upgrade chart: %w", err)
	}

	return nil
}

func (b *FluxInstaller) newHelmClient() (*helmclient.HelmClient, error) {
	kubeconfigPath, _ := pathutils.ExpandHomePath(b.kubeconfig)

	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	opts := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace: "flux-system",
			// Only set fields that have valid zero values for their types
		},
		KubeConfig:  data,
		KubeContext: b.context,
	}

	client, err := helmclient.NewClientFromKubeConf(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Helm client from kubeconfig: %w", err)
	}

	// Type assert to concrete type since we know NewClientFromKubeConf returns *HelmClient
	helmClient, ok := client.(*helmclient.HelmClient)
	if !ok {
		return nil, ErrUnexpectedClientType
	}

	return helmClient, nil
}