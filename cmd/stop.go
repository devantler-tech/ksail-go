/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/inputs"
	"github.com/devantler-tech/ksail-go/internal/managers"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop an existing Kubernetes cluster",
	Long:  "Stop an existing Kubernetes cluster specified by --name or by the loaded kind config.",
	RunE: func(_ *cobra.Command, _ []string) error {
		return handleStop()
	},
}

// -- internals ---

// handleStop handles the stop command.
func handleStop() error {
	if err := InitServices(); err != nil {
		return err
	}

	return stop()
}

func stop() error {
	params := managers.OperationParams{
		ActionMsg: "⏹️ Stopping",
		VerbMsg:   "stopping",
		PastMsg:   "stopped",
	}
	return clusterManager.ExecuteOperation(params, func(provisioner clusterprovisioner.ClusterProvisioner, name string) error {
		return provisioner.Stop(name)
	})
}

func init() {
	rootCmd.AddCommand(stopCmd)
	inputs.AddNameFlag(stopCmd)
	inputs.AddDistributionFlag(stopCmd)
}
