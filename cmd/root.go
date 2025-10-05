package cmd

import (
	"fmt"

	cluster "github.com/devantler-tech/ksail-go/cmd/cluster"
	"github.com/devantler-tech/ksail-go/cmd/workload"
	"github.com/devantler-tech/ksail-go/pkg/errorhandler"
	"github.com/devantler-tech/ksail-go/pkg/ui/asciiart"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command with version info and subcommands.
func NewRootCmd(version, commit, date string) *cobra.Command {
	// Create the command using the helper (no field selectors needed for root command)
	cmd := &cobra.Command{
		Use:   "ksail",
		Short: "SDK for operating and managing K8s clusters and workloads",
		Long: `KSail helps you easily create, manage, and test local Kubernetes clusters and workloads ` +
			`from one simple command line tool.`,
		RunE: handleRootRunE,
	}

	// Set version if available
	cmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)

	// Add all subcommands
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(cluster.NewClusterCmd())
	cmd.AddCommand(workload.NewWorkloadCmd())

	return cmd
}

// Execute runs the provided root command and handles errors.
func Execute(cmd *cobra.Command) error {
	executor := errorhandler.NewExecutor()

	err := executor.Execute(cmd)
	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// --- internals ---

// handleRootRunE handles the root command.
func handleRootRunE(
	cmd *cobra.Command,
	_ []string,
) error {
	asciiart.PrintKSailLogo(cmd.OutOrStdout())

	// The err can safely be ignored, as it can never fail at runtime.
	_ = cmd.Help()

	return nil
}
