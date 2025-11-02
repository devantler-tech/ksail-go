package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewConfigMapCmd creates the gen configmap command.
func NewConfigMapCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for configmap
	generator := kubernetes.NewGenerator("configmap")

	// Use the generator to create the command
	return generator.Generate()
}
