package ciliuminstaller

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/spf13/cobra"
)

// installParameters holds the parameters for the install command.
type installParameters struct {
	client     helm.Interface
	kubeconfig string
	context    string
	timeout    time.Duration
	writer     io.Writer
}

// NewInstallCommand creates a Cobra command for installing Cilium.
func NewInstallCommand(
	client helm.Interface,
	kubeconfig, kubectx string,
	timeout time.Duration,
	writer io.Writer,
) *cobra.Command {
	params := &installParameters{
		client:     client,
		kubeconfig: kubeconfig,
		context:    kubectx,
		timeout:    timeout,
		writer:     writer,
	}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Cilium using Helm",
		Long:  "Install Cilium in a Kubernetes cluster using Helm",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return params.runInstall(cmd.Context())
		},
	}

	return cmd
}

func (p *installParameters) runInstall(ctx context.Context) error {
	repoEntry := &helm.RepositoryEntry{
		Name: "cilium",
		URL:  "https://helm.cilium.io",
	}

	addRepoErr := p.client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add cilium repository: %w", addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName: "cilium",
		ChartName:   "cilium/cilium",
		Namespace:   "kube-system",
		RepoURL:     "https://helm.cilium.io",
		Atomic:      true,
		Silent:      true,
		UpgradeCRDs: true,
		Timeout:     p.timeout,
		Wait:        true,
		WaitForJobs: true,
	}

	applyDefaultValues(spec)

	timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	_, err := p.client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install cilium chart: %w", err)
	}

	return nil
}

func applyDefaultValues(spec *helm.ChartSpec) {
	if spec.SetJSONVals == nil {
		spec.SetJSONVals = make(map[string]string, 1)
	}

	if _, ok := spec.SetJSONVals["operator.replicas"]; !ok {
		spec.SetJSONVals["operator.replicas"] = "1"
	}
}

// uninstallParameters holds the parameters for the uninstall command.
type uninstallParameters struct {
	client     helm.Interface
	kubeconfig string
	context    string
	writer     io.Writer
}

// NewUninstallCommand creates a Cobra command for uninstalling Cilium.
func NewUninstallCommand(
	client helm.Interface,
	kubeconfig, kubectx string,
	writer io.Writer,
) *cobra.Command {
	params := &uninstallParameters{
		client:     client,
		kubeconfig: kubeconfig,
		context:    kubectx,
		writer:     writer,
	}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Cilium using Helm",
		Long:  "Uninstall Cilium from a Kubernetes cluster using Helm",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return params.runUninstall(cmd.Context())
		},
	}

	return cmd
}

func (p *uninstallParameters) runUninstall(ctx context.Context) error {
	err := p.client.UninstallRelease(ctx, "cilium", "kube-system")
	if err != nil {
		return fmt.Errorf("failed to uninstall cilium release: %w", err)
	}

	return nil
}
