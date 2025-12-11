package gen

import (
	"github.com/spf13/cobra"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

// NewSecretCmd creates the gen secret command.
func NewSecretCmd(rt *runtime.Runtime) *cobra.Command {
	return createGenCmd(rt, (*kubectl.Client).CreateSecretCmd)
}
