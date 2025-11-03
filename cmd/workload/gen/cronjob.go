package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewCronJobCmd creates the gen cronjob command.
func NewCronJobCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.DefaultClient()
	cmd, err := client.CreateCronJobCmd()
	cobra.CheckErr(err)

	return cmd
}
