package cluster

import (
	"context"

	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/spf13/cobra"
)

// NewCreateCmd wires the cluster create command using the shared runtime container.
func NewCreateCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return NewLifecycleCmd(
		runtimeContainer,
		"create",
		"Create a cluster",
		`Create a Kubernetes cluster as defined by configuration.`,
		HandleCreateRunE,
	)
}

// HandleCreateRunE executes the cluster creation workflow.
func HandleCreateRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps LifecycleDeps,
) error {
	config := LifecycleConfig{
		TitleEmoji:      "🚀",
		TitleContent:    "Create cluster...",
		ActivityContent: "creating cluster",
		SuccessContent:  "cluster created",
		ErrorPrefix:     "failed to create cluster",
	}

	return ExecuteLifecycleCommand(cmd, cfgManager, deps, config, func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
		return provisioner.Create(ctx, clusterName)
	})
}
