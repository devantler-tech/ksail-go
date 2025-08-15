package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/inputs"
	"github.com/devantler-tech/ksail-go/internal/managers"
	"github.com/spf13/cobra"
)

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an existing Kubernetes cluster",
	Long:  "Start an existing Kubernetes cluster specified by --name or by the loaded kind config.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleStart()
	},
}

// --- internals ---

// handleStart handles the start command.
func handleStart() error {
	if err := InitServices(); err != nil {
		return err
	}

	return start()
}

func start() error {
	inputs.SetInputsOrFallback(&ksailConfig)

	clusterManager := managers.NewClusterManager(&ksailConfig)
	return clusterManager.StartOrStopCluster(managers.Start)
}

func init() {
	rootCmd.AddCommand(startCmd)
	inputs.AddNameFlag(startCmd)
	inputs.AddDistributionFlag(startCmd)
}
