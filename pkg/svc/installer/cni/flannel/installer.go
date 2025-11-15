package flannel

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/k8s"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"
)

const (
	// flannelRepoName is the Helm repository identifier.
	flannelRepoName = "flannel"

	// flannelRepoURL is the Helm repository URL used for Flannel installations.
	flannelRepoURL = "https://flannel-io.github.io/flannel/"

	// flannelChartName is the full Helm chart name for Flannel.
	flannelChartName = "flannel/flannel"

	// flannelReleaseName is the Helm release name used for Flannel.
	flannelReleaseName = "flannel"

	// flannelNamespace is the Kubernetes namespace where Flannel is deployed.
	flannelNamespace = "kube-flannel"

	// flannelDaemonSetName is the name of the Flannel DaemonSet.
	flannelDaemonSetName = "kube-flannel-ds"
)

// Installer implements the installer.Installer interface for Flannel CNI using Helm.
type Installer struct {
	*cni.InstallerBase
}

// NewFlannelInstaller creates a new Flannel CNI installer backed by a Helm client.
func NewFlannelInstaller(
	client helm.Interface,
	kubeconfig, context string,
	timeout time.Duration,
) *Installer {
	flannelInstaller := &Installer{}
	flannelInstaller.InstallerBase = cni.NewInstallerBase(
		client,
		kubeconfig,
		context,
		timeout,
		flannelInstaller.waitForReadiness,
	)

	return flannelInstaller
}

// Install installs or upgrades Flannel via its Helm chart.
func (f *Installer) Install(ctx context.Context) error {
	err := f.helmInstallOrUpgradeFlannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to install Flannel: %w", err)
	}

	return nil
}

// SetWaitForReadinessFunc overrides the readiness wait function. Primarily used for testing.
func (f *Installer) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
	f.InstallerBase.SetWaitForReadinessFunc(waitFunc, f.waitForReadiness)
}

// Uninstall removes the Helm release for Flannel.
func (f *Installer) Uninstall(ctx context.Context) error {
	client, err := f.GetClient()
	if err != nil {
		return fmt.Errorf("get helm client: %w", err)
	}

	err = client.UninstallRelease(ctx, flannelReleaseName, flannelNamespace)
	if err != nil {
		return fmt.Errorf("failed to uninstall flannel release: %w", err)
	}

	return nil
}

// --- internals ---

func (f *Installer) helmInstallOrUpgradeFlannel(ctx context.Context) error {
	client, err := f.GetClient()
	if err != nil {
		return fmt.Errorf("get helm client: %w", err)
	}

	repoConfig := helm.RepoConfig{
		Name:     flannelRepoName,
		URL:      flannelRepoURL,
		RepoName: flannelRepoName,
	}

	chartConfig := helm.ChartConfig{
		ReleaseName:     flannelReleaseName,
		ChartName:       flannelChartName,
		Namespace:       flannelNamespace,
		RepoURL:         flannelRepoURL,
		CreateNamespace: true,
	}

	err = helm.InstallOrUpgradeChart(ctx, client, repoConfig, chartConfig, f.GetTimeout())
	if err != nil {
		return fmt.Errorf("install or upgrade flannel: %w", err)
	}

	return nil
}

func (f *Installer) waitForReadiness(ctx context.Context) error {
	checks := []k8s.ReadinessCheck{
		{Type: "daemonset", Namespace: flannelNamespace, Name: flannelDaemonSetName},
	}

	err := installer.WaitForResourceReadiness(
		ctx,
		f.GetKubeconfig(),
		f.GetContext(),
		checks,
		f.GetTimeout(),
		"flannel",
	)
	if err != nil {
		return fmt.Errorf("wait for flannel readiness: %w", err)
	}

	return nil
}
