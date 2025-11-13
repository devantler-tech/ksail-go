package cluster

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	cmdhelpers "github.com/devantler-tech/ksail-go/pkg/cmd"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registries"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/docker/docker/client"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/yaml"
)

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
