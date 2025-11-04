package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewIngressCmd creates the gen ingress command.
func NewIngressCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.NewClientWithStdio()
	cmd, err := client.CreateIngressCmd()
	cobra.CheckErr(err)

	return cmd
}
