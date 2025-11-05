// Package gen provides generation commands for workload manifests.
package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewGenCmd creates and returns the gen command group namespace.
func NewGenCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate workload manifests",
		Long:  "Generate workload manifests for Kubernetes resources and Flux CD resources.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	cmd.AddCommand(NewHelmReleaseCmd(runtimeContainer))

	return cmd
}
