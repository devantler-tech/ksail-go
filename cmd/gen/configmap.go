package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewConfigMapCmd creates the gen configmap command.
func NewConfigMapCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewConfigMapGenerator()

	return generator.Generate()
}
