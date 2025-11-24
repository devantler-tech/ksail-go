package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	cmdhelpers "github.com/devantler-tech/ksail-go/pkg/cmd"
	registry "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/registry"
	"github.com/docker/docker/client"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const localRegistryResourceName = "local-registry"

var (
	localRegistryProvisionStageInfo = registryStageInfo{
		title:         "Provision local registry...",
		emoji:         "🗄️",
		success:       "local registry provisioned",
		failurePrefix: "failed to provision local registry",
	}
	localRegistryConnectStageInfo = registryStageInfo{
		title:         "Attach local registry...",
		emoji:         "🔌",
		success:       "local registry attached",
		failurePrefix: "failed to attach local registry",
	}
	localRegistryCleanupStageInfo = registryStageInfo{
		title:         "Cleanup local registry...",
		emoji:         "🧹",
		success:       "local registry cleaned up",
		failurePrefix: "failed to cleanup local registry",
	}
	registryServiceFactory = func(cfg registry.Config) (registry.Service, error) {
		return registry.NewService(cfg)
	}
)

type localRegistryContext struct {
	clusterName string
	networkName string
}

func runLocalRegistryAction(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	kindConfig *kindv1alpha4.Cluster,
	k3dConfig *k3dv1alpha5.SimpleConfig,
	info registryStageInfo,
	action func(context.Context, registry.Service, localRegistryContext) error,
) error {
	if clusterCfg.Spec.LocalRegistry != v1alpha1.LocalRegistryEnabled {
		return nil
	}

	ctx := newLocalRegistryContext(clusterCfg, kindConfig, k3dConfig)

	return runLocalRegistryStage(
		cmd,
		deps,
		info,
		func(execCtx context.Context, svc registry.Service) error {
			return action(execCtx, svc, ctx)
		},
	)
}

func ensureLocalRegistryProvisioned(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	kindConfig *kindv1alpha4.Cluster,
	k3dConfig *k3dv1alpha5.SimpleConfig,
) error {
	return runLocalRegistryAction(
		cmd,
		clusterCfg,
		deps,
		kindConfig,
		k3dConfig,
		localRegistryProvisionStageInfo,
		func(execCtx context.Context, svc registry.Service, ctx localRegistryContext) error {
			createOpts := newLocalRegistryCreateOptions(clusterCfg, ctx)
			if _, err := svc.Create(execCtx, createOpts); err != nil {
				return err
			}

			_, err := svc.Start(execCtx, registry.StartOptions{Name: createOpts.Name})

			return err
		},
	)
}

func connectLocalRegistryToClusterNetwork(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	kindConfig *kindv1alpha4.Cluster,
	k3dConfig *k3dv1alpha5.SimpleConfig,
) error {
	return runLocalRegistryAction(
		cmd,
		clusterCfg,
		deps,
		kindConfig,
		k3dConfig,
		localRegistryConnectStageInfo,
		func(execCtx context.Context, svc registry.Service, ctx localRegistryContext) error {
			startOpts := registry.StartOptions{
				Name:        buildLocalRegistryName(),
				NetworkName: ctx.networkName,
			}

			_, err := svc.Start(execCtx, startOpts)

			return err
		},
	)
}

func cleanupLocalRegistry(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps cmdhelpers.LifecycleDeps,
	deleteVolumes bool,
) error {
	if clusterCfg.Spec.LocalRegistry != v1alpha1.LocalRegistryEnabled {
		return nil
	}

	kindConfig, k3dConfig, err := loadDistributionConfigs(clusterCfg, deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load distribution config: %w", err)
	}

	return runLocalRegistryAction(
		cmd,
		clusterCfg,
		deps,
		kindConfig,
		k3dConfig,
		localRegistryCleanupStageInfo,
		func(execCtx context.Context, svc registry.Service, ctx localRegistryContext) error {
			stopOpts := registry.StopOptions{
				Name:         buildLocalRegistryName(),
				ClusterName:  ctx.clusterName,
				NetworkName:  ctx.networkName,
				DeleteVolume: deleteVolumes,
			}

			return svc.Stop(execCtx, stopOpts)
		},
	)
}

func newLocalRegistryContext(
	clusterCfg *v1alpha1.Cluster,
	kindConfig *kindv1alpha4.Cluster,
	k3dConfig *k3dv1alpha5.SimpleConfig,
) localRegistryContext {
	clusterName := resolveLocalRegistryClusterName(clusterCfg, kindConfig, k3dConfig)
	networkName := resolveLocalRegistryNetworkName(clusterCfg, clusterName)

	return localRegistryContext{clusterName: clusterName, networkName: networkName}
}

func resolveLocalRegistryClusterName(
	clusterCfg *v1alpha1.Cluster,
	kindConfig *kindv1alpha4.Cluster,
	k3dConfig *k3dv1alpha5.SimpleConfig,
) string {
	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		if kindConfig != nil {
			if name := strings.TrimSpace(kindConfig.Name); name != "" {
				return name
			}
		}
	case v1alpha1.DistributionK3d:
		return resolveK3dClusterName(clusterCfg, k3dConfig)
	}

	if name := strings.TrimSpace(clusterCfg.Spec.Connection.Context); name != "" {
		return name
	}

	return "ksail"
}

func resolveLocalRegistryNetworkName(
	clusterCfg *v1alpha1.Cluster,
	clusterName string,
) string {
	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return "kind"
	case v1alpha1.DistributionK3d:
		trimmed := strings.TrimSpace(clusterName)
		if trimmed == "" {
			trimmed = "k3d"
		}

		return "k3d-" + trimmed
	default:
		return ""
	}
}

func newLocalRegistryCreateOptions(
	clusterCfg *v1alpha1.Cluster,
	ctx localRegistryContext,
) registry.CreateOptions {
	return registry.CreateOptions{
		Name:        buildLocalRegistryName(),
		Host:        registry.DefaultEndpointHost,
		Port:        resolveLocalRegistryPort(clusterCfg),
		ClusterName: ctx.clusterName,
		VolumeName:  buildLocalRegistryName(),
	}
}

func buildLocalRegistryName() string {
	return localRegistryResourceName
}

func resolveLocalRegistryPort(clusterCfg *v1alpha1.Cluster) int {
	if clusterCfg.Spec.Options.LocalRegistry.HostPort > 0 {
		return int(clusterCfg.Spec.Options.LocalRegistry.HostPort)
	}

	return dockerclient.DefaultRegistryPort
}

func runLocalRegistryStage(
	cmd *cobra.Command,
	deps cmdhelpers.LifecycleDeps,
	info registryStageInfo,
	handler func(context.Context, registry.Service) error,
) error {
	return runRegistryStage(
		cmd,
		deps,
		info,
		func(ctx context.Context, dockerClient client.APIClient) error {
			service, err := registryServiceFactory(registry.Config{DockerClient: dockerClient})
			if err != nil {
				return fmt.Errorf("create registry service: %w", err)
			}

			if ctx == nil {
				ctx = context.Background()
			}

			return handler(ctx, service)
		},
	)
}
