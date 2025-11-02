package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewIngressCmd creates the gen ingress command.
func NewIngressCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewIngressGenerator()

	return generator.Generate()
}
