package cluster

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	cmdhelpers "github.com/devantler-tech/ksail-go/pkg/cmd"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer"
	calicoinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/calico"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/cilium"
	fluxinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/flux"
	metricsserverinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/metrics-server"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/registry"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/docker/docker/client"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const (
	// k3sDisableMetricsServerFlag is the K3s flag to disable metrics-server.
	k3sDisableMetricsServerFlag = "--disable=metrics-server"
	fluxResourcesActivity       = "applying default resources"
	fluxResourcesSuccess        = "FluxInstance applied"
)

// ErrUnsupportedCNI is returned when an unsupported CNI type is encountered.
var ErrUnsupportedCNI = errors.New("unsupported CNI type")

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

	outputTimer := cmdhelpers.MaybeTimer(cmd, deps.Timer)

	clusterCfg, kindConfig, k3dConfig, err := loadClusterConfiguration(cfgManager, outputTimer)
	if err != nil {
		return err
	}

	err = ensureLocalRegistriesReady(
		cmd,
		clusterCfg,
		deps,
		cfgManager,
		kindConfig,
		k3dConfig,
	)
	if err != nil {
		return err
	}

	// Configure metrics-server for K3d before cluster creation
	setupK3dMetricsServer(clusterCfg, k3dConfig)

	err = executeClusterLifecycle(cmd, clusterCfg, deps)
	if err != nil {
		return err
	}

	connectMirrorRegistriesWithWarning(
		cmd,
		clusterCfg,
		deps,
		cfgManager,
		kindConfig,
		k3dConfig,
	)

	err = executeLocalRegistryStage(
		cmd,
		clusterCfg,
		deps,
		kindConfig,
		k3dConfig,
		localRegistryStageConnect,
	)
	if err != nil {
		return fmt.Errorf("failed to connect local registry: %w", err)
	}

	return handlePostCreationSetup(cmd, clusterCfg, deps.Timer)
}

func loadClusterConfiguration(
	cfgManager *ksailconfigmanager.ConfigManager,
	tmr timer.Timer,
) (*v1alpha1.Cluster, *v1alpha4.Cluster, *v1alpha5.SimpleConfig, error) {
	clusterCfg, err := cfgManager.LoadConfig(tmr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	kindConfig, k3dConfig, err := loadDistributionConfigs(clusterCfg, tmr)
	if err != nil {
		return nil, nil, nil, err
	}

	return clusterCfg, kindConfig, k3dConfig, nil
}

func ensureLocalRegistriesReady(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
) error {
	err := executeLocalRegistryStage(
		cmd,
		clusterCfg,
		deps,
		kindConfig,
		k3dConfig,
		localRegistryStageProvision,
	)
	if err != nil {
		return fmt.Errorf("failed to provision local registry: %w", err)
	}

	err = setupMirrorRegistries(
		cmd,
		clusterCfg,
		deps,
		cfgManager,
		kindConfig,
		k3dConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to setup mirror registries: %w", err)
	}

	return nil
}

func executeClusterLifecycle(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
) error {
	deps.Timer.NewStage()

	err := cmdhelpers.RunLifecycleWithConfig(cmd, deps, newCreateLifecycleConfig(), clusterCfg)
	if err != nil {
		return fmt.Errorf("failed to execute cluster lifecycle: %w", err)
	}

	return nil
}

func connectMirrorRegistriesWithWarning(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
) {
	err := connectRegistriesToClusterNetwork(
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
}

// handlePostCreationSetup installs CNI and metrics-server after cluster creation.
// Order depends on CNI configuration to resolve dependencies.
func handlePostCreationSetup(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	tmr timer.Timer,
) error {
	var err error

	// For custom CNI (Cilium or Calico), install CNI first as metrics-server needs networking
	// For default CNI, install metrics-server first as it's independent
	switch clusterCfg.Spec.CNI {
	case v1alpha1.CNICilium:
		err = installCustomCNIAndMetrics(cmd, clusterCfg, tmr, installCiliumCNI)
	case v1alpha1.CNICalico:
		err = installCustomCNIAndMetrics(cmd, clusterCfg, tmr, installCalicoCNI)
	case v1alpha1.CNIDefault, "":
		err = handleMetricsServer(cmd, clusterCfg, tmr)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedCNI, clusterCfg.Spec.CNI)
	}

	if err != nil {
		return err
	}

	return installFluxIfConfigured(cmd, clusterCfg, tmr)
}

// installCustomCNIAndMetrics installs a custom CNI and then metrics-server.
func installCustomCNIAndMetrics(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	tmr timer.Timer,
	installFunc func(*cobra.Command, *v1alpha1.Cluster, timer.Timer) error,
) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	tmr.NewStage()

	err := installFunc(cmd, clusterCfg, tmr)
	if err != nil {
		return err
	}

	// Install metrics-server after CNI is ready
	return handleMetricsServer(cmd, clusterCfg, tmr)
}

