package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the workload install command.
func NewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install Helm charts",
		Long:  "Install Helm charts to provision workloads through KSail.",
		RunE:  utils.HandleConfigLoadRunE,
	}
}
