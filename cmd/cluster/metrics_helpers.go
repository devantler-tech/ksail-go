package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	metricsserverinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/metrics-server"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// handleMetricsServer manages metrics-server installation based on cluster configuration.
// For K3d, metrics-server should be disabled via config (handled in setupK3dMetricsServer), not uninstalled.
func handleMetricsServer(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
	// Check if distribution provides metrics-server by default
	hasMetricsByDefault := distributionProvidesMetricsByDefault(clusterCfg.Spec.Distribution)

	// Enabled: Install if not present by default
	if clusterCfg.Spec.MetricsServer == v1alpha1.MetricsServerEnabled {
		if hasMetricsByDefault {
			// Already present, no action needed
			return nil
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout())

		tmr.NewStage()

		return installMetricsServer(cmd, clusterCfg, tmr)
	}

	// Disabled: For K3d, this is handled via config before cluster creation (setupK3dMetricsServer)
	// No post-creation action needed for K3d
	if clusterCfg.Spec.MetricsServer == v1alpha1.MetricsServerDisabled {
		if clusterCfg.Spec.Distribution == v1alpha1.DistributionK3d {
			// K3d metrics-server is disabled via config, no action needed here
			return nil
		}

		if !hasMetricsByDefault {
			// Not present, no action needed
			return nil
		}

		// For other distributions that have it by default, we would uninstall here
		// But currently only K3d has it by default, and that's handled via config
	}

	return nil
}

// distributionProvidesMetricsByDefault returns true if the distribution includes metrics-server by default.
// K3d (based on K3s) includes metrics-server, Kind does not.
func distributionProvidesMetricsByDefault(distribution v1alpha1.Distribution) bool {
	switch distribution {
	case v1alpha1.DistributionK3d:
		return true
	case v1alpha1.DistributionKind:
		return false
	default:
		return false
	}
}

// installMetricsServer installs metrics-server on the cluster.
func installMetricsServer(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Install Metrics Server...",
		Emoji:   "📊",
		Writer:  cmd.OutOrStdout(),
	})

	helmClient, kubeconfig, err := createHelmClientForCluster(clusterCfg)
	if err != nil {
		return err
	}

	timeout := getInstallTimeout(clusterCfg)
	installer := metricsserverinstaller.NewMetricsServerInstaller(
		helmClient,
		kubeconfig,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)

	return runMetricsServerInstallation(cmd, installer, tmr)
}

// runMetricsServerInstallation performs the metrics-server installation.
func runMetricsServerInstallation(
	cmd *cobra.Command,
	installer *metricsserverinstaller.MetricsServerInstaller,
	tmr timer.Timer,
) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "installing metrics-server",
		Writer:  cmd.OutOrStdout(),
	})

	installErr := installer.Install(cmd.Context())
	if installErr != nil {
		return fmt.Errorf("metrics-server installation failed: %w", installErr)
	}

	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, true)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "Metrics Server installed " + timingStr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}
