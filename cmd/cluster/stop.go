package cluster

import (
	"context"

	"github.com/devantler-tech/ksail-go/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/spf13/cobra"
)

// newStopLifecycleConfig creates the lifecycle configuration for cluster stop.
func newStopLifecycleConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "🛑",
		TitleContent:       "Stop cluster...",
		ActivityContent:    "stopping cluster",
		SuccessContent:     "cluster stopped",
		ErrorMessagePrefix: "failed to stop cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Stop(ctx, clusterName)
		},
	}
}

// NewStopCmd creates and returns the stop command.
func NewStopCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "stop",
		Short:        "Stop a running cluster",
		Long:         `Stop a running Kubernetes cluster.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = shared.NewLifecycleCommandWrapper(
		runtimeContainer,
		cfgManager,
		newStopLifecycleConfig(),
	)

	return cmd
}
