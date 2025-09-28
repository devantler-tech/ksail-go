package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

const installMessage = "Workload install coming soon."

// NewInstallCommand creates the workload install command.
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a Helm chart",
		Long:  "Install a Helm chart to your cluster.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			notify.Infoln(cmd.OutOrStdout(), installMessage)

			return nil
		},
	}

	applyCommonCommandConfig(cmd)

	return cmd
}
