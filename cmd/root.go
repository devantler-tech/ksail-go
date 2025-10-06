package cmd

import (
	"fmt"

	cluster "github.com/devantler-tech/ksail-go/cmd/cluster"
	"github.com/devantler-tech/ksail-go/cmd/workload"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/errorhandler"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/asciiart"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command with version info and subcommands.
func NewRootCmd(version, commit, date string) *cobra.Command {
	rt := runtime.New(
		func(i do.Injector) error {
			do.Provide(i, func(do.Injector) (timer.Timer, error) {
				return timer.New(), nil
			})

			return nil
		},
		func(i do.Injector) error {
			do.Provide(i, func(do.Injector) (clusterprovisioner.Factory, error) {
				return clusterprovisioner.DefaultFactory{}, nil
			})

			return nil
		},
	)

	// Create the command using the helper (no field selectors needed for root command)
	cmd := &cobra.Command{
		Use:   "ksail",
		Short: "SDK for operating and managing K8s clusters and workloads",
		Long: `KSail helps you easily create, manage, and test local Kubernetes clusters and workloads ` +
			`from one simple command line tool.`,
		RunE:         handleRootRunE,
		SilenceUsage: true,
	}

	// Set version if available
	cmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)

	// Add all subcommands
	cmd.AddCommand(NewInitCmd(rt))
	cmd.AddCommand(cluster.NewClusterCmd(rt))
	cmd.AddCommand(workload.NewWorkloadCmd(rt))

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
