package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func NewCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "create",
		Short:        "Create a cluster",
		Long:         `Create a Kubernetes cluster as defined by configuration.`,
		RunE:         HandleCreateRunE,
		SilenceUsage: true,
	}
}

// HandleCreateRunE handles the create command.
// Exported for testing purposes.
func HandleCreateRunE(
	cmd *cobra.Command,
	_ []string,
) error {
	// Create command utils
	utils, err := utils.NewCommandUtils(
		cmd,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardDistributionConfigFieldSelector(),
	)
	if err != nil {
		return fmt.Errorf("failed to create command utils: %w", err)
	}
	// Start timing
	utils.Timer.Start()

	// Load the configuration
	err = utils.ConfigManager.LoadConfig(utils.Timer)
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

	if deps == nil {
		return fmt.Errorf("missing resolved dependencies")
	}

	clusterProvisioner := deps.Provisioner
	if clusterProvisioner == nil {
		return fmt.Errorf("missing cluster provisioner dependency")
	}

	clusterName := getClusterNameFromDistributionConfig(deps.DistributionConfig)
	if clusterName == "" {
		return fmt.Errorf("missing cluster name in resolved dependencies")
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

func getClusterNameFromDistributionConfig(any any) string {
	switch cfg := any.(type) {
	case *v1alpha4.Cluster:
		return cfg.Name
	case *v1alpha5.SimpleConfig:
		return cfg.Name
	default:
		return ""
	}
}
