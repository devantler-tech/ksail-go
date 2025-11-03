package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewServiceCmd creates the gen service command.
func NewServiceCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.DefaultClient()
	cmd, err := client.CreateServiceCmd()
	cobra.CheckErr(err)

	return cmd
}
