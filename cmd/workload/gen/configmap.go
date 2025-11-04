package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewConfigMapCmd creates the gen configmap command.
func NewConfigMapCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.NewClientWithStdio()
	cmd, err := client.CreateConfigMapCmd()
	cobra.CheckErr(err)

	return cmd
}