func loadDistributionConfigs(
	clusterCfg *v1alpha1.Cluster,
	lifecycleTimer timer.Timer,
) (*v1alpha4.Cluster, *v1alpha5.SimpleConfig, error) {
	defaultConfigPath := defaultDistributionConfigPath(clusterCfg.Spec.Distribution)

	configPath := strings.TrimSpace(clusterCfg.Spec.DistributionConfig)
	if configPath == "" || strings.EqualFold(configPath, "auto") {
		clusterCfg.Spec.DistributionConfig = defaultConfigPath
		configPath = defaultConfigPath
	}

	// If distribution is K3d but the config path still points to the kind default,
	// switch to the k3d default so we don’t try to read kind.yaml.
	if clusterCfg.Spec.Distribution == v1alpha1.DistributionK3d &&
		configPath == defaultDistributionConfigPath(v1alpha1.DistributionKind) {
		clusterCfg.Spec.DistributionConfig = defaultConfigPath
		configPath = defaultConfigPath
	}

	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		manager := kindconfigmanager.NewConfigManager(configPath)

		kindConfig, err := manager.LoadConfig(lifecycleTimer)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load kind config: %w", err)
		}

		return kindConfig, nil, nil
	case v1alpha1.DistributionK3d:
		manager := k3dconfigmanager.NewConfigManager(configPath)

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

const (
	mirrorStageTitle    = "Create mirror registry..."
	mirrorStageEmoji    = "🪞"
	mirrorStageActivity = "creating mirror registries"
	mirrorStageSuccess  = "mirror registries created"
	mirrorStageFailure  = "failed to setup registries"

	connectStageTitle    = "Connect registry..."
	connectStageEmoji    = "🔗"
	connectStageActivity = "connecting registries"
	connectStageSuccess  = "registries connected"
	connectStageFailure  = "failed to connect registries"

	fluxStageTitle    = "Install Flux..."
	fluxStageEmoji    = "☸️"
	fluxStageActivity = "installing Flux controllers"
)

var (
	//nolint:gochecknoglobals // Shared stage configuration used by lifecycle helpers.
	mirrorRegistryStageInfo = registryStageInfo{
		title:         mirrorStageTitle,
		emoji:         mirrorStageEmoji,
		activity:      mirrorStageActivity,
		success:       mirrorStageSuccess,
		failurePrefix: mirrorStageFailure,
	}
	//nolint:gochecknoglobals // Shared stage configuration used by lifecycle helpers.
	connectRegistryStageInfo = registryStageInfo{
		title:         connectStageTitle,
		emoji:         connectStageEmoji,
		activity:      connectStageActivity,
		success:       connectStageSuccess,
		failurePrefix: connectStageFailure,
	}
	//nolint:gochecknoglobals // Stage action definitions reused across lifecycle flows.
	registryStageDefinitions = map[registryStageRole]registryStageDefinition{
		registryStageRoleMirror: {
			info:       mirrorRegistryStageInfo,
			kindAction: kindRegistryActionFor(registryStageRoleMirror),
			k3dAction:  k3dRegistryActionFor(registryStageRoleMirror),
		},
		registryStageRoleConnect: {
			info:       connectRegistryStageInfo,
			kindAction: kindRegistryActionFor(registryStageRoleConnect),
			k3dAction:  k3dRegistryActionFor(registryStageRoleConnect),
		},
	}
	// setupMirrorRegistries configures mirror registries before cluster creation.
	//nolint:gochecknoglobals // Function reused by tests and runtime flow.
	setupMirrorRegistries = makeRegistryStageRunner(registryStageRoleMirror)
	// connectRegistriesToClusterNetwork attaches mirror registries to the cluster network after creation.
	//nolint:gochecknoglobals // Function reused by tests and runtime flow.
	connectRegistriesToClusterNetwork = makeRegistryStageRunner(registryStageRoleConnect)
	// fluxInstallerFactory is overridden in tests to stub Flux installer creation.
	//nolint:gochecknoglobals // dependency injection for tests
	fluxInstallerFactory = func(client helm.Interface, timeout time.Duration) installer.Installer {
		return fluxinstaller.NewFluxInstaller(client, timeout)
	}
	// ensureFluxResourcesFunc enforces default Flux resources post-install.
	//nolint:gochecknoglobals // dependency injection for tests
	ensureFluxResourcesFunc = fluxinstaller.EnsureDefaultResources
	// dockerClientInvoker can be overridden in tests to avoid real Docker connections.
	//nolint:gochecknoglobals // dependency injection for tests
	dockerClientInvoker = cmdhelpers.WithDockerClient
	// dockerClientInvokerMu protects concurrent access to dockerClientInvoker in tests.
	//nolint:gochecknoglobals // protects dockerClientInvoker global variable
	dockerClientInvokerMu sync.RWMutex
)

