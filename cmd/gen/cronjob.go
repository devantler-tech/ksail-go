package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewCronJobCmd creates the gen cronjob command.
func NewCronJobCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return createGenCommand(runtimeContainer, "cronjob")
}
