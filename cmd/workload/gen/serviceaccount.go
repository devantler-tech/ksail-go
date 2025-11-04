package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewServiceAccountCmd creates the gen serviceaccount command.
func NewServiceAccountCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.NewClientWithStdio()
	cmd, err := client.CreateServiceAccountCmd()
	cobra.CheckErr(err)

	return cmd
}
