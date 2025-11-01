// Package gen provides the gen command namespace for generating Kubernetes resources.
package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewGenCmd creates and returns the gen command group namespace.
func NewGenCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate Kubernetes resource manifests",
		Long: "Generate Kubernetes resource manifests using kubectl create with --dry-run=client -o yaml. " +
			"The generated YAML is printed to stdout and can be redirected to a file using shell redirection (> file.yaml).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	cmd.AddCommand(NewClusterRoleCmd(runtimeContainer))
	cmd.AddCommand(NewClusterRoleBindingCmd(runtimeContainer))
	cmd.AddCommand(NewConfigMapCmd(runtimeContainer))
	cmd.AddCommand(NewCronJobCmd(runtimeContainer))
	cmd.AddCommand(NewDeploymentCmd(runtimeContainer))
	cmd.AddCommand(NewIngressCmd(runtimeContainer))
	cmd.AddCommand(NewJobCmd(runtimeContainer))
	cmd.AddCommand(NewNamespaceCmd(runtimeContainer))
	cmd.AddCommand(NewPodDisruptionBudgetCmd(runtimeContainer))
	cmd.AddCommand(NewPriorityClassCmd(runtimeContainer))
	cmd.AddCommand(NewQuotaCmd(runtimeContainer))
	cmd.AddCommand(NewRoleCmd(runtimeContainer))
	cmd.AddCommand(NewRoleBindingCmd(runtimeContainer))
	cmd.AddCommand(NewSecretCmd(runtimeContainer))
	cmd.AddCommand(NewServiceCmd(runtimeContainer))
	cmd.AddCommand(NewServiceAccountCmd(runtimeContainer))
	cmd.AddCommand(NewTokenCmd(runtimeContainer))

	return cmd
}
