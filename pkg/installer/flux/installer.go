// Package fluxinstaller provides a Flux installer implementation.
package fluxinstaller

import (
	"context"
	"os"
	"time"

	pathutils "github.com/devantler-tech/ksail-go/internal/utils/path"
	helmclient "github.com/mittwald/go-helm-client"
)

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
		return err
	}

	// TODO: Apply FluxInstance that syncs with local 'ksail-registry'
	return nil
}

// Uninstall removes the Helm release for the Flux Operator.
func (b *FluxInstaller) Uninstall() error {
	client, err := b.newHelmClient()
	if err != nil {
		return err
	}

	return client.UninstallReleaseByName("flux-operator")
}

// --- internals ---

func (b *FluxInstaller) helmInstallOrUpgradeFluxOperator() error {
	client, err := b.newHelmClient()
	if err != nil {
		return err
	}

	spec := helmclient.ChartSpec{
		ReleaseName:     "flux-operator",
		ChartName:       "oci://ghcr.io/controlplaneio-fluxcd/charts/flux-operator",
		Namespace:       "flux-system",
		CreateNamespace: true,
		Atomic:          true,
		UpgradeCRDs:     true,
		Timeout:         b.timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	_, err = client.InstallOrUpgradeChart(ctx, &spec, nil)

	return err
}

func (b *FluxInstaller) newHelmClient() (helmclient.Client, error) {
	kubeconfigPath, _ := pathutils.ExpandHomePath(b.kubeconfig)

	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	opts := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace: "flux-system",
		},
		KubeConfig:  data,
		KubeContext: b.context,
	}

	return helmclient.NewClientFromKubeConf(opts)
}