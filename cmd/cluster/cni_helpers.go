package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// createHelmClientForCluster creates a Helm client configured for the cluster.
func createHelmClientForCluster(clusterCfg *v1alpha1.Cluster) (*helm.Client, string, error) {
	kubeconfig, err := loadKubeconfig(clusterCfg)
	if err != nil {
		return nil, "", err
	}

	helmClient, err := helm.NewClient(kubeconfig, clusterCfg.Spec.Connection.Context)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create Helm client: %w", err)
	}

	return helmClient, kubeconfig, nil
}

// installCiliumCNI installs Cilium CNI on the cluster.
func installCiliumCNI(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Install CNI...",
		Emoji:   "🌐",
		Writer:  cmd.OutOrStdout(),
	})

	helmClient, kubeconfig, err := createHelmClientForCluster(clusterCfg)
	if err != nil {
		return err
	}

	err = helmClient.AddRepository(cmd.Context(), &helm.RepositoryEntry{
		Name: "cilium",
		URL:  "https://helm.cilium.io/",
	})
	if err != nil {
		return fmt.Errorf("failed to add Cilium Helm repository: %w", err)
	}

	installer := newCiliumInstaller(helmClient, kubeconfig, clusterCfg)

	return runCiliumInstallation(cmd, installer, tmr)
}

func newCiliumInstaller(
	helmClient *helm.Client,
	kubeconfig string,
	clusterCfg *v1alpha1.Cluster,
) *ciliuminstaller.CiliumInstaller {
	timeout := getInstallTimeout(clusterCfg)

	return ciliuminstaller.NewCiliumInstaller(
		helmClient,
		kubeconfig,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)
}

func runCiliumInstallation(
	cmd *cobra.Command,
	installer *ciliuminstaller.CiliumInstaller,
	tmr timer.Timer,
) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "installing cilium",
		Writer:  cmd.OutOrStdout(),
	})

	installErr := installer.Install(cmd.Context())
	if installErr != nil {
		return fmt.Errorf("cilium installation failed: %w", installErr)
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "awaiting cilium to be ready",
		Writer:  cmd.OutOrStdout(),
	})

	readinessErr := installer.WaitForReadiness(cmd.Context())
	if readinessErr != nil {
		return fmt.Errorf("cilium readiness check failed: %w", readinessErr)
	}

	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, true)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "CNI installed " + timingStr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}

// loadKubeconfig loads and returns the kubeconfig path.
func loadKubeconfig(clusterCfg *v1alpha1.Cluster) (string, error) {
	kubeconfig, err := expandKubeconfigPath(clusterCfg.Spec.Connection.Kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to expand kubeconfig path: %w", err)
	}

	// Validate file exists
	_, err = os.Stat(kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to access kubeconfig file: %w", err)
	}

	return kubeconfig, nil
}

// getInstallTimeout determines the timeout for component installation (Cilium, metrics-server, etc.).
// Uses cluster connection timeout if configured, otherwise defaults to 5 minutes.
func getInstallTimeout(clusterCfg *v1alpha1.Cluster) time.Duration {
	const defaultTimeout = 5

	timeout := defaultTimeout * time.Minute
	if clusterCfg.Spec.Connection.Timeout.Duration > 0 {
		timeout = clusterCfg.Spec.Connection.Timeout.Duration
	}

	return timeout
}

// expandKubeconfigPath expands tilde (~) in kubeconfig paths to the user's home directory.
func expandKubeconfigPath(kubeconfig string) (string, error) {
	if len(kubeconfig) == 0 || kubeconfig[0] != '~' {
		return kubeconfig, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, kubeconfig[1:]), nil
}
