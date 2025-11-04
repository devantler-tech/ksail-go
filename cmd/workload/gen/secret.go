package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewSecretCmd creates the gen secret command.
func NewSecretCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.NewClientWithStdio()
	cmd, err := client.CreateSecretCmd()
	cobra.CheckErr(err)

	return cmd
}
