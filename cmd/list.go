// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	var viperInstance = config.InitializeViper()

	return factory.NewCobraCommandWithFlags(
		"list",
		"List Kubernetes clusters",
		`List all Kubernetes clusters managed by KSail.`,
		func(cmd *cobra.Command, _ []string) error {
			return handleListRunE(cmd, viperInstance)
		},
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
			_ = viperInstance.BindPFlag("all", cmd.Flags().Lookup("all"))
		},
	)
}

// handleListRunE handles the list command.
func handleListRunE(cmd *cobra.Command, viperInstance *viper.Viper) error {
	all := viperInstance.GetBool("all")
	if all {
		notify.Successln(cmd.OutOrStdout(), "Listing all clusters (stub implementation)")
	} else {
		notify.Successln(cmd.OutOrStdout(), "Listing running clusters (stub implementation)")
	}

	return nil
}
