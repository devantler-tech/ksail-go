package workload

import (
	"os"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/kubectl"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// NewExposeCmd creates the workload expose command.
func NewExposeCmd(_ *runtime.Runtime) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := getKubeconfigPathSilently()

	// Create IO streams for kubectl
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Create kubectl client and get the expose command directly
	client := kubectl.NewClient(ioStreams)
	exposeCmd := client.CreateExposeCommand(kubeconfigPath)

	return exposeCmd
}
