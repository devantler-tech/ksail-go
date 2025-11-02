package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewJobCmd creates the gen job command.
func NewJobCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for job
	generator := kubernetes.NewGenerator("job")

	// Use the generator to create the command
	return generator.Generate()
}
