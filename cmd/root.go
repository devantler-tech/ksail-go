package cmd

import (
	"fmt"

	cluster "github.com/devantler-tech/ksail-go/cmd/cluster"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/cmd/ui/asciiart"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command with version info and subcommands.
func NewRootCmd(version, commit, date string) *cobra.Command {
	// Create the command using the helper (no field selectors needed for root command)
	cmd := cmdhelpers.NewCobraCommand(
		"ksail",
		"SDK for operating and managing K8s clusters and workloads",
		`KSail helps you easily create, manage, and test local Kubernetes clusters and workloads `+
			`from one simple command line tool.`,
		handleRootRunE,
	)

	// Silence errors and usage
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	// Set version if available
	cmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)

	// Add all subcommands
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(cluster.NewClusterCmd())
	cmd.AddCommand(NewReconcileCmd())

	return cmd
}

// Execute runs the provided root command and handles errors.
func Execute(cmd *cobra.Command) error {
	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// --- internals ---

// handleRootRunE handles the root command.
func handleRootRunE(
	cmd *cobra.Command,
	_ *configmanager.ConfigManager,
	_ []string,
) error {
	asciiart.PrintKSailLogo(cmd.OutOrStdout())

	// The err can safely be ignored, as it can never fail at runtime.
	_ = cmd.Help()

	return nil
}
