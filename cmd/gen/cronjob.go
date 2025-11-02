package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewCronJobCmd creates the gen cronjob command.
func NewCronJobCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewCronJobGenerator()

	return generator.Generate()
}