type registryStageRole int

const (
	registryStageRoleMirror registryStageRole = iota
	registryStageRoleConnect
)

type registryStageInfo struct {
	title         string
	emoji         string
	activity      string
	success       string
	failurePrefix string
}

type registryStageHandler struct {
	prepare func() bool
	action  func(context.Context, client.APIClient) error
}

type registryStageContext struct {
	cmd         *cobra.Command
	clusterCfg  *v1alpha1.Cluster
	kindConfig  *v1alpha4.Cluster
	k3dConfig   *v1alpha5.SimpleConfig
	mirrorSpecs []registry.MirrorSpec
}

type registryStageDefinition struct {
	info       registryStageInfo
	kindAction func(*registryStageContext) func(context.Context, client.APIClient) error
	k3dAction  func(*registryStageContext) func(context.Context, client.APIClient) error
}

type registryAction func(context.Context, *registryStageContext, client.APIClient) error

func registryActionFor(
	role registryStageRole,
	selectAction func(registryStageRole) registryAction,
) func(*registryStageContext) func(context.Context, client.APIClient) error {
	return func(ctx *registryStageContext) func(context.Context, client.APIClient) error {
		action := selectAction(role)

		if action == nil {
			return func(context.Context, client.APIClient) error {
				return nil
			}
		}

		return func(execCtx context.Context, dockerClient client.APIClient) error {
			return action(execCtx, ctx, dockerClient)
		}
	}
}

func makeRegistryStageRunner(role registryStageRole) func(
	*cobra.Command,
	*v1alpha1.Cluster,
	cmdhelpers.LifecycleDeps,
	*ksailconfigmanager.ConfigManager,
	*v1alpha4.Cluster,
	*v1alpha5.SimpleConfig,
) error {
	return func(
		cmd *cobra.Command,
		clusterCfg *v1alpha1.Cluster,
		deps cmdhelpers.LifecycleDeps,
		cfgManager *ksailconfigmanager.ConfigManager,
		kindConfig *v1alpha4.Cluster,
		k3dConfig *v1alpha5.SimpleConfig,
	) error {
		return runRegistryStageWithRole(
			cmd,
			clusterCfg,
			deps,
			cfgManager,
			kindConfig,
			k3dConfig,
			role,
		)
	}
}

func kindRegistryActionFor(
	role registryStageRole,
) func(*registryStageContext) func(context.Context, client.APIClient) error {
	return registryActionFor(role, func(currentRole registryStageRole) registryAction {
		switch currentRole {
		case registryStageRoleMirror:
			return runKindMirrorAction
		case registryStageRoleConnect:
			return runKindConnectAction
		default:
			return nil
		}
	})
}

func runKindMirrorAction(
	execCtx context.Context,
	ctx *registryStageContext,
	dockerClient client.APIClient,
) error {
	writer := ctx.cmd.OutOrStdout()
	clusterName := ctx.kindConfig.Name

	err := kindprovisioner.SetupRegistries(
		execCtx,
		ctx.kindConfig,
		clusterName,
		dockerClient,
		ctx.mirrorSpecs,
		writer,
	)
	if err != nil {
		return fmt.Errorf("failed to setup kind registries: %w", err)
	}

	return nil
}

