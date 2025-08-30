// Package fluxinstaller provides a Flux installer implementation.
package fluxinstaller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	pathutils "github.com/devantler-tech/ksail-go/internal/utils/path"
	"github.com/devantler-tech/ksail-go/pkg/io"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
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
		ReleaseName:            "flux-operator",
		ChartName:              "oci://ghcr.io/controlplaneio-fluxcd/charts/flux-operator",
		Namespace:              "flux-system",
		CreateNamespace:        true,
		Atomic:                 true,
		UpgradeCRDs:            true,
		Timeout:                b.timeout,
		ValuesYaml:             "",
		ValuesOptions: values.Options{
			ValueFiles:   nil,
			StringValues: nil,
			Values:       nil,
			FileValues:   nil,
			JSONValues:   nil,
		},
		Version:                "",
		DisableHooks:           false,
		Replace:                false,
		Wait:                   false,
		WaitForJobs:            false,
		DependencyUpdate:       false,
		GenerateName:           false,
		NameTemplate:           "",
		SkipCRDs:               false,
		SubNotes:               false,
		Force:                  false,
		ResetValues:            false,
		ReuseValues:            false,
		ResetThenReuseValues:   false,
		Recreate:               false,
		MaxHistory:             0,
		CleanupOnFail:          false,
		DryRun:                 false,
		DryRunOption:           "",
		Description:            "",
		KeepHistory:            false,
		Labels:                 nil,
		IgnoreNotFound:         false,
		DeletionPropagation:    "",
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	kubeconfigPath, err := pathutils.ExpandHomePath(b.kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to expand kubeconfig path: %w", err)
	}

	data, err := io.ReadFileSafe(homeDir, kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	opts := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        "flux-system",
			RepositoryConfig: "",
			RepositoryCache:  "",
			Debug:            false,
			Linting:          false,
			DebugLog:         nil,
			RegistryConfig:   "",
			Output:           nil,
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
