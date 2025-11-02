package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewCronJobCmd creates the gen cronjob command.
func NewCronJobCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for cronjob
	generator := kubernetes.NewGenerator("cronjob")

	// Use the generator to create the command
	return generator.Generate()
}
