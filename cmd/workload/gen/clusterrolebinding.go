package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewClusterRoleBindingCmd creates the gen clusterrolebinding command.
func NewClusterRoleBindingCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.DefaultClient()
	cmd, err := client.CreateClusterRoleBindingCmd()
	cobra.CheckErr(err)

	return cmd
}
