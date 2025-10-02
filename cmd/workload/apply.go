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
		Long:  "Apply local Kubernetes manifests to your cluster.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			notify.WriteMessage(notify.Message{
				Type:    notify.InfoType,
				Content: applyMessage,
				Writer:  cmd.OutOrStdout(),
			})

			return nil
		},
	}

	applyCommonCommandConfig(cmd)

	return cmd
}
