package cluster

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	cmdhelpers "github.com/devantler-tech/ksail-go/pkg/cmd"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/k8s"
	calicoinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/calico"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/cilium"
	flannelinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/flannel"
	metricsserverinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/metrics-server"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registries"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/docker/docker/client"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	ys "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/yaml"
)

const (
	// k3sDisableMetricsServerFlag is the K3s flag to disable metrics-server.
	k3sDisableMetricsServerFlag = "--disable=metrics-server"
	applyFieldManager           = "ksail-cni-bootstrap"
)

// ErrUnsupportedCNI is returned when an unsupported CNI type is encountered.
var ErrUnsupportedCNI = errors.New("unsupported CNI type")

// flannelInstallHook allows tests to override the Flannel installation flow.
//
//nolint:gochecknoglobals // Intentional test hook for failure simulation.
var (
	flannelInstallHook = installFlannelCNI
	flannelHookMu      sync.RWMutex
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

	return handlePostCreationSetup(cmd, clusterCfg, deps)
}

// handlePostCreationSetup installs CNI and metrics-server after cluster creation.
// Order depends on CNI configuration to resolve dependencies.
func handlePostCreationSetup(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
) error {
	tmr := deps.Timer
	// For custom CNI (Cilium, Calico, Flannel), install CNI first as metrics-server needs networking
	// For default CNI, install metrics-server first as it's independent
	switch clusterCfg.Spec.CNI {
	case v1alpha1.CNICilium:
		return installCustomCNIAndMetrics(cmd, clusterCfg, tmr, installCiliumCNI)
	case v1alpha1.CNICalico:
		return installCustomCNIAndMetrics(cmd, clusterCfg, tmr, installCalicoCNI)
	case v1alpha1.CNIFlannel:
		// K3d has native Flannel support, only install for Kind
		if clusterCfg.Spec.Distribution == v1alpha1.DistributionK3d {
			return handleMetricsServer(cmd, clusterCfg, tmr)
		}

		flannelHookMu.RLock()

		hook := flannelInstallHook

		flannelHookMu.RUnlock()

		err := installCustomCNIAndMetrics(cmd, clusterCfg, tmr, hook)
		if err != nil {
			rollbackErr := rollbackCluster(cmd, clusterCfg, deps)
			if rollbackErr != nil {
				notify.WriteMessage(notify.Message{
					Type:    notify.ErrorType,
					Content: fmt.Sprintf("rollback failed: %v", rollbackErr),
					Writer:  cmd.OutOrStdout(),
				})

				return fmt.Errorf("%w (rollback failed: %w)", err, rollbackErr)
			}

			notify.WriteMessage(notify.Message{
				Type:    notify.WarningType,
				Content: "cluster deleted after Flannel installation failure",
				Writer:  cmd.OutOrStdout(),
			})

			return err
		}

		return nil
	case v1alpha1.CNIDefault, "":
		return handleMetricsServer(cmd, clusterCfg, tmr)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedCNI, clusterCfg.Spec.CNI)
	}
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

func rollbackCluster(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.WarningType,
		Content: "attempting cluster rollback after Flannel installation failure",
		Writer:  cmd.OutOrStdout(),
	})

	provisioner, distributionConfig, err := deps.Factory.Create(cmd.Context(), clusterCfg)
	if err != nil {
		return fmt.Errorf("resolve provisioner for rollback: %w", err)
	}

	clusterName, err := configmanager.GetClusterName(distributionConfig)
	if err != nil {
		return fmt.Errorf("determine cluster name for rollback: %w", err)
	}

	deleteErr := provisioner.Delete(cmd.Context(), clusterName)
	if deleteErr != nil {
		return fmt.Errorf("delete cluster %s during rollback: %w", clusterName, deleteErr)
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.WarningType,
		Content: fmt.Sprintf("rollback complete: cluster %s removed", clusterName),
		Writer:  cmd.OutOrStdout(),
	})

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

