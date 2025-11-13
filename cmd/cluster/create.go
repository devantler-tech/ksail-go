package cluster

import (
	"context"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	cmdhelpers "github.com/devantler-tech/ksail-go/pkg/cmd"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const (
	// k3sDisableMetricsServerFlag is the K3s flag to disable metrics-server.
	k3sDisableMetricsServerFlag = "--disable=metrics-server"
)

// newCreateLifecycleConfig creates the lifecycle configuration for cluster creation.
func newCreateLifecycleConfig() cmdhelpers.LifecycleConfig {
	return cmdhelpers.LifecycleConfig{
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

	// Create field selectors including metrics-server
	fieldSelectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	fieldSelectors = append(fieldSelectors, ksailconfigmanager.DefaultMetricsServerFieldSelector())

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		fieldSelectors,
	)

	cmd.Flags().
		StringSlice("mirror-registry", []string{},
			"Configure mirror registries with format 'host=upstream' (e.g., docker.io=https://registry-1.docker.io)")
	_ = cfgManager.Viper.BindPFlag("mirror-registry", cmd.Flags().Lookup("mirror-registry"))

	cmd.RunE = cmdhelpers.WrapLifecycleHandler(runtimeContainer, cfgManager, handleCreateRunE)

	return cmd
}

// handleCreateRunE executes cluster creation with mirror registry setup and CNI installation.
func handleCreateRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps cmdhelpers.LifecycleDeps,
) error {
	deps.Timer.Start()

	clusterCfg, err := cfgManager.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	kindConfig, k3dConfig, err := loadDistributionConfigs(clusterCfg, deps.Timer)
	if err != nil {
		return err
	}

	err = setupMirrorRegistries(cmd, clusterCfg, deps, cfgManager, kindConfig, k3dConfig)
	if err != nil {
		return fmt.Errorf("failed to setup mirror registries: %w", err)
	}

	// Configure metrics-server for K3d before cluster creation
	setupK3dMetricsServer(clusterCfg, k3dConfig)

	deps.Timer.NewStage()

	err = cmdhelpers.RunLifecycleWithConfig(cmd, deps, newCreateLifecycleConfig(), clusterCfg)
	if err != nil {
		return fmt.Errorf("failed to execute cluster lifecycle: %w", err)
	}

	err = connectRegistriesToClusterNetwork(
		cmd,
		clusterCfg,
		deps,
		cfgManager,
		kindConfig,
		k3dConfig,
	)
	if err != nil {
		notify.WriteMessage(notify.Message{
			Type:    notify.WarningType,
			Content: fmt.Sprintf("failed to connect registries to cluster network: %v", err),
			Writer:  cmd.OutOrStdout(),
		})
	}

	return handlePostCreationSetup(cmd, clusterCfg, deps.Timer)
}

// handlePostCreationSetup installs CNI and metrics-server after cluster creation.
// Order depends on CNI configuration to resolve dependencies.
func handlePostCreationSetup(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	tmr timer.Timer,
) error {
	// For custom CNI (Cilium), install CNI first as metrics-server needs networking
	// For default CNI, install metrics-server first as it's independent
	if clusterCfg.Spec.CNI == v1alpha1.CNICilium {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())

		tmr.NewStage()

		err := installCiliumCNI(cmd, clusterCfg, tmr)
		if err != nil {
			return fmt.Errorf("failed to install Cilium CNI: %w", err)
		}

		// Install metrics-server after CNI is ready
		err = handleMetricsServer(cmd, clusterCfg, tmr)
		if err != nil {
			return fmt.Errorf("failed to handle metrics server: %w", err)
		}
	} else {
		// For default CNI, install metrics-server first
		err := handleMetricsServer(cmd, clusterCfg, tmr)
		if err != nil {
			return fmt.Errorf("failed to handle metrics server: %w", err)
		}
	}

	return nil
}

func loadDistributionConfigs(
	clusterCfg *v1alpha1.Cluster,
	lifecycleTimer timer.Timer,
) (*v1alpha4.Cluster, *v1alpha5.SimpleConfig, error) {
	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		manager := kindconfigmanager.NewConfigManager(clusterCfg.Spec.DistributionConfig)

		kindConfig, err := manager.LoadConfig(lifecycleTimer)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load kind config: %w", err)
		}

		return kindConfig, nil, nil
	case v1alpha1.DistributionK3d:
		manager := k3dconfigmanager.NewConfigManager(clusterCfg.Spec.DistributionConfig)

		k3dConfig, err := manager.LoadConfig(lifecycleTimer)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load k3d config: %w", err)
		}

		return nil, k3dConfig, nil
	default:
		return nil, nil, nil
	}
}

// setupK3dMetricsServer configures metrics-server for K3d clusters by adding K3s flags.
// K3s includes metrics-server by default, so we add --disable=metrics-server flag when disabled.
// This function is called during cluster creation to handle cases where:
// 1. The user overrides --metrics-server flag at create time (different from init-time config).
// 2. The k3d.yaml was manually edited and the flag needs to be added.
// 3. Ensures consistency even if the scaffolder-generated config was modified.
func setupK3dMetricsServer(clusterCfg *v1alpha1.Cluster, k3dConfig *v1alpha5.SimpleConfig) {
	// Only apply to K3d distribution
	if clusterCfg.Spec.Distribution != v1alpha1.DistributionK3d || k3dConfig == nil {
		return
	}

	// Only add disable flag if explicitly disabled
	if clusterCfg.Spec.MetricsServer != v1alpha1.MetricsServerDisabled {
		return
	}

	// Check if --disable=metrics-server is already present
	for _, arg := range k3dConfig.Options.K3sOptions.ExtraArgs {
		if arg.Arg == k3sDisableMetricsServerFlag {
			// Already configured, no action needed
			return
		}
	}

	// Add --disable=metrics-server flag
	k3dConfig.Options.K3sOptions.ExtraArgs = append(
		k3dConfig.Options.K3sOptions.ExtraArgs,
		v1alpha5.K3sArgWithNodeFilters{
			Arg:         k3sDisableMetricsServerFlag,
			NodeFilters: []string{"server:*"},
		},
	)
}
