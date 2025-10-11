package cluster

import (
	"context"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/spf13/cobra"
)

// newDeleteLifecycleConfig creates the lifecycle configuration for cluster deletion.
func newDeleteLifecycleConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "🗑️",
		TitleContent:       "Delete cluster...",
		ActivityContent:    "deleting cluster",
		SuccessContent:     "cluster deleted",
		ErrorMessagePrefix: "failed to delete cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Delete(ctx, clusterName)
		},
	}
}

// NewDeleteCmd creates and returns the delete command.
func NewDeleteCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Destroy a cluster",
		Long:         `Destroy a cluster.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = shared.NewLifecycleCommandWrapper(
		runtimeContainer,
		cfgManager,
		newDeleteLifecycleConfig(),
	)

	return cmd
}
