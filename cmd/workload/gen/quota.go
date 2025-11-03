package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewQuotaCmd creates the gen quota command.
func NewQuotaCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.DefaultClient()
	cmd, err := client.CreateQuotaCmd()
	cobra.CheckErr(err)

	return cmd
}
