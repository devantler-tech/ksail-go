package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

const applyMessage = "Workload apply coming soon."

// NewApplyCommand creates the workload apply command.
func NewApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply manifests",
		Long:  "Apply local Kubernetes manifests to your cluster.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			notify.Infoln(cmd.OutOrStdout(), applyMessage)

			return nil
		},
	}

	applyCommonCommandConfig(cmd)

	return cmd
}
