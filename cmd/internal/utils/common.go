// Package utils provides common utilities for KSail commands.
package utils

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// CommandConfig holds the configuration for creating a command.
type CommandConfig struct {
	Use        string
	Short      string
	Long       string
	RunEFunc   func(cmd *cobra.Command, configManager configmanager.ConfigManager[v1alpha1.Cluster], args []string) error
	FieldsFunc func(c *v1alpha1.Cluster) []any
}

// NewSimpleClusterCommand creates a new command with common cluster management pattern.
func NewSimpleClusterCommand(cfg CommandConfig) *cobra.Command {
	// Create field selectors if FieldsFunc is provided
	var fieldSelectors []ksail.FieldSelector[v1alpha1.Cluster]
	if cfg.FieldsFunc != nil {
		dummyCluster := &v1alpha1.Cluster{}
		fieldPointers := cfg.FieldsFunc(dummyCluster)

		// Parse the flat array: field, defaultValue, description, field, defaultValue, description, ...
		for i := 0; i < len(fieldPointers); i += 3 {
			if i+2 >= len(fieldPointers) {
				break // Not enough elements for a complete triplet
			}

			fieldPtr := fieldPointers[i]
			defaultValue := fieldPointers[i+1]
			description := fieldPointers[i+2].(string)

			// Create a field selector for this field
			fieldSelectors = append(fieldSelectors, ksail.FieldSelector[v1alpha1.Cluster]{
				Selector: func(ptr any) func(c *v1alpha1.Cluster) any {
					return func(c *v1alpha1.Cluster) any {
						// Need to re-evaluate the field pointer using the actual cluster
						return cfg.FieldsFunc(c)[i] // Return the same position in the array
					}
				}(fieldPtr),
				Description:  description,
				DefaultValue: defaultValue,
			})
		}
	}

	// Create configuration manager with field selectors
	configManager := ksail.NewManager(fieldSelectors...)

	// Create the command
	cmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cfg.RunEFunc(cmd, configManager, args)
		},
	}

	// Add flags for the field selectors
	configManager.AddFlagsFromFields(cmd)

	return cmd
}

// ClusterInfoField represents a field to log from cluster information.
type ClusterInfoField struct {
	Label string
	Value string
}

// LogClusterInfo logs cluster information fields to the command output.
func LogClusterInfo(cmd *cobra.Command, fields []ClusterInfoField) {
	for _, field := range fields {
		notify.Activityln(cmd.OutOrStdout(), field.Label+": "+field.Value)
	}
}

// LoadClusterWithErrorHandling provides common error handling pattern for loading cluster configuration.
// Exported for testing purposes.
func LoadClusterWithErrorHandling(
	cmd *cobra.Command,
	configManager configmanager.ConfigManager[v1alpha1.Cluster],
) (*v1alpha1.Cluster, error) {
	cluster, err := configManager.LoadConfig()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return nil, fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	return cluster, nil
}

// HandleSimpleClusterCommand provides common error handling and cluster loading for simple commands.
func HandleSimpleClusterCommand(
	cmd *cobra.Command,
	configManager configmanager.ConfigManager[v1alpha1.Cluster],
	successMessage string,
) (*v1alpha1.Cluster, error) {
	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadConfig()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return nil, fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(), successMessage)
	LogClusterInfo(cmd, []ClusterInfoField{
		{"Distribution", string(cluster.Spec.Distribution)},
		{"Context", cluster.Spec.Connection.Context},
	})

	return cluster, nil
}
