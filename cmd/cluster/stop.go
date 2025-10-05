package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop a running cluster",
		Long:  `Stop a running Kubernetes cluster.`,
		RunE:  utils.HandleConfigLoadRunE,
	}
}
