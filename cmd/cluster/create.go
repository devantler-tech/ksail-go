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

	cmd.RunE = shared.NewLifecycleCommandWrapper(
		runtimeContainer,
		cfgManager,
		newCreateLifecycleConfig(),
	)

	return cmd
}

// CreateDeps contains the dependencies required to handle the create command.
// Deprecated: Use shared.LifecycleDeps instead.
type CreateDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

var errMissingClusterProvisioner = shared.ErrMissingClusterProvisionerDependency

// HandleCreateRunE executes the cluster creation workflow.
// Deprecated: This function is kept for backward compatibility with tests.
func HandleCreateRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps CreateDeps,
) error {
	lifecycleDeps := shared.LifecycleDeps{
		Timer:   deps.Timer,
		Factory: deps.Factory,
	}

	return shared.HandleLifecycleRunE(cmd, cfgManager, lifecycleDeps, newCreateLifecycleConfig())
}