func runKindConnectAction(
	execCtx context.Context,
	ctx *registryStageContext,
	dockerClient client.APIClient,
) error {
	err := kindprovisioner.ConnectRegistriesToNetwork(
		execCtx,
		ctx.kindConfig,
		dockerClient,
		ctx.cmd.OutOrStdout(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect kind registries to network: %w", err)
	}

	return nil
}

type k3dRegistryAction func(context.Context, *v1alpha5.SimpleConfig, string, client.APIClient, io.Writer) error

func k3dRegistryActionFor(
	role registryStageRole,
) func(*registryStageContext) func(context.Context, client.APIClient) error {
	return registryActionFor(role, func(currentRole registryStageRole) registryAction {
		switch currentRole {
		case registryStageRoleMirror:
			return func(execCtx context.Context, ctx *registryStageContext, dockerClient client.APIClient) error {
				return runK3DRegistryAction(
					execCtx,
					ctx,
					dockerClient,
					"setup k3d registries",
					k3dprovisioner.SetupRegistries,
				)
			}
		case registryStageRoleConnect:
			return func(execCtx context.Context, ctx *registryStageContext, dockerClient client.APIClient) error {
				return runK3DRegistryAction(
					execCtx,
					ctx,
					dockerClient,
					"connect k3d registries to network",
					k3dprovisioner.ConnectRegistriesToNetwork,
				)
			}
		default:
			return nil
		}
	})
}

func runK3DRegistryAction(
	execCtx context.Context,
	ctx *registryStageContext,
	dockerClient client.APIClient,
	description string,
	action k3dRegistryAction,
) error {
	if action == nil {
		return nil
	}

	targetName := k3dconfigmanager.ResolveClusterName(ctx.clusterCfg, ctx.k3dConfig)
	writer := ctx.cmd.OutOrStdout()

	err := action(execCtx, ctx.k3dConfig, targetName, dockerClient, writer)
	if err != nil {
		return fmt.Errorf("failed to %s: %w", description, err)
	}

	return nil
}

func newRegistryHandlers(
	clusterCfg *v1alpha1.Cluster,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
	mirrorSpecs []registry.MirrorSpec,
	kindAction func(context.Context, client.APIClient) error,
	k3dAction func(context.Context, client.APIClient) error,
) map[v1alpha1.Distribution]registryStageHandler {
	return map[v1alpha1.Distribution]registryStageHandler{
		v1alpha1.DistributionKind: {
			prepare: func() bool { return prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig) },
			action:  kindAction,
		},
		v1alpha1.DistributionK3d: {
			prepare: func() bool { return prepareK3dConfigWithMirrors(clusterCfg, k3dConfig, mirrorSpecs) },
			action:  k3dAction,
		},
	}
}

func handleRegistryStage(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
	info registryStageInfo,
	mirrorSpecs []registry.MirrorSpec,
	kindAction func(context.Context, client.APIClient) error,
	k3dAction func(context.Context, client.APIClient) error,
) error {
	handlers := newRegistryHandlers(
		clusterCfg,
		cfgManager,
		kindConfig,
		k3dConfig,
		mirrorSpecs,
		kindAction,
		k3dAction,
	)

	handler, ok := handlers[clusterCfg.Spec.Distribution]
	if !ok {
		return nil
	}

	return executeRegistryStage(cmd, deps, info, handler.prepare, handler.action)
}

func runRegistryStageWithRole(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	cfgManager *ksailconfigmanager.ConfigManager,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *v1alpha5.SimpleConfig,
	role registryStageRole,
) error {
	mirrorSpecs := registry.ParseMirrorSpecs(
		cfgManager.Viper.GetStringSlice("mirror-registry"),
	)

	definition, ok := registryStageDefinitions[role]
	if !ok {
		return nil
	}

	stageCtx := &registryStageContext{
		cmd:         cmd,
		clusterCfg:  clusterCfg,
		kindConfig:  kindConfig,
		k3dConfig:   k3dConfig,
		mirrorSpecs: mirrorSpecs,
	}

	kindAction := definition.kindAction(stageCtx)
	k3dAction := definition.k3dAction(stageCtx)

	return handleRegistryStage(
		cmd,
		clusterCfg,
		deps,
		cfgManager,
		kindConfig,
		k3dConfig,
		definition.info,
		mirrorSpecs,
		kindAction,
		k3dAction,
	)
}

