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
	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
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

	cmd.Flags().
		StringSlice("mirror-registry", []string{},
			"Configure mirror registries with format 'host=upstream' (e.g., docker.io=https://registry-1.docker.io)."+
				" Use dots in registry host, not hyphens")
	_ = cfgManager.Viper.BindPFlag("mirror-registry", cmd.Flags().Lookup("mirror-registry"))

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

// handleCreateRunE executes cluster creation with mirror registry setup and CNI installation.
func handleCreateRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps shared.LifecycleDeps,
) error {
	// Start timer
	deps.Timer.Start()

	// Load config first
	clusterCfg, err := cfgManager.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Load distribution config for Kind to check for mirror registries
	var kindConfig *v1alpha4.Cluster

	if clusterCfg.Spec.Distribution == v1alpha1.DistributionKind {
		kindConfigMgr := kindconfigmanager.NewConfigManager(clusterCfg.Spec.DistributionConfig)

		kindConfig, err = kindConfigMgr.LoadConfig(deps.Timer)
		if err != nil {
			return fmt.Errorf("failed to load kind config: %w", err)
		}
	}

	// Set up mirror registries before cluster creation if enabled
	err = setupMirrorRegistries(cmd, clusterCfg, deps, cfgManager, kindConfig)
	if err != nil {
		return fmt.Errorf("failed to setup mirror registries: %w", err)
	}

	// Create cluster using standard lifecycle
	deps.Timer.NewStage()

	err = executeClusterCreation(cmd, clusterCfg, deps)
	if err != nil {
		return err
	}

	// Connect registries to the Kind network after cluster is created
	err = connectRegistriesToKindNetwork(cmd, clusterCfg, deps, cfgManager, kindConfig)
	if err != nil {
		// Log warning but don't fail - registries can still work via localhost
		notify.WriteMessage(notify.Message{
			Type:    notify.WarningType,
			Content: fmt.Sprintf("failed to connect registries to kind network: %v", err),
			Writer:  cmd.OutOrStdout(),
		})
	}

	// Install CNI if Cilium is configured
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

// executeClusterCreation handles the cluster provisioning and lifecycle.
func executeClusterCreation(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps shared.LifecycleDeps,
) error {
	config := newCreateLifecycleConfig()

	// Resolve cluster provisioner
	provisioner, distributionConfig, err := deps.Factory.Create(cmd.Context(), clusterCfg)
	if err != nil {
		return fmt.Errorf("failed to resolve cluster provisioner: %w", err)
	}

	if provisioner == nil {
		return shared.ErrMissingClusterProvisionerDependency
	}

	clusterName, err := configmanager.GetClusterName(distributionConfig)
	if err != nil {
		return fmt.Errorf("failed to get cluster name from config: %w", err)
	}

	// Show title for cluster creation
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: config.TitleContent,
		Emoji:   config.TitleEmoji,
		Writer:  cmd.OutOrStdout(),
	})

	// Show activity message
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: config.ActivityContent,
		Writer:  cmd.OutOrStdout(),
	})

	// Execute cluster creation
	err = config.Action(cmd.Context(), provisioner, clusterName)
	if err != nil {
		return fmt.Errorf("%s: %w", config.ErrorMessagePrefix, err)
	}

	// Show success message with timing
	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    config.SuccessContent,
		Timer:      deps.Timer,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// installCiliumCNI installs Cilium CNI on the cluster.
func installCiliumCNI(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Install CNI...",
		Emoji:   "🌐",
		Writer:  cmd.OutOrStdout(),
	})

	kubeconfig, _, err := loadKubeconfig(clusterCfg)
	if err != nil {
		return err
	}

	helmClient, err := helm.NewClient(kubeconfig, clusterCfg.Spec.Connection.Context)
	if err != nil {
		return fmt.Errorf("failed to create Helm client: %w", err)
	}

	repoErr := addCiliumRepository(cmd.Context(), helmClient)
	if repoErr != nil {
		return repoErr
	}

	installer := newCiliumInstaller(helmClient, kubeconfig, clusterCfg)

	return runCiliumInstallation(cmd, installer, tmr)
}

func addCiliumRepository(ctx context.Context, client helm.Interface) error {
	repoErr := client.AddRepository(ctx, &helm.RepositoryEntry{
		Name: "cilium",
		URL:  "https://helm.cilium.io/",
	})
	if repoErr != nil {
		return fmt.Errorf("failed to add Cilium Helm repository: %w", repoErr)
	}

	return nil
}

