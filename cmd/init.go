// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	var viperInstance = config.InitializeViper()

	return factory.NewCobraCommandWithFlags(
		"init",
		"Initialize a new KSail project",
		`Initialize a new KSail project with the specified configuration options.`,
		func(cmd *cobra.Command, _ []string) error {
			return handleInitRunE(cmd, viperInstance)
		},
		func(cmd *cobra.Command) {
			cmd.Flags().String("distribution", "Kind", "Kubernetes distribution to use (Kind, K3d, EKS)")
			_ = viperInstance.BindPFlag("distribution", cmd.Flags().Lookup("distribution"))
		},
	)
}

// handleInitRunE handles the init command.
func handleInitRunE(cmd *cobra.Command, viperInstance *viper.Viper) error {
	distribution := viperInstance.GetString("distribution")
	notify.Successln(cmd.OutOrStdout(), 
		"Project initialized successfully with "+distribution+" distribution (stub implementation)")

	return nil
}
