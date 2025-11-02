package gen

import (
	"github.com/devantler-tech/ksail-go/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubectl"
	"github.com/spf13/cobra"
)

// NewClusterRoleBindingCmd creates the gen clusterrolebinding command.
func NewClusterRoleBindingCmd(_ *runtime.Runtime) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := shared.GetKubeconfigPathSilently()

	// Create a kubectl generator for clusterrolebinding
	generator := kubectl.NewGenerator(kubeconfigPath, "clusterrolebinding")

	// Use the generator to create the command
	return generator.Generate()
}
