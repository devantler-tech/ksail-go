package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewSecretCmd creates the gen secret command.
func NewSecretCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewSecretGenerator()

	return generator.Generate()
}
