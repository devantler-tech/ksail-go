package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewRoleCmd creates the gen role command.
func NewRoleCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for role
	generator := kubernetes.NewGenerator("role")

	// Use the generator to create the command
	return generator.Generate()
}
