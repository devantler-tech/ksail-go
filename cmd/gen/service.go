package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewServiceCmd creates the gen service command.
func NewServiceCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for service
	generator := kubernetes.NewGenerator("service")

	// Use the generator to create the command
	return generator.Generate()
}
