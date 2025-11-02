package gen

import (
	"os"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// NewJobCmd creates the gen job command.
func NewJobCmd(rt *runtime.Runtime) *cobra.Command {
	return newResourceCmd(rt, func(client *kubectl.Client) *cobra.Command {
		return client.NewJobCmd()
	})
}
