package cluster

import (
	"context"
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/cluster"
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
	cmd := cmdhelpers.NewCobraCommand(
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
	// Bind the --all flag manually since it's added after command creation
	_ = configManager.Viper.BindPFlag("all", cmd.Flags().Lookup("all"))

	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadConfig()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// List clusters based on distribution
	clusters, err := listClustersForDistribution(cmd.Context(), cluster.Spec.Distribution)
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to list clusters: "+err.Error())

		return fmt.Errorf("failed to list clusters: %w", err)
	}

	// Display results
	all := configManager.Viper.GetBool("all")
	if all {
		notify.Successln(cmd.OutOrStdout(), "Listing all clusters")
	} else {
		notify.Successln(cmd.OutOrStdout(), "Listing running clusters")
	}

	notify.Activityln(cmd.OutOrStdout(),
		"Distribution filter: "+string(cluster.Spec.Distribution))

	// Print cluster names
	if len(clusters) == 0 {
		notify.Activityln(cmd.OutOrStdout(), "No clusters found")
	} else {
		for _, clusterName := range clusters {
			notify.Activityln(cmd.OutOrStdout(), "  - "+clusterName)
		}
	}

	return nil
}

// listClustersForDistribution lists clusters for a given distribution.
func listClustersForDistribution(
	ctx context.Context,
	distribution v1alpha1.Distribution,
) ([]string, error) {
	switch distribution {
	case v1alpha1.DistributionKind:
		return listKindClusters()
	case v1alpha1.DistributionK3d:
		return listK3dClusters(ctx)
	case v1alpha1.DistributionEKS:
		return listEKSClusters(ctx)
	default:
		return nil, fmt.Errorf("unsupported distribution: %s", distribution)
	}
}

// listKindClusters lists all kind clusters.
func listKindClusters() ([]string, error) {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())
	clusters, err := provider.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	return clusters, nil
}

// listK3dClusters lists all k3d clusters.
func listK3dClusters(ctx context.Context) ([]string, error) {
	runtime := runtimes.SelectedRuntime
	clusters, err := k3dclient.ClusterList(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to list k3d clusters: %w", err)
	}

	clusterNames := make([]string, 0, len(clusters))
	for _, c := range clusters {
		clusterNames = append(clusterNames, c.Name)
	}

	return clusterNames, nil
}

// listEKSClusters lists all EKS clusters.
func listEKSClusters(ctx context.Context) ([]string, error) {
	// For EKS, we need minimal config to create provider
	// We'll just return an informative message since EKS requires AWS credentials
	return []string{}, fmt.Errorf("EKS cluster listing requires AWS credentials configuration")
}
