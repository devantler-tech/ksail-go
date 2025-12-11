package gen

import (
	"github.com/spf13/cobra"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

// NewServiceAccountCmd creates the gen serviceaccount command.
func NewServiceAccountCmd(rt *runtime.Runtime) *cobra.Command {
	return createGenCmd(rt, (*kubectl.Client).CreateServiceAccountCmd)
}