const (
	mirrorStageTitle   = "Create mirror registries..."
	mirrorStageEmoji   = "🪞"
	mirrorStageSuccess = "mirror registries created"
	mirrorStageFailure = "failed to setup registries"

	connectStageTitle   = "Connect registries..."
	connectStageEmoji   = "🔗"
	connectStageSuccess = "registries connected"
	connectStageFailure = "failed to connect registries"
)

var (
	//nolint:gochecknoglobals // Shared stage configuration used by lifecycle helpers.
	mirrorRegistryStageInfo = registryStageInfo{
		title:         mirrorStageTitle,
		emoji:         mirrorStageEmoji,
		success:       mirrorStageSuccess,
		failurePrefix: mirrorStageFailure,
	}
	//nolint:gochecknoglobals // Shared stage configuration used by lifecycle helpers.
	connectRegistryStageInfo = registryStageInfo{
		title:         connectStageTitle,
		emoji:         connectStageEmoji,
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
)

type registryStageRole int

const (
	registryStageRoleMirror registryStageRole = iota
	registryStageRoleConnect
)

type registryStageInfo struct {
	title         string
	emoji         string
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
	mirrorSpecs []registries.MirrorSpec
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

	targetName := resolveK3dClusterName(ctx.clusterCfg, ctx.k3dConfig)
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
	mirrorSpecs []registries.MirrorSpec,
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
	mirrorSpecs []registries.MirrorSpec,
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
	mirrorSpecs := registries.ParseMirrorSpecs(
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

	err := cmdhelpers.WithDockerClient(cmd, func(dockerClient client.APIClient) error {
		err := action(cmd.Context(), dockerClient)
		if err != nil {
			return fmt.Errorf("%s: %w", info.failurePrefix, err)
		}

		notify.WriteMessage(notify.Message{
			Type:       notify.SuccessType,
			Content:    info.success,
			Timer:      deps.Timer,
			Writer:     cmd.OutOrStdout(),
			MultiStage: true,
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
func installFlannelCNI(cmd *cobra.Command, clusterCfg *v1alpha1.Cluster, tmr timer.Timer) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Install CNI...",
		Emoji:   "🌐",
		Writer:  cmd.OutOrStdout(),
	})

	kubeconfigPath, err := loadKubeconfig(clusterCfg)
	if err != nil {
		return err
	}

	kubectlClient, err := kubectl.NewManifestClient(
		kubeconfigPath,
		clusterCfg.Spec.Connection.Context,
	)
	if err != nil {
		return fmt.Errorf("failed to create kubectl client: %w", err)
	}

	// For Kind + Flannel we must ensure base CNI plugins (bridge, etc.) exist before Flannel.
	if clusterCfg.Spec.Distribution == v1alpha1.DistributionKind {
		installErr := ensureKindBaseCNIPlugins(
			cmd.Context(),
			kubeconfigPath,
			clusterCfg.Spec.Connection.Context,
			getInstallTimeout(clusterCfg),
			cmd.OutOrStdout(),
		)
		if installErr != nil {
			return fmt.Errorf("failed to bootstrap base CNI plugins: %w", installErr)
		}
	}

	installer := newFlannelInstaller(kubectlClient, kubeconfigPath, clusterCfg)

	return runFlannelInstallation(cmd, installer, tmr)
}

// ensureKindBaseCNIPlugins applies the bootstrap DaemonSet and waits for readiness (all pods running).
func ensureKindBaseCNIPlugins(
	ctx context.Context,
	kubeconfigPath, contextName string,
	timeout time.Duration,
	writer io.Writer,
) error {
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "bootstrapping base CNI plugins",
		Writer:  writer,
	})

	config, mapper, dynClient, err := prepareBootstrapClients(kubeconfigPath, contextName)
	if err != nil {
		return err
	}

	err = applyBaseCNIManifest(ctx, mapper, dynClient)
	if err != nil {
		return err
	}

	err = waitForBootstrapReadiness(ctx, config, timeout)
	if err != nil {
		return err
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "base CNI plugins ready",
		Writer:  writer,
	})

	return nil
}

func prepareBootstrapClients(
	kubeconfigPath, contextName string,
) (*rest.Config, *restmapper.DeferredDiscoveryRESTMapper, *dynamic.DynamicClient, error) {
	config, err := buildRESTConfigWithContext(kubeconfigPath, contextName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create discovery client: %w", err)
	}

	cached := memory.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cached)

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create dynamic client: %w", err)
	}

	return config, mapper, dynClient, nil
}

