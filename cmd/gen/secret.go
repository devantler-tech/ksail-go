package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewSecretCmd creates the gen secret command.
func NewSecretCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for secret
	generator := kubernetes.NewGenerator("secret")

	// Use the generator to create the command
	return generator.Generate()
}
