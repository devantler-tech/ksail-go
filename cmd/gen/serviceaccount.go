package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewServiceAccountCmd creates the gen serviceaccount command.
func NewServiceAccountCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for serviceaccount
	generator := kubernetes.NewGenerator("serviceaccount")

	// Use the generator to create the command
	return generator.Generate()
}
