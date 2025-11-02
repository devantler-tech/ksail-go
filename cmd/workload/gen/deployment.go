package gen

import (
	"os"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// NewDeploymentCmd creates the gen deployment command.
func NewDeploymentCmd(rt *runtime.Runtime) *cobra.Command {
	return newResourceCmd(rt, func(client *kubectl.Client) *cobra.Command {
		return client.NewDeploymentCmd()
	})
}
