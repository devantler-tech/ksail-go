package gen

import (
	"os"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// newResourceCmd is a helper to create a gen resource command.
func newResourceCmd(
	_ *runtime.Runtime,
	constructor func(client *kubectl.Client) (*cobra.Command, error),
) (*cobra.Command, error) {
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	client := kubectl.NewClient(ioStreams)

	return constructor(client)
}
