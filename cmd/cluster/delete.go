package cluster

import (
	"context"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

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

	config := shared.LifecycleConfig{
		TitleEmoji:         "🗑️",
		TitleContent:       "Delete cluster...",
		ActivityContent:    "deleting cluster",
		SuccessContent:     "cluster deleted",
		ErrorMessagePrefix: "failed to delete cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Delete(ctx, clusterName)
		},
	}

	cmd.RunE = shared.NewLifecycleCommandWrapper(runtimeContainer, cfgManager, config)

	return cmd
}

// DeleteDeps contains the dependencies required to handle the delete command.
// Deprecated: Use shared.LifecycleDeps instead.
type DeleteDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

var errMissingClusterProvisionerForDelete = shared.ErrMissingClusterProvisionerDependency

// HandleDeleteRunE executes the cluster deletion workflow.
// Deprecated: This function is kept for backward compatibility with tests.
func HandleDeleteRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps DeleteDeps,
) error {
	lifecycleDeps := shared.LifecycleDeps{
		Timer:   deps.Timer,
		Factory: deps.Factory,
	}

	config := shared.LifecycleConfig{
		TitleEmoji:         "🗑️",
		TitleContent:       "Delete cluster...",
		ActivityContent:    "deleting cluster",
		SuccessContent:     "cluster deleted",
		ErrorMessagePrefix: "failed to delete cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Delete(ctx, clusterName)
		},
	}

	return shared.HandleLifecycleRunE(cmd, cfgManager, lifecycleDeps, config)
}
