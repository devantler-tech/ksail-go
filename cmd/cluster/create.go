package cluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/devantler-tech/ksail-go/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailio "github.com/devantler-tech/ksail-go/pkg/io"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registries"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/docker/docker/client"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/yaml"
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
			"Configure mirror registries with format 'host=upstream' (e.g., docker.io=https://registry-1.docker.io)")
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

	deps.Timer.NewStage()

	err = shared.RunLifecycleWithConfig(cmd, deps, newCreateLifecycleConfig(), clusterCfg)
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

	if clusterCfg.Spec.CNI == v1alpha1.CNICilium {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())

		deps.Timer.NewStage()

		err = installCiliumCNI(cmd, clusterCfg, deps.Timer)
		if err != nil {
			return fmt.Errorf("failed to install Cilium CNI: %w", err)
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

// setupMirrorRegistries configures mirror registries before cluster creation.
func setupMirrorRegistries(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps shared.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
) error {
	mirrorSpecs := registries.ParseMirrorSpecs(
		cfgManager.Viper.GetStringSlice("mirror-registry"),
	)

	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return setupKindMirrorRegistries(cmd, deps, clusterCfg, cfgManager, kindConfig, mirrorSpecs)
	case v1alpha1.DistributionK3d:
		return setupK3dMirrorRegistries(cmd, deps, clusterCfg, k3dConfig, mirrorSpecs)
	default:
		return nil
	}
}

func setupKindMirrorRegistries(
	cmd *cobra.Command,
	deps shared.LifecycleDeps,
	clusterCfg *v1alpha1.Cluster,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	mirrorSpecs []registries.MirrorSpec,
) error {
	if !prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig) {
		return nil
	}

	stage := registryStage{
		title:         "Create mirror registries...",
		emoji:         "🪞",
		success:       "mirror registries created",
		failurePrefix: "failed to setup registries",
		action: func(ctx context.Context, dockerClient client.APIClient) error {
			return kindprovisioner.SetupRegistries(
				ctx,
				kindConfig,
				kindConfig.Name,
				dockerClient,
				mirrorSpecs,
				cmd.OutOrStdout(),
			)
		},
	}

	return runRegistryStage(cmd, deps, stage)
}

func setupK3dMirrorRegistries(
	cmd *cobra.Command,
	deps shared.LifecycleDeps,
	clusterCfg *v1alpha1.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
	mirrorSpecs []registries.MirrorSpec,
) error {
	if !prepareK3dConfigWithMirrors(clusterCfg, k3dConfig, mirrorSpecs) {
		return nil
	}

	stage := registryStage{
		title:         "Create mirror registries...",
		emoji:         "🪞",
		success:       "mirror registries created",
		failurePrefix: "failed to setup registries",
		action: func(ctx context.Context, dockerClient client.APIClient) error {
			return k3dprovisioner.SetupRegistries(
				ctx,
				k3dConfig,
				resolveK3dClusterName(clusterCfg, k3dConfig),
				dockerClient,
				cmd.OutOrStdout(),
			)
		},
	}

	return runRegistryStage(cmd, deps, stage)
}

// connectRegistriesToClusterNetwork attaches mirror registries to the cluster network after creation.
func connectRegistriesToClusterNetwork(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps shared.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
) error {
	mirrorSpecs := registries.ParseMirrorSpecs(
		cfgManager.Viper.GetStringSlice("mirror-registry"),
	)

	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return connectKindRegistriesToNetwork(cmd, deps, clusterCfg, cfgManager, kindConfig)
	case v1alpha1.DistributionK3d:
		return connectK3dRegistriesToNetwork(cmd, deps, clusterCfg, k3dConfig, mirrorSpecs)
	default:
		return nil
	}
}

func connectKindRegistriesToNetwork(
	cmd *cobra.Command,
	deps shared.LifecycleDeps,
	clusterCfg *v1alpha1.Cluster,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
) error {
	if !prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig) {
		return nil
	}

	stage := registryStage{
		title:         "Connect registries...",
		emoji:         "🔗",
		success:       "registries connected",
		failurePrefix: "failed to connect registries",
		action: func(ctx context.Context, dockerClient client.APIClient) error {
			return kindprovisioner.ConnectRegistriesToNetwork(
				ctx,
				kindConfig,
				dockerClient,
				cmd.OutOrStdout(),
			)
		},
	}

	return runRegistryStage(cmd, deps, stage)
}

