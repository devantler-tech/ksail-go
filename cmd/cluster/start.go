package cluster

import (
	"context"

	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return NewLifecycleCmd(
		runtimeContainer,
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		HandleStartRunE,
	)
}

// HandleStartRunE executes the cluster start workflow.
func HandleStartRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps LifecycleDeps,
) error {
	config := LifecycleConfig{
		TitleEmoji:      "▶️",
		TitleContent:    "Start cluster...",
		ActivityContent: "starting cluster",
		SuccessContent:  "cluster started",
		ErrorPrefix:     "failed to start cluster",
	}

	return ExecuteLifecycleCommand(
		cmd,
		cfgManager,
		deps,
		config,
		func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Start(ctx, clusterName)
		},
	)
}
