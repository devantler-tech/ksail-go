package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewConfigMapCmd creates the gen configmap command.
func NewConfigMapCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return createGenCommand(runtimeContainer, "configmap")
}
