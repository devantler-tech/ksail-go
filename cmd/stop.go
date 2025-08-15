/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/inputs"
	"github.com/devantler-tech/ksail-go/internal/managers"
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
	inputs.SetInputsOrFallback(&ksailConfig)

	clusterManager := managers.NewClusterManager(&ksailConfig)
	return clusterManager.StartOrStopCluster(managers.Stop)
}

func init() {
	rootCmd.AddCommand(stopCmd)
	inputs.AddNameFlag(stopCmd)
	inputs.AddDistributionFlag(stopCmd)
}
