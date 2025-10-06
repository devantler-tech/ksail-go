package cluster

import (
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd(rt *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Get the status of a cluster",
		Long:         `Get the current status of a Kubernetes cluster.`,
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(rt)

	return cmd
}
