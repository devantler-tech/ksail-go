package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewQuotaCmd creates the gen quota command.
func NewQuotaCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for quota
	generator := kubernetes.NewGenerator("quota")

	// Use the generator to create the command
	return generator.Generate()
}
