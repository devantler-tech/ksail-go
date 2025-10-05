package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start a stopped cluster",
		Long:  `Start a previously stopped cluster.`,
		RunE:  utils.HandleConfigLoadRunE,
	}
}
