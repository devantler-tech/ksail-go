package cluster

import (
	"context"
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	cmd := cmdhelpers.NewCobraCommand(
		"list",
		"List clusters",
		`List all Kubernetes clusters managed by KSail.`,
		HandleListRunE,
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to list clusters for",
			DefaultValue: v1alpha1.DistributionKind,
		},
		cmdhelpers.StandardDistributionConfigFieldSelector(),
	)

	cmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	return cmd
}

// HandleListRunE handles the list command.
// Exported for testing purposes.
func HandleListRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	ctx := context.Background()

	_ = manager.Viper.BindPFlag("all", cmd.Flags().Lookup("all"))

	config, err := manager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	engine, err := containerengine.GetAutoDetectedClient()
	if err != nil {
		return fmt.Errorf("failed to get container engine client: %w", err)
	}

	provisioner, err := newProvisioner(config, *engine)
	if err != nil {
		return fmt.Errorf("failed to create provisioner: %w", err)
	}

	fmt.Fprintln(manager.Writer)
	notify.TitleMessage(manager.Writer, "📋", notify.NewMessage("Listing clusters..."))

	clusters, err := provisioner.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	if len(clusters) == 0 {
		notify.ActivityMessage(manager.Writer, notify.NewMessage("no clusters found"))
		return nil
	}

	notify.SuccessMessage(
		manager.Writer,
		notify.NewMessage(fmt.Sprintf("found %d cluster(s):", len(clusters))),
	)

	for _, cluster := range clusters {
		fmt.Fprintf(manager.Writer, "  • %s\n", cluster)
	}

	return nil
}