func executeRegistryStage(
	cmd *cobra.Command,
	deps cmdhelpers.LifecycleDeps,
	info registryStageInfo,
	shouldPrepare func() bool,
	action func(context.Context, client.APIClient) error,
) error {
	if !shouldPrepare() {
		return nil
	}

	return runRegistryStage(cmd, deps, info, action)
}

func runRegistryStage(
	cmd *cobra.Command,
	deps cmdhelpers.LifecycleDeps,
	info registryStageInfo,
	action func(context.Context, client.APIClient) error,
) error {
	deps.Timer.NewStage()

	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: info.title,
		Emoji:   info.emoji,
		Writer:  cmd.OutOrStdout(),
	})

	if info.activity != "" {
		notify.WriteMessage(notify.Message{
			Type:    notify.ActivityType,
			Content: info.activity,
			Writer:  cmd.OutOrStdout(),
		})
	}

	dockerClientInvokerMu.RLock()

	invoker := dockerClientInvoker

	dockerClientInvokerMu.RUnlock()

	err := invoker(cmd, func(dockerClient client.APIClient) error {
		err := action(cmd.Context(), dockerClient)
		if err != nil {
			return fmt.Errorf("%s: %w", info.failurePrefix, err)
		}

		outputTimer := cmdhelpers.MaybeTimer(cmd, deps.Timer)

		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: info.success,
			Timer:   outputTimer,
			Writer:  cmd.OutOrStdout(),
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to execute registry stage: %w", err)
	}

	return nil
}

// createHelmClientForCluster creates a Helm client configured for the cluster.
func createHelmClientForCluster(clusterCfg *v1alpha1.Cluster) (*helm.Client, string, error) {
	kubeconfig, err := cmdhelpers.GetKubeconfigPathFromConfig(clusterCfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get kubeconfig path: %w", err)
	}

	// Validate file exists
	_, err = os.Stat(kubeconfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to access kubeconfig file: %w", err)
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
	timeout := installer.GetInstallTimeout(clusterCfg)

	return ciliuminstaller.NewCiliumInstaller(
		helmClient,
		kubeconfig,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)
}

// cniInstaller defines the interface for CNI installers.
type cniInstaller interface {
	Install(ctx context.Context) error
	WaitForReadiness(ctx context.Context) error
}

func runCiliumInstallation(
	cmd *cobra.Command,
	installer *ciliuminstaller.CiliumInstaller,
	tmr timer.Timer,
) error {
	return runCNIInstallation(cmd, installer, "cilium", tmr)
}

// installCalicoCNI installs Calico CNI on the cluster.
func installCalicoCNI(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
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

	installer := newCalicoInstaller(helmClient, kubeconfig, clusterCfg)

	return runCalicoInstallation(cmd, installer, tmr)
}

func newCalicoInstaller(
	helmClient *helm.Client,
	kubeconfig string,
	clusterCfg *v1alpha1.Cluster,
) *calicoinstaller.CalicoInstaller {
	timeout := installer.GetInstallTimeout(clusterCfg)

	return calicoinstaller.NewCalicoInstaller(
		helmClient,
		kubeconfig,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)
}

func runCalicoInstallation(
	cmd *cobra.Command,
	installer *calicoinstaller.CalicoInstaller,
	tmr timer.Timer,
) error {
	return runCNIInstallation(cmd, installer, "calico", tmr)
}

// runCNIInstallation is the generic implementation for running CNI installation.
func runCNIInstallation(
	cmd *cobra.Command,
	installer cniInstaller,
	cniName string,
	tmr timer.Timer,
) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "installing " + cniName,
		Writer:  cmd.OutOrStdout(),
	})

	installErr := installer.Install(cmd.Context())
	if installErr != nil {
		return fmt.Errorf("%s installation failed: %w", cniName, installErr)
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "awaiting " + cniName + " to be ready",
		Writer:  cmd.OutOrStdout(),
	})

	readinessErr := installer.WaitForReadiness(cmd.Context())
	if readinessErr != nil {
		return fmt.Errorf("%s readiness check failed: %w", cniName, readinessErr)
	}

	outputTimer := cmdhelpers.MaybeTimer(cmd, tmr)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "CNI installed",
		Timer:   outputTimer,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
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
	mirrorSpecs []registry.MirrorSpec,
) bool {
	if clusterCfg.Spec.Distribution != v1alpha1.DistributionK3d || k3dConfig == nil {
		return false
	}

	original := k3dConfig.Registries.Config

	hostEndpoints := k3dconfigmanager.ParseRegistryConfig(original)

	updatedMap, _ := registry.BuildHostEndpointMap(mirrorSpecs, "", hostEndpoints)
	if len(updatedMap) == 0 {
		return false
	}

	rendered := registry.RenderK3dMirrorConfig(updatedMap)

	if strings.TrimSpace(rendered) == strings.TrimSpace(original) {
		return strings.TrimSpace(original) != ""
	}

	k3dConfig.Registries.Config = rendered

	return true
}