func applyBaseCNIManifest(
	ctx context.Context,
	mapper meta.RESTMapper,
	dynClient dynamic.Interface,
) error {
	decoder := ys.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	manifestDocs := strings.Split(flannelinstaller.KindBaseCNIPluginsManifest, "\n---\n")

	for docIndex, rawDoc := range manifestDocs {
		err := applyBootstrapDocument(ctx, decoder, mapper, dynClient, rawDoc, docIndex)
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForBootstrapReadiness(
	ctx context.Context,
	config *rest.Config,
	timeout time.Duration,
) error {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("create kube client: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	waitErr := k8s.WaitForDaemonSetReady(
		waitCtx,
		kubeClient,
		"kube-system",
		"cni-bootstrap",
		timeout,
	)
	if waitErr != nil {
		return fmt.Errorf("cni-bootstrap daemonset not ready: %w", waitErr)
	}

	return nil
}

func buildRESTConfigWithContext(kubeconfigPath, contextName string) (*rest.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}

	overrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("build REST config for context %q: %w", contextName, err)
	}

	return config, nil
}

func applyBootstrapDocument(
	ctx context.Context,
	decoder apiruntime.Decoder,
	mapper meta.RESTMapper,
	dynClient dynamic.Interface,
	rawDoc string,
	docIndex int,
) error {
	document := strings.TrimSpace(rawDoc)
	if document == "" {
		return nil
	}

	obj := &unstructured.Unstructured{}

	_, gvk, decodeErr := decoder.Decode([]byte(document), nil, obj)
	if decodeErr != nil {
		return fmt.Errorf("decode bootstrap manifest doc %d: %w", docIndex, decodeErr)
	}

	if obj.GetKind() == "" {
		return nil
	}

	mapping, mapErr := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if mapErr != nil {
		return fmt.Errorf("REST mapping error for %s: %w", gvk.String(), mapErr)
	}

	namespaceable := dynClient.Resource(mapping.Resource)

	resourceClient := func() dynamic.ResourceInterface {
		if mapping.Scope.Name() != meta.RESTScopeNameNamespace {
			return namespaceable
		}

		namespace := obj.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}

		return namespaceable.Namespace(namespace)
	}()

	return patchBootstrapResource(ctx, resourceClient, obj)
}

func patchBootstrapResource(
	ctx context.Context,
	resourceClient dynamic.ResourceInterface,
	obj *unstructured.Unstructured,
) error {
	data, marshalErr := obj.MarshalJSON()
	if marshalErr != nil {
		return fmt.Errorf("marshal bootstrap object: %w", marshalErr)
	}

	_, patchErr := resourceClient.Patch(
		ctx,
		obj.GetName(),
		types.ApplyPatchType,
		data,
		metav1.PatchOptions{FieldManager: applyFieldManager},
	)
	if patchErr != nil {
		return fmt.Errorf(
			"apply bootstrap resource %s/%s: %w",
			obj.GetKind(),
			obj.GetName(),
			patchErr,
		)
	}

	return nil
}

func newFlannelInstaller(
	kubectlClient kubectl.Interface,
	kubeconfigPath string,
	clusterCfg *v1alpha1.Cluster,
) *flannelinstaller.Installer {
	timeout := getInstallTimeout(clusterCfg)

	return flannelinstaller.NewFlannelInstaller(
		kubectlClient,
		kubeconfigPath,
		clusterCfg.Spec.Connection.Context,
		timeout,
	)
}

func runFlannelInstallation(
	cmd *cobra.Command,
	installer *flannelinstaller.Installer,
	tmr timer.Timer,
) error {
	return runCNIInstallation(cmd, installer, "Flannel", tmr)
}

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
	timeout := getInstallTimeout(clusterCfg)

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

// getInstallTimeout determines the timeout for component installation (Cilium, Calico, metrics-server, etc.).
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
