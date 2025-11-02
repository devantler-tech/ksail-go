package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewClusterRoleCmd creates the gen clusterrole command.
func NewClusterRoleCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for clusterrole
	generator := kubernetes.NewGenerator("clusterrole")

	// Use the generator to create the command
	return generator.Generate()
}