// generateContainerdPatchesFromSpecs generates containerd config patches from mirror registry specs.
// Input format: "registry=endpoint" (e.g., "docker.io=http://localhost:5000")
func generateContainerdPatchesFromSpecs(mirrorSpecs []string) []string {
	parsed := registry.ParseMirrorSpecs(mirrorSpecs)
	if len(parsed) == 0 {
		return nil
	}

	entries := registry.BuildMirrorEntries(parsed, "", nil, nil, nil)

	patches := make([]string, 0, len(entries))
	for _, entry := range entries {
		patches = append(patches, registry.KindPatch(entry))
	}

	return patches
}

// handleMetricsServer manages metrics-server installation based on cluster configuration.
// For K3d, metrics-server should be disabled via config (handled in setupK3dMetricsServer), not uninstalled.
func handleMetricsServer(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
	// Check if distribution provides metrics-server by default
	hasMetricsByDefault := clusterCfg.Spec.Distribution.ProvidesMetricsServerByDefault()

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

	timeout := installer.GetInstallTimeout(clusterCfg)
	msInstaller := metricsserverinstaller.NewMetricsServerInstaller(
		helmClient,
		kubeconfig,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)

	return runMetricsServerInstallation(cmd, msInstaller, tmr)
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

	outputTimer := cmdhelpers.MaybeTimer(cmd, tmr)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "Metrics Server installed",
		Timer:   outputTimer,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}

func installFluxIfConfigured(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	tmr timer.Timer,
) error {
	if clusterCfg.Spec.GitOpsEngine != v1alpha1.GitOpsEngineFlux {
		return nil
	}

	helmClient, kubeconfig, err := createHelmClientForCluster(clusterCfg)
	if err != nil {
		return err
	}

	fluxInstaller := newFluxInstallerForCluster(clusterCfg, helmClient)

	err = runFluxInstallation(cmd, fluxInstaller, tmr)
	if err != nil {
		return err
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: fluxResourcesActivity,
		Writer:  cmd.OutOrStdout(),
	})

	err = ensureFluxResourcesFunc(cmd.Context(), kubeconfig, clusterCfg)
	if err != nil {
		return fmt.Errorf("failed to configure Flux resources: %w", err)
	}

	outputTimer := cmdhelpers.MaybeTimer(cmd, tmr)

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: fluxResourcesSuccess,
		Timer:   outputTimer,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}

// newFluxInstallerForCluster returns an installer tuned for the cluster context.
//
//nolint:ireturn // returns interface for dependency injection in tests
func newFluxInstallerForCluster(
	clusterCfg *v1alpha1.Cluster,
	helmClient helm.Interface,
) installer.Installer {
	timeout := installer.GetInstallTimeout(clusterCfg)

	return fluxInstallerFactory(helmClient, timeout)
}

func runFluxInstallation(
	cmd *cobra.Command,
	installer installer.Installer,
	tmr timer.Timer,
) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: fluxStageTitle,
		Emoji:   fluxStageEmoji,
		Writer:  cmd.OutOrStdout(),
	})

	tmr.NewStage()

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: fluxStageActivity,
		Writer:  cmd.OutOrStdout(),
	})

	err := installer.Install(ctx)
	if err != nil {
		return fmt.Errorf("failed to install flux controllers: %w", err)
	}

	return nil
}
