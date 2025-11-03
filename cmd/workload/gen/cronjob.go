package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewCronJobCmd creates the gen cronjob command.
func NewCronJobCmd(_ *runtime.Runtime) *cobra.Command {
	return kubectl.MustNewCommand((*kubectl.Client).NewCronJobCmd)
}
