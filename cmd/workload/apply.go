package workload

import (
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

const applyMessage = "Workload apply coming soon."

// NewApplyCommand creates the workload apply command.
func NewApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply manifests",
		Long:  "Apply manifests to the Kubernetes cluster using Kubectl.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			notify.InfoMessage(cmd.OutOrStdout(), notify.NewMessage(applyMessage))

			return nil
		},
	}

	applyCommonCommandConfig(cmd)

	return cmd
}
