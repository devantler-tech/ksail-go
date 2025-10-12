package ciliuminstaller

import (
	"context"
	"errors"
	"fmt"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
)

// ErrUnexpectedClientType is returned when the helm client constructor returns an unexpected type.
var ErrUnexpectedClientType = errors.New(
	"unexpected client type returned from helm client constructor",
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
func (c *CiliumInstaller) Uninstall(_ context.Context) error {
	err := c.client.UninstallReleaseByName("cilium")
	if err != nil {
		return fmt.Errorf("failed to uninstall cilium release: %w", err)
	}

	return nil
}

// --- internals ---

func (c *CiliumInstaller) helmInstallOrUpgradeCilium(ctx context.Context) error {
	spec := &helmclient.ChartSpec{
		ReleaseName:     "cilium",
		ChartName:       "cilium/cilium",
		Namespace:       "kube-system",
		CreateNamespace: false,
		Atomic:          true,
		UpgradeCRDs:     true,
		Timeout:         c.timeout,
		ValuesYaml:      "",
		ValuesOptions: values.Options{
			ValueFiles:   nil,
			StringValues: nil,
			Values:       nil,
			FileValues:   nil,
			JSONValues:   nil,
		},
		Version:              "",
		DisableHooks:         false,
		Replace:              false,
		Wait:                 false,
		WaitForJobs:          false,
		DependencyUpdate:     false,
		GenerateName:         false,
		NameTemplate:         "",
		SkipCRDs:             false,
		SubNotes:             false,
		Force:                false,
		ResetValues:          false,
		ReuseValues:          false,
		ResetThenReuseValues: false,
		Recreate:             false,
		MaxHistory:           0,
		CleanupOnFail:        false,
		DryRun:               false,
		DryRunOption:         "",
		Description:          "",
		KeepHistory:          false,
		Labels:               nil,
		IgnoreNotFound:       false,
		DeletionPropagation:  "",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.InstallOrUpgradeChart(timeoutCtx, spec, nil)
	if err != nil {
		return fmt.Errorf("failed to install cilium chart: %w", err)
	}

	return nil
}
