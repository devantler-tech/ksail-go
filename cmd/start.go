package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/inputs"
	"github.com/devantler-tech/ksail-go/internal/managers"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
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
	params := managers.OperationParams{
		ActionMsg: "▶️ Starting",
		VerbMsg:   "starting",
		PastMsg:   "started",
	}
	return clusterManager.ExecuteOperation(params, func(provisioner clusterprovisioner.ClusterProvisioner, name string) error {
		return provisioner.Start(name)
	})
}

func init() {
	rootCmd.AddCommand(startCmd)
	inputs.AddNameFlag(startCmd)
	inputs.AddDistributionFlag(startCmd)
}
