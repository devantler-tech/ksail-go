package cluster

import (
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to list clusters for",
			DefaultValue: v1alpha1.DistributionKind,
		},
	}

	// Create the command using the helper
	cmd := helpers.NewCobraCommand(
		"list",
		"List clusters",
		`List all Kubernetes clusters managed by KSail.`,
		HandleListRunE,
		fieldSelectors...,
	)

	// Add the special --all flag manually since it's CLI-only
	cmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	return cmd
}

// HandleListRunE handles the list command.
// Exported for testing purposes.
func HandleListRunE(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	_ []string,
) error {
	return handleListRunEWithProvisioner(cmd, configManager, nil)
}

// handleListRunEWithProvisioner is the internal implementation that accepts an optional provisioner for testing.
func handleListRunEWithProvisioner(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	provisioner provisionerFactory,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Bind the --all flag manually since it's added after command creation
	_ = configManager.Viper.BindPFlag("all", cmd.Flags().Lookup("all"))

	// Load cluster configuration
	cluster, err := configManager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	all := configManager.Viper.GetBool("all")

	// List clusters using the provisioner
	clusters, err := listClustersUsingProvisioner(cmd, cluster, provisioner)
	if err != nil {
		return err
	}

	// Display results
	displayListResults(cmd, cluster, clusters, all, tmr)

	return nil
}

// listClustersUsingProvisioner creates a provisioner and lists clusters.
func listClustersUsingProvisioner(
	cmd *cobra.Command,
	cluster *v1alpha1.Cluster,
	provisioner provisionerFactory,
) ([]string, error) {
	// Create provisioner based on distribution
	var clusterProvisioner clusterprovisioner.ClusterProvisioner

	var err error

	if provisioner != nil {
		clusterProvisioner, _, err = provisioner(
			cmd.Context(),
			cluster.Spec.Distribution,
			cluster.Spec.DistributionConfig,
			cluster.Spec.Connection.Kubeconfig,
		)
	} else {
		clusterProvisioner, _, err = clusterprovisioner.CreateClusterProvisioner(
			cmd.Context(),
			cluster.Spec.Distribution,
			cluster.Spec.DistributionConfig,
			cluster.Spec.Connection.Kubeconfig,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provisioner: %w", err)
	}

	// List clusters using the provisioner
	clusters, err := clusterProvisioner.List(cmd.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	return clusters, nil
}

// displayListResults formats and displays the list results.
func displayListResults(
	cmd *cobra.Command,
	cluster *v1alpha1.Cluster,
	clusters []string,
	all bool,
	tmr timer.Timer,
) {
	// Display the appropriate success message
	if all {
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "Listing all clusters",
			Timer:   tmr,
			Writer:  cmd.OutOrStdout(),
		})
	} else {
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "Listing running clusters",
			Timer:   tmr,
			Writer:  cmd.OutOrStdout(),
		})
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "Distribution filter: %s",
		Args:    []any{string(cluster.Spec.Distribution)},
		Writer:  cmd.OutOrStdout(),
	})

	// Display cluster names
	if len(clusters) == 0 {
		notify.WriteMessage(notify.Message{
			Type:    notify.ActivityType,
			Content: "No clusters found",
			Writer:  cmd.OutOrStdout(),
		})
	} else {
		for _, name := range clusters {
			notify.WriteMessage(notify.Message{
				Type:    notify.ActivityType,
				Content: "- %s",
				Args:    []any{name},
				Writer:  cmd.OutOrStdout(),
			})
		}
	}
}
