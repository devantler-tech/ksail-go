package gen

import (
	"github.com/devantler-tech/ksail-go/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubectl"
	"github.com/spf13/cobra"
)

// NewPriorityClassCmd creates the gen priorityclass command.
func NewPriorityClassCmd(_ *runtime.Runtime) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := shared.GetKubeconfigPathSilently()

	// Create a kubectl generator for priorityclass
	generator := kubectl.NewGenerator(kubeconfigPath, "priorityclass")

	// Use the generator to create the command
	return generator.Generate()
}