func newCiliumInstaller(
	helmClient *helm.Client,
	kubeconfig string,
	clusterCfg *v1alpha1.Cluster,
) *ciliuminstaller.CiliumInstaller {
	timeout := getCiliumInstallTimeout(clusterCfg)

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

// prepareKindConfigWithMirrors prepares the Kind config by adding mirror registry patches if needed.
// Returns true if there are containerd patches to process, false otherwise.
func prepareKindConfigWithMirrors(
	clusterCfg *v1alpha1.Cluster,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
) bool {
	// Only for Kind distribution
	if clusterCfg.Spec.Distribution != v1alpha1.DistributionKind || kindConfig == nil {
		return false
	}

	// Check for --mirror-registry flag overrides
	mirrorRegistries := cfgManager.Viper.GetStringSlice("mirror-registry")
	if len(mirrorRegistries) > 0 {
		// Add containerd patches from flag
		kindConfig.ContainerdConfigPatches = append(
			kindConfig.ContainerdConfigPatches,
			generateContainerdPatchesFromSpecs(mirrorRegistries)...,
		)
	}

	// Return true if there are containerd patches to process
	return len(kindConfig.ContainerdConfigPatches) > 0
}

// setupMirrorRegistries sets up mirror registries for Kind based on the cluster configuration.
// K3d handles registries natively through its own configuration, so no setup is needed.
// The --mirror-registry flag can be used to add/override mirror registry configurations.
func setupMirrorRegistries(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps shared.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
) error {
	// Prepare Kind config with mirror registries
	if !prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig) {
		return nil
	}

	// Start timing for registry setup
	deps.Timer.NewStage()

	// Display title
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Create mirror registries...",
		Emoji:   "🪞",
		Writer:  cmd.OutOrStdout(),
	})

	// Set up registries using Docker client
	return withDockerClient(cmd, func(dockerClient client.APIClient) error {
		err := kindprovisioner.SetupRegistries(
			cmd.Context(),
			kindConfig,
			kindConfig.Name,
			dockerClient,
			cmd.OutOrStdout(),
		)
		if err != nil {
			return fmt.Errorf("failed to setup registries: %w", err)
		}

		// Display success message with timing
		notify.WriteMessage(notify.Message{
			Type:       notify.SuccessType,
			Content:    "mirror registries created",
			Timer:      deps.Timer,
			Writer:     cmd.OutOrStdout(),
			MultiStage: true,
		})

		return nil
	})
}

// connectRegistriesToKindNetwork connects registry containers to the Kind network after cluster creation.
// This is necessary because the Kind network doesn't exist until after the cluster is created.
func connectRegistriesToKindNetwork(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps shared.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
) error {
	// Prepare Kind config with mirror registries
	if !prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig) {
		return nil
	}

	// Start timing for registry network connection
	deps.Timer.NewStage()

	// Display title
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Connect registries to cluster network...",
		Emoji:   "🔗",
		Writer:  cmd.OutOrStdout(),
	})

	// Connect registries to Kind network using Docker client
	return withDockerClient(cmd, func(dockerClient client.APIClient) error {
		err := kindprovisioner.ConnectRegistriesToNetwork(
			cmd.Context(),
			kindConfig,
			dockerClient,
			cmd.OutOrStdout(),
		)
		if err != nil {
			return fmt.Errorf("failed to connect registries to network: %w", err)
		}

		// Display success message with timing
		notify.WriteMessage(notify.Message{
			Type:       notify.SuccessType,
			Content:    "registries connected to cluster network",
			Timer:      deps.Timer,
			Writer:     cmd.OutOrStdout(),
			MultiStage: true,
		})

		return nil
	})
}

// generateContainerdPatchesFromSpecs generates containerd config patches from mirror registry specs.
// Input format: "registry=endpoint" (e.g., "docker.io=http://localhost:5000")
func generateContainerdPatchesFromSpecs(mirrorSpecs []string) []string {
	patches := make([]string, 0, len(mirrorSpecs))

	for _, spec := range mirrorSpecs {
		parts := splitMirrorSpec(spec)
		if parts == nil {
			continue
		}

		patch := fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
  endpoint = ["%s"]`, parts[0], parts[1])

		patches = append(patches, patch)
	}

	return patches
}

// splitMirrorSpec splits a mirror specification into registry and endpoint parts.
// Returns nil if the spec is invalid.
func splitMirrorSpec(spec string) []string {
	for idx, char := range spec {
		if char == '=' {
			if idx == 0 || idx == len(spec)-1 {
				return nil
			}

			return []string{spec[:idx], spec[idx+1:]}
		}
	}

	return nil
}
