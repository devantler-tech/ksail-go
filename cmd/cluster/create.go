package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a cluster",
		Long:         `Create a Kubernetes cluster as defined by configuration.`,
		SilenceUsage: true,
	}

	utils, _ := utils.NewCommandUtils(
		cmd,
		ksailconfigmanager.DefaultDistributionFieldSelector(),
		ksailconfigmanager.DefaultDistributionConfigFieldSelector(),
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return HandleCreateRunE(cmd, utils, args)
	}

	return cmd
}

// HandleCreateRunE handles the create command.
// Exported for testing purposes.
func HandleCreateRunE(
	cmd *cobra.Command,
	utils *utils.CommandUtils,
	_ []string,
) error {
	// Start timing
	utils.Timer.Start()

	// Load the configuration
	err := utils.ConfigManager.LoadConfig(utils.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Create the cluster
	err = createCluster(utils, cmd)
	if err != nil {
		return fmt.Errorf("failed to handle create cluster: %w", err)
	}

	return nil
}

func createCluster(utils *utils.CommandUtils, cmd *cobra.Command) error {
	utils.Timer.NewStage()
	deps, err := utils.Resolver.Resolve()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	showProvisioningTitle(cmd)

	clusterProvisioner := deps.Provisioner
	if clusterProvisioner == nil {
		return fmt.Errorf("missing cluster provisioner dependency")
	}

	clusterName, err := configmanager.GetClusterName(deps.DistributionConfig)
	if err != nil {
		return fmt.Errorf("failed to get cluster name from config: %w", err)
	}

	// Show activity message
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "creating cluster",
		Writer:  cmd.OutOrStdout(),
	})

	// Provision the cluster
	err = clusterProvisioner.Create(cmd.Context(), clusterName)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "cluster created",
		Timer:      utils.Timer,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})
	return nil
}

// showProvisioningTitle displays the provisioning stage title.
func showProvisioningTitle(cmd *cobra.Command) {
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Create cluster...",
		Emoji:   "🚀",
		Writer:  cmd.OutOrStdout(),
	})
}
