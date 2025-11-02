package gen

import (
	"os"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// NewClusterRoleCmd creates the gen clusterrole command.
func NewClusterRoleCmd(_ *runtime.Runtime) *cobra.Command {
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	client := kubectl.NewClient(ioStreams)

	return client.NewClusterRoleCmd()
}
