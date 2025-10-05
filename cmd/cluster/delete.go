package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/spf13/cobra"
)

// NewDeleteCmd creates and returns the delete command.
func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Destroy a cluster",
		Long:  `Destroy a cluster.`,
		RunE:  utils.HandleConfigLoadRunE,
	}
}
