package cluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailio "github.com/devantler-tech/ksail-go/pkg/io"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// newCreateLifecycleConfig creates the lifecycle configuration for cluster creation.
func newCreateLifecycleConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "🚀",
		TitleContent:       "Create cluster...",
		ActivityContent:    "creating cluster",
		SuccessContent:     "cluster created",
		ErrorMessagePrefix: "failed to create cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Create(ctx, clusterName)
		},
	}
}

// NewCreateCmd wires the cluster create command using the shared runtime container.
func NewCreateCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a cluster",
		Long:         `Create a Kubernetes cluster as defined by configuration.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = newCreateCommandRunE(runtimeContainer, cfgManager)

	return cmd
}

// newCreateCommandRunE creates the RunE handler for cluster creation with CNI installation support.
func newCreateCommandRunE(
	runtimeContainer *runtime.Runtime,
	cfgManager *ksailconfigmanager.ConfigManager,
) func(*cobra.Command, []string) error {
	return shared.WrapLifecycleHandler(runtimeContainer, cfgManager, handleCreateRunE)
}

// handleCreateRunE executes cluster creation with CNI installation.
func handleCreateRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps shared.LifecycleDeps,
) error {
	config := newCreateLifecycleConfig()

	// Reuse the standard lifecycle logic but extend with CNI installation
	err := shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err != nil {
		return fmt.Errorf("cluster creation failed: %w", err)
	}

	// Install CNI if Cilium is configured
	clusterCfg := cfgManager.GetConfig()
	if clusterCfg.Spec.CNI == v1alpha1.CNICilium {
		// Add newline separator before CNI installation
		_, _ = fmt.Fprintln(cmd.OutOrStdout())

		// Start new stage for CNI installation
		deps.Timer.NewStage()

		err = installCiliumCNI(cmd, clusterCfg, deps.Timer)
		if err != nil {
			return fmt.Errorf("failed to install Cilium CNI: %w", err)
		}
	}

	return nil
}

// installCiliumCNI installs Cilium CNI on the cluster.
func installCiliumCNI(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
	// Display title for CNI installation
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Install CNI...",
		Emoji:   "🌐",
		Writer:  cmd.OutOrStdout(),
	})

	// Get kubeconfig path and data
	kubeconfig, _, err := loadKubeconfig(clusterCfg)
	if err != nil {
		return err
	}

	// Create Helm client
	helmClient, err := helm.NewClient(kubeconfig, clusterCfg.Spec.Connection.Context)
	if err != nil {
		return fmt.Errorf("failed to create Helm client: %w", err)
	}

	timeout := getCiliumInstallTimeout(clusterCfg)

	// Add Cilium Helm repository
	err = helmClient.AddRepository(cmd.Context(), &helm.RepositoryEntry{
		Name: "cilium",
		URL:  "https://helm.cilium.io/",
	})
	if err != nil {
		return fmt.Errorf("failed to add Cilium Helm repository: %w", err)
	}

	// Create and run Cilium installer
	installer := ciliuminstaller.NewCiliumInstaller(
		helmClient,
		kubeconfig,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "installing cilium",
		Writer:  cmd.OutOrStdout(),
	})

	err = installer.Install(cmd.Context())
	if err != nil {
		return fmt.Errorf("cilium installation failed: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "awaiting cilium to be ready",
		Writer:  cmd.OutOrStdout(),
	})

	err = installer.WaitForReadiness(cmd.Context())
	if err != nil {
		return fmt.Errorf("cilium readiness check failed: %w", err)
	}

	// Display success message with timing
	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, true)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "CNI installed " + timingStr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}

// loadKubeconfig loads and returns the kubeconfig path and data.
func loadKubeconfig(clusterCfg *v1alpha1.Cluster) (string, []byte, error) {
	kubeconfig, err := expandKubeconfigPath(clusterCfg.Spec.Connection.Kubeconfig)
	if err != nil {
		return "", nil, fmt.Errorf("failed to expand kubeconfig path: %w", err)
	}

	kubeconfigData, err := ksailio.ReadFileSafe(filepath.Dir(kubeconfig), kubeconfig)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read kubeconfig file: %w", err)
	}

	return kubeconfig, kubeconfigData, nil
}

// getCiliumInstallTimeout determines the timeout for Cilium installation.
func getCiliumInstallTimeout(clusterCfg *v1alpha1.Cluster) time.Duration {
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