func connectK3dRegistriesToNetwork(
	cmd *cobra.Command,
	deps shared.LifecycleDeps,
	clusterCfg *v1alpha1.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
	mirrorSpecs []registries.MirrorSpec,
) error {
	if !prepareK3dConfigWithMirrors(clusterCfg, k3dConfig, mirrorSpecs) {
		return nil
	}

	stage := registryStage{
		title:         "Connect registries...",
		emoji:         "🔗",
		success:       "registries connected",
		failurePrefix: "failed to connect registries",
		action: func(ctx context.Context, dockerClient client.APIClient) error {
			return k3dprovisioner.ConnectRegistriesToNetwork(
				ctx,
				k3dConfig,
				resolveK3dClusterName(clusterCfg, k3dConfig),
				dockerClient,
				cmd.OutOrStdout(),
			)
		},
	}

	return runRegistryStage(cmd, deps, stage)
}

type registryStage struct {
	title         string
	emoji         string
	success       string
	failurePrefix string
	action        func(context.Context, client.APIClient) error
}

func runRegistryStage(
	cmd *cobra.Command,
	deps shared.LifecycleDeps,
	stage registryStage,
) error {
	deps.Timer.NewStage()

	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: stage.title,
		Emoji:   stage.emoji,
		Writer:  cmd.OutOrStdout(),
	})

	return withDockerClient(cmd, func(dockerClient client.APIClient) error {
		err := stage.action(cmd.Context(), dockerClient)
		if err != nil {
			return fmt.Errorf("%s: %w", stage.failurePrefix, err)
		}

		notify.WriteMessage(notify.Message{
			Type:       notify.SuccessType,
			Content:    stage.success,
			Timer:      deps.Timer,
			Writer:     cmd.OutOrStdout(),
			MultiStage: true,
		})

		return nil
	})
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

func prepareK3dConfigWithMirrors(
	clusterCfg *v1alpha1.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
	mirrorSpecs []registries.MirrorSpec,
) bool {
	if clusterCfg.Spec.Distribution != v1alpha1.DistributionK3d || k3dConfig == nil {
		return false
	}

	original := k3dConfig.Registries.Config

	hostEndpoints := parseK3dRegistryConfig(original)

	updatedMap, _ := registries.BuildHostEndpointMap(mirrorSpecs, "", hostEndpoints)
	if len(updatedMap) == 0 {
		return false
	}

	rendered := registries.RenderK3dMirrorConfig(updatedMap)

	if strings.TrimSpace(rendered) == strings.TrimSpace(original) {
		return strings.TrimSpace(original) != ""
	}

	k3dConfig.Registries.Config = rendered

	return true
}

type mirrorConfigEntry struct {
	Endpoint []string `yaml:"endpoint"`
}

type k3dMirrorConfig struct {
	Mirrors map[string]mirrorConfigEntry `yaml:"mirrors"`
}

func parseK3dRegistryConfig(raw string) map[string][]string {
	result := make(map[string][]string)

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return result
	}

	var cfg k3dMirrorConfig

	err := yaml.Unmarshal([]byte(trimmed), &cfg)
	if err != nil {
		return result
	}

	for host, entry := range cfg.Mirrors {
		if len(entry.Endpoint) == 0 {
			continue
		}

		filtered := make([]string, 0, len(entry.Endpoint))
		for _, endpoint := range entry.Endpoint {
			endpoint = strings.TrimSpace(endpoint)
			if endpoint == "" {
				continue
			}

			filtered = append(filtered, endpoint)
		}

		if len(filtered) == 0 {
			continue
		}

		result[host] = filtered
	}

	return result
}

func resolveK3dClusterName(
	clusterCfg *v1alpha1.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
) string {
	if k3dConfig != nil {
		if name := strings.TrimSpace(k3dConfig.Name); name != "" {
			return name
		}
	}

	if name := strings.TrimSpace(clusterCfg.Spec.Connection.Context); name != "" {
		return name
	}

	return "k3d"
}

// generateContainerdPatchesFromSpecs generates containerd config patches from mirror registry specs.
// Input format: "registry=endpoint" (e.g., "docker.io=http://localhost:5000")
func generateContainerdPatchesFromSpecs(mirrorSpecs []string) []string {
	parsed := registries.ParseMirrorSpecs(mirrorSpecs)
	if len(parsed) == 0 {
		return nil
	}

	entries := registries.BuildMirrorEntries(parsed, "", nil, nil, nil)

	patches := make([]string, 0, len(entries))
	for _, entry := range entries {
		patches = append(patches, registries.KindPatch(entry))
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
