package cluster

import (
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/spf13/cobra"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "status",
		Short:        "Get the status of a cluster",
		Long:         `Get the current status of a Kubernetes cluster.`,
		RunE:         utils.HandleConfigLoadRunE,
		SilenceUsage: true,
	}
}
