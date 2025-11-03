package gen

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewRoleCmd creates the gen role command.
func NewRoleCmd(rt *runtime.Runtime) *cobra.Command {
	cmd, err := newResourceCmd(rt, (*kubectl.Client).NewRoleCmd)
	if err != nil {
		panic(fmt.Sprintf("failed to create role command: %v", err))
	}

	return cmd
}
