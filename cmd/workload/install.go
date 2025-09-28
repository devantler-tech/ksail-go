package workload

import (
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

const installMessage = "Workload install coming soon."

// NewInstallCommand creates the workload install command.
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Helm charts",
		Long:  "Install Helm charts to provision workloads through KSail.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			notify.Infoln(cmd.OutOrStdout(), installMessage)

			return nil
		},
	}

	applyCommonCommandConfig(cmd)

	return cmd
}
