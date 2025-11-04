package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewJobCmd creates the gen job command.
func NewJobCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.NewClientWithStdio()
	cmd, err := client.CreateJobCmd()
	cobra.CheckErr(err)

	return cmd
}
