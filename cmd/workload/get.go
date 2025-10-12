package workload

import (
	"os"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// NewGetCmd creates the workload get command.
// The runtime parameter is kept for consistency with other workload command constructors,
// though it's currently unused as this command wraps kubectl directly.
func NewGetCmd(_ *runtime.Runtime) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := shared.GetKubeconfigPathSilently()

	// Create IO streams for kubectl
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Create kubectl client and get the get command directly
	client := kubectl.NewClient(ioStreams)
	getCmd := client.CreateGetCommand(kubeconfigPath)

	return getCmd
}
