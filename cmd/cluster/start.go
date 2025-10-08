package cluster

import (
	"context"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/spf13/cobra"
)

// newStartLifecycleConfig creates the lifecycle configuration for cluster start.
func newStartLifecycleConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "▶️",
		TitleContent:       "Start cluster...",
		ActivityContent:    "starting cluster",
		SuccessContent:     "cluster started",
		ErrorMessagePrefix: "failed to start cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Start(ctx, clusterName)
		},
	}
}

// NewStartCmd creates and returns the start command.
func NewStartCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start a stopped cluster",
		Long:         `Start a previously stopped cluster.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = shared.NewLifecycleCommandWrapper(
		runtimeContainer,
		cfgManager,
		newStartLifecycleConfig(),
	)

	return cmd
}
